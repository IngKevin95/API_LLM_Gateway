package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter"
)

// Processor maneja la lógica de negocio subyacente.
type Processor interface {
	ProcessChat(ctx context.Context, req *adapter.Request) (*adapter.Response, error)
}

// Handler HTTP para los endpoints de OpenAI.
type Handler struct {
	processor Processor
}

func NewHandler(p Processor) *Handler {
	return &Handler{processor: p}
}

func (h *Handler) HandleChatCompletions(w http.ResponseWriter, r *http.Request) {
	var req ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":{"message":"Invalid JSON payload","type":"invalid_request_error"}}`, http.StatusBadRequest)
		return
	}

	// Traducir a internal request
	internalReq := &adapter.Request{
		Model:     req.Model,
		MaxTokens: req.MaxTokens,
		Stream:    req.Stream,
	}
	for _, m := range req.Messages {
		internalReq.Messages = append(internalReq.Messages, adapter.Message{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	// Procesar sincrónicamente (no streaming)
	// Para este subslice, si stream=true, por ahora fallamos o lo ignoramos.
	// Asumimos que SS1 es solo sin streaming.
	resp, err := h.processor.ProcessChat(r.Context(), internalReq)
	if err != nil {
		http.Error(w, `{"error":{"message":"Internal server error","type":"server_error"}}`, http.StatusInternalServerError)
		return
	}

	// Convertir a respuesta OpenAI
	openaiResp := ChatCompletionResponse{
		ID:      "chatcmpl-mock",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model, // En un caso real, el modelo real usado.
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: resp.Content,
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     0, // TODO: usar tokenizer o info del adapter
			CompletionTokens: 0,
			TotalTokens:      0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(openaiResp)
}
