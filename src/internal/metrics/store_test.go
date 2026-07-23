package metrics

import (
	"context"
	"testing"
)

func TestInMemoryStore_RollingWindow(t *testing.T) {
	store := NewInMemoryStore(2)

	store.Record(RequestMetric{Provider: "openai", Model: "gpt-4", LatencyMs: 100, Status: 200})
	store.Record(RequestMetric{Provider: "anthropic", Model: "claude-3", LatencyMs: 200, Status: 200})
	store.Record(RequestMetric{Provider: "google", Model: "gemini", LatencyMs: 300, Status: 200}) // should overwrite the first one

	if len(store.metrics) != 2 {
		t.Errorf("Expected exactly 2 metrics in store, got %d", len(store.metrics))
	}

	metrics, err := store.GetMetrics(context.Background(), "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(metrics) != 2 {
		t.Errorf("Expected 2 aggregated metrics, got %d", len(metrics))
	}

	// Because we sort by Provider + Model: anthropic comes first, google second. openai was evicted.
	if metrics[0].Provider != "anthropic" {
		t.Errorf("Expected anthropic at index 0, got %s", metrics[0].Provider)
	}
	if metrics[1].Provider != "google" {
		t.Errorf("Expected google at index 1, got %s", metrics[1].Provider)
	}
}

func TestInMemoryStore_Aggregation(t *testing.T) {
	store := NewInMemoryStore(100)

	store.Record(RequestMetric{Provider: "openai", Model: "gpt-4", LatencyMs: 100, Status: 200, Tokens: 10, Cost: 0.1})
	store.Record(RequestMetric{Provider: "openai", Model: "gpt-4", LatencyMs: 300, Status: 500, Tokens: 10, Cost: 0.1})
	store.Record(RequestMetric{Provider: "openai", Model: "gpt-4", LatencyMs: 200, Status: 200, Tokens: 10, Cost: 0.1})

	metrics, err := store.GetMetrics(context.Background(), "openai")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(metrics) != 1 {
		t.Fatalf("Expected 1 aggregated metric for openai/gpt-4, got %d", len(metrics))
	}

	m := metrics[0]
	if m.Tokens != 30 {
		t.Errorf("Expected 30 tokens, got %d", m.Tokens)
	}
	// Success rate: 2/3 = 66.66%
	if m.SuccessRate < 66.0 || m.SuccessRate > 67.0 {
		t.Errorf("Expected success rate ~66.6%%, got %f", m.SuccessRate)
	}
	// Avg latency: (100+300+200)/3 = 200
	if m.AvgLatencyMs != 200.0 {
		t.Errorf("Expected avg latency 200, got %f", m.AvgLatencyMs)
	}
	// P95 latency: array [100, 200, 300], idx = 3 * 95 / 100 = 2 => 300
	if m.P95LatencyMs != 300.0 {
		t.Errorf("Expected p95 latency 300, got %f", m.P95LatencyMs)
	}
}
