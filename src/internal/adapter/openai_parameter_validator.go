package adapter

import "fmt"

// OpenAIParameterValidator validates OpenAI API parameters
type OpenAIParameterValidator struct{}

// NewOpenAIParameterValidator creates a new validator
func NewOpenAIParameterValidator() *OpenAIParameterValidator {
	return &OpenAIParameterValidator{}
}

// ValidateTemperature checks if temperature is in valid range [0, 2]
func (v *OpenAIParameterValidator) ValidateTemperature(temp float64) bool {
	return temp >= 0.0 && temp <= 2.0
}

// ValidateTopP checks if top_p is in valid range [0, 1]
func (v *OpenAIParameterValidator) ValidateTopP(topP float64) bool {
	return topP >= 0.0 && topP <= 1.0
}

// ValidateSeed checks if seed is non-negative
func (v *OpenAIParameterValidator) ValidateSeed(seed int) bool {
	return seed >= 0
}

// ValidateToolChoice checks if tool_choice is one of allowed values
func (v *OpenAIParameterValidator) ValidateToolChoice(choice string) bool {
	valid := map[string]bool{
		"none":     true,
		"auto":     true,
		"required": true,
	}
	return valid[choice]
}

// ValidateResponseFormat checks if response_format is valid
func (v *OpenAIParameterValidator) ValidateResponseFormat(format interface{}) bool {
	if format == nil {
		return true // nil is valid (use default)
	}

	if s, ok := format.(string); ok {
		return s == "text" || s == "json_object"
	}

	return false
}

// ClampTemperature clamps temperature to [0, 2]
func (v *OpenAIParameterValidator) ClampTemperature(temp float64) float64 {
	if temp < 0.0 {
		return 0.0
	}
	if temp > 2.0 {
		return 2.0
	}
	return temp
}

// ClampTopP clamps top_p to [0, 1]
func (v *OpenAIParameterValidator) ClampTopP(topP float64) float64 {
	if topP < 0.0 {
		return 0.0
	}
	if topP > 1.0 {
		return 1.0
	}
	return topP
}

// ValidateMapParameters validates all parameters in a map and returns errors
func (v *OpenAIParameterValidator) ValidateMapParameters(params map[string]interface{}) []string {
	var errors []string

	for key, value := range params {
		switch key {
		case "temperature":
			if f, ok := value.(float64); ok {
				if !v.ValidateTemperature(f) {
					errors = append(errors, fmt.Sprintf("temperature out of range [0, 2]: %f", f))
				}
			}

		case "top_p":
			if f, ok := value.(float64); ok {
				if !v.ValidateTopP(f) {
					errors = append(errors, fmt.Sprintf("top_p out of range [0, 1]: %f", f))
				}
			}

		case "seed":
			if i, ok := value.(int); ok {
				if !v.ValidateSeed(i) {
					errors = append(errors, fmt.Sprintf("seed must be non-negative: %d", i))
				}
			}

		case "tool_choice":
			if s, ok := value.(string); ok {
				if !v.ValidateToolChoice(s) {
					errors = append(errors, fmt.Sprintf("invalid tool_choice: %s", s))
				}
			}

		case "response_format":
			if !v.ValidateResponseFormat(value) {
				errors = append(errors, fmt.Sprintf("invalid response_format: %v", value))
			}
		}
	}

	return errors
}

// GetUnknownParameters returns list of parameters not recognized by OpenAI API
func (v *OpenAIParameterValidator) GetUnknownParameters(params map[string]interface{}) []string {
	knownParams := map[string]bool{
		"model":             true,
		"messages":          true,
		"temperature":       true,
		"top_p":             true,
		"n":                 true,
		"stream":            true,
		"stop":              true,
		"max_tokens":        true,
		"presence_penalty":  true,
		"frequency_penalty": true,
		"logit_bias":        true,
		"user":              true,
		"tools":             true,
		"tool_choice":       true,
		"response_format":   true,
		"seed":              true,
	}

	var unknown []string
	for key := range params {
		if !knownParams[key] {
			unknown = append(unknown, key)
		}
	}

	return unknown
}
