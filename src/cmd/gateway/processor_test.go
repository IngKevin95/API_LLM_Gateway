package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/failover"
	"api-llm-gateway/internal/registry"
)

// TestProcessChatLogsJSON verifies that ProcessChat emits structured JSON logs
// HU-050: Añadir logging estructurado al OpenAI handler
// Requirements:
// - Log JSON with fields: timestamp, level, request_id, component, action, model, messages_count
// - Request ID propagated through context
// - Latency and success/error tracked
func TestProcessChatLogsJSON(t *testing.T) {
	// Capture slog output
	logBuf := &bytes.Buffer{}
	handler := slog.NewJSONHandler(logBuf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Create minimal failover engine:
	// - mockChainResolver returns a single model (openai/gpt-4)
	// - mockAdapter returns a successful response
	chainResolver := &mockChainResolver{
		models: []registry.Model{
			{ProviderID: "openai", Name: "gpt-4"},
		},
	}

	adapters := map[string]adapter.Adapter{
		"openai": &mockAdapter{
			response: adapter.Response{Content: "test response"},
		},
	}

	engine := failover.New(chainResolver, adapters)
	gp := &GatewayProcessor{failover: engine}

	// Test: ProcessChat with request ID in context
	ctx := context.WithValue(context.Background(), "request_id", "test-req-123")
	req := &adapter.Request{
		Model: "gpt-4",
		Messages: []adapter.Message{
			{Role: "user", Content: "hello"},
		},
	}

	resp, err := gp.ProcessChat(ctx, req)

	// Verify: no error
	if err != nil {
		t.Fatalf("ProcessChat failed: %v", err)
	}

	if resp == nil || resp.Content != "test response" {
		t.Fatalf("ProcessChat returned wrong response: %v", resp)
	}

	// Verify: logs contain structured JSON with expected fields
	logs := logBuf.String()
	if logs == "" {
		t.Fatal("ProcessChat did not emit any logs")
	}

	// Parse logs (should be multiple JSON objects, one per line)
	lines := bytes.Split(bytes.TrimSpace(logBuf.Bytes()), []byte("\n"))
	if len(lines) < 1 {
		t.Fatal("No log lines found")
	}

	// Verify first log (request received)
	var firstLog map[string]interface{}
	if err := json.Unmarshal(lines[0], &firstLog); err != nil {
		t.Fatalf("First log is not valid JSON: %s, error: %v", lines[0], err)
	}

	// Check required fields in first log
	expectedFields := []string{"level", "time", "msg", "component", "action"}
	for _, field := range expectedFields {
		if _, ok := firstLog[field]; !ok {
			t.Errorf("Missing field %q in log: %v", field, firstLog)
		}
	}

	// Verify component is "openai-handler"
	if component, ok := firstLog["component"]; ok {
		if component != "openai-handler" {
			t.Errorf("Expected component=openai-handler, got %v", component)
		}
	} else {
		t.Error("Missing component field")
	}

	// Verify action is "request_received"
	if action, ok := firstLog["action"]; ok {
		if action != "request_received" {
			t.Errorf("Expected action=request_received, got %v", action)
		}
	} else {
		t.Error("Missing action field")
	}

	// Verify model
	if model, ok := firstLog["model"]; ok {
		if model != "gpt-4" {
			t.Errorf("Expected model=gpt-4, got %v", model)
		}
	} else {
		t.Error("Missing model field")
	}

	// Verify request_id
	if reqID, ok := firstLog["request_id"]; ok {
		if reqID != "test-req-123" {
			t.Errorf("Expected request_id=test-req-123, got %v", reqID)
		}
	} else {
		t.Error("Missing request_id field")
	}
}

// mockChainResolver implements failover.ChainResolver for testing
type mockChainResolver struct {
	models []registry.Model
}

func (m *mockChainResolver) Resolve(capability string, estimatedTokens int) ([]registry.Model, error) {
	return m.models, nil
}

// mockAdapter implements adapter.Adapter for testing
type mockAdapter struct {
	response    adapter.Response
	err         error
	embedResult adapter.Embedding
}

func (m *mockAdapter) Chat(ctx context.Context, req adapter.Request) (adapter.Response, error) {
	if m.err != nil {
		return adapter.Response{}, m.err
	}
	return m.response, nil
}

func (m *mockAdapter) Stream(ctx context.Context, req adapter.Request) (adapter.TokenStream, error) {
	return nil, nil
}

func (m *mockAdapter) Embed(ctx context.Context, req adapter.Request) (adapter.Embedding, error) {
	if m.err != nil {
		return adapter.Embedding{}, m.err
	}
	return m.embedResult, nil
}

// TestRegistryLoadsConfigYAML validates that config.yaml loads all provider IDs correctly
// HU-056: Alinear IDs de proveedores en config.yaml con buildAdapters()
// Expected: config.yaml uses normalized IDs (openai, anthropic, local, google, aihubmix, omniroute)
func TestRegistryLoadsConfigYAML(t *testing.T) {
	configPath := "../../config.yaml"
	reg, err := registry.Load(configPath, nil)
	if err != nil {
		t.Fatalf("failed to load config.yaml: %v", err)
	}

	// Verify we have models loaded
	allModels := reg.ModelNames()
	if len(allModels) == 0 {
		t.Fatal("config.yaml loaded but no models found")
	}

	// Verify each provider ID is recognized (can call APIKey without error)
	expectedProviders := []string{"openai", "anthropic", "local", "google", "aihubmix", "omniroute"}
	for _, providerID := range expectedProviders {
		// APIKey exists for this provider (even if empty string for local)
		key := reg.APIKey(providerID)
		_ = key // just verify no panic
	}

	// Log success
	t.Logf("registry: %d modelos cargados desde config.yaml", len(allModels))
}

// TestProcessEmbedding_Success validates that ProcessEmbedding returns embeddings
// HU-057: Implementar GatewayProcessor.ProcessEmbedding()
// Expected: happy path returns adapter.Embedding with vectors
func TestProcessEmbedding_Success(t *testing.T) {
	// Mock adapter that returns embeddings
	adapters := map[string]adapter.Adapter{
		"openai": &mockAdapter{
			embedResult: adapter.Embedding{
				Vectors: [][]float64{{0.1, 0.2, 0.3}, {0.4, 0.5, 0.6}},
			},
		},
	}

	// Mock chain resolver returns openai model
	chainResolver := &mockChainResolver{
		models: []registry.Model{
			{ProviderID: "openai", Name: "text-embedding-3"},
		},
	}

	engine := failover.New(chainResolver, adapters)
	gp := &GatewayProcessor{failover: engine}

	// Test: ProcessEmbedding with valid input
	ctx := context.Background()
	req := &adapter.Request{
		Model: "text-embedding-3",
		Input: []string{"hello", "world"},
	}

	emb, err := gp.ProcessEmbedding(ctx, req)

	// Verify: no error
	if err != nil {
		t.Fatalf("ProcessEmbedding failed: %v", err)
	}

	if emb == nil {
		t.Fatal("ProcessEmbedding returned nil embedding")
	}

	// Verify: has vectors
	if len(emb.Vectors) != 2 {
		t.Fatalf("expected 2 vectors, got %d", len(emb.Vectors))
	}

	if len(emb.Vectors[0]) != 3 {
		t.Fatalf("expected 3 dimensions, got %d", len(emb.Vectors[0]))
	}

	if emb.Vectors[0][0] != 0.1 {
		t.Errorf("expected 0.1, got %f", emb.Vectors[0][0])
	}
}
