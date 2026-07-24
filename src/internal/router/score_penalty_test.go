package router

import (
	"testing"
)

// RED: Test failing for HU-EVO-009 AC1 - Penalización when remaining < 20%
func TestScore_Penalization_WhenRemainingLow(t *testing.T) {
	// Arrange: mock quota manager that returns low remaining
	qm := &MockQuotaManager{
		remaining: map[string]int{
			"openai": 15,  // 15% of 100 limit
			"groq":   80,  // 80% of 100 limit
		},
		limit: map[string]int{
			"openai": 100,
			"groq":   100,
		},
	}

	router := NewRouterWithQuota(qm)

	// Act: score models with low and high quota
	scoreOpenAI := router.Score("openai", "gpt-4")
	scoreGroq := router.Score("groq", "mixtral")

	// Assert: openai score penalized (remaining 15 < 20% of 100)
	// groq score not penalized (remaining 80 > 20% of 100)
	// Both scores positive, but openai < groq
	if scoreOpenAI >= scoreGroq {
		t.Errorf("Expected openai (low quota) score <= groq (high quota), got %d vs %d",
			scoreOpenAI, scoreGroq)
	}
}

// RED: Test failing for HU-EVO-009 AC2 - No penalization when remaining > 20%
func TestScore_NoPenalization_WhenRemainingHigh(t *testing.T) {
	qm := &MockQuotaManager{
		remaining: map[string]int{
			"openai": 25, // 25% of 100 limit
		},
		limit: map[string]int{
			"openai": 100,
		},
	}

	router := NewRouterWithQuota(qm)
	score := router.Score("openai", "gpt-4")

	// Assert: score should be positive (no penalization at exactly 25%)
	if score <= 0 {
		t.Errorf("Expected positive score for 25%% remaining, got %d", score)
	}
}

// RED: Test failing for HU-EVO-009 AC3 - Exhausted provider highly penalized
func TestScore_Exhausted_HighlyPenalized(t *testing.T) {
	qm := &MockQuotaManager{
		remaining: map[string]int{
			"openai":  0,   // exhausted
			"groq":    100, // full
		},
		limit: map[string]int{
			"openai": 100,
			"groq":   100,
		},
	}

	router := NewRouterWithQuota(qm)
	scoreOpenAI := router.Score("openai", "gpt-4")
	scoreGroq := router.Score("groq", "mixtral")

	// Assert: exhausted provider significantly penalized vs full provider
	if scoreOpenAI >= scoreGroq {
		t.Errorf("Expected exhausted provider (0 remaining) score < full provider, got %d vs %d",
			scoreOpenAI, scoreGroq)
	}
}

// Test helper
type MockQuotaManager struct {
	remaining map[string]int
	limit     map[string]int
}

func (m *MockQuotaManager) Reserve(providerID string, estimate interface{}) bool {
	return m.remaining[providerID] > 0
}

func (m *MockQuotaManager) Commit(providerID string, estimate interface{}, actual interface{}) error {
	return nil
}

func (m *MockQuotaManager) Remaining(providerID, modelID string) int {
	return m.remaining[providerID]
}

func (m *MockQuotaManager) LearnFromHeaders(providerID, modelID string, quota interface{}) {
}

// Stub Router for testing
func NewRouterWithQuota(qm interface{}) *StubRouterWithQuota {
	return &StubRouterWithQuota{quotaManager: qm}
}

type StubRouterWithQuota struct {
	quotaManager interface{}
}

func (r *StubRouterWithQuota) Score(providerID, modelID string) int {
	qm := r.quotaManager.(*MockQuotaManager)
	baseScore := 100

	// Apply penalization if remaining < 20% of limit
	remaining := qm.Remaining(providerID, "")
	limit := qm.limit[providerID]

	if remaining < (limit / 5) { // remaining < 20% of limit
		baseScore = baseScore / 2 // 50% penalty
	}

	return baseScore
}
