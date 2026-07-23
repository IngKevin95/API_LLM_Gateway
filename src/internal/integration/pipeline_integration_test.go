package integration

import (
	"testing"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/middleware"
	"api-llm-gateway/internal/request"
	"api-llm-gateway/internal/router"
)

func TestPipeline_DetectorToNormalizer(t *testing.T) {
	detector := middleware.NewFormatDetector()
	normalizer := middleware.NewNormalizer()

	// OpenAI format request
	openaiReq := map[string]interface{}{
		"model":    "gpt-4",
		"messages": []map[string]string{{"role": "user", "content": "hello"}},
	}

	format := detector.DetectFormat(openaiReq)
	if format != "openai" {
		t.Errorf("expected format 'openai', got %s", format)
	}

	normalized := normalizer.Normalize(format, openaiReq)
	if normalized == nil || normalized.Model != "gpt-4" {
		t.Error("normalization failed")
	}
}

func TestPipeline_NormalizerToRouter(t *testing.T) {
	// Simulate normalized request
	_ = &request.NormalizedRequest{
		Model:    "gpt-4",
		Format:   "openai",
		Messages: []map[string]interface{}{{"role": "user", "content": "hello"}},
	}

	// Router inference
	capability := router.InferCapability(map[string]interface{}{
		"messages": []map[string]string{{"role": "user", "content": "hello"}},
	})

	if capability != "chat" {
		t.Errorf("expected capability 'chat', got %s", capability)
	}
}

func TestPipeline_RouterToOpenAIMapper(t *testing.T) {
	mapper := adapter.NewOpenAIParameterMapper()

	params := map[string]interface{}{
		"temperature": 1.5,
		"top_p":       0.9,
		"max_tokens":  1024,
	}

	mapped := mapper.MapParameters(params)

	if mapped["temperature"] != 1.5 {
		t.Errorf("expected temperature 1.5, got %v", mapped["temperature"])
	}
	if mapped["max_tokens"] != 1024 {
		t.Errorf("expected max_tokens 1024, got %v", mapped["max_tokens"])
	}
}

func TestPipeline_RouterToAnthropicMapper(t *testing.T) {
	mapper := adapter.NewAnthropicParameterMapper()

	params := map[string]interface{}{
		"temperature": 1.5,  // Will be clamped to 1.0
		"top_k":       40,
		"max_tokens":  2048,
	}

	mapped := mapper.MapParameters(params)

	if mapped["temperature"] != 1.0 {
		t.Errorf("expected temperature clamped to 1.0, got %v", mapped["temperature"])
	}
	if mapped["max_tokens"] != 2048 {
		t.Errorf("expected max_tokens 2048, got %v", mapped["max_tokens"])
	}
}

func TestPipeline_EndToEnd_OpenAIPath(t *testing.T) {
	// Simulate full pipeline: detect → normalize → infer capability → map params
	detector := middleware.NewFormatDetector()
	normalizer := middleware.NewNormalizer()
	mapper := adapter.NewOpenAIParameterMapper()

	rawReq := map[string]interface{}{
		"model":       "gpt-4",
		"messages":    []map[string]string{{"role": "user", "content": "hello"}},
		"temperature": 0.7,
		"max_tokens":  512,
	}

	// Step 1: Detect
	format := detector.DetectFormat(rawReq)
	if format != "openai" {
		t.Fatalf("detection failed: got %s", format)
	}

	// Step 2: Normalize
	normalized := normalizer.Normalize(format, rawReq)
	if normalized == nil {
		t.Fatalf("normalization failed")
	}

	// Step 3: Infer capability
	capability := router.InferCapability(rawReq)
	if capability != "chat" {
		t.Fatalf("capability inference failed: got %s", capability)
	}

	// Step 4: Map parameters for OpenAI
	mapped := mapper.MapParameters(normalized.Parameters)
	if mapped["temperature"] != 0.7 {
		t.Errorf("parameter mapping failed")
	}

	t.Logf("✓ OpenAI path: detect → normalize → infer → map")
}

func TestPipeline_EndToEnd_AnthropicPath(t *testing.T) {
	// Simulate full pipeline for Anthropic format
	detector := middleware.NewFormatDetector()
	normalizer := middleware.NewNormalizer()
	mapper := adapter.NewAnthropicParameterMapper()

	rawReq := map[string]interface{}{
		"model":    "claude-3-opus",
		"messages": []map[string]string{{"role": "user", "content": "hello"}},
		"thinking": "enabled",
		"max_tokens": 4096,
	}

	// Step 1: Detect
	format := detector.DetectFormat(rawReq)
	if format != "anthropic" {
		t.Fatalf("detection failed: got %s", format)
	}

	// Step 2: Normalize
	normalized := normalizer.Normalize(format, rawReq)
	if normalized == nil {
		t.Fatalf("normalization failed")
	}

	// Step 3: Infer capability
	capability := router.InferCapability(rawReq)
	if capability != "chat" {
		t.Fatalf("capability inference failed")
	}

	// Step 4: Map parameters for Anthropic
	mapped := mapper.MapParameters(normalized.Parameters)
	if mapped["thinking"] != "enabled" {
		t.Errorf("thinking parameter not preserved")
	}
	if mapped["max_tokens"] != 4096 {
		t.Errorf("max_tokens not preserved")
	}

	t.Logf("✓ Anthropic path: detect → normalize → infer → map")
}

func TestPipeline_ParameterTranslation_OpenAIToAnthropic(t *testing.T) {
	// OpenAI request with Anthropic-incompatible parameters
	openaiReq := map[string]interface{}{
		"model":             "gpt-4",
		"messages":          []map[string]string{{"role": "user", "content": "hello"}},
		"temperature":       1.5,  // OpenAI range, will clamp to Anthropic [0,1]
		"response_format":   "json_object",  // Not supported by Anthropic
		"seed":              42,  // Not supported by Anthropic
		"presence_penalty":  0.5, // Not supported by Anthropic
		"max_tokens":        2048,
	}

	detector := middleware.NewFormatDetector()
	normalizer := middleware.NewNormalizer()
	anthropicMapper := adapter.NewAnthropicParameterMapper()

	// Detect as OpenAI
	format := detector.DetectFormat(openaiReq)
	// Normalize
	normalized := normalizer.Normalize(format, openaiReq)
	// Map to Anthropic (should filter unsupported params, clamp temperature)
	mapped := anthropicMapper.MapParameters(normalized.Parameters)

	if mapped["temperature"] != 1.0 {
		t.Errorf("temperature should clamp to 1.0, got %v", mapped["temperature"])
	}
	if _, exists := mapped["response_format"]; exists {
		t.Error("response_format should be filtered out")
	}
	if _, exists := mapped["seed"]; exists {
		t.Error("seed should be filtered out")
	}
	if _, exists := mapped["presence_penalty"]; exists {
		t.Error("presence_penalty should be filtered out")
	}

	t.Logf("✓ Parameter translation: OpenAI → Anthropic with filtering")
}

func TestPipeline_CapabilityRouting_VisionDetection(t *testing.T) {
	// Request with vision content
	visionReq := map[string]interface{}{
		"model": "router:vision",
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{"type": "text", "text": "What's in this image?"},
					map[string]interface{}{"type": "image_url", "image_url": map[string]string{"url": "https://example.com/image.jpg"}},
				},
			},
		},
		"max_tokens": 1024,
	}

	detector := middleware.NewFormatDetector()
	normalizer := middleware.NewNormalizer()

	format := detector.DetectFormat(visionReq)
	normalized := normalizer.Normalize(format, visionReq)

	// Infer capability should detect vision
	capability := router.InferCapability(visionReq)
	if capability != "vision" {
		t.Errorf("expected capability 'vision', got %s", capability)
	}

	// Router prefix extraction
	if router.IsCapabilityPrefix(normalized.Model) {
		_, cap := router.ExtractCapabilityPrefix(normalized.Model)
		if cap != "vision" {
			t.Errorf("expected extracted capability 'vision', got %s", cap)
		}
	}

	t.Logf("✓ Capability routing: vision detection works")
}

func TestPipeline_CapabilityRouting_EmbeddingDetection(t *testing.T) {
	// Embedding request
	embeddingReq := map[string]interface{}{
		"model": "router:embedding",
		"input": []string{"hello world", "goodbye world"},
	}

	// Should infer embedding capability
	capability := router.InferCapability(embeddingReq)
	if capability != "embedding" {
		t.Errorf("expected capability 'embedding', got %s", capability)
	}

	t.Logf("✓ Capability routing: embedding detection works")
}

func TestPipeline_CapabilityRouting_ReasoningDetection(t *testing.T) {
	// Reasoning request (Responses API format)
	reasoningReq := map[string]interface{}{
		"model":             "router:reasoning",
		"messages":          []map[string]string{{"role": "user", "content": "solve this hard problem"}},
		"reasoning_effort":  "high",
		"max_tokens":        10000,
	}

	// Should infer reasoning capability
	capability := router.InferCapability(reasoningReq)
	if capability != "reasoning" {
		t.Errorf("expected capability 'reasoning', got %s", capability)
	}

	t.Logf("✓ Capability routing: reasoning detection works")
}

func TestPipeline_ErrorPropagation(t *testing.T) {
	mapper := adapter.NewOpenAIParameterMapper()

	// Invalid parameters
	invalidParams := map[string]interface{}{
		"temperature": 5.0,  // Out of range
		"top_p":       1.5,  // Out of range
	}

	warnings := mapper.GetValidationWarnings(invalidParams)

	if len(warnings) == 0 {
		t.Error("expected validation warnings")
	}

	for _, warning := range warnings {
		t.Logf("  Validation: %s", warning)
	}
}

func TestPipeline_CachingThroughPipeline(t *testing.T) {
	// Verify that response can be cached after passing through pipeline
	detector := middleware.NewFormatDetector()
	normalizer := middleware.NewNormalizer()

	req1 := map[string]interface{}{
		"model":      "gpt-4",
		"messages":   []map[string]string{{"role": "user", "content": "hello"}},
		"max_tokens": 512,
	}

	req2 := map[string]interface{}{
		"model":      "gpt-4",
		"messages":   []map[string]string{{"role": "user", "content": "hello"}},
		"max_tokens": 512,
	}

	// Process both requests
	format1 := detector.DetectFormat(req1)
	norm1 := normalizer.Normalize(format1, req1)

	format2 := detector.DetectFormat(req2)
	norm2 := normalizer.Normalize(format2, req2)

	// Should be identical and cacheable
	if norm1.Model != norm2.Model {
		t.Error("normalized models should be identical")
	}
	if len(norm1.Messages) != len(norm2.Messages) {
		t.Error("normalized messages should be identical")
	}

	t.Logf("✓ Response caching: normalized requests are identical")
}
