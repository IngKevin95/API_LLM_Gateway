package metrics

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"api-llm-gateway/internal/auth"
)

// HU-060 AC2: uptime + requests breakdown
type RequestMetrics struct {
	Total      int `json:"total"`
	ByHandler  map[string]int `json:"by_handler"`
	Errors     int `json:"errors"`
}

// HU-060 AC3: provider status
type ProviderStatus struct {
	Name                string `json:"name"`
	Available           bool   `json:"available"`
	LastSuccess         string `json:"last_success,omitempty"`
	CircuitBreakerOpen  bool   `json:"circuit_breaker_open"`
}

// HU-060 AC3: latency percentiles
type LatencyMetrics struct {
	P50Ms  float64 `json:"p50_ms"`
	P95Ms  float64 `json:"p95_ms"`
	P99Ms  float64 `json:"p99_ms"`
}

// HU-060: Gateway Metrics response structure (AC2+AC3)
type GatewayMetrics struct {
	UptimeSeconds int                `json:"uptime_seconds"`
	Requests      RequestMetrics     `json:"requests"`
	Providers     []ProviderStatus   `json:"providers"`
	Latency       LatencyMetrics     `json:"latency"`
	Models        []ModelMetric      `json:"models"`
	// Quota (HU-EVO-011) expone el desglose de cuota remanente por
	// proveedor/modelo, leído en vivo de quota.Manager.Snapshot() en cada
	// request (sin cache). Filtrado por scope del requester cuando no es
	// admin (AC5).
	Quota         []QuotaSnapshot    `json:"quota,omitempty"`
}

// QuotaSnapshot es la proyección pública (JSON) de quota.Snapshot para
// /metrics (HU-EVO-011).
type QuotaSnapshot struct {
	Provider  string     `json:"provider"`
	Model     string     `json:"model"`
	Limit     int64      `json:"limit"`
	Remaining int64      `json:"remaining"`
	ResetAt   *time.Time `json:"reset_at,omitempty"`
	Healthy   bool       `json:"healthy"`
	LearnedAt *time.Time `json:"learned_at"`
}

type ModelMetric struct {
	Model        string  `json:"model"`
	Provider     string  `json:"provider"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	P95LatencyMs float64 `json:"p95_latency_ms"`
	SuccessRate  float64 `json:"success_rate"`
	Tokens       int     `json:"tokens"`
	Cost         float64 `json:"cost"`
}

type Store interface {
	GetMetrics(ctx context.Context, providerFilter string) ([]ModelMetric, error)
	GetGatewayMetrics(ctx context.Context) (*GatewayMetrics, error)
}

type Handler struct {
	store Store
	// capabilityLookup resuelve las capacidades de un (provider, model) para
	// filtrar el bloque Quota por scope del requester (HU-EVO-011 AC5).
	// Cuando es nil, no se filtra (comportamiento admin/legacy).
	capabilityLookup func(provider, model string) []string
}

func NewHandler(store Store) *Handler {
	return &Handler{
		store: store,
	}
}

// SetCapabilityLookup inyecta la función que mapea (provider, model) a sus
// capacidades declaradas en el Registry, usada para filtrar el bloque Quota
// cuando el requester no es admin (HU-EVO-011 AC5).
func (h *Handler) SetCapabilityLookup(f func(provider, model string) []string) {
	h.capabilityLookup = f
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// HU-060: Return full gateway metrics (AC2+AC3)
	metrics, err := h.store.GetGatewayMetrics(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if metrics == nil {
		metrics = &GatewayMetrics{
			Requests: RequestMetrics{ByHandler: make(map[string]int)},
			Providers: []ProviderStatus{},
			Models: []ModelMetric{},
		}
	}
	metrics.Quota = h.filterQuota(r.Context(), metrics.Quota)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		// Just log or ignore since we already wrote 200
	}
}

// filterQuota aplica HU-EVO-011 AC5: si el request trae una identidad
// autenticada (no-admin, resuelta por un middleware previo como
// apikey.Middleware) y hay un capabilityLookup configurado, se descartan las
// filas de Quota cuyo modelo el requester no puede usar. Sin identidad en el
// contexto (ej. ruta admin-only por token estático) o sin capabilityLookup,
// no se filtra nada.
func (h *Handler) filterQuota(ctx context.Context, in []QuotaSnapshot) []QuotaSnapshot {
	id, ok := auth.FromContext(ctx)
	if !ok || h.capabilityLookup == nil {
		return in
	}
	out := make([]QuotaSnapshot, 0, len(in))
	for _, q := range in {
		for _, cap := range h.capabilityLookup(q.Provider, q.Model) {
			if id.HasScope("capability:" + cap) {
				out = append(out, q)
				break
			}
		}
	}
	return out
}
