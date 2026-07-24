package main

import (
	"context"
	"testing"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/quota"
)

// stubAdapter always returns a fixed response carrying QuotaInfo headers,
// simulating what generic.Chat now populates from real HTTP headers.
type stubAdapter struct {
	resp adapter.Response
	err  error
}

func (s *stubAdapter) Chat(ctx context.Context, req adapter.Request) (adapter.Response, error) {
	return s.resp, s.err
}
func (s *stubAdapter) Stream(ctx context.Context, req adapter.Request) (adapter.TokenStream, error) {
	return nil, nil
}
func (s *stubAdapter) Embed(ctx context.Context, req adapter.Request) (adapter.Embedding, error) {
	return adapter.Embedding{}, nil
}

// INT-quota-learn (reapertura) — wrapWithQuotaMiddleware es la función que
// main() invoca para envolver los adapters reales; este test verifica que la
// cadena adapter.Chat -> Middleware.Chat -> Manager.LearnFromHeaders queda
// efectivamente cableada, no solo declarada.
func TestWrapWithQuotaMiddleware_LearnsFromRealAdapterResponse(t *testing.T) {
	qm := quota.NewInMemoryManager()
	qm.InitFromRegistry(map[string]*int{"groq": nil})
	qm.SetLimit("groq", quota.Consumption{Tokens: quota.DefaultQuotaHint, Requests: 1000})

	adapters := map[string]adapter.Adapter{
		"groq": &stubAdapter{resp: adapter.Response{
			Content:   "ok",
			QuotaInfo: adapter.QuotaInfo{Limit: 100, Remaining: 7},
		}},
	}

	wrapped := wrapWithQuotaMiddleware(adapters, qm)
	mw, ok := wrapped["groq"].(*quota.Middleware)
	if !ok {
		t.Fatalf("esperaba *quota.Middleware, obtuve %T", wrapped["groq"])
	}

	_, err := mw.Chat(context.Background(), adapter.Request{Model: "mixtral", Messages: []adapter.Message{{Role: "user", Content: "hi"}}})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}

	if got := qm.Remaining("groq", "mixtral"); got != 7 {
		t.Errorf("Remaining tras LearnFromHeaders: esperaba 7 (aprendido de QuotaInfo), obtuve %d", got)
	}
}
