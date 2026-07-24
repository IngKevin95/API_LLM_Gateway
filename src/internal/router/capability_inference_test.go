package router

import (
	"testing"
)

func TestCapabilityInference_DefaultToChat(t *testing.T) {
	// Empty request defaults to "chat"
	req := map[string]interface{}{}
	capability := InferCapability(req)
	if capability != "chat" {
		t.Errorf("expected default capability 'chat', got '%s'", capability)
	}
}

func TestCapabilityInference_ChatFromMessages(t *testing.T) {
	// Request with messages is "chat"
	req := map[string]interface{}{
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "hello"},
		},
	}
	capability := InferCapability(req)
	if capability != "chat" {
		t.Errorf("expected 'chat' from messages, got '%s'", capability)
	}
}

func TestCapabilityInference_VisionFromImage(t *testing.T) {
	// Request with image content is "vision"
	req := map[string]interface{}{
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{
						"type": "image_url",
						"image_url": map[string]interface{}{
							"url": "https://example.com/image.png",
						},
					},
				},
			},
		},
	}
	capability := InferCapability(req)
	if capability != "vision" {
		t.Errorf("expected 'vision' from image content, got '%s'", capability)
	}
}

func TestCapabilityInference_VisionFromImageURL(t *testing.T) {
	// Request with image_url field is "vision"
	req := map[string]interface{}{
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{
						"type": "image_url",
					},
				},
			},
		},
	}
	capability := InferCapability(req)
	if capability != "vision" {
		t.Errorf("expected 'vision' from image_url, got '%s'", capability)
	}
}

func TestCapabilityInference_VisionFromImageBase64(t *testing.T) {
	// Request with base64 encoded image is "vision"
	req := map[string]interface{}{
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{
						"type":  "image",
						"image": "data:image/png;base64,iVBORw0KG...",
					},
				},
			},
		},
	}
	capability := InferCapability(req)
	if capability != "vision" {
		t.Errorf("expected 'vision' from base64 image, got '%s'", capability)
	}
}

func TestCapabilityInference_EmbeddingFromEmbeddingInput(t *testing.T) {
	// Request with "input" field for embeddings
	req := map[string]interface{}{
		"input": "text to embed",
	}
	capability := InferCapability(req)
	if capability != "embedding" {
		t.Errorf("expected 'embedding' from input field, got '%s'", capability)
	}
}

func TestCapabilityInference_EmbeddingFromMultipleInputs(t *testing.T) {
	// Request with array of inputs for embeddings
	req := map[string]interface{}{
		"input": []interface{}{
			"text 1",
			"text 2",
		},
	}
	capability := InferCapability(req)
	if capability != "embedding" {
		t.Errorf("expected 'embedding' from input array, got '%s'", capability)
	}
}

func TestCapabilityInference_ReasoningFromReasoningEffort(t *testing.T) {
	// Responses API with reasoning_effort is "reasoning"
	req := map[string]interface{}{
		"reasoning_effort": "high",
		"input": []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": "solve this math problem",
			},
		},
	}
	capability := InferCapability(req)
	if capability != "reasoning" {
		t.Errorf("expected 'reasoning' from reasoning_effort, got '%s'", capability)
	}
}

func TestCapabilityInference_PrioritizeReasoningOverChat(t *testing.T) {
	// reasoning_effort takes priority over messages
	req := map[string]interface{}{
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "hello"},
		},
		"reasoning_effort": "medium",
	}
	capability := InferCapability(req)
	if capability != "reasoning" {
		t.Errorf("expected 'reasoning' priority over chat, got '%s'", capability)
	}
}

func TestCapabilityInference_TextOnlyIsChat(t *testing.T) {
	// Text-only messages are chat, not vision
	req := map[string]interface{}{
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": "just plain text",
			},
		},
	}
	capability := InferCapability(req)
	if capability != "chat" {
		t.Errorf("expected 'chat' for text-only, got '%s'", capability)
	}
}

func TestCapabilityInference_MixedContentWithImage(t *testing.T) {
	// Mixed content with image -> vision
	req := map[string]interface{}{
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "What's in this image?",
					},
					map[string]interface{}{
						"type": "image_url",
						"image_url": map[string]interface{}{
							"url": "https://example.com/image.png",
						},
					},
				},
			},
		},
	}
	capability := InferCapability(req)
	if capability != "vision" {
		t.Errorf("expected 'vision' for mixed content with image, got '%s'", capability)
	}
}
