package google_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter"
	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter/google"
)

func geminiOKBody(text string) string {
	return `{"candidates":[{"content":{"parts":[{"text":"` + text + `"}]}}]}`
}

// HU-030 AC1 — Chat básico exitoso
func TestGoogle_Chat_Happy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(geminiOKBody("hola desde gemini")))
	}))
	defer srv.Close()

	ad := google.New(srv.URL, "key-test")
	resp, err := ad.Chat(context.Background(), adapter.Request{
		Model:    "gemini-1.5-pro",
		Messages: []adapter.Message{{Role: "user", Content: "hola"}},
	})
	if err != nil {
		t.Fatalf("Chat error inesperado: %v", err)
	}
	if resp.Content != "hola desde gemini" {
		t.Errorf("content esperado 'hola desde gemini', obtuve %q", resp.Content)
	}
}

// HU-030 AC4 — System prompt extraído a systemInstruction
func TestGoogle_Chat_SystemPromptExtracted(t *testing.T) {
	var captured map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(geminiOKBody("ok")))
	}))
	defer srv.Close()

	ad := google.New(srv.URL, "key-test")
	_, err := ad.Chat(context.Background(), adapter.Request{
		Model: "gemini-1.5-pro",
		Messages: []adapter.Message{
			{Role: "system", Content: "Eres un asistente útil"},
			{Role: "user", Content: "hola"},
		},
	})
	if err != nil {
		t.Fatalf("Chat error: %v", err)
	}

	// Verificar que systemInstruction está presente
	si, ok := captured["systemInstruction"]
	if !ok {
		t.Fatal("esperaba campo 'systemInstruction' en el payload Gemini")
	}
	siMap := si.(map[string]any)
	parts := siMap["parts"].([]any)
	if len(parts) == 0 {
		t.Fatal("systemInstruction.parts vacío")
	}
	text := parts[0].(map[string]any)["text"].(string)
	if !strings.Contains(text, "asistente") {
		t.Errorf("systemInstruction.parts[0].text esperado 'asistente...', obtuve %q", text)
	}

	// Verificar que el mensaje system NO aparece en contents
	contents := captured["contents"].([]any)
	for _, c := range contents {
		cMap := c.(map[string]any)
		if cMap["role"] == "system" {
			t.Error("el rol 'system' no debe aparecer en contents de Gemini")
		}
	}
}

// HU-030 AC3 — 429 cuota concurrente → ProviderError retryable
func TestGoogle_Chat_RateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"message":"quota exceeded"}}`))
	}))
	defer srv.Close()

	ad := google.New(srv.URL, "key-test")
	_, err := ad.Chat(context.Background(), adapter.Request{
		Model:    "gemini-1.5-pro",
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

// HU-030 AC2 — 400 payload grande → ProviderError NON-retryable
func TestGoogle_Chat_PayloadTooLarge_NonRetryable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"image too large"}}`))
	}))
	defer srv.Close()

	ad := google.New(srv.URL, "key-test")
	_, err := ad.Chat(context.Background(), adapter.Request{
		Model:    "gemini-1.5-pro",
		Messages: []adapter.Message{{Role: "user", Content: "imagen grande"}},
	})
	var pe *adapter.ProviderError
	if !errors.As(err, &pe) {
		t.Fatalf("esperaba *adapter.ProviderError, obtuve %v", err)
	}
	if pe.Status != 400 || pe.Retryable {
		t.Errorf("400 debe ser non-retryable, obtuve status=%d retryable=%v", pe.Status, pe.Retryable)
	}
}
