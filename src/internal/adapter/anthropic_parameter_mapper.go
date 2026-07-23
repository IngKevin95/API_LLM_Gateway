package adapter

// AnthropicParameterMapper translates and validates Anthropic parameters
type AnthropicParameterMapper struct {
	validator *AnthropicParameterValidator
}

// NewAnthropicParameterMapper creates a new Anthropic parameter mapper
func NewAnthropicParameterMapper() *AnthropicParameterMapper {
	return &AnthropicParameterMapper{
		validator: NewAnthropicParameterValidator(),
	}
}

// MapParameters processes and validates Anthropic parameters
// Returns a cleaned map with validated and clamped parameters
func (m *AnthropicParameterMapper) MapParameters(params map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Known parameters that we allow
	for key, value := range params {
		switch key {
		case "temperature":
			if f, ok := value.(float64); ok {
				result["temperature"] = m.validator.ClampTemperature(f)
			}

		case "top_k":
			if i, ok := value.(int); ok {
				if m.validator.ValidateTopK(i) {
					result["top_k"] = i
				}
				// Invalid top_k values are skipped (not included in result)
			}

		case "top_p":
			if f, ok := value.(float64); ok {
				result["top_p"] = f
			}

		case "max_tokens":
			// max_tokens is required for Anthropic, pass through
			result["max_tokens"] = value

		case "thinking":
			if s, ok := value.(string); ok {
				if m.validator.ValidateThinking(s) {
					result["thinking"] = s
				}
			}

		case "tool_use":
			if s, ok := value.(string); ok {
				if m.validator.ValidateToolUse(s) {
					result["tool_use"] = s
				}
			}

		case "model", "messages", "system":
			// These are handled separately, pass through
			result[key] = value

		case "stop", "tools":
			// Other standard parameters pass through
			result[key] = value

		// Unsupported OpenAI features are silently dropped:
		// response_format, seed, presence_penalty, frequency_penalty, n, etc.
		}
	}

	return result
}

// GetValidationWarnings returns list of parameters that failed validation
func (m *AnthropicParameterMapper) GetValidationWarnings(params map[string]interface{}) []string {
	warnings := m.validator.ValidateMapParameters(params)
	unknowns := m.validator.GetUnknownParameters(params)

	// Combine validation errors and unknown parameters
	warnings = append(warnings, unknowns...)

	return warnings
}
