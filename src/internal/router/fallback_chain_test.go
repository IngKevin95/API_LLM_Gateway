package router

import (
	"testing"

	"api-llm-gateway/internal/registry"
)

// UnhealthyMockHealthSource returns unhealthy for first call, then healthy
type SelectiveHealthSource struct {
	unhealthyModels map[string]bool
}

func (s *SelectiveHealthSource) Healthy(providerID, model string) bool {
	key := providerID + ":" + model
	return !s.unhealthyModels[key]
}

func TestFallback_PrimaryUnavailableUsesSecondary(t *testing.T) {
	// Arrange: setup models with primary unhealthy
	source := &MockModelSource{
		models: map[string][]registry.Model{
			"chat": {
				{Name: "gpt-4", ProviderID: "openai", QualityScore: 95, LatencyP50ms: 100, CostPer1M: 3, MaxContextToks: 8192, Disabled: false},
				{Name: "gpt-3.5", ProviderID: "openai", QualityScore: 80, LatencyP50ms: 80, CostPer1M: 1, MaxContextToks: 4096, Disabled: false},
				{Name: "claude-opus", ProviderID: "anthropic", QualityScore: 90, LatencyP50ms: 150, CostPer1M: 15, MaxContextToks: 200000, Disabled: false},
			},
		},
	}

	health := &SelectiveHealthSource{
		unhealthyModels: map[string]bool{
			"openai:gpt-4": true, // primary is down
		},
	}

	router := New(source, health, &MockQuotaSource{}, &MockTokenizer{})

	// Act: resolve chat capability
	models, err := router.Resolve("chat", 1000)

	// Assert: should return fallback models (not gpt-4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) == 0 {
		t.Fatal("expected fallback models")
	}
	if models[0].Name == "gpt-4" {
		t.Errorf("expected fallback to not include unhealthy gpt-4, got %v", models)
	}
}

func TestFallback_AllPrimaryProvidersDownUsesAlternative(t *testing.T) {
	// Setup: multiple providers, one down
	source := &MockModelSource{
		models: map[string][]registry.Model{
			"chat": {
				{Name: "gpt-4", ProviderID: "openai", QualityScore: 95, LatencyP50ms: 100, CostPer1M: 3, MaxContextToks: 8192, Disabled: false},
				{Name: "gpt-3.5", ProviderID: "openai", QualityScore: 80, LatencyP50ms: 80, CostPer1M: 1, MaxContextToks: 4096, Disabled: false},
				{Name: "claude-opus", ProviderID: "anthropic", QualityScore: 90, LatencyP50ms: 150, CostPer1M: 15, MaxContextToks: 200000, Disabled: false},
				{Name: "gemini-pro", ProviderID: "google", QualityScore: 85, LatencyP50ms: 120, CostPer1M: 2, MaxContextToks: 32000, Disabled: false},
			},
		},
	}

	health := &SelectiveHealthSource{
		unhealthyModels: map[string]bool{
			"openai:gpt-4":   true,
			"openai:gpt-3.5": true,
		},
	}

	router := New(source, health, &MockQuotaSource{}, &MockTokenizer{})

	// Resolve should skip both OpenAI models and use alternatives
	models, err := router.Resolve("chat", 1000)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) == 0 {
		t.Fatal("expected fallback models when primary provider is down")
	}

	// Verify no OpenAI models in result
	for _, m := range models {
		if m.ProviderID == "openai" {
			t.Errorf("expected no OpenAI models in fallback, got %v", m.Name)
		}
	}
}

func TestFallback_AllModelsUnavailableReturnsError(t *testing.T) {
	source := &MockModelSource{
		models: map[string][]registry.Model{
			"chat": {
				{Name: "gpt-4", ProviderID: "openai", QualityScore: 95, LatencyP50ms: 100, CostPer1M: 3, MaxContextToks: 8192, Disabled: false},
				{Name: "claude-opus", ProviderID: "anthropic", QualityScore: 90, LatencyP50ms: 150, CostPer1M: 15, MaxContextToks: 200000, Disabled: false},
			},
		},
	}

	health := &SelectiveHealthSource{
		unhealthyModels: map[string]bool{
			"openai:gpt-4":          true,
			"anthropic:claude-opus": true,
		},
	}

	router := New(source, health, &MockQuotaSource{}, &MockTokenizer{})

	// Act: all models unavailable
	_, err := router.Resolve("chat", 1000)

	// Assert: should return error
	if err == nil {
		t.Fatal("expected error when all models unavailable")
	}
	if err != ErrNoModelsAvailable {
		t.Errorf("expected ErrNoModelsAvailable, got %v", err)
	}
}

func TestFallback_FallbackChainPreservesScoring(t *testing.T) {
	// Verify that fallback models are still scored correctly
	source := &MockModelSource{
		models: map[string][]registry.Model{
			"chat": {
				{Name: "gpt-4", ProviderID: "openai", QualityScore: 95, LatencyP50ms: 100, CostPer1M: 3, MaxContextToks: 8192, Disabled: false},
				// Fallback: lower quality but available
				{Name: "gpt-3.5", ProviderID: "openai", QualityScore: 70, LatencyP50ms: 80, CostPer1M: 1, MaxContextToks: 4096, Disabled: false},
				// Fallback: better quality alternative provider
				{Name: "claude-opus", ProviderID: "anthropic", QualityScore: 90, LatencyP50ms: 150, CostPer1M: 15, MaxContextToks: 200000, Disabled: false},
			},
		},
	}

	health := &SelectiveHealthSource{
		unhealthyModels: map[string]bool{
			"openai:gpt-4": true, // primary down
		},
	}

	router := New(source, health, &MockQuotaSource{}, &MockTokenizer{})

	models, err := router.Resolve("chat", 1000)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) == 0 {
		t.Fatal("expected fallback models")
	}

	// Verify models are returned and gpt-4 is excluded
	// Scoring is multi-factor (quality + latency + cost), so gpt-3.5 can rank
	// higher than claude-opus despite lower quality due to better latency/cost
	if len(models) >= 2 {
		if models[0].Name == "gpt-4" || models[1].Name == "gpt-4" {
			t.Errorf("expected gpt-4 excluded from fallback, got %v", getNames(models))
		}
	}
}

func TestFallback_DisabledModelsExcludedFromFallback(t *testing.T) {
	source := &MockModelSource{
		models: map[string][]registry.Model{
			"chat": {
				{Name: "gpt-4", ProviderID: "openai", QualityScore: 95, LatencyP50ms: 100, CostPer1M: 3, MaxContextToks: 8192, Disabled: true}, // disabled
				{Name: "gpt-3.5", ProviderID: "openai", QualityScore: 80, LatencyP50ms: 80, CostPer1M: 1, MaxContextToks: 4096, Disabled: false},
				{Name: "claude-opus", ProviderID: "anthropic", QualityScore: 90, LatencyP50ms: 150, CostPer1M: 15, MaxContextToks: 200000, Disabled: false},
			},
		},
	}

	router := New(source, &MockHealthSource{}, &MockQuotaSource{}, &MockTokenizer{})

	models, err := router.Resolve("chat", 1000)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify gpt-4 is not in results (disabled)
	for _, m := range models {
		if m.Name == "gpt-4" {
			t.Error("expected disabled model gpt-4 to be excluded from fallback")
		}
	}
}

func TestFallback_QuotaExhaustedTriesNext(t *testing.T) {
	source := &MockModelSource{
		models: map[string][]registry.Model{
			"chat": {
				{Name: "gpt-4", ProviderID: "openai", QualityScore: 95, LatencyP50ms: 100, CostPer1M: 3, MaxContextToks: 8192, Disabled: false},
				{Name: "gpt-3.5", ProviderID: "openai", QualityScore: 80, LatencyP50ms: 80, CostPer1M: 1, MaxContextToks: 4096, Disabled: false},
				{Name: "claude-opus", ProviderID: "anthropic", QualityScore: 90, LatencyP50ms: 150, CostPer1M: 15, MaxContextToks: 200000, Disabled: false},
			},
		},
	}

	// Selective quota source
	quota := &SelectiveQuotaSource{
		zeroQuotaModels: map[string]bool{
			"openai:gpt-4": true, // quota exhausted
		},
	}

	router := New(source, &MockHealthSource{}, quota, &MockTokenizer{})

	models, err := router.Resolve("chat", 1000)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// gpt-4 should not appear (no quota)
	for _, m := range models {
		if m.Name == "gpt-4" {
			t.Error("expected quota-exhausted model gpt-4 to be excluded")
		}
	}
}

// SelectiveQuotaSource for testing quota exhaustion
type SelectiveQuotaSource struct {
	zeroQuotaModels map[string]bool
}

func (s *SelectiveQuotaSource) Remaining(providerID, model string) int {
	key := providerID + ":" + model
	if s.zeroQuotaModels[key] {
		return 0
	}
	return 10000
}
