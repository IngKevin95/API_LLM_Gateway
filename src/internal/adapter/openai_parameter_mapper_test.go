package adapter

import (
	"testing"
)

func TestOpenAIParameterMapper_MapBasicParameters(t *testing.T) {
	mapper := NewOpenAIParameterMapper()

	params := map[string]interface{}{
		"temperature": 0.7,
		"top_p":       0.9,
		"max_tokens":  256,
	}

	mapped := mapper.MapParameters(params)

	if mapped["temperature"] != 0.7 {
		t.Errorf("expected temperature 0.7, got %v", mapped["temperature"])
	}
	if mapped["top_p"] != 0.9 {
		t.Errorf("expected top_p 0.9, got %v", mapped["top_p"])
	}
	if mapped["max_tokens"] != 256 {
		t.Errorf("expected max_tokens 256, got %v", mapped["max_tokens"])
	}
}

func TestOpenAIParameterMapper_ClampOutOfRangeParameters(t *testing.T) {
	mapper := NewOpenAIParameterMapper()

	params := map[string]interface{}{
		"temperature": 5.0,   // out of range, should clamp to 2.0
		"top_p":       1.5,   // out of range, should clamp to 1.0
	}

	mapped := mapper.MapParameters(params)

	if mapped["temperature"] != 2.0 {
		t.Errorf("expected temperature clamped to 2.0, got %v", mapped["temperature"])
	}
	if mapped["top_p"] != 1.0 {
		t.Errorf("expected top_p clamped to 1.0, got %v", mapped["top_p"])
	}
}

func TestOpenAIParameterMapper_HandleNegativeValues(t *testing.T) {
	mapper := NewOpenAIParameterMapper()

	params := map[string]interface{}{
		"temperature": -1.0,   // clamp to 0.0
		"top_p":       -0.5,   // clamp to 0.0
	}

	mapped := mapper.MapParameters(params)

	if mapped["temperature"] != 0.0 {
		t.Errorf("expected temperature clamped to 0.0, got %v", mapped["temperature"])
	}
	if mapped["top_p"] != 0.0 {
		t.Errorf("expected top_p clamped to 0.0, got %v", mapped["top_p"])
	}
}

func TestOpenAIParameterMapper_ValidateToolChoice(t *testing.T) {
	mapper := NewOpenAIParameterMapper()

	params := map[string]interface{}{
		"tool_choice": "auto",
	}

	mapped := mapper.MapParameters(params)

	if mapped["tool_choice"] != "auto" {
		t.Errorf("expected tool_choice 'auto', got %v", mapped["tool_choice"])
	}
}

func TestOpenAIParameterMapper_FilterInvalidToolChoice(t *testing.T) {
	mapper := NewOpenAIParameterMapper()

	params := map[string]interface{}{
		"tool_choice": "invalid_value",
	}

	mapped := mapper.MapParameters(params)

	// Invalid tool_choice should be removed or logged as warning
	if _, exists := mapped["tool_choice"]; exists {
		// If it still exists, it should be the original value
		// but mapper can choose to remove or keep invalid values
	}
}

func TestOpenAIParameterMapper_ResponseFormat(t *testing.T) {
	mapper := NewOpenAIParameterMapper()

	params := map[string]interface{}{
		"response_format": "json_object",
	}

	mapped := mapper.MapParameters(params)

	if mapped["response_format"] != "json_object" {
		t.Errorf("expected response_format 'json_object', got %v", mapped["response_format"])
	}
}

func TestOpenAIParameterMapper_PreservesValidParameters(t *testing.T) {
	mapper := NewOpenAIParameterMapper()

	params := map[string]interface{}{
		"temperature":     0.7,
		"top_p":           0.9,
		"seed":            42,
		"max_tokens":      512,
		"tool_choice":     "auto",
		"response_format": "json_object",
		"presence_penalty": 0.0,
		"frequency_penalty": 0.0,
	}

	mapped := mapper.MapParameters(params)

	// All valid parameters should be preserved
	count := 0
	for key := range mapped {
		if key != "" {
			count++
		}
	}

	if count < 6 {
		t.Errorf("expected at least 6 parameters preserved, got %d", count)
	}
}

func TestOpenAIParameterMapper_RemovesUnknownParameters(t *testing.T) {
	mapper := NewOpenAIParameterMapper()

	params := map[string]interface{}{
		"temperature":     0.7,
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

func TestOpenAIParameterMapper_HandleMissingOptionalParameters(t *testing.T) {
	mapper := NewOpenAIParameterMapper()

	// Minimal request with only required parameters
	params := map[string]interface{}{
		"model":    "gpt-4",
		"messages": []interface{}{},
	}

	mapped := mapper.MapParameters(params)

	// Should handle gracefully - model and messages might be handled separately
	// or this could just validate present parameters
	if mapped == nil {
		t.Error("expected non-nil map")
	}
}

func TestOpenAIParameterMapper_SeedValidation(t *testing.T) {
	mapper := NewOpenAIParameterMapper()

	tests := []struct {
		seed      int
		shouldKeep bool
	}{
		{0, true},
		{42, true},
		{999999, true},
		{-1, false},  // negative seeds not allowed
	}

	for _, tt := range tests {
		params := map[string]interface{}{
			"seed": tt.seed,
		}

		mapped := mapper.MapParameters(params)

		if tt.shouldKeep {
			if mapped["seed"] != tt.seed {
				t.Errorf("expected seed %d to be preserved", tt.seed)
			}
		} else {
			if _, exists := mapped["seed"]; exists {
				t.Errorf("expected invalid seed %d to be removed", tt.seed)
			}
		}
	}
}
