package router_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/IngKevin95/API_LLM_Gateway/internal/registry"
	"github.com/IngKevin95/API_LLM_Gateway/internal/router"
	"github.com/IngKevin95/API_LLM_Gateway/internal/tokenizer"
)

// journey_smoke sub-slice 2: catálogo real (config.yaml) + Router resuelven
// una capacidad end-to-end en memoria, con las fuentes stub de EP-001.
func TestIntegration_RegistryPlusRouter_ResolvesChat(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-x")
	t.Setenv("ANTHROPIC_API_KEY", "sk-y")

	cfg := filepath.Join("..", "..", "config.yaml") // src/config.yaml desde src/internal/router
	reg, err := registry.Load(cfg)
	if err != nil {
		t.Fatalf("cargar config.yaml: %v", err)
	}

	r := router.New(reg, router.StaticHealth{}, router.StaticQuota{}, tokenizer.NewHeuristic())
	chain, err := r.Resolve("chat", 1000)
	if err != nil {
		t.Fatalf("Resolve chat error inesperado: %v", err)
	}
	if len(chain) == 0 {
		t.Fatal("esperaba una cadena de fallback no vacía para 'chat'")
	}
	// El primer eslabón debe ser un modelo real del catálogo con ProviderID resuelto.
	if chain[0].Name == "" || chain[0].ProviderID == "" {
		t.Errorf("primer modelo mal formado: %+v", chain[0])
	}
	t.Logf("cadena chat: %v", func() []string {
		out := make([]string, len(chain))
		for i, m := range chain {
			out[i] = m.ProviderID + "/" + m.Name
		}
		return out
	}())

	// Modo explícito (HU-003 AC1): modelo real sano → exactamente ese modelo.
	exp, err := r.ResolveExplicit("chat", "gpt-4o", false, 1000)
	if err != nil {
		t.Fatalf("ResolveExplicit gpt-4o: %v", err)
	}
	if len(exp) != 1 || exp[0].Name != "gpt-4o" {
		t.Errorf("explícito esperaba [gpt-4o], obtuve %v", exp)
	}

	// Ruta de error (HU-002b AC1): capacidad desconocida → ErrUnknownCapability.
	if _, err := r.Resolve("quantum", 1000); !errors.Is(err, router.ErrUnknownCapability) {
		t.Errorf("capacidad desconocida esperaba ErrUnknownCapability, obtuve %v", err)
	}
}
