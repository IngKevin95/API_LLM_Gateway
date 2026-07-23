package request

// NormalizedRequest represents a request normalized to internal format.
// All incoming requests (OpenAI, Anthropic, Responses API) are translated
// to this structure for uniform processing in the router and adapters.
type NormalizedRequest struct {
	// Format is the detected input format: "openai", "anthropic", "responses"
	Format string

	// Model is the requested model name
	Model string

	// Messages is the conversation history in normalized format
	Messages []map[string]interface{}

	// Parameters contains request-specific parameters (temperature, top_p, etc.)
	Parameters map[string]interface{}
}
