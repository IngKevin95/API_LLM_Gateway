package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"api-llm-gateway/internal/auth"
)

// AlertRow es la proyección pública (JSON) de una fila de provider_alerts
// devuelta por GET /alerts (HU-EVO-013).
type AlertRow struct {
	ID         int64      `json:"id"`
	Provider   string     `json:"provider"`
	Model      string     `json:"model"`
	Severity   string     `json:"severity"`
	Message    string     `json:"message"`
	AlertTime  time.Time  `json:"alert_time"`
	UpdatedAt  time.Time  `json:"updated_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

// AlertsResponse es el envelope paginado de GET /alerts.
type AlertsResponse struct {
	Data  []AlertRow `json:"data"`
	Page  int        `json:"page"`
	Limit int        `json:"limit"`
	Total int        `json:"total"`
}

// CapabilityLookup resuelve las capacidades declaradas de un (provider,
// model) — inyectado desde el Registry (mismo mecanismo que
// metrics.Handler.SetCapabilityLookup) para filtrar alertas por scope
// (HU-EVO-013 AC4).
type CapabilityLookup func(provider, model string) []string

// AlertsHandler expone GET /alerts con filtrado RBAC: admin (token estático
// GATEWAY_ADMIN_TOKEN, resuelto por el caller vía adminCtxKey) ve todo;
// cualquier otra identidad autenticada (auth.Identity en contexto, inyectada
// por apikey.Middleware) solo ve alertas de modelos cubiertos por sus scopes
// (HU-EVO-013 AC1/AC2/AC4). Sin identidad en contexto y sin admin: 401.
type AlertsHandler struct {
	db               *sql.DB
	capabilityLookup CapabilityLookup
}

func NewAlertsHandler(db *sql.DB, capabilityLookup CapabilityLookup) *AlertsHandler {
	return &AlertsHandler{db: db, capabilityLookup: capabilityLookup}
}

type adminCtxKey struct{}

// WithAdmin marca el contexto de un request como autenticado con el token de
// administrador (HU-EVO-013 AC3: ve todas las alertas sin filtro de scope).
func WithAdmin(ctx context.Context) context.Context {
	return context.WithValue(ctx, adminCtxKey{}, true)
}

func isAdmin(ctx context.Context) bool {
	v, _ := ctx.Value(adminCtxKey{}).(bool)
	return v
}

func (h *AlertsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	admin := isAdmin(r.Context())
	id, hasIdentity := auth.FromContext(r.Context())
	if !admin && !hasIdentity {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	if h.db == nil {
		// Sin PostgreSQL configurada (fail-soft, ver design.md Migration Plan):
		// responde lista vacía en vez de 500, para no tumbar clientes de /alerts
		// en despliegues sin DB.
		writeJSON(w, http.StatusOK, AlertsResponse{Data: []AlertRow{}, Page: 1, Limit: 50, Total: 0})
		return
	}

	page, limit := parsePagination(r)

	rows, err := h.db.QueryContext(r.Context(), `
		SELECT id, provider_id, model_id, severity, message, alert_time, updated_at, resolved_at
		FROM provider_alerts
		WHERE resolved_at IS NULL
		ORDER BY alert_time DESC
	`)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var all []AlertRow
	for rows.Next() {
		var a AlertRow
		var resolvedAt sql.NullTime
		if err := rows.Scan(&a.ID, &a.Provider, &a.Model, &a.Severity, &a.Message, &a.AlertTime, &a.UpdatedAt, &resolvedAt); err != nil {
			http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
			return
		}
		if resolvedAt.Valid {
			t := resolvedAt.Time
			a.ResolvedAt = &t
		}
		all = append(all, a)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	// HU-EVO-013 AC1/AC2/AC4: filtrado RBAC en memoria por scope (admin
	// no filtra). No hay concepto de tenant propio en provider_alerts (la
	// alerta pertenece al proveedor/modelo, no a un tenant); el aislamiento
	// multi-tenant se aplica vía scope de capacidad del modelo, consistente
	// con el resto del Gateway (authz.go) -- ver progress_log de la sesión que
	// implementó este endpoint para el detalle de esta decisión.
	filtered := all
	if !admin {
		filtered = make([]AlertRow, 0, len(all))
		for _, a := range all {
			if h.allowed(id, a) {
				filtered = append(filtered, a)
			}
		}
	}

	total := len(filtered)
	start := (page - 1) * limit
	end := start + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	pageData := filtered[start:end]
	if pageData == nil {
		pageData = []AlertRow{}
	}

	writeJSON(w, http.StatusOK, AlertsResponse{Data: pageData, Page: page, Limit: limit, Total: total})
}

func (h *AlertsHandler) allowed(id auth.Identity, a AlertRow) bool {
	if h.capabilityLookup == nil {
		return true // sin lookup configurado: no se puede evaluar scope, no bloquea (fail-open documentado)
	}
	for _, cap := range h.capabilityLookup(a.Provider, a.Model) {
		if id.HasScope("capability:" + cap) {
			return true
		}
	}
	return false
}

func parsePagination(r *http.Request) (page, limit int) {
	page = 1
	limit = 50
	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 200 {
		limit = l
	}
	return page, limit
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
