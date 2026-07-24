package quota

import (
	"sync"
	"testing"
	"time"

	"api-llm-gateway/internal/adapter"
)

// RED: Test failing for HU-EVO-007 AC1 - Atomic RAM update
func TestLearnFromHeaders_UpdatesRemaining_Atomically(t *testing.T) {
	m := NewInMemoryManager()
	m.SetLimit("openai", Consumption{Tokens: 10000})

	// Arrange: quota learned from header
	quotaInfo := adapter.QuotaInfo{
		Limit:     10000,
		Remaining: 9950,
		ResetAt:   nil,
	}

	// Act: learn from headers
	m.LearnFromHeaders("openai", "", quotaInfo)

	// Assert: Remaining reflects learned value immediately
	remaining := m.Remaining("openai", "")
	if remaining != 9950 {
		t.Errorf("Expected Remaining=9950, got %d", remaining)
	}
}

// RED: Test failing for HU-EVO-007 AC2 - Learned limit > hint
func TestLearnFromHeaders_LearnedLimitPrecedence(t *testing.T) {
	m := NewInMemoryManager()
	// Initialize with hint of 500000
	m.SetLimit("openai", Consumption{Tokens: 500000})

	// Act: learn higher limit from headers (1000000)
	quotaInfo := adapter.QuotaInfo{
		Limit:     1000000,
		Remaining: 999000,
		ResetAt:   nil,
	}
	m.LearnFromHeaders("openai", "", quotaInfo)

	// Assert: limit is now 1000000 (learned), not 500000 (hint)
	// We need a method to check limit
	remaining := m.Remaining("openai", "")
	if remaining != 999000 {
		t.Errorf("Expected learned limit applied, Remaining=999000, got %d", remaining)
	}
}

// RED: Test failing for HU-EVO-007 AC3 - Clamp negative to 0
func TestLearnFromHeaders_ClampNegativeToZero(t *testing.T) {
	m := NewInMemoryManager()
	m.SetLimit("openai", Consumption{Tokens: 10000})

	// Act: learn negative remaining (overshoot)
	quotaInfo := adapter.QuotaInfo{
		Limit:     10000,
		Remaining: -100,
		ResetAt:   nil,
	}
	m.LearnFromHeaders("openai", "", quotaInfo)

	// Assert: remaining clamped to 0
	remaining := m.Remaining("openai", "")
	if remaining != 0 {
		t.Errorf("Expected Remaining=0 (clamped), got %d", remaining)
	}
}

// RED: Test failing for HU-EVO-007 AC4 - Thread-safe concurrent updates
func TestLearnFromHeaders_ConcurrentUpdates_NoRaces(t *testing.T) {
	m := NewInMemoryManager()
	m.SetLimit("openai", Consumption{Tokens: 10000})

	var wg sync.WaitGroup
	numGoroutines := 10

	// Act: 10 concurrent goroutines each call LearnFromHeaders
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			quotaInfo := adapter.QuotaInfo{
				Limit:     10000,
				Remaining: int64(10000 - (idx * 100)),
				ResetAt:   nil,
			}
			m.LearnFromHeaders("openai", "", quotaInfo)
		}(i)
	}

	wg.Wait()

	// Assert: final state is consistent (last write wins)
	remaining := m.Remaining("openai", "")
	if remaining < 0 || remaining > 10000 {
		t.Errorf("Expected consistent final state, got Remaining=%d", remaining)
	}
}

// RED: Test failing for HU-EVO-007 AC5 - Reset detection
func TestLearnFromHeaders_DetectsQuotaReset(t *testing.T) {
	now := time.Now()
	m := NewInMemoryManagerWithClock(func() time.Time { return now })

	m.SetLimit("openai", Consumption{Tokens: 10000})

	// Arrange: first learn with exhausted state
	resetYesterday := now.Add(-24 * time.Hour)
	quotaInfo1 := adapter.QuotaInfo{
		Limit:     10000,
		Remaining: 0,
		ResetAt:   &resetYesterday,
	}
	m.LearnFromHeaders("openai", "", quotaInfo1)

	// Verify exhausted
	if m.Remaining("openai", "") != 0 {
		t.Fatalf("Setup failed: provider should be exhausted")
	}

	// Act: new reset time detected (24h in future = window reset)
	resetTomorrow := now.Add(24 * time.Hour)
	quotaInfo2 := adapter.QuotaInfo{
		Limit:     10000,
		Remaining: 10000,
		ResetAt:   &resetTomorrow,
	}
	m.LearnFromHeaders("openai", "", quotaInfo2)

	// Assert: provider reactivated with new quota
	remaining := m.Remaining("openai", "")
	if remaining != 10000 {
		t.Errorf("Expected provider reactivated with Remaining=10000, got %d", remaining)
	}
}
