package adapter

import (
	"testing"
)

func TestOpenAIParameterValidator_ValidTemperature(t *testing.T) {
	validator := NewOpenAIParameterValidator()

	tests := []struct {
		temp  float64
		valid bool
	}{
		{0.0, true},   // minimum
		{0.5, true},   // mid-range
		{1.0, true},   // middle
		{1.5, true},   // high
		{2.0, true},   // maximum
		{-0.1, false}, // below minimum
		{2.1, false},  // above maximum
		{3.0, false},  // way above
	}

	for _, tt := range tests {
		result := validator.ValidateTemperature(tt.temp)
		if result != tt.valid {
			t.Errorf("ValidateTemperature(%f) = %v, want %v", tt.temp, result, tt.valid)
		}
	}
}

func TestOpenAIParameterValidator_ValidTopP(t *testing.T) {
	validator := NewOpenAIParameterValidator()

	tests := []struct {
		topP  float64
		valid bool
	}{
		{0.0, true},   // minimum
		{0.5, true},   // mid-range
		{1.0, true},   // maximum
		{-0.1, false}, // below minimum
		{1.1, false},  // above maximum
		{0.95, true},  // typical value
	}

	for _, tt := range tests {
		result := validator.ValidateTopP(tt.topP)
		if result != tt.valid {
			t.Errorf("ValidateTopP(%f) = %v, want %v", tt.topP, result, tt.valid)
		}
	}
}

func TestOpenAIParameterValidator_ValidSeed(t *testing.T) {
	validator := NewOpenAIParameterValidator()

	tests := []struct {
		seed  int
		valid bool
	}{
		{0, true},
		{1, true},
		{12345, true},
		{2147483647, true}, // max int32
		{-1, false},        // negative
		{-100, false},      // negative
	}

	for _, tt := range tests {
		result := validator.ValidateSeed(tt.seed)
		if result != tt.valid {
			t.Errorf("ValidateSeed(%d) = %v, want %v", tt.seed, result, tt.valid)
		}
	}
}

func TestOpenAIParameterValidator_ValidToolChoice(t *testing.T) {
	validator := NewOpenAIParameterValidator()

	tests := []struct {
		choice string
		valid  bool
	}{
		{"none", true},
		{"auto", true},
		{"required", true},
		{"invalid", false},
		{"NONE", false}, // case-sensitive
		{"", false},     // empty
	}

	for _, tt := range tests {
		result := validator.ValidateToolChoice(tt.choice)
		if result != tt.valid {
			t.Errorf("ValidateToolChoice(%q) = %v, want %v", tt.choice, result, tt.valid)
		}
	}
}

func TestOpenAIParameterValidator_ValidResponseFormat(t *testing.T) {
	validator := NewOpenAIParameterValidator()

	tests := []struct {
		format interface{}
		valid  bool
	}{
		{"text", true},
		{"json_object", true},
		{"invalid_format", false},
		{123, false}, // wrong type
		{nil, true},  // nil is valid (means use default)
	}

	for _, tt := range tests {
		result := validator.ValidateResponseFormat(tt.format)
		if result != tt.valid {
			t.Errorf("ValidateResponseFormat(%v) = %v, want %v", tt.format, result, tt.valid)
		}
	}
}

func TestOpenAIParameterValidator_ClampTemperature(t *testing.T) {
	validator := NewOpenAIParameterValidator()

	tests := []struct {
		input    float64
		expected float64
	}{
		{0.5, 0.5},  // in range
		{-1.0, 0.0}, // clamp to min
		{3.0, 2.0},  // clamp to max
		{0.0, 0.0},  // boundary
		{2.0, 2.0},  // boundary
	}

	for _, tt := range tests {
		result := validator.ClampTemperature(tt.input)
		if result != tt.expected {
			t.Errorf("ClampTemperature(%f) = %f, want %f", tt.input, result, tt.expected)
		}
	}
}

func TestOpenAIParameterValidator_ClampTopP(t *testing.T) {
	validator := NewOpenAIParameterValidator()

	tests := []struct {
		input    float64
		expected float64
	}{
		{0.5, 0.5},  // in range
		{-0.5, 0.0}, // clamp to min
		{1.5, 1.0},  // clamp to max
	}

	for _, tt := range tests {
		result := validator.ClampTopP(tt.input)
		if result != tt.expected {
			t.Errorf("ClampTopP(%f) = %f, want %f", tt.input, result, tt.expected)
		}
	}
}

func TestOpenAIParameterValidator_ValidateMapParameters(t *testing.T) {
	validator := NewOpenAIParameterValidator()

	// Valid parameters
	validParams := map[string]interface{}{
		"temperature": 0.7,
		"top_p":       0.9,
		"seed":        42,
		"tool_choice": "auto",
	}

	errors := validator.ValidateMapParameters(validParams)
	if len(errors) > 0 {
		t.Errorf("ValidateMapParameters with valid params returned errors: %v", errors)
	}

	// Invalid parameters
	invalidParams := map[string]interface{}{
		"temperature": 5.0, // out of range
		"top_p":       1.5, // out of range
		"tool_choice": "invalid",
	}

	errors = validator.ValidateMapParameters(invalidParams)
	if len(errors) == 0 {
		t.Error("ValidateMapParameters with invalid params should return errors")
	}
}

func TestOpenAIParameterValidator_UnknownParametersLogged(t *testing.T) {
	validator := NewOpenAIParameterValidator()

	params := map[string]interface{}{
		"temperature":     0.7,
		"unknown_param":   "value",
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
