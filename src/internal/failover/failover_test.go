package failover_test

import (
	"context"
	"errors"
	"testing"

	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter"
	"github.com/IngKevin95/API_LLM_Gateway/internal/failover"
	"github.com/IngKevin95/API_LLM_Gateway/internal/registry"
)

// --- stubs ---

type fakeChain struct {
	models []registry.Model
	err    error
}

func (f fakeChain) Resolve(_ string, _ int) ([]registry.Model, error) { return f.models, f.err }

type fakeAdapter struct {
	resp  adapter.Response
	err   error
	calls *int
}

func (f fakeAdapter) Chat(_ context.Context, _ adapter.Request) (adapter.Response, error) {
	if f.calls != nil {
		*f.calls++
	}
	return f.resp, f.err
}
func (f fakeAdapter) Stream(context.Context, adapter.Request) (adapter.TokenStream, error) {
	return nil, f.err
}
func (f fakeAdapter) Embed(context.Context, adapter.Request) (adapter.Embedding, error) {
	return adapter.Embedding{}, f.err
}

func model(name, provider string) registry.Model {
	return registry.Model{Name: name, ProviderID: provider, Capabilities: []string{"chat"}, MaxContextToks: 100000}
}

func provErr(status int) error {
	return &adapter.ProviderError{Provider: "x", Status: status, Retryable: adapter.RetryableStatus(status)}
}

// HU-004a AC1 — Happy: providerA 503 → failover a providerB.
func TestComplete_FailoverSimple(t *testing.T) {
	aCalls, bCalls := 0, 0
	eng := failover.New(
		fakeChain{models: []registry.Model{model("gpt", "A"), model("claude", "B")}},
		map[string]adapter.Adapter{
			"A": fakeAdapter{err: provErr(503), calls: &aCalls},
			"B": fakeAdapter{resp: adapter.Response{Content: "desde B"}, calls: &bCalls},
		},
	)
	resp, err := eng.Complete(context.Background(), "chat", adapter.Request{})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if resp.Content != "desde B" {
		t.Errorf("esperaba respuesta de B, obtuve %q", resp.Content)
	}
	if aCalls != 1 || bCalls != 1 {
		t.Errorf("esperaba A=1 B=1 llamadas, obtuve A=%d B=%d", aCalls, bCalls)
	}
}

// HU-004a AC2 — Error: pool agotado → *failover.Error status 502.
func TestComplete_PoolExhausted(t *testing.T) {
	eng := failover.New(
		fakeChain{models: []registry.Model{model("gpt", "A"), model("claude", "B")}},
		map[string]adapter.Adapter{
			"A": fakeAdapter{err: provErr(503)},
			"B": fakeAdapter{err: provErr(500)},
		},
	)
	_, err := eng.Complete(context.Background(), "chat", adapter.Request{})
	var fe *failover.Error
	if !errors.As(err, &fe) || fe.Status != 502 {
		t.Fatalf("esperaba *failover.Error 502, obtuve %v", err)
	}
}

// HU-004a AC3 — Edge: 429 → failover (no retry al mismo).
func TestComplete_429Failover(t *testing.T) {
	aCalls, bCalls := 0, 0
	eng := failover.New(
		fakeChain{models: []registry.Model{model("gpt", "A"), model("claude", "B")}},
		map[string]adapter.Adapter{
			"A": fakeAdapter{err: provErr(429), calls: &aCalls},
			"B": fakeAdapter{resp: adapter.Response{Content: "B"}, calls: &bCalls},
		},
	)
	resp, err := eng.Complete(context.Background(), "chat", adapter.Request{})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if resp.Content != "B" || aCalls != 1 || bCalls != 1 {
		t.Errorf("429 debe failover a B sin reintentar A; A=%d B=%d resp=%q", aCalls, bCalls, resp.Content)
	}
}

// HU-004a AC4 — Edge: degradación a modelo local.
func TestComplete_DegradeToLocal(t *testing.T) {
	eng := failover.New(
		fakeChain{models: []registry.Model{model("gpt", "remote"), model("mistral", "local")}},
		map[string]adapter.Adapter{
			"remote": fakeAdapter{err: provErr(503)},
			"local":  fakeAdapter{resp: adapter.Response{Content: "local"}},
		},
	)
	resp, err := eng.Complete(context.Background(), "chat", adapter.Request{})
	if err != nil || resp.Content != "local" {
		t.Fatalf("esperaba degradación a local, obtuve resp=%q err=%v", resp.Content, err)
	}
}

// HU-004a AC5 — Error: 400 del cliente → sin failover, retorna 400.
func TestComplete_400NoFailover(t *testing.T) {
	aCalls, bCalls := 0, 0
	eng := failover.New(
		fakeChain{models: []registry.Model{model("gpt", "A"), model("claude", "B")}},
		map[string]adapter.Adapter{
			"A": fakeAdapter{err: provErr(400), calls: &aCalls},
			"B": fakeAdapter{resp: adapter.Response{Content: "B"}, calls: &bCalls},
		},
	)
	_, err := eng.Complete(context.Background(), "chat", adapter.Request{})
	var pe *adapter.ProviderError
	if !errors.As(err, &pe) || pe.Status != 400 {
		t.Fatalf("esperaba ProviderError 400 sin failover, obtuve %v", err)
	}
	if aCalls != 1 || bCalls != 0 {
		t.Errorf("400 no debe hacer failover; A=%d B=%d", aCalls, bCalls)
	}
}
