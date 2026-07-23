package router

import (
	"testing"

	"api-llm-gateway/internal/registry"
)

func TestBackwardCompat_ExplicitModelNameWorks(t *testing.T) {
	// Old behavior: request with explicit model name should work
	source := &MockModelSource{
		models: map[string][]registry.Model{
			"chat": {
				{Name: "gpt-4", ProviderID: "openai", QualityScore: 95, LatencyP50ms: 100, CostPer1M: 3, MaxContextToks: 8192, Disabled: false},
				{Name: "claude-opus", ProviderID: "anthropic", QualityScore: 90, LatencyP50ms: 150, CostPer1M: 15, MaxContextToks: 200000, Disabled: false},
			},
		},
	}

	router := New(source, &MockHealthSource{}, &MockQuotaSource{}, &MockTokenizer{})

	// Old API: resolve explicit model by name (via ResolveExplicit)
	models, err := router.ResolveExplicit("chat", "gpt-4", false, 1000)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) == 0 {
		t.Fatal("expected explicit model resolution")
	}
	if models[0].Name != "gpt-4" {
		t.Errorf("expected gpt-4, got %s", models[0].Name)
	}
}

func TestBackwardCompat_CapabilityPrefixNotBroken(t *testing.T) {
	// New feature: "router:capability" should also work
	source := &MockModelSource{
		models: map[string][]registry.Model{
			"chat": {
				{Name: "gpt-4", ProviderID: "openai", QualityScore: 95, LatencyP50ms: 100, CostPer1M: 3, MaxContextToks: 8192, Disabled: false},
			},
		},
	}

	router := New(source, &MockHealthSource{}, &MockQuotaSource{}, &MockTokenizer{})

	// New API: "router:chat" should resolve by capability
	models, err := router.ResolveCapabilityPrefix("router:chat", 1000)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) == 0 {
		t.Fatal("expected capability-based resolution")
	}
}

func TestBackwardCompat_BothAPIsMixable(t *testing.T) {
	// Can use both old (explicit model) and new (capability prefix) in same request flow
	source := &MockModelSource{
		models: map[string][]registry.Model{
			"chat": {
				{Name: "gpt-4", ProviderID: "openai", QualityScore: 95, LatencyP50ms: 100, CostPer1M: 3, MaxContextToks: 8192, Disabled: false},
				{Name: "claude-opus", ProviderID: "anthropic", QualityScore: 90, LatencyP50ms: 150, CostPer1M: 15, MaxContextToks: 200000, Disabled: false},
			},
		},
	}

	router := New(source, &MockHealthSource{}, &MockQuotaSource{}, &MockTokenizer{})

	// Request 1: old API - explicit model
	models1, err1 := router.ResolveExplicit("chat", "gpt-4", false, 1000)
	if err1 != nil || len(models1) == 0 {
		t.Fatal("explicit model resolution failed")
	}

	// Request 2: new API - capability prefix
	models2, err2 := router.ResolveCapabilityPrefix("router:chat", 1000)
	if err2 != nil || len(models2) == 0 {
		t.Fatal("capability resolution failed")
	}

	// Both should work independently
	if models1[0].Name != "gpt-4" {
		t.Errorf("explicit model should return gpt-4")
	}
	if len(models2) == 0 {
		t.Errorf("capability resolution should return models")
	}
}

func TestBackwardCompat_ExplicitModelNotAffectedByCapabilityInference(t *testing.T) {
	// When client sends explicit model (no router: prefix),
	// capability inference should not interfere
	source := &MockModelSource{
		models: map[string][]registry.Model{
			"chat": {
				{Name: "gpt-4", ProviderID: "openai", QualityScore: 95, LatencyP50ms: 100, CostPer1M: 3, MaxContextToks: 8192, Disabled: false},
			},
			"vision": {
				{Name: "gpt-4-vision", ProviderID: "openai", QualityScore: 90, LatencyP50ms: 150, CostPer1M: 5, MaxContextToks: 8192, Disabled: false},
			},
		},
	}

	router := New(source, &MockHealthSource{}, &MockQuotaSource{}, &MockTokenizer{})

	// Explicit model name should bypass capability inference entirely
	models, err := router.ResolveCapabilityPrefix("gpt-4", 1000)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) == 0 {
		t.Fatal("explicit model resolution failed")
	}
	if models[0].Name != "gpt-4" {
		t.Errorf("expected gpt-4 (chat model), not inferred capability")
	}
}

func TestBackwardCompat_UnknownExplicitModelStillErrors(t *testing.T) {
	// Old error behavior should be preserved
	source := &MockModelSource{
		models: map[string][]registry.Model{
			"chat": {
				{Name: "gpt-4", ProviderID: "openai", QualityScore: 95, LatencyP50ms: 100, CostPer1M: 3, MaxContextToks: 8192, Disabled: false},
			},
		},
	}

	router := New(source, &MockHealthSource{}, &MockQuotaSource{}, &MockTokenizer{})

	// Request non-existent explicit model
	_, err := router.ResolveExplicit("chat", "nonexistent-model", false, 1000)

	// Should still get ModelNotFoundError (backward compat)
	if err == nil {
		t.Fatal("expected error for unknown model")
	}
	if _, ok := err.(*ModelNotFoundError); !ok {
		t.Errorf("expected ModelNotFoundError, got %T", err)
	}
}

func TestBackwardCompat_OldStyleModelNameViaResolveCapabilityPrefix(t *testing.T) {
	// ResolveCapabilityPrefix should handle both "router:X" and plain model names
	source := &MockModelSource{
		models: map[string][]registry.Model{
			"chat": {
				{Name: "gpt-4", ProviderID: "openai", QualityScore: 95, LatencyP50ms: 100, CostPer1M: 3, MaxContextToks: 8192, Disabled: false},
			},
		},
	}

	router := New(source, &MockHealthSource{}, &MockQuotaSource{}, &MockTokenizer{})

	// Plain model name (no "router:" prefix) should still work
	models, err := router.ResolveCapabilityPrefix("gpt-4", 1000)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) == 0 {
		t.Fatal("expected models for explicit name")
	}
	if models[0].Name != "gpt-4" {
		t.Errorf("expected gpt-4, got %s", models[0].Name)
	}
}
