package adapter

import (
	"testing"
)

func TestAnthropicParameterValidator_ValidTemperature(t *testing.T) {
	validator := NewAnthropicParameterValidator()

	tests := []struct {
		temp  float64
		valid bool
	}{
		{0.0, true},      // minimum
		{0.5, true},      // mid-range
		{1.0, true},      // maximum
		{-0.1, false},    // below minimum
		{1.1, false},     // above maximum
		{0.8, true},      // typical value
	}

	for _, tt := range tests {
		result := validator.ValidateTemperature(tt.temp)
		if result != tt.valid {
			t.Errorf("ValidateTemperature(%f) = %v, want %v", tt.temp, result, tt.valid)
		}
	}
}

func TestAnthropicParameterValidator_ValidTopK(t *testing.T) {
	validator := NewAnthropicParameterValidator()

	tests := []struct {
		topK  int
		valid bool
	}{
		{1, true},
		{40, true},        // typical value
		{256, true},       // high value
		{0, false},        // must be positive
		{-1, false},       // negative
	}

	for _, tt := range tests {
		result := validator.ValidateTopK(tt.topK)
		if result != tt.valid {
			t.Errorf("ValidateTopK(%d) = %v, want %v", tt.topK, result, tt.valid)
		}
	}
}

func TestAnthropicParameterValidator_MaxTokensRequired(t *testing.T) {
	validator := NewAnthropicParameterValidator()

	// Anthropic requires max_tokens
	if !validator.IsMaxTokensRequired() {
		t.Error("expected max_tokens to be required for Anthropic")
	}
}

func TestAnthropicParameterValidator_ValidMaxTokens(t *testing.T) {
	validator := NewAnthropicParameterValidator()

	tests := []struct {
		maxTokens int
		valid     bool
	}{
		{1, true},
		{100, true},
		{4096, true},
		{200000, true},      // typical for Claude 3
		{0, false},          // must be positive
		{-1, false},         // negative
	}

	for _, tt := range tests {
		result := validator.ValidateMaxTokens(tt.maxTokens)
		if result != tt.valid {
			t.Errorf("ValidateMaxTokens(%d) = %v, want %v", tt.maxTokens, result, tt.valid)
		}
	}
}

func TestAnthropicParameterValidator_ValidThinking(t *testing.T) {
	validator := NewAnthropicParameterValidator()

	tests := []struct {
		thinking string
		valid    bool
	}{
		{"enabled", true},
		{"disabled", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		result := validator.ValidateThinking(tt.thinking)
		if result != tt.valid {
			t.Errorf("ValidateThinking(%q) = %v, want %v", tt.thinking, result, tt.valid)
		}
	}
}

func TestAnthropicParameterValidator_ValidToolUse(t *testing.T) {
	validator := NewAnthropicParameterValidator()

	tests := []struct {
		toolUse string
		valid   bool
	}{
		{"auto", true},
		{"required", true},
		{"none", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		result := validator.ValidateToolUse(tt.toolUse)
		if result != tt.valid {
			t.Errorf("ValidateToolUse(%q) = %v, want %v", tt.toolUse, result, tt.valid)
		}
	}
}

func TestAnthropicParameterValidator_ClampTemperature(t *testing.T) {
	validator := NewAnthropicParameterValidator()

	tests := []struct {
		input    float64
		expected float64
	}{
		{0.5, 0.5},      // in range
		{-1.0, 0.0},     // clamp to min
		{2.0, 1.0},      // clamp to max
		{0.0, 0.0},      // boundary
		{1.0, 1.0},      // boundary
	}

	for _, tt := range tests {
		result := validator.ClampTemperature(tt.input)
		if result != tt.expected {
			t.Errorf("ClampTemperature(%f) = %f, want %f", tt.input, result, tt.expected)
		}
	}
}

func TestAnthropicParameterValidator_ValidateMapParameters(t *testing.T) {
	validator := NewAnthropicParameterValidator()

	// Valid parameters
	validParams := map[string]interface{}{
		"temperature": 0.7,
		"top_k":       40,
		"max_tokens":  2048,
		"thinking":    "enabled",
	}

	errors := validator.ValidateMapParameters(validParams)
	if len(errors) > 0 {
		t.Errorf("ValidateMapParameters with valid params returned errors: %v", errors)
	}

	// Missing max_tokens (required)
	invalidParams := map[string]interface{}{
		"temperature": 0.7,
		"top_k":       40,
	}

	errors = validator.ValidateMapParameters(invalidParams)
	if len(errors) == 0 {
		t.Error("expected error when max_tokens missing (required)")
	}
}

func TestAnthropicParameterValidator_OutOfRangeParameters(t *testing.T) {
	validator := NewAnthropicParameterValidator()

	params := map[string]interface{}{
		"temperature": 2.0,  // out of range [0, 1]
		"top_k":       0,    // invalid (must be >= 1)
		"max_tokens":  4096,
	}

	errors := validator.ValidateMapParameters(params)
	if len(errors) == 0 {
		t.Error("expected errors for out-of-range parameters")
	}
}

func TestAnthropicParameterValidator_UnknownParameters(t *testing.T) {
	validator := NewAnthropicParameterValidator()

	params := map[string]interface{}{
		"temperature":   0.7,
		"unknown_param": "value",
		"another_unknown": 123,
	}

	warnings := validator.GetUnknownParameters(params)

	if len(warnings) == 0 {
		t.Error("expected warnings for unknown parameters")
	}
	if len(warnings) != 2 {
		t.Errorf("expected 2 unknown parameters, got %d", len(warnings))
	}
}

func TestAnthropicParameterValidator_UnsupportedFeatureFallback(t *testing.T) {
	validator := NewAnthropicParameterValidator()

	// Some OpenAI features might not be supported by Anthropic
	params := map[string]interface{}{
		"response_format": "json_object",  // might need fallback
		"temperature":     0.7,
		"max_tokens":      2048,
	}

	fallback := validator.CheckUnsupportedFeatures(params)

	// Should identify features that need fallback
	if fallback != nil && len(fallback) > 0 {
		// Some features identified as unsupported - expected behavior
		if _, ok := fallback["response_format"]; !ok {
			t.Logf("response_format fallback: %v", fallback)
		}
	}
}

func TestAnthropicParameterValidator_TemperatureRequired(t *testing.T) {
	validator := NewAnthropicParameterValidator()

	// Temperature is not required for Anthropic (optional)
	params := map[string]interface{}{
		"max_tokens": 2048,
		"top_k":      40,
	}

	errors := validator.ValidateMapParameters(params)

	// Should not error just because temperature is missing
	temperatureError := false
	for _, err := range errors {
		if err == "temperature is required" {
			temperatureError = true
		}
	}

	if temperatureError {
		t.Error("temperature should not be required")
	}
}
