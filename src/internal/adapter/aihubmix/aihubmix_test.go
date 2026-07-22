package aihubmix_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter"
	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter/aihubmix"
)

// HU-029 AC1 — Chat básico exitoso
func TestAIHubMix_Chat_Happy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path esperado /v1/chat/completions, obtuve %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"respuesta aihubmix"}}]}`))
	}))
	defer srv.Close()

	ad := aihubmix.New(srv.URL, "sk-test")
	resp, err := ad.Chat(context.Background(), adapter.Request{
		Model:    "gpt-4o",
		Messages: []adapter.Message{{Role: "user", Content: "hola"}},
	})
	if err != nil {
		t.Fatalf("Chat error inesperado: %v", err)
	}
	if resp.Content != "respuesta aihubmix" {
		t.Errorf("content esperado 'respuesta aihubmix', obtuve %q", resp.Content)
	}
}

// HU-029 AC2 — Rate limit 429 → ProviderError retryable
func TestAIHubMix_Chat_RateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"rate limit"}`))
	}))
	defer srv.Close()

	ad := aihubmix.New(srv.URL, "sk-test")
	_, err := ad.Chat(context.Background(), adapter.Request{
		Model:    "gpt-4o",
		Messages: []adapter.Message{{Role: "user", Content: "x"}},
	})
	var pe *adapter.ProviderError
	if !errors.As(err, &pe) {
		t.Fatalf("esperaba *adapter.ProviderError, obtuve %v", err)
	}
	if pe.Status != 429 || !pe.Retryable {
		t.Errorf("esperaba 429 retryable, obtuve status=%d retryable=%v", pe.Status, pe.Retryable)
	}
}

// HU-029 AC3 — Upstream 500/503 → ProviderError retryable
func TestAIHubMix_Chat_Upstream500(t *testing.T) {
	for _, code := range []int{500, 503} {
		code := code
		t.Run(http.StatusText(code), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(code)
			}))
			defer srv.Close()

			ad := aihubmix.New(srv.URL, "sk-test")
			_, err := ad.Chat(context.Background(), adapter.Request{
				Model:    "gpt-4o",
				Messages: []adapter.Message{{Role: "user", Content: "x"}},
			})
			var pe *adapter.ProviderError
			if !errors.As(err, &pe) {
				t.Fatalf("esperaba *adapter.ProviderError, obtuve %v", err)
			}
			if !pe.Retryable {
				t.Errorf("status %d debe ser retryable", code)
			}
		})
	}
}

// HU-029 AC4 — Params no soportados ignorados (respuesta exitosa de todas formas)
func TestAIHubMix_Chat_UnknownParamsIgnored(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`))
	}))
	defer srv.Close()

	ad := aihubmix.New(srv.URL, "sk-test")
	resp, err := ad.Chat(context.Background(), adapter.Request{
		Model:    "gpt-4o",
		Messages: []adapter.Message{{Role: "user", Content: "test"}},
		Params:   map[string]any{"logprobs": true, "top_logprobs": 5}, // no soportados
	})
	if err != nil {
		t.Fatalf("Chat con params exóticos falló: %v", err)
	}
	if resp.Content != "ok" {
		t.Errorf("content esperado 'ok', obtuve %q", resp.Content)
	}
}
