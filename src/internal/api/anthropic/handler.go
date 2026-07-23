package anthropic

import (
	"context"
	"encoding/json"
	"net/http"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/dlp"
)

type Processor interface {
	ProcessChat(ctx context.Context, req *adapter.Request) (*adapter.Response, error)
	ProcessChatStream(ctx context.Context, req *adapter.Request) (adapter.TokenStream, error)
}

type Handler struct {
	processor Processor
}

func NewHandler(p Processor) *Handler {
	return &Handler{processor: p}
}

func (h *Handler) HandleMessages(w http.ResponseWriter, r *http.Request) {
	var req MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"type":"error","error":{"type":"invalid_request_error","message":"Invalid JSON payload"}}`, http.StatusBadRequest)
		return
	}

	internalReq := &adapter.Request{
		Model:             req.Model,
		MaxTokens:         req.MaxTokens,
		Stream:            req.Stream,
		StreamScannerFunc: dlp.ScannerFromContext(r.Context()),
	}

	for _, m := range req.Messages {
		if len(m.Content) > 0 && m.Content[0].Type == "text" {
			internalReq.Messages = append(internalReq.Messages, adapter.Message{
				Role:    m.Role,
				Content: m.Content[0].Text,
			})
		}
	}

	if req.Stream {
		h.handleStream(w, r, internalReq, req.Model)
		return
	}

	resp, err := h.processor.ProcessChat(r.Context(), internalReq)
	if err != nil || resp == nil {
		http.Error(w, `{"type":"error","error":{"type":"invalid_request_error","message":"Invalid request"}}`, http.StatusBadRequest)
		return
	}

	anthropicResp := MessageResponse{
		ID:    "msg_mock",
		Type:  "message",
		Role:  "assistant",
		Model: req.Model,
		Content: []Content{
			{Type: "text", Text: resp.Content},
		},
		StopReason: "end_turn",
		Usage: Usage{
			InputTokens:  0,
			OutputTokens: 0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(anthropicResp)
}

func (h *Handler) handleStream(w http.ResponseWriter, r *http.Request, internalReq *adapter.Request, model string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	stream, err := h.processor.ProcessChatStream(r.Context(), internalReq)
	if err != nil {
		http.Error(w, `{"type":"error","error":{"type":"api_error","message":"Internal server error"}}`, http.StatusInternalServerError)
		return
	}
	defer stream.Close()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	encoder := json.NewEncoder(w)

	// message_start
	startEvent := MessageStartEvent{
		Type: "message_start",
		Message: MessageResponse{
			ID:    "msg_stream_mock",
			Type:  "message",
			Role:  "assistant",
			Model: model,
			Usage: Usage{InputTokens: 0, OutputTokens: 0},
		},
	}
	w.Write([]byte("event: message_start\ndata: "))
	encoder.Encode(startEvent)
	w.Write([]byte("\n"))
	flusher.Flush()

	// content_block_start
	w.Write([]byte("event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n"))
	flusher.Flush()

	for {
		token, ok, err := stream.Next()
		if err != nil || !ok {
			break
		}

		delta := ContentBlockDeltaEvent{
			Type:  "content_block_delta",
			Index: 0,
			Delta: Delta{Type: "text_delta", Text: token},
		}
		w.Write([]byte("event: content_block_delta\ndata: "))
		encoder.Encode(delta)
		w.Write([]byte("\n"))
		flusher.Flush()
	}

	// content_block_stop
	w.Write([]byte("event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":0}\n\n"))

	// message_stop
	w.Write([]byte("event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n"))
	flusher.Flush()
}
