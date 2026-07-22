package failover_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter"
	"github.com/IngKevin95/API_LLM_Gateway/internal/breaker"
	"github.com/IngKevin95/API_LLM_Gateway/internal/failover"
	"github.com/IngKevin95/API_LLM_Gateway/internal/health"
	"github.com/IngKevin95/API_LLM_Gateway/internal/registry"
	"github.com/IngKevin95/API_LLM_Gateway/internal/router"
	"github.com/IngKevin95/API_LLM_Gateway/internal/tokenizer"
)

// El breaker hace fast-fail de un proveedor tripped: el engine lo saltea.
func TestComplete_BreakerSkipsTripped(t *testing.T) {
	aCalls, bCalls := 0, 0
	eng := failover.New(
		fakeChain{models: []registry.Model{model("gpt", "A"), model("claude", "B")}},
		map[string]adapter.Adapter{
			"A": fakeAdapter{resp: adapter.Response{Content: "A"}, calls: &aCalls},
			"B": fakeAdapter{resp: adapter.Response{Content: "B"}, calls: &bCalls},
		},
	)
	b := breaker.New(breaker.Config{MaxInFlight: 10, Cooldown: time.Second})
	b.Trip("A") // A inalcanzable
	eng.Breaker = b

	resp, err := eng.Complete(context.Background(), "chat", adapter.Request{})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if resp.Content != "B" || aCalls != 0 || bCalls != 1 {
		t.Errorf("breaker debe saltear A (fast-fail) y usar B; A=%d B=%d resp=%q", aCalls, bCalls, resp.Content)
	}
}

// journey_smoke SS4: stack completo — registry + health monitor (HealthSource vivo)
// + breaker + router + failover. Un proveedor marcado no-sano sale de la cadena.
func TestIntegration_FullStack_HealthAndBreaker(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-x")
	t.Setenv("ANTHROPIC_API_KEY", "sk-y")

	reg, err := registry.Load(filepath.Join("..", "..", "config.yaml"), nil)
	if err != nil {
		t.Fatalf("cargar config: %v", err)
	}

	// Health Monitor real como HealthSource del router: marca 'local-ollama' caído.
	down := map[string]bool{"local-ollama": true}
	mon := health.New([]string{"openai", "anthropic", "local-ollama"},
		func(p string) bool { return !down[p] }, 1, 1)
	mon.CheckOnce() // aplica el estado

	rt := router.New(reg, mon, router.StaticQuota{}, tokenizer.NewHeuristic())

	// El router debe excluir a local-ollama (no-sano) de la cadena chat.
	chain, _ := rt.Resolve("chat", 10)
	for _, m := range chain {
		if m.ProviderID == "local-ollama" {
			t.Fatalf("local-ollama no-sano no debe estar en la cadena: %v", chain)
		}
	}

	// Failover con breaker sobre adapters: el primero falla, degrada al siguiente.
	adapters := map[string]adapter.Adapter{}
	for i, m := range chain {
		if i == 0 {
			adapters[m.ProviderID] = fakeAdapter{err: provErr(503)}
		} else {
			adapters[m.ProviderID] = fakeAdapter{resp: adapter.Response{Content: "ok:" + m.ProviderID}}
		}
	}
	eng := failover.New(rt, adapters)
	eng.Breaker = breaker.New(breaker.Config{MaxInFlight: 300, Cooldown: 30 * time.Second})

	resp, err := eng.Complete(context.Background(), "chat", adapter.Request{Messages: []adapter.Message{{Role: "user", Content: "hola"}}})
	if err != nil {
		t.Fatalf("Complete error: %v", err)
	}
	if resp.Content == "" {
		t.Error("esperaba respuesta completada vía failover con salud viva")
	}
	t.Logf("stack completo respondió: %s (cadena %d proveedores sanos)", resp.Content, len(chain))
}
