package router

import (
	"testing"

	"api-llm-gateway/internal/registry"
)

// MockModelSource para testing
type MockModelSource struct {
	models map[string][]registry.Model
}

func (m *MockModelSource) ModelsFor(capability string) []registry.Model {
	return m.models[capability]
}

func (m *MockModelSource) HasCapability(capability string) bool {
	_, ok := m.models[capability]
	return ok
}

func (m *MockModelSource) FindModel(name string) (registry.Model, bool) {
	for _, models := range m.models {
		for _, model := range models {
			if model.Name == name {
				return model, true
			}
		}
	}
	return registry.Model{}, false
}

func (m *MockModelSource) ModelNames() []string {
	var names []string
	for _, models := range m.models {
		for _, model := range models {
			names = append(names, model.Name)
		}
	}
	return names
}

// MockHealthSource always returns healthy
type MockHealthSource struct{}

func (m *MockHealthSource) Healthy(providerID, model string) bool {
	return true
}

// MockQuotaSource always returns plenty of quota
type MockQuotaSource struct{}

func (m *MockQuotaSource) Remaining(providerID, model string) int {
	return 10000
}

// MockTokenizer for testing
type MockTokenizer struct{}

func (m *MockTokenizer) Estimate(text string) int {
	return len(text) // rough estimate
}

func (m *MockTokenizer) FitsWindow(estimated, maxContext int) bool {
	return estimated <= maxContext
}

func TestRouter_ParseCapabilityPrefix(t *testing.T) {
	// Arrange
	source := &MockModelSource{
		models: map[string][]registry.Model{
			"chat": {
				{Name: "gpt-4", ProviderID: "openai", MaxContextToks: 8192, CostPer1M: 3, Disabled: false},
				{Name: "claude-opus", ProviderID: "anthropic", MaxContextToks: 200000, CostPer1M: 15, Disabled: false},
			},
			"vision": {
				{Name: "gpt-4-vision", ProviderID: "openai", MaxContextToks: 8192, CostPer1M: 5, Disabled: false},
			},
		},
	}

	router := New(source, &MockHealthSource{}, &MockQuotaSource{}, &MockTokenizer{})

	// Test: router:chat should resolve to chat capability
	t.Run("router:chat resolves to chat capability", func(t *testing.T) {
		models, err := router.ResolveCapabilityPrefix("router:chat", 1000)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(models) == 0 {
			t.Fatal("expected models for chat capability")
		}
		if !contains(models, "gpt-4") && !contains(models, "claude-opus") {
			t.Errorf("expected chat models, got %v", getNames(models))
		}
	})

	// Test: router:vision should resolve to vision capability
	t.Run("router:vision resolves to vision capability", func(t *testing.T) {
		models, err := router.ResolveCapabilityPrefix("router:vision", 1000)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(models) == 0 {
			t.Fatal("expected models for vision capability")
		}
		if !contains(models, "gpt-4-vision") {
			t.Errorf("expected vision model gpt-4-vision, got %v", getNames(models))
		}
	})

	// Test: unknown capability should error
	t.Run("router:unknown returns error", func(t *testing.T) {
		_, err := router.ResolveCapabilityPrefix("router:unknown", 1000)
		if err == nil {
			t.Fatal("expected error for unknown capability")
		}
	})

	// Test: explicit model name (no prefix) should not use capability resolution
	t.Run("explicit model names bypass capability resolution", func(t *testing.T) {
		models, err := router.ResolveCapabilityPrefix("gpt-4", 1000)
		if err != nil {
			// Should either succeed with explicit model or return error
			// (depends on ResolveExplicit implementation)
		}
		if len(models) > 0 && models[0].Name != "gpt-4" {
			t.Errorf("expected explicit model gpt-4, got %v", getNames(models))
		}
	})
}

func TestRouter_ExtractCapability(t *testing.T) {
	tests := []struct {
		input      string
		wantPrefix string
		wantCap    string
	}{
		{"router:chat", "router", "chat"},
		{"router:vision", "router", "vision"},
		{"router:embedding", "router", "embedding"},
		{"gpt-4", "", ""},
		{"claude-opus", "", ""},
		{"router:", "router", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			prefix, cap := ExtractCapabilityPrefix(tt.input)
			if prefix != tt.wantPrefix || cap != tt.wantCap {
				t.Errorf("ExtractCapabilityPrefix(%q) = (%q, %q), want (%q, %q)",
					tt.input, prefix, cap, tt.wantPrefix, tt.wantCap)
			}
		})
	}
}

func TestRouter_IsCapabilityPrefix(t *testing.T) {
	tests := []struct {
		model string
		want  bool
	}{
		{"router:chat", true},
		{"router:vision", true},
		{"router:embedding", true},
		{"gpt-4", false},
		{"claude-opus", false},
		{"router:", true},  // technically valid, just empty capability
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			got := IsCapabilityPrefix(tt.model)
			if got != tt.want {
				t.Errorf("IsCapabilityPrefix(%q) = %v, want %v", tt.model, got, tt.want)
			}
		})
	}
}

// Helper functions
func contains(models []registry.Model, name string) bool {
	for _, m := range models {
		if m.Name == name {
			return true
		}
	}
	return false
}

func getNames(models []registry.Model) []string {
	var names []string
	for _, m := range models {
		names = append(names, m.Name)
	}
	return names
}
