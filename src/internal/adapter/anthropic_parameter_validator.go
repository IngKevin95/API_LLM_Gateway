package adapter

import "fmt"

// AnthropicParameterValidator validates Anthropic API parameters
type AnthropicParameterValidator struct{}

// NewAnthropicParameterValidator creates a new Anthropic parameter validator
func NewAnthropicParameterValidator() *AnthropicParameterValidator {
	return &AnthropicParameterValidator{}
}

// ValidateTemperature checks if temperature is in valid range [0, 1] for Anthropic
func (v *AnthropicParameterValidator) ValidateTemperature(temp float64) bool {
	return temp >= 0.0 && temp <= 1.0
}

// ValidateTopK checks if top_k is valid (must be >= 1)
func (v *AnthropicParameterValidator) ValidateTopK(topK int) bool {
	return topK >= 1
}

// IsMaxTokensRequired returns true because Anthropic requires max_tokens
func (v *AnthropicParameterValidator) IsMaxTokensRequired() bool {
	return true
}

// ValidateMaxTokens checks if max_tokens is positive
func (v *AnthropicParameterValidator) ValidateMaxTokens(maxTokens int) bool {
	return maxTokens > 0
}

// ValidateThinking checks if thinking mode is valid
func (v *AnthropicParameterValidator) ValidateThinking(thinking string) bool {
	valid := map[string]bool{
		"enabled":  true,
		"disabled": true,
	}
	return valid[thinking]
}

// ValidateToolUse checks if tool_use mode is valid
func (v *AnthropicParameterValidator) ValidateToolUse(toolUse string) bool {
	valid := map[string]bool{
		"auto":     true,
		"required": true,
		"none":     true,
	}
	return valid[toolUse]
}

// ClampTemperature clamps temperature to [0, 1]
func (v *AnthropicParameterValidator) ClampTemperature(temp float64) float64 {
	if temp < 0.0 {
		return 0.0
	}
	if temp > 1.0 {
		return 1.0
	}
	return temp
}

// ValidateMapParameters validates all parameters in a map and returns errors
func (v *AnthropicParameterValidator) ValidateMapParameters(params map[string]interface{}) []string {
	var errors []string

	// max_tokens is required for Anthropic
	if _, exists := params["max_tokens"]; !exists {
		errors = append(errors, "max_tokens is required for Anthropic")
	}

	for key, value := range params {
		switch key {
		case "temperature":
			if f, ok := value.(float64); ok {
				if !v.ValidateTemperature(f) {
					errors = append(errors, fmt.Sprintf("temperature out of range [0, 1]: %f", f))
				}
			}

		case "top_k":
			if i, ok := value.(int); ok {
				if !v.ValidateTopK(i) {
					errors = append(errors, fmt.Sprintf("top_k must be >= 1: %d", i))
				}
			}

		case "max_tokens":
			if i, ok := value.(int); ok {
				if !v.ValidateMaxTokens(i) {
					errors = append(errors, fmt.Sprintf("max_tokens must be positive: %d", i))
				}
			}

		case "thinking":
			if s, ok := value.(string); ok {
				if !v.ValidateThinking(s) {
					errors = append(errors, fmt.Sprintf("invalid thinking mode: %s", s))
				}
			}

		case "tool_use":
			if s, ok := value.(string); ok {
				if !v.ValidateToolUse(s) {
					errors = append(errors, fmt.Sprintf("invalid tool_use mode: %s", s))
				}
			}
		}
	}

	return errors
}

// GetUnknownParameters returns list of parameters not recognized by Anthropic API
func (v *AnthropicParameterValidator) GetUnknownParameters(params map[string]interface{}) []string {
	knownParams := map[string]bool{
		"model":       true,
		"messages":    true,
		"system":      true,
		"temperature": true,
		"top_k":       true,
		"top_p":       true,
		"max_tokens":  true,
		"stop":        true,
		"thinking":    true,
		"tool_use":    true,
		"tools":       true,
	}

	var unknown []string
	for key := range params {
		if !knownParams[key] {
			unknown = append(unknown, key)
		}
	}

	return unknown
}

// CheckUnsupportedFeatures identifies OpenAI features not supported by Anthropic
func (v *AnthropicParameterValidator) CheckUnsupportedFeatures(params map[string]interface{}) map[string]interface{} {
	unsupported := make(map[string]interface{})

	// response_format is not directly supported by Anthropic like OpenAI
	if val, exists := params["response_format"]; exists {
		unsupported["response_format"] = val
	}

	// seed is not supported by Anthropic
	if val, exists := params["seed"]; exists {
		unsupported["seed"] = val
	}

	// presence_penalty and frequency_penalty not supported
	if val, exists := params["presence_penalty"]; exists {
		unsupported["presence_penalty"] = val
	}
	if val, exists := params["frequency_penalty"]; exists {
		unsupported["frequency_penalty"] = val
	}

	// n (number of completions) not supported
	if val, exists := params["n"]; exists {
		unsupported["n"] = val
	}

	return unsupported
}
