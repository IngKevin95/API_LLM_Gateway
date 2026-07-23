package middleware

import (
	"testing"
)

// Test: FormatDetector interface exists and can detect OpenAI format
func TestFormatDetector_DetectsOpenAIFormat(t *testing.T) {
	detector := NewFormatDetector()

	request := map[string]interface{}{
		"messages": []interface{}{},
		"model":    "gpt-4",
	}

	format := detector.DetectFormat(request)

	if format != "openai" {
		t.Errorf("expected 'openai', got '%s'", format)
	}
}

// Test: FormatDetector detects Anthropic format
func TestFormatDetector_DetectsAnthropicFormat(t *testing.T) {
	detector := NewFormatDetector()

	request := map[string]interface{}{
		"model":    "claude-opus",
		"messages": []interface{}{},
	}

	format := detector.DetectFormat(request)

	if format != "anthropic" {
		t.Errorf("expected 'anthropic', got '%s'", format)
	}
}

// Test: FormatDetector detects Responses API format
func TestFormatDetector_DetectsResponsesAPIFormat(t *testing.T) {
	detector := NewFormatDetector()

	request := map[string]interface{}{
		"input":             []interface{}{},
		"reasoning_effort":  "medium",
	}

	format := detector.DetectFormat(request)

	if format != "responses" {
		t.Errorf("expected 'responses', got '%s'", format)
	}
}

// Test: Unknown format returns empty string
func TestFormatDetector_UnknownFormat(t *testing.T) {
	detector := NewFormatDetector()

	request := map[string]interface{}{
		"unknown_field": "value",
	}

	format := detector.DetectFormat(request)

	if format != "" {
		t.Errorf("expected empty string for unknown format, got '%s'", format)
	}
}
