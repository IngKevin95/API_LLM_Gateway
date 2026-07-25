// Command gateway is the entrypoint del API LLM Gateway.
// Scaffold mínimo runnable: levanta un servidor HTTP con /health y /metrics.
// Registry, Router, Adapters y demás componentes se cablean por slice (EP-XXX).
package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"

	"api-llm-gateway/internal/adapter"
	adapteraihubmix "api-llm-gateway/internal/adapter/aihubmix"
	adapteranthropc "api-llm-gateway/internal/adapter/anthropic"
	adaptergeneric "api-llm-gateway/internal/adapter/generic"
	adaptergoogle "api-llm-gateway/internal/adapter/google"
	adapterlocal "api-llm-gateway/internal/adapter/local"
	adapteromniroute "api-llm-gateway/internal/adapter/omniroute"
	adapteropenai "api-llm-gateway/internal/adapter/openai"
	"api-llm-gateway/internal/alert"
	apianthropic "api-llm-gateway/internal/api/anthropic"
	apimcp "api-llm-gateway/internal/api/mcp"
	apiopenai "api-llm-gateway/internal/api/openai"
	"api-llm-gateway/internal/auth"
	"api-llm-gateway/internal/auth/apikey"
	"api-llm-gateway/internal/failover"
	"api-llm-gateway/internal/handler"
	"api-llm-gateway/internal/health"
	"api-llm-gateway/internal/metrics"
	"api-llm-gateway/internal/middleware"
	"api-llm-gateway/internal/quota"
	"api-llm-gateway/internal/registry"
	"api-llm-gateway/internal/router"
	"api-llm-gateway/internal/tokenizer"
	"api-llm-gateway/internal/user"
)

// freeTierConfigPath es la ruta por defecto del catálogo de proveedores
// gratuitos curados (HU-EVO-002); override vía GATEWAY_FREETIER_CONFIG.
const freeTierConfigPath = "config/providers/free-tier.yaml"

func main() {
	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}

	// Carga del Registry en boot. Fail-fast si el config existe pero es inválido;
	// si no hay config declarado, arranca en modo scaffold (solo /health).
	cfgPath := os.Getenv("GATEWAY_CONFIG")
	if cfgPath == "" {
		if _, err := os.Stat("config.yaml"); err == nil {
			cfgPath = "config.yaml"
		}
	}

	// HU-060: metrics store (shared by processor and handler)
	metricsStore := metrics.NewInMemoryStore(10000)

	var processor *GatewayProcessor
	var alertDB *sql.DB
	var metricsHandlerCapabilityLookup func(provider, model string) []string
	var reg *registry.Registry
	var pgPersister *quota.PostgresPersister
	if cfgPath != "" {
		var err error
		reg, err = registry.Load(cfgPath, nil)
		if err != nil {
			log.Fatalf("registry: %v", err) // fail-fast, no arranca en estado parcial
		}

		// HU-EVO-002: merge del catálogo de proveedores gratuitos curados sobre
		// el catálogo base. Override de ruta vía GATEWAY_FREETIER_CONFIG; si no
		// existe el archivo (ni el default ni el override explícito) se sigue
		// sin free-tier, salvo que el override haya sido declarado explícito
		// (en ese caso, fail-fast).
		freeTierPath := os.Getenv("GATEWAY_FREETIER_CONFIG")
		explicitFreeTier := freeTierPath != ""
		if freeTierPath == "" {
			freeTierPath = freeTierConfigPath
		}
		if _, statErr := os.Stat(freeTierPath); statErr == nil {
			if err := reg.MergeFreeTier(freeTierPath, nil); err != nil {
				log.Fatalf("registry: merge free-tier %s: %v", freeTierPath, err)
			}
			log.Printf("INFO gateway: free-tier catalog cargado desde %s", freeTierPath)
		} else if explicitFreeTier {
			log.Fatalf("registry: GATEWAY_FREETIER_CONFIG=%s no encontrado: %v", freeTierPath, statErr)
		}

		// HU-EVO-008: Persister real a PostgreSQL, opt-in vía
		// GATEWAY_QUOTA_POSTGRES_DSN. Si no está declarado, sigue usando
		// NoPersister (no-op) — comportamiento sin cambios. Si está declarado
		// pero la conexión falla, se loguea warning y se sigue solo con
		// memoria (AC3: la persistencia nunca debe tumbar el boot).
		qm := quota.NewInMemoryManager()
		if dsn := os.Getenv("GATEWAY_QUOTA_POSTGRES_DSN"); dsn != "" {
			var err error
			pgPersister, err = quota.NewPostgresPersister(dsn, 1000)
			if err != nil {
				log.Printf("WARN gateway: quota postgres persister no disponible, sigue solo en RAM: %v", err)
			} else {
				qm = quota.NewInMemoryManagerWithPersister(time.Now, pgPersister)
				if restored, err := pgPersister.LoadRemaining(context.Background()); err != nil {
					log.Printf("WARN gateway: quota postgres LoadRemaining falló, arranca sin restaurar: %v", err)
				} else {
					for providerID, remaining := range restored {
						qm.RestoreRemaining(providerID, int(remaining)) // AC5: precedencia sobre quota_hint
					}
					log.Printf("INFO gateway: quota restaurada desde PostgreSQL para %d proveedor(es)", len(restored))
				}
			}
		}

		// HU-EVO-005: Quota Manager inicializado desde los quota_hint del Registry
		// (incluye los proveedores free-tier recién mergeados). InitFromRegistry
		// no pisa estado ya restaurado desde el persister (AC5).
		qm.InitFromRegistry(reg.QuotaHints())

		// HU-EVO-004: Health Monitor real; RetireOn429 lo invoca el Failover al
		// recibir un 429 de un adapter (retiro temporal con backoff/Retry-After).
		providerIDs := make([]string, 0, len(reg.Providers()))
		for _, p := range reg.Providers() {
			providerIDs = append(providerIDs, p.ID)
		}
		hm := health.New(providerIDs, func(string) bool { return true }, 3, 2)

		// Build Router (EP-001) con Health/Quota reales en vez de los stubs estáticos.
		rt := router.New(reg, hm, qm, tokenizer.NewHeuristic())

		// Build Adapters (EP-002, EP-008, EP-EVO-001), envueltos por el
		// Quota Middleware (HU-EVO-006/007/009) para aprender cuota real
		// desde los headers de respuesta e imponer reserva/commit por proveedor.
		adapters := wrapWithQuotaMiddleware(buildAdapters(reg), qm)

		// Build Failover Engine (EP-002); RetireOn429 se conecta al recibir 429s.
		fe := failover.New(rt, adapters)
		fe.OnRateLimited = hm.RetireOn429

		// Create Processor that uses Failover and metrics
		processor = NewGatewayProcessor(fe, metricsStore)

		// Lista de proveedores configurados para /metrics, visible sin tráfico previo.
		metricsStore.SetProviderSnapshot(providerIDs)

		// HU-EVO-011: Quota Manager real como fuente en vivo del bloque
		// "quota" de /metrics (leído por request, sin cache -- ver design.md).
		metricsStore.SetQuotaSource(quotaSourceAdapter{qm: qm})
		capabilityLookup := buildCapabilityLookup(reg)
		metricsHandlerCapabilityLookup = capabilityLookup

		// HU-EVO-012: Alert Manager, opt-in vía la misma PostgreSQL de cuota
		// (GATEWAY_QUOTA_POSTGRES_DSN). Fail-soft: sin DSN, no arranca el
		// worker (WARN, no bloquea boot) y /alerts responde lista vacía.
		if dsn := os.Getenv("GATEWAY_QUOTA_POSTGRES_DSN"); dsn != "" {
			if adb, err := sql.Open("postgres", dsn); err != nil {
				log.Printf("WARN gateway: alert manager sin DB (sql.Open): %v", err)
			} else if err := adb.PingContext(context.Background()); err != nil {
				log.Printf("WARN gateway: alert manager sin DB (ping): %v", err)
				_ = adb.Close()
			} else {
				threshold := alert.DefaultThreshold
				if raw := os.Getenv("GATEWAY_ALERT_THRESHOLD"); raw != "" {
					if v, err := strconv.ParseFloat(raw, 64); err == nil && v > 0 {
						threshold = v
					}
				}
				am, err := alert.NewManager(adb, alertQuotaAdapter{qm: qm}, threshold)
				if err != nil {
					log.Printf("WARN gateway: alert manager migración falló, worker deshabilitado: %v", err)
					_ = adb.Close()
				} else {
					alertDB = adb
					interval := time.Minute
					if raw := os.Getenv("GATEWAY_ALERT_INTERVAL"); raw != "" {
						if d, err := time.ParseDuration(raw); err == nil && d > 0 {
							interval = d
						}
					}
					go am.Run(context.Background(), interval)
					log.Printf("INFO gateway: alert manager arrancado (threshold=%.2f, interval=%s)", threshold, interval)
				}
			}
		} else {
			log.Printf("WARN gateway: GATEWAY_QUOTA_POSTGRES_DSN no configurado, alert manager deshabilitado (/alerts responde vacío)")
		}
	} else {
		log.Printf("WARN gateway: sin config.yaml, arrancando en modo scaffold (solo /health)")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// apiKeyStore resuelve identidades no-admin legacy (HU-EVO-013, y
	// HU-EVO-011 AC5 para /metrics), seedeadas vía GATEWAY_API_KEYS.
	// HU-EVO-018 agrega el store de PostgreSQL (userKeys, cableado más abajo)
	// como fuente adicional/reemplazo gradual: identityMiddleware prueba
	// userKeys primero y cae a apiKeyStore/env si no matchea, así una
	// instalación puede migrar sin invalidar las keys legacy de un día para
	// el otro.
	apiKeyStore := apikey.NewStore()
	loadAPIKeysFromEnv(apiKeyStore)

	// HU-EVO-017/HU-EVO-018: store de usuarios + API keys en PostgreSQL,
	// opt-in vía GATEWAY_USERS_POSTGRES_DSN (fallback: la misma DSN de cuota,
	// GATEWAY_QUOTA_POSTGRES_DSN, para no exigir una segunda variable en
	// despliegues con una sola instancia PostgreSQL). Fail-soft: sin DSN o si
	// la conexión falla, /users y /users/{id}/api-keys quedan deshabilitados
	// (503) en vez de tumbar el boot.
	var userStore *user.Store
	var userKeys *user.KeyStore
	var sessionStore *user.SessionStore
	usersDSN := os.Getenv("GATEWAY_USERS_POSTGRES_DSN")
	if usersDSN == "" {
		usersDSN = os.Getenv("GATEWAY_QUOTA_POSTGRES_DSN")
	}
	if usersDSN != "" {
		if udb, err := sql.Open("postgres", usersDSN); err != nil {
			log.Printf("WARN gateway: users store sin DB (sql.Open): %v", err)
		} else if err := udb.PingContext(context.Background()); err != nil {
			log.Printf("WARN gateway: users store sin DB (ping): %v", err)
			_ = udb.Close()
		} else if us, err := user.NewStore(udb); err != nil {
			log.Printf("WARN gateway: users store migración falló: %v", err)
			_ = udb.Close()
		} else if ks, err := user.NewKeyStore(udb, us); err != nil {
			log.Printf("WARN gateway: api_keys store migración falló: %v", err)
			_ = udb.Close()
		} else if ss, err := user.NewSessionStore(udb, us); err != nil {
			log.Printf("WARN gateway: sessions store migración falló: %v", err)
			_ = udb.Close()
		} else {
			userStore, userKeys, sessionStore = us, ks, ss
			log.Printf("INFO gateway: users/api_keys/sessions store conectado a PostgreSQL")
		}
	} else {
		log.Printf("WARN gateway: GATEWAY_USERS_POSTGRES_DSN no configurado, /users deshabilitado")
	}
	jwtSecret := []byte(os.Getenv("GATEWAY_JWT_SECRET"))
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("default-dev-secret-do-not-use-in-prod")
	}

	// identityMiddleware (HU-EVO-018): intenta resolver la identidad primero
	// contra JWT local (Fase 4), luego userKeys (PostgreSQL), finalmente apiKeyStore (legacy).
	identityMiddleware := func(next http.Handler) http.Handler {
		legacy := apikey.Middleware(apiKeyStore)(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearer(r)
			if token != "" {
				// JWT Local (HS256) check
				claims := jwt.MapClaims{}
				_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
					return jwtSecret, nil
				}, jwt.WithValidMethods([]string{"HS256"}))

				if err == nil {
					sid, _ := claims["sid"].(string)
					if sid != "" && sessionStore != nil {
						if ok, _ := sessionStore.IsValid(r.Context(), sid); ok {
							sub, _ := claims["sub"].(string)
							tenant, _ := claims["tenant"].(string)
							id := auth.Identity{Subject: sub, Tenant: tenant, SessionID: sid}
							if rawScopes, ok := claims["scopes"].([]any); ok {
								for _, s := range rawScopes {
									if str, ok := s.(string); ok {
										id.Scopes = append(id.Scopes, str)
									}
								}
							}
							next.ServeHTTP(w, r.WithContext(auth.WithIdentity(r.Context(), id)))
							return
						}
					}
				}

				if userKeys != nil {
					if id, ok := userKeys.Authenticate(r.Context(), token); ok {
						next.ServeHTTP(w, r.WithContext(auth.WithIdentity(r.Context(), id)))
						return
					}
				}
			}
			legacy.ServeHTTP(w, r)
		})
	}

	// registerAuthRoutes se cablea recién ahora que identityMiddleware existe:
	// /sessions y /auth/mfa/* dependen de auth.FromContext, que solo se
	// popula si la ruta pasa por identityMiddleware (JWT local emitido por
	// POST /auth/login, o userKeys/apiKeyStore legacy). Antes de este fix las
	// rutas quedaban registradas sin ese wrap y devolvían 401 siempre
	// (hueco de wiring detectado en EP-EVO-004-SS3, corregido acá).
	registerAuthRoutes(mux, userStore, sessionStore, os.Getenv("GATEWAY_ADMIN_TOKEN"), jwtSecret, identityMiddleware)
	registerUsersRoutes(mux, userStore, userKeys, os.Getenv("GATEWAY_ADMIN_TOKEN"), identityMiddleware)

	// HU-060: /metrics endpoint con datos reales en memoria. Admin (token
	// estático GATEWAY_ADMIN_TOKEN) ve todo sin filtrar; una identidad
	// resuelta por identityMiddleware (no-admin) ve el bloque quota filtrado
	// por scope (AC5). Sin admin y sin key válida: 401.
	metricsHandler := metrics.NewHandler(metricsStore)
	if metricsHandlerCapabilityLookup != nil {
		metricsHandler.SetCapabilityLookup(metricsHandlerCapabilityLookup)
	}
	metricsAuthenticated := identityMiddleware(metricsHandler)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		adminToken := os.Getenv("GATEWAY_ADMIN_TOKEN")
		if adminToken != "" && r.Header.Get("Authorization") == "Bearer "+adminToken {
			metricsHandler.ServeHTTP(w, r)
			return
		}
		metricsAuthenticated.ServeHTTP(w, r)
	})

	// HU-EVO-013: GET /alerts, RBAC-aware. Admin (mismo token estático que
	// /metrics) ve todas las alertas sin filtro; cualquier otra identidad
	// resuelta por identityMiddleware (userKeys o apiKeyStore legacy) solo ve
	// alertas de modelos cubiertos por sus scopes. Sin identidad y sin admin:
	// 401.
	alertsHandler := handler.NewAlertsHandler(alertDB, handler.CapabilityLookup(metricsHandlerCapabilityLookup))
	alertsAuthenticated := identityMiddleware(alertsHandler)
	mux.HandleFunc("/alerts", func(w http.ResponseWriter, r *http.Request) {
		adminToken := os.Getenv("GATEWAY_ADMIN_TOKEN")
		if adminToken != "" && r.Header.Get("Authorization") == "Bearer "+adminToken {
			alertsHandler.ServeHTTP(w, r.WithContext(handler.WithAdmin(r.Context())))
			return
		}
		alertsAuthenticated.ServeHTTP(w, r)
	})

	// Register OpenAI-compatible endpoints (HU-012a, HU-012b, HU-012c)
	if processor != nil {
		openaiHandler := apiopenai.NewHandler(processor)
		mux.HandleFunc("POST /v1/chat/completions", openaiHandler.HandleChatCompletions)
		mux.HandleFunc("POST /v1/embeddings", openaiHandler.HandleEmbeddings)

		// Register Anthropic-compatible endpoints (HU-013, HU-016)
		anthropicHandler := apianthropic.NewHandler(processor)
		mux.HandleFunc("POST /v1/messages", anthropicHandler.HandleMessages)

		// Register Universal Compatibility endpoint (HU-EVO-006)
		responsesHandler := handler.NewResponsesHandler(processor)
		mux.Handle("POST /responses", responsesHandler)

		// Register MCP integration (HU-033)
		mcpHandler := apimcp.NewHandler(os.Getenv("GATEWAY_ADMIN_TOKEN"), reg)
		mux.Handle("POST /mcp", mcpHandler)
	}

	var readHeaderTimeout, writeTimeout time.Duration
	if cfgPath != "" {
		// Re-load the config just for timeouts
		reg, err := registry.Load(cfgPath, nil)
		if err == nil {
			rMs, wMs := reg.ServerTimeouts()
			if rMs > 0 {
				readHeaderTimeout = time.Duration(rMs) * time.Millisecond
			}
			if wMs > 0 {
				writeTimeout = time.Duration(wMs) * time.Millisecond
			}
		}
	}
	if readHeaderTimeout == 0 {
		readHeaderTimeout = 5 * time.Second
	}
	if writeTimeout == 0 {
		writeTimeout = 30 * time.Second
	}

	// Aplicar middleware de request ID
	var handler http.Handler = mux
	handler = middleware.RequestID(handler)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
	}

	go func() {
		log.Printf("gateway escuchando en :%s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
	if pgPersister != nil {
		if err := pgPersister.Close(); err != nil {
			log.Printf("quota persister close: %v", err)
		} else {
			log.Printf("INFO gateway: quota persister flushed and closed")
		}
	}
}

// quotaSourceAdapter adapta quota.Manager.Snapshot() ([]quota.Snapshot) al
// metrics.QuotaSource ([]metrics.QuotaEntry) que espera metrics.Store, sin
// que el paquete metrics dependa directamente del paquete quota
// (HU-EVO-011).
type quotaSourceAdapter struct {
	qm quota.Manager
}

func (a quotaSourceAdapter) Snapshot() []metrics.QuotaEntry {
	snap := a.qm.Snapshot()
	out := make([]metrics.QuotaEntry, 0, len(snap))
	for _, s := range snap {
		out = append(out, metrics.QuotaEntry{
			Provider:  s.Provider,
			Model:     s.Model,
			Limit:     s.Limit,
			Remaining: s.Remaining,
			ResetAt:   s.ResetAt,
			Healthy:   s.Healthy,
			LearnedAt: s.LearnedAt,
		})
	}
	return out
}

// alertQuotaAdapter adapta quota.Manager.Snapshot() al alert.QuotaSnapshotter
// que espera alert.Manager (mismo patrón que quotaSourceAdapter para metrics).
type alertQuotaAdapter struct {
	qm quota.Manager
}

func (a alertQuotaAdapter) Snapshot() []alert.QuotaEntry {
	snap := a.qm.Snapshot()
	out := make([]alert.QuotaEntry, 0, len(snap))
	for _, s := range snap {
		out = append(out, alert.QuotaEntry{
			Provider:  s.Provider,
			Model:     s.Model,
			Limit:     s.Limit,
			Remaining: s.Remaining,
		})
	}
	return out
}

// buildCapabilityLookup resuelve las capacidades declaradas de un (provider,
// model) desde el Registry, usado para filtrar /metrics#quota (HU-EVO-011
// AC5) y /alerts (HU-EVO-013 AC4) por scope del requester. model=="" (fila
// agregada sin desglose, ver quota.Snapshot) devuelve la unión de
// capacidades de todos los modelos de ese proveedor.
func buildCapabilityLookup(reg *registry.Registry) func(provider, model string) []string {
	return func(provider, model string) []string {
		var caps []string
		for _, p := range reg.Providers() {
			if p.ID != provider {
				continue
			}
			for _, m := range p.Models {
				if model == "" || m.Name == model {
					caps = append(caps, m.Capabilities...)
				}
			}
		}
		return caps
	}
}

// registerUsersRoutes cablea /users, /users/{id} y /users/{id}/api-keys*
// (HU-EVO-017/HU-EVO-018). Si userStore es nil (sin PostgreSQL configurada),
// responde 503 a todas las rutas en vez de nil-pointer panic.
func registerUsersRoutes(mux *http.ServeMux, userStore *user.Store, userKeys *user.KeyStore, adminToken string, identity func(http.Handler) http.Handler) {
	unavailable := func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"users store not configured"}`, http.StatusServiceUnavailable)
	}
	if userStore == nil || userKeys == nil {
		mux.HandleFunc("/users", unavailable)
		mux.HandleFunc("GET /users/me", unavailable)
		mux.HandleFunc("PATCH /users/{id}", unavailable)
		mux.HandleFunc("/users/{id}/api-keys", unavailable)
		mux.HandleFunc("DELETE /users/{id}/api-keys/{keyId}", unavailable)
		return
	}

	usersHandler := handler.NewUsersHandler(userStore)
	apiKeysHandler := handler.NewAPIKeysHandler(userKeys)

	// wrap prueba primero el token estatico GATEWAY_ADMIN_TOKEN (bypass total,
	// no es una auth.Identity real y el legacy apikey.Middleware dentro de
	// identity() lo rechazaria con 401 antes de llegar a resolveUserAuth);
	// si no matchea, recien ahi pasa por identity() (JWT de sesion / API key
	// / legacy) + resolveUserAuth.
	wrap := func(h http.Handler) http.HandlerFunc {
		withAuth := identity(resolveUserAuth(adminToken, userStore, h))
		return func(w http.ResponseWriter, r *http.Request) {
			if adminToken != "" && r.Header.Get("Authorization") == "Bearer "+adminToken {
				ac := handler.AdminContext{IsAdmin: true, GlobalAdmin: true}
				h.ServeHTTP(w, handler.WithAdminContextValue(r, ac))
				return
			}
			withAuth.ServeHTTP(w, r)
		}
	}

	// GET /users/me (HU-EVO-022 AC1) requiere identityMiddleware, no
	// resolveUserAuth: cualquier usuario autenticado (admin o no) resuelve su
	// propio perfil vía auth.Identity, sin pasar por AdminContext. El token
	// estático GATEWAY_ADMIN_TOKEN no es una fila real de `users` (no tiene
	// Subject), así que no puede pasar por identity()+Me -- se sintetiza un
	// perfil admin fijo para que el Dashboard (que usa este token como Bearer
	// por defecto, ver EP-EVO-003) resuelva la tab "Team" igual que con un
	// JWT de sesión real (hueco de wiring detectado por
	// wiring-adversarial-verifier en EP-EVO-004-SS3: antes devolvia 401 y
	// dejaba Team/Profile & Security vacíos para ese camino de auth).
	meHandler := func(w http.ResponseWriter, r *http.Request) {
		if adminToken != "" && r.Header.Get("Authorization") == "Bearer "+adminToken {
			handler.WriteStaticAdminProfile(w)
			return
		}
		identity(http.HandlerFunc(usersHandler.Me)).ServeHTTP(w, r)
	}
	mux.HandleFunc("GET /users/me", meHandler)
	mux.HandleFunc("/users", wrap(usersHandler))
	mux.HandleFunc("PATCH /users/{id}", wrap(http.HandlerFunc(usersHandler.PatchUser)))
	mux.HandleFunc("/users/{id}/api-keys", wrap(apiKeysHandler))
	mux.HandleFunc("DELETE /users/{id}/api-keys/{keyId}", wrap(http.HandlerFunc(apiKeysHandler.RevokeAPIKey)))
}

// registerAuthRoutes cablea /sessions, /auth/login y /auth/mfa*. identity
// envuelve cada ruta protegida con identityMiddleware (JWT local o
// userKeys/apiKeyStore legacy) para que auth.FromContext resuelva el
// Subject -- sin este wrap las rutas devuelven 401 siempre (bug de wiring
// corregido en EP-EVO-004-SS3, ver llamador en main()).
func registerAuthRoutes(mux *http.ServeMux, userStore *user.Store, sessionStore *user.SessionStore, adminToken string, jwtSecret []byte, identity func(http.Handler) http.Handler) {
	unavailable := func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"auth store not configured"}`, http.StatusServiceUnavailable)
	}
	if userStore == nil || sessionStore == nil {
		mux.HandleFunc("/sessions", unavailable)
		mux.HandleFunc("DELETE /sessions/{id}", unavailable)
		mux.HandleFunc("/auth/mfa/enroll", unavailable)
		mux.HandleFunc("/auth/mfa/verify", unavailable)
		mux.HandleFunc("/auth/mfa/disable", unavailable)
		return
	}

	sessionsHandler := handler.NewSessionsHandler(sessionStore)
	mfaHandler := handler.NewMfaHandler(userStore)
	authHandler := handler.NewAuthHandler(userStore, sessionStore, jwtSecret)

	wrap := func(h http.Handler) http.HandlerFunc {
		return identity(h).ServeHTTP
	}

	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("/sessions", wrap(sessionsHandler))
	mux.HandleFunc("DELETE /sessions/{id}", wrap(http.HandlerFunc(sessionsHandler.RevokeSession)))
	mux.HandleFunc("/auth/mfa/enroll", wrap(http.HandlerFunc(mfaHandler.Enroll)))
	mux.HandleFunc("/auth/mfa/verify", wrap(http.HandlerFunc(mfaHandler.Verify)))
	mux.HandleFunc("/auth/mfa/disable", wrap(http.HandlerFunc(mfaHandler.Disable)))
}

// resolveUserAuth resuelve AdminContext para el dominio de usuarios: token
// estático GATEWAY_ADMIN_TOKEN = admin global; si no, lee la auth.Identity ya
// resuelta por identityMiddleware (JWT de sesión, API key o legacy -- las
// tres vías quedan cubiertas porque este handler siempre se monta detrás de
// identity()) y, si el usuario resuelto tiene role=admin, promueve a admin de
// su propio tenant (no global). Antes de este fix solo reconocía API keys,
// dejando 403 a cualquier admin autenticado por sesión JWT vía /auth/login
// (hueco de wiring detectado en EP-EVO-004-SS3 al probar el journey real).
// Delega el 401/403 final a cada handler (algunos endpoints son admin-only,
// otros owner-or-admin).
func resolveUserAuth(adminToken string, userStore *user.Store, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ac := handler.AdminContext{}

		if adminToken != "" && r.Header.Get("Authorization") == "Bearer "+adminToken {
			ac = handler.AdminContext{IsAdmin: true, GlobalAdmin: true}
		} else if id, ok := auth.FromContext(ctx); ok {
			ac.Tenant = id.Tenant
			if u, err := userStore.Get(ctx, id.Subject); err == nil && u.Role == user.RoleAdmin {
				ac.IsAdmin = true
			}
		}

		req := r.WithContext(ctx)
		req = handler.WithAdminContextValue(req, ac)
		next.ServeHTTP(w, req)
	}
}

func extractBearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
	}
	return ""
}

// loadAPIKeysFromEnv seedea el apikey.Store desde GATEWAY_API_KEYS
// (formato "key:tenant:cap1,cap2;key2:tenant2:cap1"), usado por /alerts para
// resolver identidades no-admin (HU-EVO-013). Sin la env var, el store queda
// vacío y todo acceso no-admin a /alerts devuelve 401 (fail-closed).
func loadAPIKeysFromEnv(store *apikey.Store) {
	raw := os.Getenv("GATEWAY_API_KEYS")
	if raw == "" {
		return
	}
	for _, entry := range strings.Split(raw, ";") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		parts := strings.SplitN(entry, ":", 3)
		if len(parts) != 3 {
			log.Printf("WARN gateway: GATEWAY_API_KEYS entrada inválida (esperado key:tenant:scopes), omitiendo")
			continue
		}
		key, tenant, scopesRaw := parts[0], parts[1], parts[2]
		var scopes []string
		for _, s := range strings.Split(scopesRaw, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				scopes = append(scopes, "capability:"+s)
			}
		}
		store.Add(key, auth.Identity{Subject: tenant, Tenant: tenant, Scopes: scopes})
	}
}

// wrapWithQuotaMiddleware envuelve cada adapter con quota.Middleware
// (HU-EVO-006/007/009): intercepta Chat/Embed para reservar/confirmar
// consumo y aprender QuotaInfo real de los headers de respuesta hacia el
// Manager compartido con el Router (para que la penalización por cuota baja
// de scoreAll opere sobre datos aprendidos en runtime, no solo el hint
// estático del Registry).
func wrapWithQuotaMiddleware(adapters map[string]adapter.Adapter, qm quota.Manager) map[string]adapter.Adapter {
	wrapped := make(map[string]adapter.Adapter, len(adapters))
	for providerID, ad := range adapters {
		wrapped[providerID] = quota.NewMiddleware(qm, providerID, ad)
	}
	return wrapped
}

// buildAdapters constructs and returns adapters for all configured providers.
func buildAdapters(reg *registry.Registry) map[string]adapter.Adapter {
	adapters := make(map[string]adapter.Adapter)

	// For MVP, build adapters for known provider types based on registry config.
	// This is a simplified approach; a fuller version would introspect registry.providers dynamically.

	// Manually add OpenAI adapter (always available if API key is set)
	if key := reg.APIKey("openai"); key != "" {
		adapters["openai"] = adapteropenai.New("https://api.openai.com/v1", key)
	}

	// Manually add Anthropic adapter
	if key := reg.APIKey("anthropic"); key != "" {
		adapters["anthropic"] = adapteranthropc.New("https://api.anthropic.com", key)
	}

	// Manually add Google adapter
	if key := reg.APIKey("google"); key != "" {
		adapters["google"] = adaptergoogle.New("", key)
	}

	// Add AIHubMix if configured
	if key := reg.APIKey("aihubmix"); key != "" {
		adapters["aihubmix"] = adapteraihubmix.New("https://api.aihubmix.com/v1", key)
	}

	// Add local Ollama if available (no API key needed)
	adapters["local"] = adapterlocal.New("http://localhost:11434")

	// Add OmniRoute adapter (local provider, no API key needed)
	adapters["omniroute"] = adapteromniroute.New(adapteromniroute.Config{
		BaseURL: "http://omniroute:20128/v1",
		APIKey:  "",
	})

	// HU-EVO-001: adapters data-driven para providers declarados type:generic
	// (catálogo free-tier: Groq, Cerebras, Mistral, Gemini, Cloudflare AI, etc.).
	// Todos hablan wire OpenAI-compatible salvo que se declare lo contrario.
	for _, p := range reg.Providers() {
		if p.Type != "generic" {
			continue
		}
		if p.APIKey == "" {
			log.Printf("WARN gateway: provider generic %q sin api_key resuelta, omitiendo adapter", p.ID)
			continue
		}
		spec := adaptergeneric.ProviderSpec{
			BaseURL: p.BaseURL,
			Format:  adaptergeneric.FormatOpenAI,
		}
		ad, err := adaptergeneric.New(spec, p.APIKey)
		if err != nil {
			log.Printf("WARN gateway: provider generic %q spec inválido, omitiendo adapter: %v", p.ID, err)
			continue
		}
		adapters[p.ID] = ad
	}

	return adapters
}
