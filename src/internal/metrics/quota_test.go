package metrics_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"api-llm-gateway/internal/auth"
	"api-llm-gateway/internal/metrics"
)

type quotaStore struct {
	quota []metrics.QuotaSnapshot
}

func (s *quotaStore) GetMetrics(ctx context.Context, providerFilter string) ([]metrics.ModelMetric, error) {
	return nil, nil
}

func (s *quotaStore) GetGatewayMetrics(ctx context.Context) (*metrics.GatewayMetrics, error) {
	return &metrics.GatewayMetrics{
		Requests:  metrics.RequestMetrics{ByHandler: map[string]int{}},
		Providers: []metrics.ProviderStatus{},
		Models:    []metrics.ModelMetric{},
		Quota:     s.quota,
	}, nil
}

// HU-EVO-011 AC1: happy path, bloque quota presente sin filtrar cuando no
// hay identidad no-admin en contexto (admin/legacy).
func TestHandler_Quota_NoIdentity_ReturnsUnfiltered(t *testing.T) {
	store := &quotaStore{quota: []metrics.QuotaSnapshot{
		{Provider: "groq", Model: "mixtral", Limit: 14400, Remaining: 14200, Healthy: true},
	}}
	h := metrics.NewHandler(store)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d", w.Code)
	}
	var got metrics.GatewayMetrics
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("json inválido: %v", err)
	}
	if len(got.Quota) != 1 || got.Quota[0].Provider != "groq" {
		t.Fatalf("esperaba quota sin filtrar, obtuve %+v", got.Quota)
	}
}

// HU-EVO-011 AC5: identidad no-admin con scope capability:coding filtra
// cuota de modelos fuera de su scope (aquí, un modelo de capability vision).
func TestHandler_Quota_NonAdminIdentity_FiltersOutOfScope(t *testing.T) {
	store := &quotaStore{quota: []metrics.QuotaSnapshot{
		{Provider: "groq", Model: "mixtral-coding", Limit: 1000, Remaining: 900, Healthy: true},
		{Provider: "openai", Model: "gpt4-vision", Limit: 1000, Remaining: 900, Healthy: true},
	}}
	h := metrics.NewHandler(store)
	h.SetCapabilityLookup(func(provider, model string) []string {
		if provider == "groq" {
			return []string{"coding"}
		}
		return []string{"vision"}
	})

	id := auth.Identity{Subject: "t1", Tenant: "t1", Scopes: []string{"capability:coding"}}
	req := httptest.NewRequest("GET", "/metrics", nil)
	req = req.WithContext(auth.WithIdentity(req.Context(), id))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	var got metrics.GatewayMetrics
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("json inválido: %v", err)
	}
	if len(got.Quota) != 1 || got.Quota[0].Provider != "groq" {
		t.Fatalf("esperaba solo quota de groq (scope coding), obtuve %+v", got.Quota)
	}
}
