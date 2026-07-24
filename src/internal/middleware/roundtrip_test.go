package middleware

import (
	"testing"
)

func TestRoundTrip_OpenAIMessagePreservation(t *testing.T) {
	// Raw OpenAI request
	raw := map[string]interface{}{
		"model": "gpt-4",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "system",
				"content": "You are helpful.",
			},
			map[string]interface{}{
				"role":    "user",
				"content": "What is 2+2?",
			},
		},
		"temperature": 0.7,
		"top_p":       0.9,
		"max_tokens":  256,
	}

	detector := NewFormatDetector()
	normalizer := NewNormalizer()

	// Detect
	format := detector.DetectFormat(raw)
	if format != "openai" {
		t.Fatalf("detection failed: expected 'openai', got '%s'", format)
	}

	// Normalize
	nr := normalizer.Normalize(format, raw)

	// Verify format
	if nr.Format != "openai" {
		t.Errorf("format not preserved: expected 'openai', got '%s'", nr.Format)
	}

	// Verify model
	if nr.Model != "gpt-4" {
		t.Errorf("model not preserved: expected 'gpt-4', got '%s'", nr.Model)
	}

	// Verify message count
	if len(nr.Messages) != 2 {
		t.Errorf("message count mismatch: expected 2, got %d", len(nr.Messages))
	}

	// Verify first message content
	if nr.Messages[0]["role"] != "system" || nr.Messages[0]["content"] != "You are helpful." {
		t.Errorf("first message corrupted: got %v", nr.Messages[0])
	}

	// Verify second message content
	if nr.Messages[1]["role"] != "user" || nr.Messages[1]["content"] != "What is 2+2?" {
		t.Errorf("second message corrupted: got %v", nr.Messages[1])
	}

	// Verify all parameters preserved
	if nr.Parameters["temperature"] != 0.7 {
		t.Error("temperature not preserved")
	}
	if nr.Parameters["top_p"] != 0.9 {
		t.Error("top_p not preserved")
	}
	if nr.Parameters["max_tokens"] != 256 {
		t.Error("max_tokens not preserved")
	}
}

func TestRoundTrip_AnthropicMessagePreservation(t *testing.T) {
	// Raw Anthropic request
	raw := map[string]interface{}{
		"model": "claude-opus",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": "Hello Claude",
			},
			map[string]interface{}{
				"role":    "assistant",
				"content": "Hello! How can I help?",
			},
		},
		"temperature": 0.5,
		"max_tokens":  1024,
		"top_k":       40,
	}

	detector := NewFormatDetector()
	normalizer := NewNormalizer()

	// Detect
	format := detector.DetectFormat(raw)
	if format != "anthropic" {
		t.Fatalf("detection failed: expected 'anthropic', got '%s'", format)
	}

	// Normalize
	nr := normalizer.Normalize(format, raw)

	// Verify format
	if nr.Format != "anthropic" {
		t.Errorf("format not preserved: got '%s'", nr.Format)
	}

	// Verify model
	if nr.Model != "claude-opus" {
		t.Errorf("model not preserved: got '%s'", nr.Model)
	}

	// Verify messages
	if len(nr.Messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(nr.Messages))
	}

	// Verify parameters
	if nr.Parameters["temperature"] != 0.5 {
		t.Error("temperature not preserved")
	}
	if nr.Parameters["max_tokens"] != 1024 {
		t.Error("max_tokens not preserved")
	}
	if nr.Parameters["top_k"] != 40 {
		t.Error("top_k not preserved")
	}
}

func TestRoundTrip_ResponsesAPIPreservation(t *testing.T) {
	// Raw Responses API request
	raw := map[string]interface{}{
		"model": "gpt-4-turbo",
		"input": []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": "What is AI?",
			},
		},
		"reasoning_effort": "high",
		"max_tokens":       2048,
	}

	detector := NewFormatDetector()
	normalizer := NewNormalizer()

	// Detect
	format := detector.DetectFormat(raw)
	if format != "responses" {
		t.Fatalf("detection failed: expected 'responses', got '%s'", format)
	}

	// Normalize
	nr := normalizer.Normalize(format, raw)

	// Verify format
	if nr.Format != "responses" {
		t.Errorf("format not preserved: got '%s'", nr.Format)
	}

	// Verify model
	if nr.Model != "gpt-4-turbo" {
		t.Errorf("model not preserved: got '%s'", nr.Model)
	}

	// Verify input was treated as messages
	if len(nr.Messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(nr.Messages))
	}

	// Verify parameters
	if nr.Parameters["reasoning_effort"] != "high" {
		t.Error("reasoning_effort not preserved")
	}
	if nr.Parameters["max_tokens"] != 2048 {
		t.Error("max_tokens not preserved")
	}
}

func TestRoundTrip_UnknownFieldsPreserved(t *testing.T) {
	// Request with custom/unknown fields
	raw := map[string]interface{}{
		"model": "gpt-4",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": "test",
			},
		},
		"custom_field_1": "value1",
		"custom_field_2": 42,
		"nested":         map[string]interface{}{"key": "value"},
	}

	detector := NewFormatDetector()
	normalizer := NewNormalizer()

	format := detector.DetectFormat(raw)
	nr := normalizer.Normalize(format, raw)

	// Verify custom fields are in parameters
	if nr.Parameters["custom_field_1"] != "value1" {
		t.Error("custom_field_1 not preserved")
	}
	if nr.Parameters["custom_field_2"] != 42 {
		t.Error("custom_field_2 not preserved")
	}
	if nested, ok := nr.Parameters["nested"].(map[string]interface{}); !ok {
		t.Error("nested field not preserved as map")
	} else if nested["key"] != "value" {
		t.Error("nested field content lost")
	}
}

func TestRoundTrip_EmptyMessagesHandled(t *testing.T) {
	// Edge case: request with empty messages array
	raw := map[string]interface{}{
		"model":    "gpt-4",
		"messages": []interface{}{},
	}

	detector := NewFormatDetector()
	normalizer := NewNormalizer()

	format := detector.DetectFormat(raw)
	nr := normalizer.Normalize(format, raw)

	if len(nr.Messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(nr.Messages))
	}
	if nr.Model != "gpt-4" {
		t.Error("model lost on empty messages")
	}
}

func TestRoundTrip_NoMessagesFieldHandled(t *testing.T) {
	// Edge case: Responses API without messages field
	raw := map[string]interface{}{
		"model":            "gpt-4",
		"reasoning_effort": "low",
	}

	detector := NewFormatDetector()

	// Should return empty string for unknown format (no messages, no input)
	format := detector.DetectFormat(raw)
	if format != "" {
		t.Errorf("expected empty format for request without messages/input, got '%s'", format)
	}
}
