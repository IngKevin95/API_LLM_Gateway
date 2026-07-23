package main

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/failover"
	"api-llm-gateway/internal/metrics"
	"api-llm-gateway/internal/middleware"
)

// MetricsRecorder interface para inyección de dependencia en tests
type MetricsRecorder interface {
	Record(m metrics.RequestMetric)
}

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
	failover     *failover.Engine
	metricsStore MetricsRecorder
}

// NewGatewayProcessor creates a processor from a failover engine.
func NewGatewayProcessor(fe *failover.Engine, ms MetricsRecorder) *GatewayProcessor {
	return &GatewayProcessor{failover: fe, metricsStore: ms}
}

// ProcessChat processes a non-streaming chat request.
// HU-050: Logging OpenAI handler — logs request/response metadata
// HU-051: Debug ProcessChat() — estructura de logging y manejo de errores
// HU-052: Validar Router.Route() — invoca failover.Complete() que usa Router.Route()
func (gp *GatewayProcessor) ProcessChat(ctx context.Context, req *adapter.Request) (*adapter.Response, error) {
	start := time.Now()
	reqID := middleware.FromContext(ctx)

	slog.InfoContext(ctx, "request received",
		slog.String("component", "openai-handler"),
		slog.String("action", "request_received"),
		slog.String("model", req.Model),
		slog.Int("messages_count", len(req.Messages)),
		slog.String("request_id", reqID),
	)

	// Use "chat" as default capability (can be overridden based on request parameters in future)
	// HU-056: Normalizar IDs proveedores — config.yaml IDs usados por Router.Route()
	resp, err := gp.failover.Complete(ctx, "chat", *req)
	latencyMs := time.Since(start).Milliseconds()

	if err != nil {
		provider := "unknown"
		status := 500
		if provErr, ok := err.(*adapter.ProviderError); ok {
			provider = provErr.Provider
			status = provErr.Status
		}

		// HU-060: registrar métrica de error
		if gp.metricsStore != nil {
			gp.metricsStore.Record(metrics.RequestMetric{
				Provider:  provider,
				Model:     req.Model,
				LatencyMs: latencyMs,
				Status:    status,
			})
		}

		logFields := []any{
			slog.String("component", "openai-handler"),
			slog.String("action", "request_failed"),
			slog.String("model", req.Model),
			slog.String("request_id", reqID),
			slog.Int64("latency_ms", latencyMs),
		}
		if provErr, ok := err.(*adapter.ProviderError); ok {
			logFields = append(logFields, slog.Int("provider_status", provErr.Status))
			logFields = append(logFields, slog.String("provider", provErr.Provider))
		} else {
			logFields = append(logFields, slog.String("error_type", "internal"))
		}
		slog.ErrorContext(ctx, "request failed", logFields...)
		return nil, err
	}

	// HU-060: registrar métrica de éxito (provider extraído de failover, asumimos success status 200)
	if gp.metricsStore != nil {
		gp.metricsStore.Record(metrics.RequestMetric{
			Provider:  "unknown", // ponytail: failover no retorna provider usado; futuro: modificar Complete para retornarlo
			Model:     req.Model,
			LatencyMs: latencyMs,
			Status:    200,
		})
	}

	slog.InfoContext(ctx, "request completed",
		slog.String("component", "openai-handler"),
		slog.String("action", "request_completed"),
		slog.String("model", req.Model),
		slog.Int("response_length", len(resp.Content)),
		slog.String("request_id", reqID),
		slog.Int64("latency_ms", latencyMs),
	)
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
// HU-057: Implementar ProcessEmbedding() — usa failover.Embed() con capability "embedding"
// Requiere: HU-056 normalización IDs proveedores
// AC HU-057: (1) happy path request->adapter->embedding, (2) error provider, (3) edge model no soporta embeddings
func (gp *GatewayProcessor) ProcessEmbedding(ctx context.Context, req *adapter.Request) (*adapter.Embedding, error) {
	start := time.Now()
	reqID := middleware.FromContext(ctx)

	slog.InfoContext(ctx, "embedding request received",
		slog.String("component", "openai-handler"),
		slog.String("action", "embedding_request_received"),
		slog.String("model", req.Model),
		slog.Int("inputs_count", len(req.Input)),
		slog.String("request_id", reqID),
	)

	// Use "embedding" capability (dedicated adapter capability, not chat)
	emb, err := gp.failover.Embed(ctx, "embedding", *req)
	latencyMs := time.Since(start).Milliseconds()

	if err != nil {
		provider := "unknown"
		status := 500
		if provErr, ok := err.(*adapter.ProviderError); ok {
			provider = provErr.Provider
			status = provErr.Status
		}

		// HU-060: registrar métrica de error
		if gp.metricsStore != nil {
			gp.metricsStore.Record(metrics.RequestMetric{
				Provider:  provider,
				Model:     req.Model,
				LatencyMs: latencyMs,
				Status:    status,
			})
		}

		logFields := []any{
			slog.String("component", "openai-handler"),
			slog.String("action", "embedding_request_failed"),
			slog.String("model", req.Model),
			slog.String("request_id", reqID),
			slog.Int64("latency_ms", latencyMs),
		}
		if provErr, ok := err.(*adapter.ProviderError); ok {
			logFields = append(logFields, slog.Int("provider_status", provErr.Status))
			logFields = append(logFields, slog.String("provider", provErr.Provider))
		} else {
			logFields = append(logFields, slog.String("error_type", "internal"))
		}
		slog.ErrorContext(ctx, "embedding request failed", logFields...)
		return nil, err
	}

	// HU-060: registrar métrica de éxito
	if gp.metricsStore != nil {
		gp.metricsStore.Record(metrics.RequestMetric{
			Provider:  "unknown", // ponytail: failover no retorna provider usado
			Model:     req.Model,
			LatencyMs: latencyMs,
			Status:    200,
		})
	}

	slog.InfoContext(ctx, "embedding request completed",
		slog.String("component", "openai-handler"),
		slog.String("action", "embedding_request_completed"),
		slog.String("model", req.Model),
		slog.Int("vectors_count", len(emb.Vectors)),
		slog.String("request_id", reqID),
		slog.Int64("latency_ms", latencyMs),
	)
	return &emb, nil
}
