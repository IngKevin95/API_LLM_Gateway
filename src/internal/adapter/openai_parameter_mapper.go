package adapter

// OpenAIParameterMapper translates and validates OpenAI parameters
type OpenAIParameterMapper struct {
	validator *OpenAIParameterValidator
}

// NewOpenAIParameterMapper creates a new parameter mapper
func NewOpenAIParameterMapper() *OpenAIParameterMapper {
	return &OpenAIParameterMapper{
		validator: NewOpenAIParameterValidator(),
	}
}

// MapParameters processes and validates OpenAI parameters
// Returns a cleaned map with validated and clamped parameters
func (m *OpenAIParameterMapper) MapParameters(params map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Known parameters that we allow
	for key, value := range params {
		switch key {
		case "temperature":
			if f, ok := value.(float64); ok {
				result["temperature"] = m.validator.ClampTemperature(f)
			}

		case "top_p":
			if f, ok := value.(float64); ok {
				result["top_p"] = m.validator.ClampTopP(f)
			}

		case "seed":
			if i, ok := value.(int); ok {
				if m.validator.ValidateSeed(i) {
					result["seed"] = i
				}
				// Invalid seeds are skipped (not included in result)
			}

		case "tool_choice":
			if s, ok := value.(string); ok {
				if m.validator.ValidateToolChoice(s) {
					result["tool_choice"] = s
				}
			}

		case "response_format":
			if m.validator.ValidateResponseFormat(value) {
				result["response_format"] = value
			}

		case "max_tokens":
			// Pass through max_tokens without transformation
			result["max_tokens"] = value

		case "model", "messages":
			// These are handled separately, pass through
			result[key] = value

		case "n", "stream", "stop", "presence_penalty", "frequency_penalty", "logit_bias", "user", "tools":
			// Other standard parameters pass through
			result[key] = value

			// Unknown parameters are silently dropped
		}
	}

	return result
}

// GetValidationWarnings returns list of parameters that failed validation
func (m *OpenAIParameterMapper) GetValidationWarnings(params map[string]interface{}) []string {
	warnings := m.validator.ValidateMapParameters(params)
	unknowns := m.validator.GetUnknownParameters(params)

	// Combine validation errors and unknown parameters
	warnings = append(warnings, unknowns...)

	return warnings
}
