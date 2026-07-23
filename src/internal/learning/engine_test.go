package learning_test

import (
	"testing"

	"api-llm-gateway/internal/learning"
)

type mockMetricProvider struct {
	data []learning.ModelStats
}

func (m *mockMetricProvider) GetRecentStats() ([]learning.ModelStats, error) {
	return m.data, nil
}

type mockWeightUpdater struct {
	weights map[string]float64
}

func (m *mockWeightUpdater) UpdateWeights(weights map[string]float64) error {
	m.weights = weights
	return nil
}

func TestEngine_AdjustsWeights(t *testing.T) {
	mp := &mockMetricProvider{
		data: []learning.ModelStats{
			{Model: "gpt-4", Samples: 150, AvgLatency: 100, SuccessRate: 0.99},
			{Model: "claude-3", Samples: 150, AvgLatency: 500, SuccessRate: 0.99},
		},
	}
	updater := &mockWeightUpdater{}
	eng := learning.NewEngine(mp, updater)

	err := eng.Evaluate()
	if err != nil {
		t.Fatal(err)
	}

	if updater.weights == nil {
		t.Fatal("Expected weights to be updated")
	}
	if updater.weights["gpt-4"] <= updater.weights["claude-3"] {
		t.Errorf("Expected gpt-4 to have higher weight, got %f vs %f", updater.weights["gpt-4"], updater.weights["claude-3"])
	}
}

func TestEngine_InsufficientData(t *testing.T) {
	mp := &mockMetricProvider{
		data: []learning.ModelStats{
			{Model: "gpt-4", Samples: 50, AvgLatency: 100, SuccessRate: 0.99}, // < 100
		},
	}
	updater := &mockWeightUpdater{}
	eng := learning.NewEngine(mp, updater)

	eng.Evaluate()

	if updater.weights != nil {
		t.Errorf("Expected no weight update for insufficient data")
	}
}

func TestEngine_Guardrails(t *testing.T) {
	mp := &mockMetricProvider{
		data: []learning.ModelStats{
			{Model: "gpt-4", Samples: 150, AvgLatency: 10, SuccessRate: 0.99}, // Super fast
			{Model: "claude-3", Samples: 150, AvgLatency: 5000, SuccessRate: 0.99}, // Super slow
		},
	}
	updater := &mockWeightUpdater{}
	eng := learning.NewEngine(mp, updater)

	eng.Evaluate()

	if updater.weights["gpt-4"] > 0.90 { // Maximum cap should be 90% or something (Guardrail)
		t.Errorf("Expected weight capped at guardrail, got %f", updater.weights["gpt-4"])
	}
}

func TestEngine_Rollback(t *testing.T) {
	mp := &mockMetricProvider{
		data: []learning.ModelStats{
			{Model: "gpt-4", Samples: 150, AvgLatency: 100, SuccessRate: 0.85}, // Degraded (< 0.90)
		},
	}
	updater := &mockWeightUpdater{weights: map[string]float64{"gpt-4": 0.8}}
	eng := learning.NewEngine(mp, updater)

	// Guardamos el snapshot original (simulando estado anterior)
	eng.SetPreviousWeights(map[string]float64{"gpt-4": 0.5, "claude-3": 0.5})

	eng.Evaluate()

	if updater.weights["gpt-4"] != 0.5 {
		t.Errorf("Expected rollback to 0.5, got %f", updater.weights["gpt-4"])
	}
}
