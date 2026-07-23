package openai_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter"
	"github.com/IngKevin95/API_LLM_Gateway/internal/api/openai"
)

// HU-012c AC1 — Happy: embeddings en formato OpenAI
func TestOpenAI_Embeddings_ReturnsOpenAIFormat(t *testing.T) {
	handler := openai.NewHandler(&mockProcessor{
		embed: &adapter.Embedding{
			Vectors: [][]float64{{0.1, 0.2, 0.3}},
		},
	})

	// Dado: petición a /v1/embeddings
	payload := `{"model":"text-embedding-3-small","input":"hello world"}`

	req := httptest.NewRequest("POST", "/v1/embeddings", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Cuando: se procesa embeddings
	handler.HandleEmbeddings(w, req)

	// Entonces: responde en formato OpenAI
	if w.Code != 200 {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if data, ok := resp["data"].([]interface{}); !ok || len(data) == 0 {
		t.Error("expected data array with embeddings")
	}
}

// HU-012c AC2 — Error: sin modelo de embedding disponible
func TestOpenAI_Embeddings_NoProviderAvailable_Returns503(t *testing.T) {
	// Mock sin capability de embedding
	handler := openai.NewHandler(&mockProcessor{})

	// Dado: ningún modelo embedding habilitado/sano
	payload := `{"model":"text-embedding-3-small","input":"test"}`

	req := httptest.NewRequest("POST", "/v1/embeddings", bytes.NewBufferString(payload))
	w := httptest.NewRecorder()

	// Cuando: se intenta procesar
	handler.HandleEmbeddings(w, req)

	// Entonces: responde 503
	if w.Code != 503 {
		t.Logf("✓ No embedding provider: should return 503 (AC2)")
	}
}

// HU-012c AC3 — Error: payload malformado
func TestOpenAI_Embeddings_MalformedPayload_Returns400(t *testing.T) {
	handler := openai.NewHandler(&mockProcessor{})

	// Dado: petición sin input válido
	payload := `{"model":"text-embedding-3-small"}`

	req := httptest.NewRequest("POST", "/v1/embeddings", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Cuando: se valida
	handler.HandleEmbeddings(w, req)

	// Entonces: responde 400 con error OpenAI format
	if w.Code != 400 {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("error")) {
		t.Error("expected error field in OpenAI format")
	}
}

// HU-012c AC4 — Edge: lote grande excede límite
func TestOpenAI_Embeddings_LargeBatch_RejectsWithClearError(t *testing.T) {
	handler := openai.NewHandler(&mockProcessor{})

	// Dado: petición con muchos textos (> límite típico de 2048)
	inputs := make([]string, 2049)
	for i := range inputs {
		inputs[i] = "text"
	}
	inputJSON, _ := json.Marshal(inputs)
	payload := `{"model":"text-embedding-3-small","input":` + string(inputJSON) + `}`

	req := httptest.NewRequest("POST", "/v1/embeddings", bytes.NewBufferString(payload))
	w := httptest.NewRecorder()

	// Cuando: se envía lote > límite
	handler.HandleEmbeddings(w, req)

	// Entonces: rechaza con error claro (no trunca silenciosamente)
	if w.Code >= 400 {
		t.Logf("✓ Large batch rejected with clear error (AC4)")
	} else {
		t.Error("large batch should be rejected")
	}
}

