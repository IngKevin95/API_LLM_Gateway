package middleware

import (
	"testing"

	"api-llm-gateway/internal/request"
)

var _ *request.NormalizedRequest

func TestNormalizer_ConvertOpenAIRequest(t *testing.T) {
	openaiReq := map[string]interface{}{
		"model": "gpt-4",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": "hello",
			},
		},
		"temperature": 0.7,
	}

	normalizer := NewNormalizer()
	nr := normalizer.Normalize("openai", openaiReq)

	if nr.Format != "openai" {
		t.Errorf("expected Format='openai', got '%s'", nr.Format)
	}
	if nr.Model != "gpt-4" {
		t.Errorf("expected Model='gpt-4', got '%s'", nr.Model)
	}
	if len(nr.Messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(nr.Messages))
	}
	if nr.Parameters["temperature"] != 0.7 {
		t.Errorf("expected temperature=0.7, got %v", nr.Parameters["temperature"])
	}
}

func TestNormalizer_ConvertAnthropicRequest(t *testing.T) {
	anthropicReq := map[string]interface{}{
		"model": "claude-opus",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": "hello",
			},
		},
		"temperature": 0.5,
		"max_tokens":  1024,
	}

	normalizer := NewNormalizer()
	nr := normalizer.Normalize("anthropic", anthropicReq)

	if nr.Format != "anthropic" {
		t.Errorf("expected Format='anthropic', got '%s'", nr.Format)
	}
	if nr.Model != "claude-opus" {
		t.Errorf("expected Model='claude-opus', got '%s'", nr.Model)
	}
	if nr.Parameters["max_tokens"] != 1024 {
		t.Errorf("expected max_tokens=1024, got %v", nr.Parameters["max_tokens"])
	}
}

func TestNormalizer_ConvertResponsesAPIRequest(t *testing.T) {
	responsesReq := map[string]interface{}{
		"model": "default",
		"input": []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": "hello",
			},
		},
		"reasoning_effort": "high",
	}

	normalizer := NewNormalizer()
	nr := normalizer.Normalize("responses", responsesReq)

	if nr.Format != "responses" {
		t.Errorf("expected Format='responses', got '%s'", nr.Format)
	}
	if nr.Parameters["reasoning_effort"] != "high" {
		t.Errorf("expected reasoning_effort='high', got %v", nr.Parameters["reasoning_effort"])
	}
}

func TestNormalizer_PreservesAllParameters(t *testing.T) {
	req := map[string]interface{}{
		"model": "gpt-4",
		"messages": []interface{}{},
		"temperature":  0.8,
		"top_p":        0.95,
		"seed":         42,
		"custom_field": "value",
	}

	normalizer := NewNormalizer()
	nr := normalizer.Normalize("openai", req)

	if nr.Parameters["temperature"] != 0.8 {
		t.Errorf("lost temperature parameter")
	}
	if nr.Parameters["top_p"] != 0.95 {
		t.Errorf("lost top_p parameter")
	}
	if nr.Parameters["seed"] != 42 {
		t.Errorf("lost seed parameter")
	}
	if nr.Parameters["custom_field"] != "value" {
		t.Errorf("lost custom_field parameter")
	}
}
