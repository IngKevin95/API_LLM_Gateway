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

	_ "github.com/lib/pq"

	"api-llm-gateway/internal/adapter"
	adapteraihubmix "api-llm-gateway/internal/adapter/aihubmix"
	adapteranthropc "api-llm-gateway/internal/adapter/anthropic"
	adaptergoogle "api-llm-gateway/internal/adapter/google"
	adapterlocal "api-llm-gateway/internal/adapter/local"
	adapteromniroute "api-llm-gateway/internal/adapter/omniroute"
	adapteropenai "api-llm-gateway/internal/adapter/openai"
	adaptergeneric "api-llm-gateway/internal/adapter/generic"
	"api-llm-gateway/internal/alert"
	apianthropic "api-llm-gateway/internal/api/anthropic"
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
	if cfgPath != "" {
		var err error
		reg, err := registry.Load(cfgPath, nil)
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
			pgPersister, err := quota.NewPostgresPersister(dsn, 1000)
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

	// apiKeyStore resuelve identidades no-admin (HU-EVO-013, y HU-EVO-011 AC5
	// para /metrics), seedeadas vía GATEWAY_API_KEYS. Compartido entre
	// /metrics y /alerts: mismas credenciales, mismo mecanismo de scope.
	apiKeyStore := apikey.NewStore()
	loadAPIKeysFromEnv(apiKeyStore)

	// HU-060: /metrics endpoint con datos reales en memoria. Admin (token
	// estático GATEWAY_ADMIN_TOKEN) ve todo sin filtrar; una identidad de
	// apiKeyStore (no-admin) ve el bloque quota filtrado por scope (AC5).
	// Sin admin y sin key válida: 401.
	metricsHandler := metrics.NewHandler(metricsStore)
	if metricsHandlerCapabilityLookup != nil {
		metricsHandler.SetCapabilityLookup(metricsHandlerCapabilityLookup)
	}
	metricsAuthenticated := apikey.Middleware(apiKeyStore)(metricsHandler)
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
	// resuelta por apikey.Middleware (mismo apiKeyStore) solo ve alertas de
	// modelos cubiertos por sus scopes. Sin identidad y sin admin: 401.
	alertsHandler := handler.NewAlertsHandler(alertDB, handler.CapabilityLookup(metricsHandlerCapabilityLookup))
	alertsAuthenticated := apikey.Middleware(apiKeyStore)(alertsHandler)
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

		// Register MCP integration (HU-033) — stub handler for now
		// TODO: Full MCP integration in Fase 2
		mux.HandleFunc("POST /mcp", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotImplemented)
			_, _ = w.Write([]byte(`{"error":"MCP integration pending"}`))
		})
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
