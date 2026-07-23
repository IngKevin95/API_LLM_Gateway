package metrics

import (
	"context"
	"encoding/json"
	"net/http"
)

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
	providerFilter := r.URL.Query().Get("provider")

	metrics, err := h.store.GetMetrics(r.Context(), providerFilter)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if metrics == nil {
		metrics = []ModelMetric{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		// Just log or ignore since we already wrote 200
	}
}
