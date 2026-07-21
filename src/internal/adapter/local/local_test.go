package local_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter"
	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter/local"
)

func isProviderErr(err error, target **adapter.ProviderError) bool { return errors.As(err, target) }

// HU-024 AC1 — Happy: servidor local OpenAI-compatible → respuesta normalizada.
func TestLocal_Happy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path esperado /v1/chat/completions, obtuve %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"local ok"}}]}`))
	}))
	defer srv.Close()

	ad := local.New(srv.URL)
	resp, err := ad.Chat(context.Background(), adapter.Request{Model: "mistral-7b", Messages: []adapter.Message{{Role: "user", Content: "hola"}}})
	if err != nil {
		t.Fatalf("Chat error: %v", err)
	}
	if resp.Content != "local ok" {
		t.Errorf("content esperado 'local ok', obtuve %q", resp.Content)
	}
}

// HU-024 AC2 — Error: servidor local colgado → timeout (ctx) → *ProviderError.
func TestLocal_Timeout(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		<-release // se cuelga
	}))
	defer srv.Close()
	defer close(release)

	ad := local.New(srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	_, err := ad.Chat(ctx, adapter.Request{Model: "mistral-7b", Messages: []adapter.Message{{Role: "user", Content: "x"}}})
	var pe *adapter.ProviderError
	if !isProviderErr(err, &pe) || !pe.Retryable {
		t.Fatalf("esperaba *ProviderError retryable por timeout, obtuve %v", err)
	}
}

// HU-024 AC3 — Edge: respuesta no OpenAI-compatible → *ProviderError sin crashear.
func TestLocal_NonCompatibleResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"unexpected":"shape","no":"choices"}`)) // 200 pero no OpenAI-compat
	}))
	defer srv.Close()

	ad := local.New(srv.URL)
	_, err := ad.Chat(context.Background(), adapter.Request{Model: "mistral-7b", Messages: []adapter.Message{{Role: "user", Content: "x"}}})
	var pe *adapter.ProviderError
	if !isProviderErr(err, &pe) {
		t.Fatalf("esperaba *ProviderError ante respuesta no compatible, obtuve %v", err)
	}
}
