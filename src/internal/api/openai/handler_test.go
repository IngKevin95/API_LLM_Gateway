package openai_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter"
	"github.com/IngKevin95/API_LLM_Gateway/internal/api/openai"
)

// mockProcessor stub for TDD
type mockProcessor struct {
	resp *adapter.Response
	err  error
}

func (m *mockProcessor) ProcessChat(ctx context.Context, req *adapter.Request) (*adapter.Response, error) {
	return m.resp, m.err
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
