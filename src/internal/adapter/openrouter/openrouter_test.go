package openrouter_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/adapter/openrouter"
)

// HU-031 AC1 — Headers HTTP-Referer y X-Title presentes en cada request
func TestOpenRouter_Chat_HeadersPresent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("HTTP-Referer"); got == "" {
			t.Errorf("esperaba HTTP-Referer, no está")
		}
		if got := r.Header.Get("X-Title"); got == "" {
			t.Errorf("esperaba X-Title, no está")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"ok openrouter"}}]}`))
	}))
	defer srv.Close()

	ad := openrouter.New(openrouter.Config{
		BaseURL: srv.URL,
		APIKey:  "sk-or-test",
		Referer: "https://api-llm-gateway",
		Title:   "API LLM Gateway",
	})
	resp, err := ad.Chat(context.Background(), adapter.Request{
		Model:    "anthropic/claude-3-haiku",
		Messages: []adapter.Message{{Role: "user", Content: "hola"}},
	})
	if err != nil {
		t.Fatalf("Chat error: %v", err)
	}
	if resp.Content != "ok openrouter" {
		t.Errorf("content esperado 'ok openrouter', obtuve %q", resp.Content)
	}
}

// HU-031 AC2 — Modelo upstream no disponible 503 → ProviderError retryable
func TestOpenRouter_Chat_UpstreamUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":"upstream unavailable"}`))
	}))
	defer srv.Close()

	ad := openrouter.New(openrouter.Config{BaseURL: srv.URL, APIKey: "sk-or-test"})
	_, err := ad.Chat(context.Background(), adapter.Request{
		Model:    "anthropic/claude-3-haiku",
		Messages: []adapter.Message{{Role: "user", Content: "x"}},
	})
	var pe *adapter.ProviderError
	if !errors.As(err, &pe) {
		t.Fatalf("esperaba *adapter.ProviderError, obtuve %v", err)
	}
	if !pe.Retryable {
		t.Errorf("503 debe ser retryable")
	}
}

// HU-031 AC3 — Timeout / ctx cancelado → error
func TestOpenRouter_Chat_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Nunca responde para simular timeout
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancelamos inmediatamente

	ad := openrouter.New(openrouter.Config{BaseURL: srv.URL, APIKey: "sk-or-test"})
	_, err := ad.Chat(ctx, adapter.Request{
		Model:    "anthropic/claude-3-haiku",
		Messages: []adapter.Message{{Role: "user", Content: "x"}},
	})
	if err == nil {
		t.Fatal("esperaba error por contexto cancelado")
	}
}
