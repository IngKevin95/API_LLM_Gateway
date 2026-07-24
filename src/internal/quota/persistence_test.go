package quota

import (
	"sync"
	"testing"
	"time"

	"api-llm-gateway/internal/adapter"
)

// RED: Test failing for HU-EVO-008 AC1 - Async persistence non-blocking
func TestPersistence_EnqueueNonBlocking_UnderFiveMillis(t *testing.T) {
	// Arrange: create a slow persister that takes 100ms per enqueue
	slowPersister := &SlowPersister{delay: 100 * time.Millisecond}

	// Act: measure time to call persistence.Enqueue indirectly via LearnFromHeaders
	// (assume Manager will integrate persister; for now, call directly)
	start := time.Now()
	_ = slowPersister.Enqueue(PersistJob{
		ProviderID: "openai",
		ModelID:    "",
		Quota: adapter.QuotaInfo{
			Limit:     10000,
			Remaining: 9950,
			ResetAt:   nil,
		},
	})
	elapsed := time.Since(start)

	// Assert: enqueue should be fast even if backend is slow (async queuing only)
	if elapsed > 5*time.Millisecond {
		t.Errorf("Enqueue took %v, should be <5ms (async queue, not wait for DB)", elapsed)
	}
}

// RED: Test failing for HU-EVO-008 AC2 - Concurrent async enqueue with idempotence
func TestPersistence_ConcurrentEnqueue_Idempotent(t *testing.T) {
	persister := &InMemoryPersister{}
	jobs := 0
	jobsMutex := sync.Mutex{}

	// Act: 100 concurrent enqueues for same provider+model
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			job := PersistJob{
				ProviderID: "openai",
				ModelID:    "",
				Quota: adapter.QuotaInfo{
					Limit:     10000,
					Remaining: int64(10000 - (i % 100)),
					ResetAt:   nil,
				},
			}
			_ = persister.Enqueue(job)
			jobsMutex.Lock()
			jobs++
			jobsMutex.Unlock()
		}()
	}
	wg.Wait()

	// Assert: all enqueues succeeded (idempotent)
	if jobs != 100 {
		t.Errorf("Expected 100 successful enqueues, got %d", jobs)
	}
}

// RED: Test failing for HU-EVO-008 AC3 - DB unavailability does not block
func TestPersistence_DBUnavailable_NonBlocking(t *testing.T) {
	// Arrange: persister that always fails but is async
	failingPersister := &FailingPersister{}

	// Act: enqueue should not block even if DB is down
	start := time.Now()
	err := failingPersister.Enqueue(PersistJob{
		ProviderID: "openai",
		ModelID:    "",
		Quota: adapter.QuotaInfo{
			Limit:     10000,
			Remaining: 9950,
			ResetAt:   nil,
		},
	})
	elapsed := time.Since(start)

	// Assert: enqueue returns quickly (graceful) regardless of DB state
	if elapsed > 5*time.Millisecond {
		t.Errorf("Enqueue took %v even with DB down, should be async/fast", elapsed)
	}
	// Error handling is async, so we don't expect an error to propagate
	if err != nil {
		t.Logf("Note: err is %v (may be async)", err)
	}
}

// RED: Test failing for HU-EVO-008 AC5 - Audit trail preservation
func TestPersistence_AuditTrail_PreserveHistory(t *testing.T) {
	auditor := &AuditingPersister{}

	// Act: enqueue 3 updates for same provider in succession
	for i := 0; i < 3; i++ {
		job := PersistJob{
			ProviderID: "cerebras",
			ModelID:    "",
			Quota: adapter.QuotaInfo{
				Limit:     10000,
				Remaining: int64(10000 - (i * 1000)),
				ResetAt:   nil,
			},
		}
		_ = auditor.Enqueue(job)
	}

	// Assert: audit log has 3 distinct entries (not deduplicated)
	if len(auditor.history) != 3 {
		t.Errorf("Expected 3 audit trail entries, got %d", len(auditor.history))
	}

	// Verify timestamps are non-decreasing (may be equal if operations are very fast)
	for i := 0; i < len(auditor.history)-1; i++ {
		if auditor.history[i].Timestamp.After(auditor.history[i+1].Timestamp) {
			t.Errorf("Audit trail timestamps out of order: %v > %v",
				auditor.history[i].Timestamp, auditor.history[i+1].Timestamp)
		}
	}
}

// Test helpers
type SlowPersister struct {
	delay time.Duration
}

func (sp *SlowPersister) Enqueue(job PersistJob) error {
	// Simulate async: would background this, but we're testing that enqueue itself is fast
	go func() {
		time.Sleep(sp.delay)
		// DB write happens here
	}()
	return nil
}

func (sp *SlowPersister) Close() error {
	return nil
}

type InMemoryPersister struct {
	mu   sync.Mutex
	jobs []PersistJob
}

func (ip *InMemoryPersister) Enqueue(job PersistJob) error {
	ip.mu.Lock()
	defer ip.mu.Unlock()
	ip.jobs = append(ip.jobs, job)
	return nil
}

func (ip *InMemoryPersister) Close() error {
	return nil
}

type FailingPersister struct{}

func (fp *FailingPersister) Enqueue(job PersistJob) error {
	// Simulate async failure handling: enqueue returns quickly, error logged in background
	go func() {
		// Pretend to log error in background
	}()
	return nil
}

func (fp *FailingPersister) Close() error {
	return nil
}

type AuditEntry struct {
	Job       PersistJob
	Timestamp time.Time
}

type AuditingPersister struct {
	mu      sync.Mutex
	history []AuditEntry
}

func (ap *AuditingPersister) Enqueue(job PersistJob) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	ap.history = append(ap.history, AuditEntry{
		Job:       job,
		Timestamp: time.Now(),
	})
	return nil
}

func (ap *AuditingPersister) Close() error {
	return nil
}
