package adapter

import (
	"testing"
)

func TestAnthropicParameterMapper_MapBasicParameters(t *testing.T) {
	mapper := NewAnthropicParameterMapper()

	params := map[string]interface{}{
		"temperature": 0.7,
		"top_k":       40,
		"max_tokens":  2048,
	}

	mapped := mapper.MapParameters(params)

	if mapped["temperature"] != 0.7 {
		t.Errorf("expected temperature 0.7, got %v", mapped["temperature"])
	}
	if mapped["top_k"] != 40 {
		t.Errorf("expected top_k 40, got %v", mapped["top_k"])
	}
	if mapped["max_tokens"] != 2048 {
		t.Errorf("expected max_tokens 2048, got %v", mapped["max_tokens"])
	}
}

func TestAnthropicParameterMapper_ClampTemperatureDifferentRange(t *testing.T) {
	mapper := NewAnthropicParameterMapper()

	params := map[string]interface{}{
		"temperature": 2.0,   // out of range [0, 1], should clamp to 1.0
		"max_tokens":  2048,
	}

	mapped := mapper.MapParameters(params)

	// Anthropic temperature is [0, 1], not [0, 2] like OpenAI
	if mapped["temperature"] != 1.0 {
		t.Errorf("expected temperature clamped to 1.0, got %v", mapped["temperature"])
	}
}

func TestAnthropicParameterMapper_EnforceMaxTokensRequired(t *testing.T) {
	mapper := NewAnthropicParameterMapper()

	// Request without max_tokens should be flagged
	params := map[string]interface{}{
		"temperature": 0.7,
		"top_k":       40,
	}

	warnings := mapper.GetValidationWarnings(params)

	if len(warnings) == 0 {
		t.Error("expected warning about missing max_tokens (required for Anthropic)")
	}
}

func TestAnthropicParameterMapper_ThinkingSupport(t *testing.T) {
	mapper := NewAnthropicParameterMapper()

	params := map[string]interface{}{
		"thinking":   "enabled",
		"max_tokens": 4096,
	}

	mapped := mapper.MapParameters(params)

	if mapped["thinking"] != "enabled" {
		t.Errorf("expected thinking 'enabled', got %v", mapped["thinking"])
	}
}

func TestAnthropicParameterMapper_ToolUseMode(t *testing.T) {
	mapper := NewAnthropicParameterMapper()

	params := map[string]interface{}{
		"tool_use":   "auto",
		"max_tokens": 2048,
	}

	mapped := mapper.MapParameters(params)

	if mapped["tool_use"] != "auto" {
		t.Errorf("expected tool_use 'auto', got %v", mapped["tool_use"])
	}
}

func TestAnthropicParameterMapper_FilterUnsupportedFeatures(t *testing.T) {
	mapper := NewAnthropicParameterMapper()

	params := map[string]interface{}{
		"temperature":      0.7,
		"max_tokens":       2048,
		"response_format":  "json_object", // not supported, should be filtered
		"seed":             42,            // not supported, should be filtered
		"presence_penalty": 0.0,           // not supported, should be filtered
	}

	mapped := mapper.MapParameters(params)

	// Unsupported features should not be in result
	if _, exists := mapped["response_format"]; exists {
		t.Error("expected response_format to be filtered out")
	}
	if _, exists := mapped["seed"]; exists {
		t.Error("expected seed to be filtered out")
	}
	if _, exists := mapped["presence_penalty"]; exists {
		t.Error("expected presence_penalty to be filtered out")
	}

	// Supported features should be preserved
	if mapped["temperature"] != 0.7 {
		t.Error("expected temperature to be preserved")
	}
	if mapped["max_tokens"] != 2048 {
		t.Error("expected max_tokens to be preserved")
	}
}

func TestAnthropicParameterMapper_ValidateOutOfRangeTemperature(t *testing.T) {
	mapper := NewAnthropicParameterMapper()

	params := map[string]interface{}{
		"temperature": 1.5, // out of range
		"max_tokens":  2048,
	}

	warnings := mapper.GetValidationWarnings(params)

	if len(warnings) == 0 {
		t.Error("expected warning about temperature out of range [0, 1]")
	}
}

func TestAnthropicParameterMapper_TopKValidation(t *testing.T) {
	mapper := NewAnthropicParameterMapper()

	tests := []struct {
		topK      int
		shouldErr bool
	}{
		{1, false},
		{40, false},
		{256, false},
		{0, true},       // invalid
		{-1, true},      // invalid
	}

	for _, tt := range tests {
		params := map[string]interface{}{
			"top_k":      tt.topK,
			"max_tokens": 2048,
		}

		warnings := mapper.GetValidationWarnings(params)

		if tt.shouldErr && len(warnings) == 0 {
			t.Errorf("expected warning for top_k=%d", tt.topK)
		}
		if !tt.shouldErr && len(warnings) > 0 {
			t.Errorf("unexpected warning for top_k=%d: %v", tt.topK, warnings)
		}
	}
}

func TestAnthropicParameterMapper_PreserveValidParameters(t *testing.T) {
	mapper := NewAnthropicParameterMapper()

	params := map[string]interface{}{
		"temperature": 0.7,
		"top_k":       40,
		"top_p":       0.9,
		"max_tokens":  2048,
		"thinking":    "enabled",
		"tool_use":    "auto",
		"stop":        []string{"\n\n"},
	}

	mapped := mapper.MapParameters(params)

	// All valid Anthropic parameters should be preserved
	if mapped["temperature"] != 0.7 {
		t.Error("expected temperature preserved")
	}
	if mapped["top_k"] != 40 {
		t.Error("expected top_k preserved")
	}
	if mapped["top_p"] != 0.9 {
		t.Error("expected top_p preserved")
	}
	if mapped["max_tokens"] != 2048 {
		t.Error("expected max_tokens preserved")
	}
	if mapped["thinking"] != "enabled" {
		t.Error("expected thinking preserved")
	}
	if mapped["tool_use"] != "auto" {
		t.Error("expected tool_use preserved")
	}
}

func TestAnthropicParameterMapper_NegativeTemperatureClamp(t *testing.T) {
	mapper := NewAnthropicParameterMapper()

	params := map[string]interface{}{
		"temperature": -0.5,  // should clamp to 0.0
		"max_tokens":  2048,
	}

	mapped := mapper.MapParameters(params)

	if mapped["temperature"] != 0.0 {
		t.Errorf("expected temperature clamped to 0.0, got %v", mapped["temperature"])
	}
}

func TestAnthropicParameterMapper_RemovesUnknownParameters(t *testing.T) {
	mapper := NewAnthropicParameterMapper()

	params := map[string]interface{}{
		"temperature":     0.7,
		"max_tokens":      2048,
		"unknown_param1":  "value",
		"another_unknown": 123,
	}

	mapped := mapper.MapParameters(params)

	// Unknown parameters should be removed
	if _, exists := mapped["unknown_param1"]; exists {
		t.Error("expected unknown_param1 to be removed")
	}
	if _, exists := mapped["another_unknown"]; exists {
		t.Error("expected another_unknown to be removed")
	}

	// Known parameters should remain
	if _, exists := mapped["temperature"]; !exists {
		t.Error("expected temperature to be preserved")
	}
}
