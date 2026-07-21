package openai_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter"
	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter/openai"
)

func isProviderError(err error, target **adapter.ProviderError) bool {
	return errors.As(err, target)
}

func readBody(r *http.Request) string {
	b, _ := io.ReadAll(r.Body)
	return string(b)
}

// HU-020a AC1 — Happy: chat básico traduce a /v1/chat/completions y normaliza.
func TestChat_Happy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path esperado /v1/chat/completions, obtuve %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer sk-test" {
			t.Errorf("Authorization esperado 'Bearer sk-test', obtuve %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"hola mundo"}}]}`))
	}))
	defer srv.Close()

	ad := openai.New(srv.URL, "sk-test")
	resp, err := ad.Chat(context.Background(), adapter.Request{
		Model:    "gpt-4o",
		Messages: []adapter.Message{{Role: "user", Content: "hola"}},
	})
	if err != nil {
		t.Fatalf("Chat error inesperado: %v", err)
	}
	if resp.Content != "hola mundo" {
		t.Errorf("content esperado 'hola mundo', obtuve %q", resp.Content)
	}
}

// HU-020a AC3 — Error: 500 del proveedor → *ProviderError (retryable) para failover.
func TestChat_ProviderError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":{"message":"boom"}}`))
	}))
	defer srv.Close()

	ad := openai.New(srv.URL, "sk-test")
	_, err := ad.Chat(context.Background(), adapter.Request{Model: "gpt-4o", Messages: []adapter.Message{{Role: "user", Content: "x"}}})
	var pe *adapter.ProviderError
	if !isProviderError(err, &pe) {
		t.Fatalf("esperaba *adapter.ProviderError, obtuve %v", err)
	}
	if pe.Status != 500 || !pe.Retryable {
		t.Errorf("esperaba status 500 retryable, obtuve status=%d retryable=%v", pe.Status, pe.Retryable)
	}
}

// HU-020a AC2 — Edge: tool calling preserva el schema en la respuesta.
func TestChat_ToolCallingPreserved(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := readBody(r)
		if !strings.Contains(body, "get_weather") {
			t.Errorf("el request debe preservar el tool 'get_weather', body=%s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"","tool_calls":[{"id":"c1","function":{"name":"get_weather","arguments":"{}"}}]}}]}`))
	}))
	defer srv.Close()

	ad := openai.New(srv.URL, "sk-test")
	resp, err := ad.Chat(context.Background(), adapter.Request{
		Model:    "gpt-4o",
		Messages: []adapter.Message{{Role: "user", Content: "clima?"}},
		Tools:    []adapter.Tool{{Name: "get_weather", Schema: `{"type":"object"}`}},
	})
	if err != nil {
		t.Fatalf("Chat error inesperado: %v", err)
	}
	if len(resp.ToolCalls) != 1 || resp.ToolCalls[0].Name != "get_weather" {
		t.Errorf("esperaba tool_call get_weather, obtuve %+v", resp.ToolCalls)
	}
}
