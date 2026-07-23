package middleware

import "api-llm-gateway/internal/request"

// Normalizer converts requests from any detected format to NormalizedRequest.
type Normalizer struct{}

// NewNormalizer creates a new Normalizer instance.
func NewNormalizer() *Normalizer {
	return &Normalizer{}
}

// Normalize converts a request of a given format to NormalizedRequest.
func (n *Normalizer) Normalize(format string, rawRequest map[string]interface{}) *request.NormalizedRequest {
	nr := &request.NormalizedRequest{
		Format:     format,
		Parameters: make(map[string]interface{}),
	}

	// Extract model (common field across all formats)
	if model, ok := rawRequest["model"].(string); ok {
		nr.Model = model
	}

	// Extract messages (common field for OpenAI/Anthropic)
	if messages, ok := rawRequest["messages"].([]interface{}); ok {
		nr.Messages = convertMessages(messages)
	}

	// Extract input (Responses API specific)
	if input, ok := rawRequest["input"].([]interface{}); ok {
		nr.Messages = convertMessages(input)
	}

	// Copy all other fields as parameters
	for key, value := range rawRequest {
		if key != "model" && key != "messages" && key != "input" {
			nr.Parameters[key] = value
		}
	}

	return nr
}

// convertMessages converts a slice of interface{} to a slice of map[string]interface{}
func convertMessages(msgs []interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(msgs))
	for _, msg := range msgs {
		if m, ok := msg.(map[string]interface{}); ok {
			result = append(result, m)
		}
	}
	return result
}
