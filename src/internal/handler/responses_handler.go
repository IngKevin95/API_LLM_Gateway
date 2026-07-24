package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/middleware"
	"api-llm-gateway/internal/router"
)

// Processor maneja la lógica de negocio subyacente.
type Processor interface {
	ProcessChat(ctx context.Context, req *adapter.Request) (*adapter.Response, error)
}

// ResponsesHandler implements universal /responses endpoint
type ResponsesHandler struct {
	processor  Processor
	detector   *middleware.FormatDetector
	normalizer *middleware.Normalizer
}

// NewResponsesHandler creates a new /responses endpoint handler
func NewResponsesHandler(p Processor) *ResponsesHandler {
	return &ResponsesHandler{
		processor:  p,
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

	// Convert normalized messages to adapter format
	adapterReq := &adapter.Request{
		Model: model,
	}
	for _, m := range normalized.Messages {
		role, _ := m["role"].(string)
		content, _ := m["content"].(string)
		if role == "" {
			role = "user"
		}
		if text, ok := m["text"].(string); ok && content == "" {
			content = text
		}
		adapterReq.Messages = append(adapterReq.Messages, adapter.Message{
			Role:    role,
			Content: content,
		})
	}

	// Route to appropriate provider (fallback chain)
	resp, err := h.processor.ProcessChat(r.Context(), adapterReq)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Provider error", err)
		return
	}

	// Format response back to universal format
	response := map[string]interface{}{
		"status":     "ok",
		"capability": capability,
		"model":      resp.ProviderID,
		"content":    resp.Content,
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
