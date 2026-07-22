package shutdown_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/IngKevin95/API_LLM_Gateway/internal/audit"
	"github.com/IngKevin95/API_LLM_Gateway/internal/shutdown"
	"github.com/IngKevin95/API_LLM_Gateway/internal/syncworker"
	"github.com/IngKevin95/API_LLM_Gateway/internal/wal"
)

type mockStore struct {
	mu     sync.Mutex
	events int
}

func (m *mockStore) Write(_ context.Context, b []audit.EncryptedEvent) error {
	m.mu.Lock()
	m.events += len(b)
	m.mu.Unlock()
	return nil
}
func (m *mockStore) count() int { m.mu.Lock(); defer m.mu.Unlock(); return m.events }

type enc struct{}

func (enc) Seal(e audit.Event) (audit.EncryptedEvent, error) {
	return audit.EncryptedEvent{Ciphertext: []byte(e.ID)}, nil
}

func newWorker(t *testing.T, store audit.Store) (*syncworker.Worker, *wal.WAL) {
	t.Helper()
	w, err := wal.New(t.TempDir(), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { w.Close() })
	return syncworker.New(store, enc{}, w, syncworker.Config{BatchSize: 1000, FlushInterval: time.Hour, ChanBuffer: 100}), w
}

// HU-040 AC1 — Happy: drena las requests en vuelo antes de flushear (sin timeout).
func TestShutdown_DrainsInFlight(t *testing.T) {
	store := &mockStore{}
	worker, _ := newWorker(t, store)
	c := shutdown.New(worker, 500*time.Millisecond)

	c.Begin() // una request en vuelo
	go func() { time.Sleep(50 * time.Millisecond); c.End() }()

	timedOut := c.Shutdown(context.Background())
	if timedOut {
		t.Error("no debía haber timeout: la request drenó a tiempo")
	}
}

// HU-040 AC2 — Error: request que excede el timeout de drain → timedOut.
func TestShutdown_DrainTimeout(t *testing.T) {
	store := &mockStore{}
	worker, _ := newWorker(t, store)
	c := shutdown.New(worker, 80*time.Millisecond)

	c.Begin() // request que nunca termina
	timedOut := c.Shutdown(context.Background())
	if !timedOut {
		t.Error("esperaba timedOut cuando una request no drena a tiempo")
	}
}

// HU-040 AC3 — Edge: flush garantizado del buffer del Sync Worker antes de salir.
func TestShutdown_FlushesBuffer(t *testing.T) {
	store := &mockStore{}
	worker, _ := newWorker(t, store) // sin Run: los eventos quedan en el channel
	c := shutdown.New(worker, 500*time.Millisecond)

	for i := 0; i < 7; i++ {
		_ = worker.Enqueue(audit.Event{ID: "e"})
	}
	c.Shutdown(context.Background())
	if store.count() != 7 {
		t.Errorf("shutdown debe flushear los 7 eventos buffereados, persistió %d", store.count())
	}
}

// blockingStore respeta el ctx: se cuelga hasta que el ctx (timeout DB) vence.
type blockingStore struct{}

func (blockingStore) Write(ctx context.Context, _ []audit.EncryptedEvent) error {
	<-ctx.Done()
	return ctx.Err()
}

// HU-040 AC5 — Prevención de deadlock: con un Store lento, el flush acotado por
// ctx (timeout DB < shutdown) no cuelga el shutdown.
func TestShutdown_SlowStoreNoDeadlock(t *testing.T) {
	worker, _ := newWorker(t, blockingStore{})
	_ = worker.Enqueue(audit.Event{ID: "e"})
	c := shutdown.New(worker, 80*time.Millisecond)

	start := time.Now()
	c.Shutdown(context.Background())
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("shutdown se colgó con Store lento (deadlock): %s", elapsed)
	}
}

// journey_smoke SS3: ciclo shutdown(flush) → boot(recover del WAL) end-to-end.
// HU-040 AC4 — Recovery: boot procesa el WAL residual y lo replica al Store.
func TestRecover_ReplaysResidualWAL(t *testing.T) {
	store := &mockStore{}
	dir := t.TempDir()
	// Simula WAL residual de una corrida previa.
	w, _ := wal.New(dir, 1<<20)
	for i := 0; i < 4; i++ {
		_ = w.Append(audit.Event{ID: "residual"})
	}
	w.Close()

	fresh, _ := wal.New(dir, 1<<20)
	defer fresh.Close()
	if err := shutdown.Recover(context.Background(), fresh, enc{}, store); err != nil {
		t.Fatalf("Recover: %v", err)
	}
	if store.count() != 4 {
		t.Errorf("boot recovery debe replicar los 4 eventos residuales, obtuve %d", store.count())
	}
}
