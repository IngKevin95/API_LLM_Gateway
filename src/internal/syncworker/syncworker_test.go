package syncworker_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/IngKevin95/API_LLM_Gateway/internal/audit"
	"github.com/IngKevin95/API_LLM_Gateway/internal/syncworker"
	"github.com/IngKevin95/API_LLM_Gateway/internal/wal"
)

type mockStore struct {
	mu      sync.Mutex
	events  int
	batches int
}

func (m *mockStore) Write(_ context.Context, batch []audit.EncryptedEvent) error {
	m.mu.Lock()
	m.events += len(batch)
	m.batches++
	m.mu.Unlock()
	return nil
}
func (m *mockStore) count() int { m.mu.Lock(); defer m.mu.Unlock(); return m.events }

type mockEncryptor struct{ seals int32 }

func (m *mockEncryptor) Seal(e audit.Event) (audit.EncryptedEvent, error) {
	atomic.AddInt32(&m.seals, 1)
	return audit.EncryptedEvent{Ciphertext: []byte("enc:" + e.ID)}, nil
}

func newWorker(t *testing.T, store audit.Store, enc audit.Encryptor, cfg syncworker.Config) *syncworker.Worker {
	t.Helper()
	w, err := wal.New(t.TempDir(), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { w.Close() })
	return syncworker.New(store, enc, w, cfg)
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatal("condición no alcanzada dentro del timeout")
}

// HU-038 AC1 — Happy: batching por tamaño → Store recibe el lote, sin bloquear al productor.
func TestWorker_BatchesBySize(t *testing.T) {
	store := &mockStore{}
	enc := &mockEncryptor{}
	w := newWorker(t, store, enc, syncworker.Config{BatchSize: 5, FlushInterval: time.Hour, ChanBuffer: 100})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Run(ctx)

	for i := 0; i < 5; i++ {
		if err := w.Enqueue(audit.Event{ID: "e"}); err != nil {
			t.Fatalf("Enqueue no debe fallar dentro del buffer: %v", err)
		}
	}
	waitFor(t, func() bool { return store.count() == 5 })
}

// HU-038 AC1 — Happy: batching por tiempo (flush interval) aunque no se llene.
func TestWorker_BatchesByTime(t *testing.T) {
	store := &mockStore{}
	w := newWorker(t, store, &mockEncryptor{}, syncworker.Config{BatchSize: 1000, FlushInterval: 20 * time.Millisecond, ChanBuffer: 100})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Run(ctx)
	_ = w.Enqueue(audit.Event{ID: "solo"})
	waitFor(t, func() bool { return store.count() == 1 })
}

// journey_smoke SS2: worker real + WAL + Store/Encryptor mock; batch cifrado persiste.
// HU-038 AC4 — KMS Envelope: cada evento se cifra vía Encryptor antes de persistir.
func TestWorker_EncryptsBeforeStore(t *testing.T) {
	store := &mockStore{}
	enc := &mockEncryptor{}
	w := newWorker(t, store, enc, syncworker.Config{BatchSize: 3, FlushInterval: time.Hour, ChanBuffer: 100})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Run(ctx)
	for i := 0; i < 3; i++ {
		_ = w.Enqueue(audit.Event{ID: "e"})
	}
	waitFor(t, func() bool { return store.count() == 3 })
	if got := atomic.LoadInt32(&enc.seals); got != 3 {
		t.Errorf("esperaba 3 Seal (cifrado por evento), obtuve %d", got)
	}
}

// HU-038 AC2 — Backpressure: channel saturado → dropea baja prioridad, sin bloquear.
func TestWorker_BackpressureDropsLowPriority(t *testing.T) {
	store := &mockStore{}
	// buffer 1 y SIN consumidor corriendo → se satura enseguida.
	w := newWorker(t, store, &mockEncryptor{}, syncworker.Config{BatchSize: 10, FlushInterval: time.Hour, ChanBuffer: 1})

	dropped := 0
	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			if err := w.Enqueue(audit.Event{ID: "low", Priority: -1}); errors.Is(err, syncworker.ErrDropped) {
				dropped++
			}
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Enqueue bloqueó al productor (debe ser no bloqueante)")
	}
	if dropped == 0 {
		t.Error("esperaba que se dropearan eventos de baja prioridad ante saturación")
	}
}
