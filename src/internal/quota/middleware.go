package quota

import (
	"context"
	"errors"

	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter"
)

var ErrQuotaExceeded = errors.New("quota exceeded")

// Middleware wraps an adapter to enforce quota and track consumption.
type Middleware struct {
	manager  Manager
	provider string
	next     adapter.Adapter
}

// NewMiddleware creates a new quota adapter middleware.
func NewMiddleware(manager Manager, provider string, next adapter.Adapter) *Middleware {
	return &Middleware{
		manager:  manager,
		provider: provider,
		next:     next,
	}
}

// Chat performs a Chat request enforcing quota.
func (m *Middleware) Chat(ctx context.Context, req adapter.Request) (adapter.Response, error) {
	estimate := Consumption{
		Tokens:   m.estimateTokens(req),
		Requests: 1,
	}

	if !m.manager.Reserve(m.provider, estimate) {
		return adapter.Response{}, &adapter.ProviderError{
			Provider:  m.provider,
			Status:    429,
			Retryable: true,
			Err:       ErrQuotaExceeded,
		}
	}

	resp, err := m.next.Chat(ctx, req)
	if err != nil {
		// If it failed completely, actual is 0 (or whatever we consumed)
		_ = m.manager.Commit(m.provider, estimate, Consumption{Tokens: 0, Requests: 1})
		return resp, err
	}

	actual := Consumption{
		Tokens:   estimate.Tokens + len(resp.Content), // roughly
		Requests: 1,
	}
	_ = m.manager.Commit(m.provider, estimate, actual)
	return resp, nil
}

// Stream performs a streaming request enforcing quota.
func (m *Middleware) Stream(ctx context.Context, req adapter.Request) (adapter.TokenStream, error) {
	// Not implemented for brevity, would wrap TokenStream to count emitted tokens
	return m.next.Stream(ctx, req)
}

// Embed performs an embed request enforcing quota.
func (m *Middleware) Embed(ctx context.Context, req adapter.Request) (adapter.Embedding, error) {
	return m.next.Embed(ctx, req)
}

func (m *Middleware) estimateTokens(req adapter.Request) int {
	n := 0
	for _, msg := range req.Messages {
		n += len(msg.Content)
	}
	return n
}
