package quota_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter"
	"github.com/IngKevin95/API_LLM_Gateway/internal/failover"
	"github.com/IngKevin95/API_LLM_Gateway/internal/quota"
	"github.com/IngKevin95/API_LLM_Gateway/internal/registry"
	"github.com/IngKevin95/API_LLM_Gateway/internal/router"
	"github.com/IngKevin95/API_LLM_Gateway/internal/tokenizer"
)

type fakeAdapter struct {
	resp adapter.Response
	err  error
}

func (f fakeAdapter) Chat(ctx context.Context, req adapter.Request) (adapter.Response, error) {
	return f.resp, f.err
}

func (f fakeAdapter) Stream(ctx context.Context, req adapter.Request) (adapter.TokenStream, error) {
	return nil, nil
}

func (f fakeAdapter) Embed(ctx context.Context, req adapter.Request) (adapter.Embedding, error) {
	return adapter.Embedding{}, nil
}

// TestIntegration_QuotaFailover ensures that quota middleware works seamlessly
// with the failover engine, deducting quotas for successful providers.
func TestIntegration_QuotaFailover(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-x")
	t.Setenv("ANTHROPIC_API_KEY", "sk-y")

	reg, err := registry.Load(filepath.Join("..", "..", "config.yaml"), nil)
	if err != nil {
		t.Fatalf("cargar config: %v", err)
	}

	qm := quota.NewInMemoryManager()
	// Set quota for all providers.
	qm.SetLimit("openai", quota.Consumption{Tokens: 1000, Requests: 10})
	qm.SetLimit("anthropic", quota.Consumption{Tokens: 1000, Requests: 10})
	qm.SetLimit("local-ollama", quota.Consumption{Tokens: 1000, Requests: 10})

	// Inject QuotaManager into the Router as well (so Router knows if global provider quota is exhausted).
	// Currently quota manager's Remaining() just takes a key, which corresponds to the provider in this test.
	type routerQuota struct{ qm quota.Manager }
	
	// Create router with StaticQuota for now, or adapt router interface if needed.
	// Since EP-003 says QuotaManager is used by Router, we'll implement it.
	rt := router.New(reg, router.StaticHealth{}, router.StaticQuota{}, tokenizer.NewHeuristic())

	chain, _ := rt.Resolve("chat", 10)
	adapters := map[string]adapter.Adapter{}

	for i, m := range chain {
		var ad adapter.Adapter
		if i == 0 {
			// First fails with a retryable error
			ad = fakeAdapter{err: &adapter.ProviderError{Provider: m.ProviderID, Status: 503, Retryable: true}}
		} else {
			// Second succeeds
			ad = fakeAdapter{resp: adapter.Response{Content: "ok:" + m.ProviderID}}
		}
		// Wrap each adapter in QuotaMiddleware
		adapters[m.ProviderID] = quota.NewMiddleware(qm, m.ProviderID, ad)
	}

	eng := failover.New(rt, adapters)
	req := adapter.Request{
		Messages: []adapter.Message{{Role: "user", Content: "hola"}},
	}
	resp, err := eng.Complete(context.Background(), "chat", req)
	
	if err != nil {
		t.Fatalf("Complete error inesperado: %v", err)
	}
	if resp.Content == "" || resp.Content[:3] != "ok:" {
		t.Errorf("esperaba respuesta de un proveedor de failover, obtuve %q", resp.Content)
	}

	// Verify quota was consumed by the successful provider and NOT by the failed one.
	firstProvider := chain[0].ProviderID
	secondProvider := chain[1].ProviderID

	t.Logf("chain: %d models, first=%s, second=%s", len(chain), firstProvider, secondProvider)

	rem1 := qm.Remaining(firstProvider, "")
	if rem1 != 1000 { // Error 503 means no tokens actually consumed!
		t.Errorf("expected primary provider quota to be fully refunded, got %d", rem1)
	}

	rem2 := qm.Remaining(secondProvider, "")
	if rem2 >= 1000 {
		t.Errorf("expected secondary provider quota to be consumed, got %d", rem2)
	}
}
