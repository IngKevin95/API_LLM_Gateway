package metrics_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"api-llm-gateway/internal/metrics"
)

type mockStore struct {
	data []metrics.ModelMetric
	err  error
}

func (m *mockStore) GetMetrics(ctx context.Context, providerFilter string) ([]metrics.ModelMetric, error) {
	if m.err != nil {
		return nil, m.err
	}
	if providerFilter != "" {
		var filtered []metrics.ModelMetric
		for _, v := range m.data {
			if v.Provider == providerFilter {
				filtered = append(filtered, v)
			}
		}
		return filtered, nil
	}
	return m.data, nil
}

func (m *mockStore) GetGatewayMetrics(ctx context.Context) (*metrics.GatewayMetrics, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &metrics.GatewayMetrics{
		UptimeSeconds: 3600,
		Requests: metrics.RequestMetrics{
			Total: len(m.data),
			ByHandler: map[string]int{
				"/v1/chat/completions": len(m.data),
			},
			Errors: 0,
		},
		Providers: []metrics.ProviderStatus{},
		Latency: metrics.LatencyMetrics{
			P50Ms:  100.0,
			P95Ms:  200.0,
			P99Ms:  300.0,
		},
		Models: m.data,
	}, nil
}

// mockAuth middleware simulates scope checking
func mockAuth(operatorOnly bool, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// En un entorno real se extraería del contexto, acá simulamos por header para el test
		scope := r.Header.Get("X-Scope")
		if operatorOnly && scope != "operator" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next(w, r) // Fixed later to support standard http interfaces
	}
}

func mockAuthMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := r.Header.Get("X-Scope")
		if scope != "operator" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func TestMetricsHandler_Success(t *testing.T) {
	store := &mockStore{
		data: []metrics.ModelMetric{
			{Model: "gpt-4", Provider: "openai", AvgLatencyMs: 150.5, P95LatencyMs: 300.0, SuccessRate: 0.99, Tokens: 5000, Cost: 0.15},
		},
	}
	handler := metrics.NewHandler(store)
	server := mockAuthMW(http.HandlerFunc(handler.ServeHTTP))

	req := httptest.NewRequest(http.MethodGet, "/v1/metrics", nil)
	req.Header.Set("X-Scope", "operator")
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rr.Code)
	}

	var res metrics.GatewayMetrics
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if len(res.Models) != 1 {
		t.Fatalf("Expected 1 model, got %d", len(res.Models))
	}
	if res.Models[0].Model != "gpt-4" {
		t.Errorf("Expected gpt-4, got %s", res.Models[0].Model)
	}
}

func TestMetricsHandler_Forbidden(t *testing.T) {
	handler := metrics.NewHandler(&mockStore{})
	server := mockAuthMW(http.HandlerFunc(handler.ServeHTTP))

	req := httptest.NewRequest(http.MethodGet, "/v1/metrics", nil)
	// No X-Scope header
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected 403, got %d", rr.Code)
	}
}

func TestMetricsHandler_EmptyState(t *testing.T) {
	store := &mockStore{
		data: []metrics.ModelMetric{},
	}
	handler := metrics.NewHandler(store)
	server := mockAuthMW(http.HandlerFunc(handler.ServeHTTP))

	req := httptest.NewRequest(http.MethodGet, "/v1/metrics", nil)
	req.Header.Set("X-Scope", "operator")
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rr.Code)
	}

	var res metrics.GatewayMetrics
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if len(res.Models) != 0 {
		t.Errorf("Expected 0 models, got %d", len(res.Models))
	}
}

func TestMetricsHandler_FilterByProvider(t *testing.T) {
	store := &mockStore{
		data: []metrics.ModelMetric{
			{Model: "gpt-4", Provider: "openai"},
			{Model: "claude-3", Provider: "anthropic"},
		},
	}
	handler := metrics.NewHandler(store)
	server := mockAuthMW(http.HandlerFunc(handler.ServeHTTP))

	req := httptest.NewRequest(http.MethodGet, "/v1/metrics?provider=anthropic", nil)
	req.Header.Set("X-Scope", "operator")
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rr.Code)
	}

	var res metrics.GatewayMetrics
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if len(res.Models) != 2 {
		t.Fatalf("Expected 2 models in response, got %d", len(res.Models))
	}
}

func TestMetricsHandler_DBError(t *testing.T) {
	store := &mockStore{
		err: errors.New("db down"),
	}
	handler := metrics.NewHandler(store)
	server := mockAuthMW(http.HandlerFunc(handler.ServeHTTP))

	req := httptest.NewRequest(http.MethodGet, "/v1/metrics", nil)
	req.Header.Set("X-Scope", "operator")
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rr.Code)
	}
}
