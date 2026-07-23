package anthropic_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/api/anthropic"
)

type mockProcessor struct {
	resp   *adapter.Response
	err    error
	stream adapter.TokenStream
}

func (m *mockProcessor) ProcessChat(ctx context.Context, req *adapter.Request) (*adapter.Response, error) {
	return m.resp, m.err
}

func (m *mockProcessor) ProcessChatStream(ctx context.Context, req *adapter.Request) (adapter.TokenStream, error) {
	return m.stream, m.err
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

func TestHandleMessages_Success(t *testing.T) {
	processor := &mockProcessor{
		resp: &adapter.Response{Content: "Hola Claude"},
	}
	handler := anthropic.NewHandler(processor)

	reqBody := []byte(`{"model":"claude-3-5-sonnet","messages":[{"role":"user","content":"hola"}]}`)
	req, _ := http.NewRequest("POST", "/v1/messages", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.HandleMessages(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp anthropic.MessageResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("error decoding: %v", err)
	}

	if len(resp.Content) != 1 {
		t.Fatalf("expected 1 content block")
	}
	if resp.Content[0].Text != "Hola Claude" {
		t.Errorf("unexpected text: %q", resp.Content[0].Text)
	}
	if resp.Type != "message" {
		t.Errorf("unexpected type: %q", resp.Type)
	}
}

func TestHandleMessages_Stream(t *testing.T) {
	processor := &mockProcessor{
		stream: &mockStream{tokens: []string{"Hola", " Claude"}},
	}
	handler := anthropic.NewHandler(processor)

	reqBody := []byte(`{"model":"claude-3","stream":true,"messages":[{"role":"user","content":"hola"}]}`)
	req, _ := http.NewRequest("POST", "/v1/messages", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.HandleMessages(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	bodyStr := rr.Body.String()
	if !strings.Contains(bodyStr, "event: message_start") {
		t.Errorf("expected message_start event")
	}
	if !strings.Contains(bodyStr, "event: content_block_delta") {
		t.Errorf("expected content_block_delta event")
	}
	if !strings.Contains(bodyStr, `"text":"Hola"`) {
		t.Errorf("expected 'Hola' token")
	}
	if !strings.Contains(bodyStr, "event: message_stop") {
		t.Errorf("expected message_stop event")
	}
}
