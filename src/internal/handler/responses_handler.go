package handler

import (
	"encoding/json"
	"net/http"

	"api-llm-gateway/internal/middleware"
	"api-llm-gateway/internal/router"
)

// ResponsesHandler implements universal /responses endpoint
type ResponsesHandler struct {
	detector   *middleware.FormatDetector
	normalizer *middleware.Normalizer
}

// NewResponsesHandler creates a new /responses endpoint handler
func NewResponsesHandler() *ResponsesHandler {
	return &ResponsesHandler{
		detector:   middleware.NewFormatDetector(),
		normalizer: middleware.NewNormalizer(),
	}
}

// ServeHTTP implements http.Handler for the /responses endpoint
func (h *ResponsesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var rawPayload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&rawPayload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON payload", err)
		return
	}

	// Detect format (OpenAI, Anthropic, or universal)
	format := h.detector.DetectFormat(rawPayload)

	// Validate minimum required fields
	if _, hasMessages := rawPayload["messages"]; !hasMessages {
		if _, hasInput := rawPayload["input"]; !hasInput {
			writeJSONError(w, http.StatusBadRequest, "missing required field: messages or input", nil)
			return
		}
	}

	// Normalize to internal format
	normalized := h.normalizer.Normalize(format, rawPayload)

	if normalized == nil {
		writeJSONError(w, http.StatusBadRequest, "failed to normalize request", nil)
		return
	}

	// Validate model is present
	if normalized.Model == "" {
		writeJSONError(w, http.StatusBadRequest, "model is required", nil)
		return
	}

	// Infer capability from payload
	capability := router.InferCapability(rawPayload)

	// Resolve router prefix if present
	model := normalized.Model
	if router.IsCapabilityPrefix(model) {
		_, capability = router.ExtractCapabilityPrefix(model)
	}

	// Route to appropriate provider (fallback chain)
	// For now, return a 200 OK stub
	response := map[string]interface{}{
		"status":     "ok",
		"capability": capability,
		"model":      model,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, status int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errResp := map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
		},
	}

	if err != nil {
		errResp["error"].(map[string]interface{})["details"] = err.Error()
	}

	json.NewEncoder(w).Encode(errResp)
}
