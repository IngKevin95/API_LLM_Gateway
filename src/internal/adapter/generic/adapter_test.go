package generic

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"api-llm-gateway/internal/adapter"
)

// HU-EVO-001 AC1 — Happy: adapter OpenAI-compatible (Groq) traduce y normaliza.
func TestChat_OpenAIFormat_Groq(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]string{"role": "assistant", "content": "hola desde groq"}}},
		})
	}))
	defer srv.Close()

	a, err := New(ProviderSpec{BaseURL: srv.URL, AuthHeader: "Authorization", Format: FormatOpenAI}, "groq-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	resp, err := a.Chat(context.Background(), adapter.Request{Model: "mixtral-8x7b-32768", Messages: []adapter.Message{{Role: "user", Content: "hola"}}})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if resp.Content != "hola desde groq" {
		t.Errorf("Content: esperaba %q, obtuve %q", "hola desde groq", resp.Content)
	}
	if gotAuth != "Bearer groq-key" {
		t.Errorf("Authorization: esperaba %q, obtuve %q", "Bearer groq-key", gotAuth)
	}
}

// HU-EVO-001 AC2 — Happy: adapter Claude-compatible traduce y normaliza.
func TestChat_ClaudeFormat(t *testing.T) {
	var gotAuth, gotVersion string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("x-api-key")
		gotVersion = r.Header.Get("anthropic-version")
		json.NewEncoder(w).Encode(map[string]any{
			"content": []map[string]any{{"type": "text", "text": "hola claude-compat"}},
		})
	}))
	defer srv.Close()

	a, err := New(ProviderSpec{BaseURL: srv.URL, AuthHeader: "x-api-key", Format: FormatClaude}, "claude-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	resp, err := a.Chat(context.Background(), adapter.Request{Model: "claude-compat-model", Messages: []adapter.Message{{Role: "user", Content: "hola"}}})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if resp.Content != "hola claude-compat" {
		t.Errorf("Content: esperaba %q, obtuve %q", "hola claude-compat", resp.Content)
	}
	if gotAuth != "claude-key" {
		t.Errorf("x-api-key: esperaba %q, obtuve %q", "claude-key", gotAuth)
	}
	if gotVersion == "" {
		t.Error("anthropic-version: esperaba header presente")
	}
}

// HU-EVO-001 AC3 — Error: spec inválido -> ErrInvalidProviderSpec.
func TestNew_InvalidSpec_ReturnsErrInvalidProviderSpec(t *testing.T) {
	cases := []ProviderSpec{
		{BaseURL: "", Format: FormatOpenAI},
		{BaseURL: "http://x", Format: "unsupported-format"},
	}
	for _, spec := range cases {
		if _, err := New(spec, "k"); err == nil {
			t.Errorf("spec %+v: esperaba error, obtuve nil", spec)
		} else if !errors.Is(err, ErrInvalidProviderSpec) {
			t.Errorf("spec %+v: esperaba ErrInvalidProviderSpec, obtuve %v", spec, err)
		}
	}
}

// HU-EVO-001 AC4 — Edge: headers extra por proveedor, sin sobrescribir auth.
func TestChat_ExtraHeaders_DoNotOverrideAuth(t *testing.T) {
	var gotAuth, gotCustom string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotCustom = r.Header.Get("X-Custom-Header")
		json.NewEncoder(w).Encode(map[string]any{"choices": []map[string]any{{"message": map[string]string{"content": "ok"}}}})
	}))
	defer srv.Close()

	a, err := New(ProviderSpec{
		BaseURL:    srv.URL,
		AuthHeader: "Authorization",
		Format:     FormatOpenAI,
		Headers:    map[string]string{"X-Custom-Header": "curated-value", "Authorization": "should-not-win"},
	}, "real-key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := a.Chat(context.Background(), adapter.Request{Model: "m"}); err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if gotAuth != "Bearer real-key" {
		t.Errorf("Authorization no debe ser sobrescrito por Headers extra: obtuve %q", gotAuth)
	}
	if gotCustom != "curated-value" {
		t.Errorf("X-Custom-Header: esperaba %q, obtuve %q", "curated-value", gotCustom)
	}
}

// HU-EVO-001 AC5 — Edge: timeout por adapter distinto del global.
func TestChat_PerAdapterTimeout_IndependentFromGlobal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		json.NewEncoder(w).Encode(map[string]any{"choices": []map[string]any{{"message": map[string]string{"content": "tarde"}}}})
	}))
	defer srv.Close()

	a, err := New(ProviderSpec{BaseURL: srv.URL, Format: FormatOpenAI, TimeoutMs: 5}, "k")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// contexto padre sin timeout (global "infinito"): el corte debe venir del TimeoutMs propio del spec.
	_, err = a.Chat(context.Background(), adapter.Request{Model: "m"})
	if err == nil {
		t.Fatal("esperaba error por timeout propio del adapter (5ms) menor al tiempo de respuesta del server (50ms)")
	}
}
