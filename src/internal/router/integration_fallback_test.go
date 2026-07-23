package router

import (
	"testing"

	"api-llm-gateway/internal/registry"
)

func TestIntegration_CapabilityRoutingWithFallback(t *testing.T) {
	// Scenario: Client uses "router:chat" with primary provider down
	source := &MockModelSource{
		models: map[string][]registry.Model{
			"chat": {
				{Name: "gpt-4", ProviderID: "openai", QualityScore: 95, LatencyP50ms: 100, CostPer1M: 3, MaxContextToks: 8192, Disabled: false},
				{Name: "claude-opus", ProviderID: "anthropic", QualityScore: 90, LatencyP50ms: 150, CostPer1M: 15, MaxContextToks: 200000, Disabled: false},
				{Name: "gemini-pro", ProviderID: "google", QualityScore: 85, LatencyP50ms: 120, CostPer1M: 2, MaxContextToks: 32000, Disabled: false},
			},
		},
	}

	health := &SelectiveHealthSource{
		unhealthyModels: map[string]bool{
			"openai:gpt-4": true, // OpenAI down
		},
	}

	router := New(source, health, &MockQuotaSource{}, &MockTokenizer{})

	// Client uses new "router:chat" prefix
	models, err := router.ResolveCapabilityPrefix("router:chat", 1000)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) == 0 {
		t.Fatal("expected fallback models")
	}

	// Should get alternatives (Claude or Gemini)
	gotOpenAI := false
	for _, m := range models {
		if m.ProviderID == "openai" {
			gotOpenAI = true
		}
	}
	if gotOpenAI {
		t.Error("expected no OpenAI models in fallback result")
	}
}

func TestIntegration_VisionRoutingWithFallback(t *testing.T) {
	// Scenario: Vision request routes automatically, with fallback
	source := &MockModelSource{
		models: map[string][]registry.Model{
			"vision": {
				{Name: "gpt-4-vision", ProviderID: "openai", QualityScore: 90, LatencyP50ms: 150, CostPer1M: 5, MaxContextToks: 8192, Disabled: false},
				{Name: "claude-vision", ProviderID: "anthropic", QualityScore: 85, LatencyP50ms: 200, CostPer1M: 10, MaxContextToks: 100000, Disabled: false},
				{Name: "gemini-vision", ProviderID: "google", QualityScore: 80, LatencyP50ms: 180, CostPer1M: 3, MaxContextToks: 50000, Disabled: false},
			},
		},
	}

	health := &SelectiveHealthSource{
		unhealthyModels: map[string]bool{
			"openai:gpt-4-vision": true,
		},
	}

	router := New(source, health, &MockQuotaSource{}, &MockTokenizer{})

	// Infer vision capability from request
	request := map[string]interface{}{
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{"type": "image_url"},
					map[string]interface{}{"type": "text", "text": "What's in this image?"},
				},
			},
		},
	}

	capability := InferCapability(request)
	if capability != "vision" {
		t.Fatalf("expected vision capability, got %s", capability)
	}

	// Resolve with fallback
	models, err := router.Resolve(capability, 1000)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) == 0 {
		t.Fatal("expected vision models")
	}

	// Primary vision provider should be excluded
	for _, m := range models {
		if m.Name == "gpt-4-vision" {
			t.Error("expected primary vision model to be excluded due to health")
		}
	}
}

func TestIntegration_EmbeddingRoutingWithFullFallback(t *testing.T) {
	// Scenario: Embedding request, multiple providers, first two down
	source := &MockModelSource{
		models: map[string][]registry.Model{
			"embedding": {
				{Name: "text-embedding-3-large", ProviderID: "openai", QualityScore: 95, LatencyP50ms: 50, CostPer1M: 130, MaxContextToks: 8191, Disabled: false},
				{Name: "text-embedding-3-small", ProviderID: "openai", QualityScore: 90, LatencyP50ms: 40, CostPer1M: 20, MaxContextToks: 8191, Disabled: false},
				{Name: "embedding-001", ProviderID: "google", QualityScore: 80, LatencyP50ms: 100, CostPer1M: 1, MaxContextToks: 2048, Disabled: false},
			},
		},
	}

	health := &SelectiveHealthSource{
		unhealthyModels: map[string]bool{
			"openai:text-embedding-3-large": true,
			"openai:text-embedding-3-small": true,
		},
	}

	router := New(source, health, &MockQuotaSource{}, &MockTokenizer{})

	// Resolve embedding capability
	models, err := router.Resolve("embedding", 1000)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) == 0 {
		t.Fatal("expected fallback embedding model")
	}

	// Should only get Google (OpenAI models down)
	if models[0].ProviderID != "google" {
		t.Errorf("expected Google provider as fallback, got %s", models[0].ProviderID)
	}
}

func TestIntegration_MixedOldAndNewClientRequests(t *testing.T) {
	// Scenario: Gateway handles both old (explicit model) and new (router:capability) clients
	source := &MockModelSource{
		models: map[string][]registry.Model{
			"chat": {
				{Name: "gpt-4", ProviderID: "openai", QualityScore: 95, LatencyP50ms: 100, CostPer1M: 3, MaxContextToks: 8192, Disabled: false},
				{Name: "claude-opus", ProviderID: "anthropic", QualityScore: 90, LatencyP50ms: 150, CostPer1M: 15, MaxContextToks: 200000, Disabled: false},
			},
		},
	}

	router := New(source, &MockHealthSource{}, &MockQuotaSource{}, &MockTokenizer{})

	// Old client: explicit model "gpt-4"
	oldModels, err1 := router.ResolveCapabilityPrefix("gpt-4", 1000)
	if err1 != nil || len(oldModels) == 0 {
		t.Fatal("old client request failed")
	}
	if oldModels[0].Name != "gpt-4" {
		t.Errorf("expected gpt-4 for old client")
	}

	// New client: capability prefix "router:chat"
	newModels, err2 := router.ResolveCapabilityPrefix("router:chat", 1000)
	if err2 != nil || len(newModels) == 0 {
		t.Fatal("new client request failed")
	}

	// Both requests should work independently
	// Success: both old and new APIs work
	t.Logf("Old client got %s, new client got %s", oldModels[0].Name, newModels[0].Name)
}

func TestIntegration_CapabilityInferenceIntegration(t *testing.T) {
	// Full integration: request parsing -> capability inference -> routing -> fallback
	source := &MockModelSource{
		models: map[string][]registry.Model{
			"chat": {
				{Name: "gpt-4", ProviderID: "openai", QualityScore: 95, LatencyP50ms: 100, CostPer1M: 3, MaxContextToks: 8192, Disabled: false},
				{Name: "claude-opus", ProviderID: "anthropic", QualityScore: 90, LatencyP50ms: 150, CostPer1M: 15, MaxContextToks: 200000, Disabled: false},
			},
			"reasoning": {
				{Name: "gpt-4-reasoning", ProviderID: "openai", QualityScore: 98, LatencyP50ms: 500, CostPer1M: 10, MaxContextToks: 8192, Disabled: false},
			},
		},
	}

	health := &SelectiveHealthSource{
		unhealthyModels: map[string]bool{
			"openai:gpt-4-reasoning": true, // reasoning model down
		},
	}

	router := New(source, health, &MockQuotaSource{}, &MockTokenizer{})

	// Request with reasoning_effort (should infer reasoning capability)
	request := map[string]interface{}{
		"reasoning_effort": "high",
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "complex problem"},
		},
	}

	// Step 1: Infer capability
	capability := InferCapability(request)
	if capability != "reasoning" {
		t.Fatalf("expected reasoning capability, got %s", capability)
	}

	// Step 2: Resolve with fallback
	_, err := router.Resolve(capability, 1000)

	// Should fail because reasoning model is down and no fallback exists
	if err == nil {
		t.Fatal("expected error when reasoning model unavailable with no fallback")
	}
	if err != ErrNoModelsAvailable {
		t.Errorf("expected ErrNoModelsAvailable, got %v", err)
	}
}
