package omniroute_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/adapter/omniroute"
)

// HU-036 AC1 — Petición exitosa vía OmniRoute
func TestOmniRoute_Chat_Happy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"ok omniroute"}}]}`))
	}))
	defer srv.Close()

	ad := omniroute.New(omniroute.Config{BaseURL: srv.URL, APIKey: "omni-key"})
	resp, err := ad.Chat(context.Background(), adapter.Request{
		Model:    "gpt-4o",
		Messages: []adapter.Message{{Role: "user", Content: "hola"}},
	})
	if err != nil {
		t.Fatalf("Chat error: %v", err)
	}
	if resp.Content != "ok omniroute" {
		t.Errorf("content esperado 'ok omniroute', obtuve %q", resp.Content)
	}
}

// HU-036 AC3 — 429 → ProviderError retryable (failover al siguiente)
func TestOmniRoute_Chat_RateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"quota exceeded"}`))
	}))
	defer srv.Close()

	ad := omniroute.New(omniroute.Config{BaseURL: srv.URL, APIKey: "omni-key"})
	_, err := ad.Chat(context.Background(), adapter.Request{
		Model:    "gpt-4o",
		Messages: []adapter.Message{{Role: "user", Content: "x"}},
	})
	var pe *adapter.ProviderError
	if !errors.As(err, &pe) {
		t.Fatalf("esperaba *adapter.ProviderError, obtuve %v", err)
	}
	if pe.Status != 429 || !pe.Retryable {
		t.Errorf("429 debe ser retryable, obtuve status=%d retryable=%v", pe.Status, pe.Retryable)
	}
}

// HU-036 AC3 — 500 → ProviderError retryable
func TestOmniRoute_Chat_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ad := omniroute.New(omniroute.Config{BaseURL: srv.URL, APIKey: "omni-key"})
	_, err := ad.Chat(context.Background(), adapter.Request{
		Model:    "gpt-4o",
		Messages: []adapter.Message{{Role: "user", Content: "x"}},
	})
	var pe *adapter.ProviderError
	if !errors.As(err, &pe) {
		t.Fatalf("esperaba *adapter.ProviderError, obtuve %v", err)
	}
	if !pe.Retryable {
		t.Errorf("500 debe ser retryable")
	}
}
