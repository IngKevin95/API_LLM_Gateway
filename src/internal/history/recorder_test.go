package history_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"api-llm-gateway/internal/history"
)

type mockPersister struct {
	mu     sync.Mutex
	events []history.Record
	slow   bool
}

func (m *mockPersister) Save(ctx context.Context, rec history.Record) error {
	if m.slow {
		time.Sleep(50 * time.Millisecond)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, rec)
	return nil
}

func (m *mockPersister) UpdateFeedback(ctx context.Context, id string, feedback int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, e := range m.events {
		if e.ID == id {
			m.events[i].Feedback = feedback
			return nil
		}
	}
	return nil
}

func TestRecorder_RecordAsync(t *testing.T) {
	mp := &mockPersister{}
	rec := history.NewRecorder(mp, nil, 10)
	rec.Start()
	defer rec.Stop()

	// Should not block
	start := time.Now()
	rec.Record(history.Record{ID: "req-1", Model: "gpt-4", LatencyMs: 120, Success: true})
	if time.Since(start) > 10*time.Millisecond {
		t.Errorf("Record blocked!")
	}

	time.Sleep(10 * time.Millisecond) // Wait for async

	mp.mu.Lock()
	defer mp.mu.Unlock()
	if len(mp.events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(mp.events))
	}
	if mp.events[0].ID != "req-1" {
		t.Errorf("Expected req-1, got %s", mp.events[0].ID)
	}
}

func TestRecorder_Feedback(t *testing.T) {
	mp := &mockPersister{}
	rec := history.NewRecorder(mp, nil, 10)
	rec.Start()
	defer rec.Stop()

	rec.Record(history.Record{ID: "req-2", Model: "gpt-3.5", Success: true})
	time.Sleep(10 * time.Millisecond)

	rec.AddFeedback("req-2", 5) // 5 stars
	time.Sleep(10 * time.Millisecond)

	mp.mu.Lock()
	defer mp.mu.Unlock()
	if mp.events[0].Feedback != 5 {
		t.Errorf("Expected feedback 5, got %d", mp.events[0].Feedback)
	}
}

func TestRecorder_SlowBackend(t *testing.T) {
	mp := &mockPersister{slow: true}
	rec := history.NewRecorder(mp, nil, 2)
	rec.Start()
	defer rec.Stop()

	// Fill queue quickly, should not block main thread even if queue full/dropped
	start := time.Now()
	for i := 0; i < 5; i++ {
		rec.Record(history.Record{ID: "req-slow", Model: "gpt-4"})
	}
	if time.Since(start) > 40*time.Millisecond {
		t.Errorf("Record blocked on full queue!")
	}
}

type mockRedactor struct{}

func (m *mockRedactor) Redact(ctx context.Context, payload []byte) ([]byte, error) {
	return []byte("[REDACTED]"), nil
}

func TestRecorder_Redaction(t *testing.T) {
	mp := &mockPersister{}
	mr := &mockRedactor{}
	rec := history.NewRecorder(mp, mr, 10)
	rec.Start()
	defer rec.Stop()

	rec.Record(history.Record{ID: "req-pii", Payload: []byte("my email is test@test.com")})
	time.Sleep(10 * time.Millisecond)

	mp.mu.Lock()
	defer mp.mu.Unlock()
	if string(mp.events[0].Payload) != "[REDACTED]" {
		t.Errorf("Expected redacted payload, got %s", mp.events[0].Payload)
	}
}
