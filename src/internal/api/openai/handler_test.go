package openai_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/api/openai"
)

// mockProcessor stub for TDD
type mockProcessor struct {
	resp   *adapter.Response
	err    error
	stream adapter.TokenStream
	embed  *adapter.Embedding
}

func (m *mockProcessor) ProcessChat(ctx context.Context, req *adapter.Request) (*adapter.Response, error) {
	return m.resp, m.err
}

func (m *mockProcessor) ProcessChatStream(ctx context.Context, req *adapter.Request) (adapter.TokenStream, error) {
	return m.stream, m.err
}

func (m *mockProcessor) ProcessEmbedding(ctx context.Context, req *adapter.Request) (*adapter.Embedding, error) {
	return m.embed, m.err
}

type mockStream struct {
	tokens []string
	pos    int
}

func (m *mockStream) Next() (string, bool, error) {
	if m.pos >= len(m.tokens) {
		return "", false, nil
	}
	t := m.tokens[m.pos]
	m.pos++
	return t, true, nil
}

func (m *mockStream) Close() error {
	return nil
}

func TestHandleChatCompletions_Success(t *testing.T) {
	processor := &mockProcessor{
		resp: &adapter.Response{Content: "Hola, soy el bot"},
	}
	handler := openai.NewHandler(processor)

	reqBody := []byte(`{"model":"gpt-4","messages":[{"role":"user","content":"hola"}]}`)
	req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.HandleChatCompletions(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp openai.ChatCompletionResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message.Content != "Hola, soy el bot" {
		t.Errorf("unexpected content: %q", resp.Choices[0].Message.Content)
	}
	if resp.Object != "chat.completion" {
		t.Errorf("unexpected object: %q", resp.Object)
	}
}

func TestHandleChatCompletions_Stream(t *testing.T) {
	processor := &mockProcessor{
		stream: &mockStream{tokens: []string{"Hola", ", ", "mundo"}},
	}
	handler := openai.NewHandler(processor)

	reqBody := []byte(`{"model":"gpt-4","stream":true,"messages":[{"role":"user","content":"hola"}]}`)
	req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.HandleChatCompletions(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	bodyStr := rr.Body.String()
	if !strings.Contains(bodyStr, "data: [DONE]") {
		t.Errorf("expected [DONE] in response")
	}
	if !strings.Contains(bodyStr, `"content":"Hola"`) {
		t.Errorf("expected 'Hola' in response, got %q", bodyStr)
	}
}

func TestHandleChatCompletions_InvalidJSON(t *testing.T) {
	handler := openai.NewHandler(&mockProcessor{})

	reqBody := []byte(`{invalid-json`)
	req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(reqBody))

	rr := httptest.NewRecorder()
	handler.HandleChatCompletions(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandleEmbeddings(t *testing.T) {
	processor := &mockProcessor{
		embed: &adapter.Embedding{Vectors: [][]float64{{0.1, 0.2}, {0.3, 0.4}}},
	}
	handler := openai.NewHandler(processor)

	reqBody := []byte(`{"model":"text-embedding-3","input":["hola","mundo"]}`)
	req, _ := http.NewRequest("POST", "/v1/embeddings", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.HandleEmbeddings(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp openai.EmbeddingResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("error decoding: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 vectors, got %d", len(resp.Data))
	}
	if resp.Data[0].Embedding[0] != 0.1 {
		t.Errorf("expected 0.1, got %f", resp.Data[0].Embedding[0])
	}
}

func TestHandleModels(t *testing.T) {
	handler := openai.NewHandler(&mockProcessor{})

	req, _ := http.NewRequest("GET", "/v1/models", nil)
	rr := httptest.NewRecorder()
	handler.HandleModels(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp map[string]any
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["object"] != "list" {
		t.Errorf("expected object=list, got %v", resp["object"])
	}
}

// HU-051 AC: Error handling — translate to correct HTTP status codes
func TestHandleChatCompletions_ErrorUnauthorized(t *testing.T) {
	// API key inválida: esperado 401, no 500
	processor := &mockProcessor{
		err: &adapter.ProviderError{
			Provider:  "openai",
			Status:    401,
			Retryable: false,
			Err:       nil,
		},
	}
	handler := openai.NewHandler(processor)

	reqBody := []byte(`{"model":"gpt-4","messages":[{"role":"user","content":"test"}]}`)
	req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.HandleChatCompletions(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestHandleChatCompletions_ErrorUnavailable(t *testing.T) {
	// No provider disponible: esperado 503, no 500
	processor := &mockProcessor{
		err: &adapter.ProviderError{
			Provider:  "gateway",
			Status:    503,
			Retryable: true,
			Err:       nil,
		},
	}
	handler := openai.NewHandler(processor)

	reqBody := []byte(`{"model":"gpt-4","messages":[{"role":"user","content":"test"}]}`)
	req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.HandleChatCompletions(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}
}

func TestHandleChatCompletions_ErrorTimeout(t *testing.T) {
	// Timeout: esperado 504, no 500
	processor := &mockProcessor{
		err: &adapter.ProviderError{
			Provider:  "openai",
			Status:    504,
			Retryable: true,
			Err:       nil,
		},
	}
	handler := openai.NewHandler(processor)

	reqBody := []byte(`{"model":"gpt-4","messages":[{"role":"user","content":"test"}]}`)
	req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.HandleChatCompletions(rr, req)

	if rr.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d", rr.Code)
	}
}
