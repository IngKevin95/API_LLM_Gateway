// Package integration_test verifica end-to-end que el header Retry-After de
// un 429 real se propaga desde el adapter genérico hasta el Health Monitor,
// cerrando el hueco de HU-EVO-004 AC2 (el retiro debe respetar Retry-After,
// no el default de 30s, cuando el proveedor lo especifica).
package integration_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/adapter/generic"
	"api-llm-gateway/internal/failover"
	"api-llm-gateway/internal/health"
	"api-llm-gateway/internal/registry"
)

// stubChain resuelve siempre a un único modelo del proveedor bajo prueba.
type stubChain struct {
	providerID string
}

func (s stubChain) Resolve(capability string, estimatedTokens int) ([]registry.Model, error) {
	return []registry.Model{{ProviderID: s.providerID, Name: "test-model"}}, nil
}

func TestRetryAfter_PropagaDesdeAdapterHastaHealthMonitor(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":"rate limited"}`))
	}))
	defer srv.Close()

	ad, err := generic.New(generic.ProviderSpec{
		BaseURL: srv.URL,
		Format:  generic.FormatOpenAI,
	}, "fake-key")
	if err != nil {
		t.Fatalf("generic.New: %v", err)
	}

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	hm := health.New([]string{"groq"}, func(string) bool { return true }, 1, 1)
	hm.SetClock(func() time.Time { return now })

	eng := failover.New(stubChain{providerID: "groq"}, map[string]adapter.Adapter{"groq": ad})
	eng.OnRateLimited = hm.RetireOn429

	_, err = eng.Complete(context.Background(), "chat", adapter.Request{Model: "test-model"})
	if err == nil {
		t.Fatal("se esperaba error 502 (pool agotado tras 429 único eslabón)")
	}
	var ferr *failover.Error
	if !errors.As(err, &ferr) || ferr.Status != 502 {
		t.Fatalf("se esperaba failover.Error status 502, got %v", err)
	}

	// Retry-After: 60 -> debe seguir retirado tras 30s (el default viejo) pero
	// activo tras 61s. Si RetryAfter no se hubiese propagado (quedando en 0),
	// el Monitor habría aplicado el default de 30s y ya estaría sano acá.
	now = now.Add(30 * time.Second)
	if hm.Healthy("groq", "") {
		t.Fatal("con Retry-After=60 no debe reactivar a los 30s (eso indicaría que RetryAfter no se propagó y se aplicó el default de 30s)")
	}
	now = now.Add(31 * time.Second) // total 61s > 60s
	if !hm.Healthy("groq", "") {
		t.Error("debe reactivar automáticamente al vencer Retry-After=60s")
	}
}
