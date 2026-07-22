package quota

import (
	"context"
	"errors"
	"testing"

	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter"
)

type mockAdapter struct {
	response adapter.Response
	err      error
}

func (m *mockAdapter) Chat(ctx context.Context, req adapter.Request) (adapter.Response, error) {
	return m.response, m.err
}

func (m *mockAdapter) Stream(ctx context.Context, req adapter.Request) (adapter.TokenStream, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAdapter) Embed(ctx context.Context, req adapter.Request) (adapter.Embedding, error) {
	return adapter.Embedding{}, errors.New("not implemented")
}

func TestMiddleware_Chat_Success(t *testing.T) {
	manager := NewInMemoryManager()
	manager.SetLimit("openai", Consumption{Tokens: 100, Requests: 5})

	mockAd := &mockAdapter{
		response: adapter.Response{Content: "Hello world!"}, // 12 chars roughly
	}

	mw := NewMiddleware(manager, "openai", mockAd)

	req := adapter.Request{
		Messages: []adapter.Message{{Content: "Hi"}},
	}
	resp, err := mw.Chat(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "Hello world!" {
		t.Errorf("expected Hello world!, got %v", resp.Content)
	}

	remaining := manager.Remaining("openai", "")
	if remaining != 100-(12+2) {
		t.Errorf("expected remaining %d, got %d", 100-(12+2), remaining)
	}
}

func TestMiddleware_Chat_QuotaExceeded(t *testing.T) {
	manager := NewInMemoryManager()
	manager.SetLimit("openai", Consumption{Tokens: 1, Requests: 5})

	mockAd := &mockAdapter{}
	mw := NewMiddleware(manager, "openai", mockAd)

	req := adapter.Request{
		Messages: []adapter.Message{{Content: "Hi"}},
	}
	_, err := mw.Chat(context.Background(), req)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var pe *adapter.ProviderError
	if !errors.As(err, &pe) || pe.Status != 429 {
		t.Errorf("expected ProviderError with status 429, got %v", err)
	}
}
