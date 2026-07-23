package metrics

import (
	"context"
	"encoding/json"
	"net/http"
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
}

func NewHandler(store Store) *Handler {
	return &Handler{
		store: store,
	}
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		// Just log or ignore since we already wrote 200
	}
}
