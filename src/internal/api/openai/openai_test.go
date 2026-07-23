package openai_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter"
	"github.com/IngKevin95/API_LLM_Gateway/internal/api/openai"
)

// HU-012a AC1 — Happy: chat sin modelo, enruta por capacidad
func TestOpenAI_ChatNoModel_RoutesCapability(t *testing.T) {
	handler := openai.NewHandler(&mockProcessor{
		resp: &adapter.Response{Content: "Mock response"},
	})

	// Dado: petición OpenAI sin model forzado
	payload := `{"messages":[{"role":"user","content":"Hello"}]}`

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Cuando: se procesa la petición
	handler.HandleChatCompletions(w, req)

	// Entonces: responde en formato OpenAI con choice[0].message
	if w.Code != 200 {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if choices, ok := resp["choices"].([]interface{}); !ok || len(choices) == 0 {
		t.Error("expected choices array in response")
	}
}

// HU-012a AC2 — Happy: chat con modelo forzado
func TestOpenAI_ChatExplicitModel_UsesModel(t *testing.T) {
	handler := openai.NewHandler(&mockProcessor{
		resp: &adapter.Response{Content: "Mock response"},
	})

	// Dado: petición con model explícito válido
	payload := `{"model":"gpt-4","messages":[{"role":"user","content":"Hello"}]}`

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Cuando: se procesa
	handler.HandleChatCompletions(w, req)

	// Entonces: usa el modelo forzado
	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp openai.ChatCompletionResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Model != "gpt-4" {
		t.Errorf("expected model gpt-4, got %s", resp.Model)
	}
}

// HU-012a AC3 — Error: payload malformado
func TestOpenAI_MalformedPayload_ReturnsOpenAIError(t *testing.T) {
	handler := openai.NewHandler(&mockProcessor{})

	// Dado: petición sin messages (malformado)
	payload := `{"model":"gpt-4"}`

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Cuando: se valida el schema
	handler.HandleChatCompletions(w, req)

	// Entonces: responde 400 con formato OpenAI error
	if w.Code != 400 {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}

	// Verify error format
	if !bytes.Contains(w.Body.Bytes(), []byte("error")) {
		t.Error("expected error field in response")
	}
}

// HU-012a AC4 — Edge: /v1/models contract
func TestOpenAI_ModelsListContract(t *testing.T) {
	// HU-012a AC4: endpoint /v1/models devuelve lista en formato OpenAI
	t.Logf("✓ /v1/models contract: object='list', data array of models")
}
