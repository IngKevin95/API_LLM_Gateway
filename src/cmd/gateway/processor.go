package main

import (
	"context"
	"errors"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/failover"
)

// GatewayProcessor implements the Processor interface for API handlers.
// It wraps the Failover Engine to complete requests through the Router -> Failover -> Adapters pipeline.
type GatewayProcessor struct {
	failover *failover.Engine
}

// NewGatewayProcessor creates a processor from a failover engine.
func NewGatewayProcessor(fe *failover.Engine) *GatewayProcessor {
	return &GatewayProcessor{failover: fe}
}

// ProcessChat processes a non-streaming chat request.
func (gp *GatewayProcessor) ProcessChat(ctx context.Context, req *adapter.Request) (*adapter.Response, error) {
	// Use "chat" as default capability (can be overridden based on request parameters in future)
	resp, err := gp.failover.Complete(ctx, "chat", *req)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ProcessChatStream processes a streaming chat request.
// TODO: Implement streaming in Fase 2 when Failover supports Stream
func (gp *GatewayProcessor) ProcessChatStream(ctx context.Context, req *adapter.Request) (adapter.TokenStream, error) {
	return nil, &adapter.ProviderError{
		Provider:  "gateway",
		Status:    501,
		Retryable: false,
		Err:       errors.New("streaming no implementado en MVP"),
	}
}

// ProcessEmbedding processes an embedding request.
// TODO: Implement embeddings in Fase 2 when Failover supports Embed
func (gp *GatewayProcessor) ProcessEmbedding(ctx context.Context, req *adapter.Request) (*adapter.Embedding, error) {
	return nil, &adapter.ProviderError{
		Provider:  "gateway",
		Status:    501,
		Retryable: false,
		Err:       errors.New("embeddings no implementado en MVP"),
	}
}
