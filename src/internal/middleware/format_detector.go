package middleware

import "strings"

// FormatDetector identifies the format of incoming requests (OpenAI, Anthropic, Responses).
type FormatDetector struct{}

// NewFormatDetector creates a new FormatDetector instance.
func NewFormatDetector() *FormatDetector {
	return &FormatDetector{}
}

// DetectFormat inspects the request structure and returns the detected format.
// Returns: "openai", "anthropic", "responses", or "" if unknown.
func (fd *FormatDetector) DetectFormat(request map[string]interface{}) string {
	// Responses API: has "input" and "reasoning_effort"
	if _, hasInput := request["input"]; hasInput {
		if _, hasReasoning := request["reasoning_effort"]; hasReasoning {
			return "responses"
		}
	}

	// Both OpenAI and Anthropic have "messages" and "model"
	if _, hasMessages := request["messages"]; !hasMessages {
		return ""
	}
	if _, hasModel := request["model"]; !hasModel {
		return ""
	}

	// Heuristic: if model name contains "claude", it's Anthropic
	if model, ok := request["model"].(string); ok {
		if strings.Contains(model, "claude") {
			return "anthropic"
		}
	}

	// Default to OpenAI for generic messages+model
	return "openai"
}
