package openai

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/dlp"
)

// Processor maneja la lógica de negocio subyacente.
type Processor interface {
	ProcessChat(ctx context.Context, req *adapter.Request) (*adapter.Response, error)
	ProcessChatStream(ctx context.Context, req *adapter.Request) (adapter.TokenStream, error)
	ProcessEmbedding(ctx context.Context, req *adapter.Request) (*adapter.Embedding, error)
}

// Handler HTTP para los endpoints de OpenAI.
type Handler struct {
	processor Processor
}

func NewHandler(p Processor) *Handler {
	return &Handler{processor: p}
}

func (h *Handler) HandleChatCompletions(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var req ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(r.Context(), "chat request decode error",
			slog.String("handler", "openai"),
			slog.String("method", "chat"),
			slog.String("error_type", "validation"),
			slog.String("error", err.Error()),
			slog.Int("status", http.StatusBadRequest),
			slog.Int64("latency_ms", time.Since(start).Milliseconds()),
		)
		http.Error(w, `{"error":{"message":"Invalid JSON payload","type":"invalid_request_error"}}`, http.StatusBadRequest)
		return
	}

	// HU-012a AC3: validar que messages sea obligatorio
	if len(req.Messages) == 0 {
		slog.ErrorContext(r.Context(), "chat request missing messages",
			slog.String("handler", "openai"),
			slog.String("method", "chat"),
			slog.String("model", req.Model),
			slog.String("error_type", "validation"),
			slog.Int("status", http.StatusBadRequest),
			slog.Int64("latency_ms", time.Since(start).Milliseconds()),
		)
		http.Error(w, `{"error":{"message":"missing messages field","type":"invalid_request_error"}}`, http.StatusBadRequest)
		return
	}

	internalReq := &adapter.Request{
		Model:             req.Model,
		MaxTokens:         req.MaxTokens,
		Stream:            req.Stream,
		StreamScannerFunc: dlp.ScannerFromContext(r.Context()),
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
		status := http.StatusInternalServerError
		message := "Internal server error"

		// Translate adapter.ProviderError to appropriate HTTP status
		if provErr, ok := err.(*adapter.ProviderError); ok {
			switch provErr.Status {
			case 401:
				status = http.StatusUnauthorized
				message = "Unauthorized: invalid API key"
			case 503:
				status = http.StatusServiceUnavailable
				message = "Service unavailable: no provider available"
			case 504:
				status = http.StatusGatewayTimeout
				message = "Gateway timeout"
			default:
				status = http.StatusInternalServerError
				message = "Provider error"
			}
		}

		errResp := map[string]any{
			"error": map[string]string{
				"message": message,
				"type":    "server_error",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(errResp)
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

	slog.InfoContext(r.Context(), "chat completed",
		slog.String("handler", "openai"),
		slog.String("method", "chat"),
		slog.String("model", req.Model),
		slog.Int("status", http.StatusOK),
		slog.Int64("latency_ms", time.Since(start).Milliseconds()),
	)
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

func (h *Handler) HandleEmbeddings(w http.ResponseWriter, r *http.Request) {
	var req EmbeddingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":{"message":"Invalid JSON payload","type":"invalid_request_error"}}`, http.StatusBadRequest)
		return
	}

	var inputs []string
	switch v := req.Input.(type) {
	case string:
		inputs = []string{v}
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				inputs = append(inputs, str)
			}
		}
	}

	if len(inputs) == 0 {
		http.Error(w, `{"error":{"message":"Input is required","type":"invalid_request_error"}}`, http.StatusBadRequest)
		return
	}

	internalReq := &adapter.Request{
		Model: req.Model,
		Input: inputs,
	}

	resp, err := h.processor.ProcessEmbedding(r.Context(), internalReq)
	if err != nil || resp == nil {
		http.Error(w, `{"error":{"message":"No embedding provider available","type":"service_unavailable"}}`, http.StatusServiceUnavailable)
		return
	}

	var data []EmbeddingData
	for i, vec := range resp.Vectors {
		data = append(data, EmbeddingData{
			Object:    "embedding",
			Index:     i,
			Embedding: vec,
		})
	}

	openaiResp := EmbeddingResponse{
		Object: "list",
		Data:   data,
		Model:  req.Model,
		Usage: Usage{
			PromptTokens: 0,
			TotalTokens:  0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(openaiResp)
}

func (h *Handler) HandleModels(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"object": "list",
		"data": []map[string]any{
			{"id": "router.coding", "object": "model", "created": time.Now().Unix(), "owned_by": "gateway"},
			{"id": "gpt-4o", "object": "model", "created": time.Now().Unix(), "owned_by": "openai"},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
