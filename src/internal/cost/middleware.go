package cost

import (
	"context"

	"api-llm-gateway/internal/adapter"
)

// Middleware wraps an adapter to track and record costs after execution.
type Middleware struct {
	tracker  *Tracker
	provider string
	next     adapter.Adapter
}

// NewMiddleware creates a new cost tracking middleware.
func NewMiddleware(tracker *Tracker, provider string, next adapter.Adapter) *Middleware {
	return &Middleware{
		tracker:  tracker,
		provider: provider,
		next:     next,
	}
}

// Chat executes a request and records its cost.
func (m *Middleware) Chat(ctx context.Context, req adapter.Request) (adapter.Response, error) {
	resp, err := m.next.Chat(ctx, req)

	// Even if there's an error, we might have consumed prompt tokens.
	// We'll calculate prompt tokens anyway.
	promptTokens := m.estimateTokens(req)
	completionTokens := 0
	if err == nil {
		completionTokens = len(resp.Content) // roughly
	}

	record := CostRecord{
		AgentID:          extractAgentID(ctx),
		ProviderID:       m.provider,
		Model:            req.Model,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
	}

	// Track the cost asynchronously or synchronously (Tracker handles it)
	_ = m.tracker.Track(ctx, record)

	return resp, err
}

func (m *Middleware) Stream(ctx context.Context, req adapter.Request) (adapter.TokenStream, error) {
	// Not implemented for brevity
	return m.next.Stream(ctx, req)
}

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

func extractAgentID(ctx context.Context) string {
	val := ctx.Value("agentID")
	if str, ok := val.(string); ok {
		return str
	}
	return "anonymous"
}
