package cost

import (
	"context"
	"testing"
)

type mockModelFinder struct {
	costPer1M int
	exists    bool
}

func (m *mockModelFinder) CostPer1M(modelName string) (int, bool) {
	return m.costPer1M, m.exists
}

func TestTracker_Track(t *testing.T) {
	tests := []struct {
		name          string
		modelCost     int
		exists        bool
		prompt        int
		completion    int
		expectedCost  float64
	}{
		{
			name:         "Cost 10 per 1M, 500k prompt, 500k completion",
			modelCost:    10,
			exists:       true,
			prompt:       500000,
			completion:   500000,
			expectedCost: 10.0,
		},
		{
			name:         "Unknown model, zero cost",
			modelCost:    0,
			exists:       false,
			prompt:       100,
			completion:   100,
			expectedCost: 0.0,
		},
		{
			name:         "Zero cost model",
			modelCost:    0,
			exists:       true,
			prompt:       1000,
			completion:   1000,
			expectedCost: 0.0,
		},
		{
			name:         "Fractional cost",
			modelCost:    15, // $15 per 1M
			exists:       true,
			prompt:       1000, // 0.015
			completion:   2000, // 0.030 -> total 0.045
			expectedCost: 0.045,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var recordedRecord CostRecord
			sink := func(ctx context.Context, r CostRecord) error {
				recordedRecord = r
				return nil
			}

			finder := &mockModelFinder{costPer1M: tt.modelCost, exists: tt.exists}
			tracker := NewTracker(finder, sink)

			err := tracker.Track(context.Background(), CostRecord{
				AgentID:          "agent-1",
				ProviderID:       "provider-1",
				Model:            "test-model",
				PromptTokens:     tt.prompt,
				CompletionTokens: tt.completion,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if recordedRecord.Cost != tt.expectedCost {
				t.Errorf("expected cost %v, got %v", tt.expectedCost, recordedRecord.Cost)
			}
		})
	}
}
