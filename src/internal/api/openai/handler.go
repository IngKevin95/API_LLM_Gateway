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
	ProcessChatStream(ctx context.Context, req *adapter.Request) (adapter.TokenStream, error)
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

	if req.Stream {
		h.handleStream(w, r, internalReq, req.Model)
		return
	}

	resp, err := h.processor.ProcessChat(r.Context(), internalReq)
	if err != nil {
		http.Error(w, `{"error":{"message":"Internal server error","type":"server_error"}}`, http.StatusInternalServerError)
		return
	}

	openaiResp := ChatCompletionResponse{
		ID:      "chatcmpl-mock",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
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
			PromptTokens:     0,
			CompletionTokens: 0,
			TotalTokens:      0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(openaiResp)
}

func (h *Handler) handleStream(w http.ResponseWriter, r *http.Request, internalReq *adapter.Request, model string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	stream, err := h.processor.ProcessChatStream(r.Context(), internalReq)
	if err != nil {
		http.Error(w, `{"error":{"message":"Internal server error","type":"server_error"}}`, http.StatusInternalServerError)
		return
	}
	defer stream.Close()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	encoder := json.NewEncoder(w)

	// Send initial role chunk
	initialChunk := ChatCompletionChunk{
		ID:      "chatcmpl-stream-mock",
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []ChunkChoice{{Index: 0, Delta: ChunkMessage{Role: "assistant"}}},
	}
	w.Write([]byte("data: "))
	encoder.Encode(initialChunk)
	w.Write([]byte("\n"))
	flusher.Flush()

	for {
		token, ok, err := stream.Next()
		if err != nil {
			// In SSE, usually we just stop or send an error chunk, but OpenAI drops connection or sends error json?
			break
		}
		if !ok {
			break
		}

		chunk := ChatCompletionChunk{
			ID:      "chatcmpl-stream-mock",
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   model,
			Choices: []ChunkChoice{{Index: 0, Delta: ChunkMessage{Content: token}}},
		}
		w.Write([]byte("data: "))
		encoder.Encode(chunk)
		w.Write([]byte("\n"))
		flusher.Flush()
	}

	w.Write([]byte("data: [DONE]\n\n"))
	flusher.Flush()
}
