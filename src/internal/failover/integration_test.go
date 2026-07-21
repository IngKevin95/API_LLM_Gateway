package failover_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter"
	"github.com/IngKevin95/API_LLM_Gateway/internal/failover"
	"github.com/IngKevin95/API_LLM_Gateway/internal/registry"
	"github.com/IngKevin95/API_LLM_Gateway/internal/router"
	"github.com/IngKevin95/API_LLM_Gateway/internal/tokenizer"
)

// journey_smoke SS3: catálogo real (config.yaml) + router (EP-001) + failover
// completan una petición degradando por la cadena cuando el primario falla.
func TestIntegration_RealChain_FailsOverToWorking(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-x")
	t.Setenv("ANTHROPIC_API_KEY", "sk-y")

	reg, err := registry.Load(filepath.Join("..", "..", "config.yaml"))
	if err != nil {
		t.Fatalf("cargar config: %v", err)
	}
	rt := router.New(reg, router.StaticHealth{}, router.StaticQuota{}, tokenizer.NewHeuristic())

	// Adapters por proveedor real del config.yaml. El primero de la cadena falla;
	// el failover degrada al siguiente que responde.
	adapters := map[string]adapter.Adapter{}
	chain, _ := rt.Resolve("chat", 10)
	if len(chain) < 2 {
		t.Fatalf("esperaba cadena chat con ≥2 proveedores, obtuve %d", len(chain))
	}
	for i, m := range chain {
		if i == 0 {
			adapters[m.ProviderID] = fakeAdapter{err: provErr(503)} // primario cae
		} else {
			adapters[m.ProviderID] = fakeAdapter{resp: adapter.Response{Content: "ok:" + m.ProviderID}}
		}
	}

	eng := failover.New(rt, adapters)
	resp, err := eng.Complete(context.Background(), "chat", adapter.Request{
		Messages: []adapter.Message{{Role: "user", Content: "hola"}},
	})
	if err != nil {
		t.Fatalf("Complete error inesperado: %v", err)
	}
	if resp.Content == "" || resp.Content[:3] != "ok:" {
		t.Errorf("esperaba respuesta de un proveedor de failover, obtuve %q", resp.Content)
	}
	t.Logf("failover completó con: %s", resp.Content)
}
