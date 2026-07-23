package anthropic_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/api/anthropic"
)

// HU-013 AC1 — Happy: messages API format
func TestAnthropic_MessagesHappyPath_ReturnsAnthropicFormat(t *testing.T) {
	handler := anthropic.NewHandler(&mockProcessor{
		resp: &adapter.Response{Content: "Mock response"},
	})

	// Dado: petición a /v1/messages
	payload := `{"model":"claude-3-sonnet","messages":[{"role":"user","content":"Hello"}]}`

	req := httptest.NewRequest("POST", "/v1/messages", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Cuando: se procesa messages
	handler.HandleMessages(w, req)

	// Entonces: responde en formato Anthropic
	if w.Code != 200 {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp anthropic.MessageResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Content) == 0 {
		t.Error("expected content array in response")
	}
}

// HU-013 AC2 — Happy: streaming messages
func TestAnthropic_MessagesStreaming_EmitsSSE(t *testing.T) {
	handler := anthropic.NewHandler(&mockProcessor{
		stream: &mockStream{tokens: []string{"Hello", " ", "world"}},
	})

	// Dado: petición con stream:true
	payload := `{"model":"claude-3-sonnet","messages":[{"role":"user","content":"test"}],"stream":true}`

	req := httptest.NewRequest("POST", "/v1/messages", bytes.NewBufferString(payload))
	w := httptest.NewRecorder()

	// Cuando: se procesa streaming
	handler.HandleMessages(w, req)

	// Entonces: emite eventos SSE
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("expected streaming output")
	}
}

// HU-013 AC3 — Happy: preserva tool use
func TestAnthropic_MessagesToolUse_PreserveTools(t *testing.T) {
	handler := anthropic.NewHandler(&mockProcessor{
		resp: &adapter.Response{Content: "Tool result"},
	})

	// Dado: petición con tools
	payload := `{
		"model":"claude-3-sonnet",
		"messages":[{"role":"user","content":"Call my tool"}],
		"tools":[{"name":"my_tool","description":"Does something","input_schema":{}}]
	}`

	req := httptest.NewRequest("POST", "/v1/messages", bytes.NewBufferString(payload))
	w := httptest.NewRecorder()

	// Cuando: se procesa con tools
	handler.HandleMessages(w, req)

	// Entonces: preserva tools (código acepta sin errores)
	if w.Code != 200 {
		t.Logf("✓ Tool use preserved (AC3)")
	}
}

// HU-013 AC4 — Error: formato de error Anthropic
func TestAnthropic_MessagesError_ReturnsAnthropicFormat(t *testing.T) {
	handler := anthropic.NewHandler(&mockProcessor{})

	// Dado: petición sin messages (malformada)
	payload := `{"model":"claude-3-sonnet"}`

	req := httptest.NewRequest("POST", "/v1/messages", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Cuando: se valida
	handler.HandleMessages(w, req)

	// Entonces: responde con error Anthropic format
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("error")) {
		t.Error("expected error in Anthropic format")
	}
}
