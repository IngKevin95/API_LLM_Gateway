package anthropic_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/adapter/anthropic"
)

// captura el body del request para verificar la traducción.
func captureServer(t *testing.T, status int, respBody string, sink *map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if sink != nil {
			_ = json.Unmarshal(b, sink)
		}
		if status != 0 && status != 200 {
			w.WriteHeader(status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(respBody))
	}))
}

// HU-021a AC1 — Happy: extrae system y mapea roles al Messages API.
func TestChat_RoleTranslation(t *testing.T) {
	var body map[string]any
	srv := captureServer(t, 200, `{"content":[{"type":"text","text":"hola"}]}`, &body)
	defer srv.Close()

	ad := anthropic.New(srv.URL, "sk-ant")
	resp, err := ad.Chat(context.Background(), adapter.Request{
		Model: "claude-opus-4",
		Messages: []adapter.Message{
			{Role: "system", Content: "sos un asistente"},
			{Role: "user", Content: "hola"},
		},
	})
	if err != nil {
		t.Fatalf("Chat error: %v", err)
	}
	if resp.Content != "hola" {
		t.Errorf("content esperado 'hola', obtuve %q", resp.Content)
	}
	if body["system"] != "sos un asistente" {
		t.Errorf("system esperado extraído, body=%v", body)
	}
	msgs, _ := body["messages"].([]any)
	if len(msgs) != 1 {
		t.Errorf("esperaba 1 message (sin el system), obtuve %v", body["messages"])
	}
}

// HU-021a AC2 — Happy: tools OpenAI → tool_use de Anthropic.
func TestChat_ToolCalling(t *testing.T) {
	srv := captureServer(t, 200, `{"content":[{"type":"tool_use","id":"t1","name":"get_weather","input":{}}]}`, nil)
	defer srv.Close()

	ad := anthropic.New(srv.URL, "sk-ant")
	resp, err := ad.Chat(context.Background(), adapter.Request{
		Model:    "claude-opus-4",
		Messages: []adapter.Message{{Role: "user", Content: "clima?"}},
		Tools:    []adapter.Tool{{Name: "get_weather", Schema: `{"type":"object"}`}},
	})
	if err != nil {
		t.Fatalf("Chat error: %v", err)
	}
	if len(resp.ToolCalls) != 1 || resp.ToolCalls[0].Name != "get_weather" {
		t.Errorf("esperaba tool_call get_weather, obtuve %+v", resp.ToolCalls)
	}
}

// HU-021a AC3 — Error: parámetro no soportado (seed) se ignora con WARN.
func TestChat_UnsupportedParamWarns(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	srv := captureServer(t, 200, `{"content":[{"type":"text","text":"ok"}]}`, nil)
	defer srv.Close()

	ad := anthropic.New(srv.URL, "sk-ant")
	_, err := ad.Chat(context.Background(), adapter.Request{
		Model:    "claude-opus-4",
		Messages: []adapter.Message{{Role: "user", Content: "x"}},
		Params:   map[string]any{"seed": 42},
	})
	if err != nil {
		t.Fatalf("no debe fallar por un param no soportado: %v", err)
	}
	if !strings.Contains(buf.String(), "seed") || !strings.Contains(strings.ToUpper(buf.String()), "WARN") {
		t.Errorf("esperaba WARN nombrando 'seed', log=%q", buf.String())
	}
}

// HU-021a AC4 — Edge: max_tokens ausente → default 4096.
func TestChat_DefaultMaxTokens(t *testing.T) {
	var body map[string]any
	srv := captureServer(t, 200, `{"content":[{"type":"text","text":"ok"}]}`, &body)
	defer srv.Close()

	ad := anthropic.New(srv.URL, "sk-ant")
	_, err := ad.Chat(context.Background(), adapter.Request{Model: "claude-opus-4", Messages: []adapter.Message{{Role: "user", Content: "x"}}})
	if err != nil {
		t.Fatalf("Chat error: %v", err)
	}
	if mt, _ := body["max_tokens"].(float64); mt != 4096 {
		t.Errorf("esperaba max_tokens default 4096, obtuve %v", body["max_tokens"])
	}
}

// HU-021a AC5 — Error: 5xx/429 → *ProviderError para failover.
func TestChat_NetworkError(t *testing.T) {
	srv := captureServer(t, 429, "", nil)
	defer srv.Close()

	ad := anthropic.New(srv.URL, "sk-ant")
	_, err := ad.Chat(context.Background(), adapter.Request{Model: "claude-opus-4", Messages: []adapter.Message{{Role: "user", Content: "x"}}})
	var pe *adapter.ProviderError
	if !isProviderErr(err, &pe) || !pe.Retryable {
		t.Fatalf("esperaba *ProviderError retryable, obtuve %v", err)
	}
}
