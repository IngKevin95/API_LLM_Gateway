package adapter_test

import (
	"testing"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/adapter/anthropic"
	"api-llm-gateway/internal/adapter/local"
	"api-llm-gateway/internal/adapter/openai"
)

// journey_smoke SS2: los tres adapters conviven bajo el contrato Adapter
// (verificación en tiempo de compilación + rótulo en runtime).
var (
	_ adapter.Adapter = (*openai.Adapter)(nil)
	_ adapter.Adapter = (*anthropic.Adapter)(nil)
	_ adapter.Adapter = local.New("http://localhost")
)

func TestAdaptersImplementContract(t *testing.T) {
	adapters := map[string]adapter.Adapter{
		"openai":    openai.New("http://x", "k"),
		"anthropic": anthropic.New("http://x", "k"),
		"local":     local.New("http://x"),
	}
	if len(adapters) != 3 {
		t.Fatalf("esperaba 3 adapters bajo el contrato, obtuve %d", len(adapters))
	}
}
