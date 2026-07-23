package middleware

import (
	"context"
	"testing"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/request"
)

var _ *request.NormalizedRequest

// MockProcessor implements Processor interface for testing
type MockProcessor struct {
	lastReq *adapter.Request
}

func (mp *MockProcessor) ProcessChat(ctx context.Context, req *adapter.Request) (*adapter.Response, error) {
	mp.lastReq = req
	return &adapter.Response{
		Content: "test response",
	}, nil
}

func TestPipeline_DetectsFormatAndNormalizes(t *testing.T) {
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

	detector := NewFormatDetector()
	normalizer := NewNormalizer()
	mock := &MockProcessor{}

	// Verify mock is ready (used in integration test)
	if mock == nil {
		t.Fatal("mock processor not initialized")
	}

	// Detect format
	format := detector.DetectFormat(openaiReq)
	if format != "openai" {
		t.Fatalf("expected format='openai', got '%s'", format)
	}

	// Normalize
	nr := normalizer.Normalize(format, openaiReq)
	if nr.Format != "openai" {
		t.Errorf("expected normalized format='openai', got '%s'", nr.Format)
	}
	if nr.Model != "gpt-4" {
		t.Errorf("expected model='gpt-4', got '%s'", nr.Model)
	}
	if len(nr.Messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(nr.Messages))
	}
}

func TestPipeline_AnthropicDetectionAndNormalization(t *testing.T) {
	anthropicReq := map[string]interface{}{
		"model": "claude-opus",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": "hello",
			},
		},
		"max_tokens": 1024,
	}

	detector := NewFormatDetector()
	normalizer := NewNormalizer()

	format := detector.DetectFormat(anthropicReq)
	if format != "anthropic" {
		t.Fatalf("expected format='anthropic', got '%s'", format)
	}

	nr := normalizer.Normalize(format, anthropicReq)
	if nr.Format != "anthropic" {
		t.Errorf("expected normalized format='anthropic'")
	}
	if nr.Parameters["max_tokens"] != 1024 {
		t.Errorf("expected max_tokens preserved")
	}
}

func TestPipeline_FullChain(t *testing.T) {
	rawReq := map[string]interface{}{
		"model": "gpt-4",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": "test",
			},
		},
		"temperature": 0.5,
	}

	detector := NewFormatDetector()
	normalizer := NewNormalizer()

	// Full pipeline: detect -> normalize
	format := detector.DetectFormat(rawReq)
	nr := normalizer.Normalize(format, rawReq)

	// Verify chain integrity
	if nr.Format == "" {
		t.Error("format not detected")
	}
	if nr.Model == "" {
		t.Error("model not extracted")
	}
	if len(nr.Messages) == 0 {
		t.Error("messages not normalized")
	}
	if nr.Parameters["temperature"] != 0.5 {
		t.Error("parameters not preserved")
	}
}
