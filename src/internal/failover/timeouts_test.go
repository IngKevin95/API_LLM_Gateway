package failover_test

import (
	"context"
	"testing"
	"time"

	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter"
	"github.com/IngKevin95/API_LLM_Gateway/internal/failover"
	"github.com/IngKevin95/API_LLM_Gateway/internal/registry"
)

// slowAdapter responde tras `delay`, o falla si el ctx (TTFT) vence antes.
type slowAdapter struct {
	delay time.Duration
	resp  adapter.Response
}

func (s slowAdapter) Chat(ctx context.Context, _ adapter.Request) (adapter.Response, error) {
	select {
	case <-time.After(s.delay):
		return s.resp, nil
	case <-ctx.Done():
		return adapter.Response{}, &adapter.ProviderError{Provider: "slow", Status: 0, Retryable: true, Err: ctx.Err()}
	}
}
func (slowAdapter) Stream(context.Context, adapter.Request) (adapter.TokenStream, error) {
	return nil, nil
}
func (slowAdapter) Embed(context.Context, adapter.Request) (adapter.Embedding, error) {
	return adapter.Embedding{}, nil
}

// HU-004c AC1 — Edge: TTFT estricto excedido (chat) → aborta primario y hace failover.
func TestComplete_TTFTFailover(t *testing.T) {
	eng := failover.New(
		fakeChain{models: []registry.Model{model("gpt", "A"), model("claude", "B")}},
		map[string]adapter.Adapter{
			"A": slowAdapter{delay: time.Second}, // supera el TTFT
			"B": fakeAdapter{resp: adapter.Response{Content: "B"}},
		},
	)
	eng.TTFT = map[string]time.Duration{"chat": 50 * time.Millisecond}
	resp, err := eng.Complete(context.Background(), "chat", adapter.Request{})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if resp.Content != "B" {
		t.Errorf("TTFT excedido debe hacer failover a B, obtuve %q", resp.Content)
	}
}

// HU-004c AC2 — Happy: reasoning con timeout relajado no dispara failover.
func TestComplete_ReasoningRelaxed(t *testing.T) {
	eng := failover.New(
		fakeChain{models: []registry.Model{model("claude", "A")}},
		map[string]adapter.Adapter{
			"A": slowAdapter{delay: 40 * time.Millisecond, resp: adapter.Response{Content: "pensado"}},
		},
	)
	eng.TTFT = map[string]time.Duration{"chat": 10 * time.Millisecond, "reasoning": time.Second}
	resp, err := eng.Complete(context.Background(), "reasoning", adapter.Request{})
	if err != nil {
		t.Fatalf("reasoning no debe fallar con timeout relajado: %v", err)
	}
	if resp.Content != "pensado" {
		t.Errorf("esperaba respuesta del modelo de reasoning, obtuve %q", resp.Content)
	}
}
