package main

import (
	"context"
	"errors"
	"log"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/failover"
)

// GatewayProcessor implements the Processor interface for API handlers.
// It wraps the Failover Engine to complete requests through the Router -> Failover -> Adapters pipeline.
// EP-011 Scope:
//   HU-050: Logging OpenAI handler
//   HU-051: Debug ProcessChat()
//   HU-052: Validar Router.Route()
//   HU-056: Normalizar IDs proveedores (config.yaml)
//   HU-057: Implementar ProcessEmbedding()
//   HU-058: Fix Anthropic /v1/messages
type GatewayProcessor struct {
	failover *failover.Engine
}

// NewGatewayProcessor creates a processor from a failover engine.
func NewGatewayProcessor(fe *failover.Engine) *GatewayProcessor {
	return &GatewayProcessor{failover: fe}
}

// ProcessChat processes a non-streaming chat request.
// HU-050: Logging OpenAI handler — logs request/response metadata
// HU-051: Debug ProcessChat() — estructura de logging y manejo de errores
// HU-052: Validar Router.Route() — invoca failover.Complete() que usa Router.Route()
func (gp *GatewayProcessor) ProcessChat(ctx context.Context, req *adapter.Request) (*adapter.Response, error) {
	log.Printf("ProcessChat: model=%s, messages=%d", req.Model, len(req.Messages))

	// Use "chat" as default capability (can be overridden based on request parameters in future)
	// HU-056: Normalizar IDs proveedores — config.yaml IDs usados por Router.Route()
	resp, err := gp.failover.Complete(ctx, "chat", *req)
	if err != nil {
		log.Printf("ProcessChat ERROR: %v (type: %T)", err, err)
		return nil, err
	}

	log.Printf("ProcessChat OK: content=%d chars", len(resp.Content))
	return &resp, nil
}

// ProcessChatStream processes a streaming chat request.
// HU-053: OmniRoute adapter será implementado en EP-011 (internal/adapter/omniroute/)
// HU-054: Registrar adaptador OmniRoute en buildAdapters()
// Streaming se construye cuando adapter OmniRoute esté listo
func (gp *GatewayProcessor) ProcessChatStream(ctx context.Context, req *adapter.Request) (adapter.TokenStream, error) {
	// Stub para streaming: será reemplazado cuando HU-053/054 creen adaptador OmniRoute
	return nil, &adapter.ProviderError{
		Provider:  "gateway",
		Status:    501,
		Retryable: false,
		Err:       errors.New("streaming no implementado en MVP, depende de HU-053/054 OmniRoute adapter"),
	}
}

// ProcessEmbedding processes an embedding request.
// HU-057: Implementar ProcessEmbedding() — se construye en EP-011 first slice
// Requiere: HU-056 normalización IDs proveedores
func (gp *GatewayProcessor) ProcessEmbedding(ctx context.Context, req *adapter.Request) (*adapter.Embedding, error) {
	// Stub para HU-057: será reemplazado por invocación failover.Embed(ctx, "embedding", *req)
	// AC HU-057: (1) happy path request->adapter->embedding, (2) error provider, (3) edge model no soporta embeddings
	return nil, &adapter.ProviderError{
		Provider:  "gateway",
		Status:    501,
		Retryable: false,
		Err:       errors.New("HU-057: embeddings no implementado en MVP, será completado en EP-011 first slice"),
	}
}
