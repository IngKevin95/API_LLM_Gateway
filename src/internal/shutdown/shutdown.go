// Package shutdown coordina el cierre ordenado: drena las requests en vuelo,
// garantiza el flush del Sync Worker, y recupera el WAL residual en el boot.
package shutdown

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"api-llm-gateway/internal/audit"
	"api-llm-gateway/internal/syncworker"
	"api-llm-gateway/internal/wal"
)

// Coordinator drena requests en vuelo y flushea el Sync Worker ante el cierre.
type Coordinator struct {
	worker       *syncworker.Worker
	drainTimeout time.Duration
	wg           sync.WaitGroup
	draining     atomic.Bool
}

// New crea un coordinador con el worker y el timeout de drenado.
func New(worker *syncworker.Worker, drainTimeout time.Duration) *Coordinator {
	return &Coordinator{worker: worker, drainTimeout: drainTimeout}
}

// Begin registra una request en vuelo; devuelve false si ya se está drenando
// (se rechazan requests nuevas durante el shutdown).
func (c *Coordinator) Begin() bool {
	if c.draining.Load() {
		return false
	}
	c.wg.Add(1)
	return true
}

// End marca el fin de una request en vuelo.
func (c *Coordinator) End() { c.wg.Done() }

// Shutdown deja de aceptar, espera a que drenen las requests en vuelo (con
// timeout) y garantiza el flush del buffer del Sync Worker. Devuelve true si el
// drenado excedió el timeout. El flush usa un ctx acotado para no colgarse si la
// DB no responde (timeout DB < timeout de shutdown → sin deadlock).
func (c *Coordinator) Shutdown(ctx context.Context) (timedOut bool) {
	c.draining.Store(true)

	drained := make(chan struct{})
	go func() { c.wg.Wait(); close(drained) }()
	select {
	case <-drained:
	case <-time.After(c.drainTimeout):
		timedOut = true
		log.Printf("WARN shutdown: timeout de drenado (%s); requests en vuelo abandonadas", c.drainTimeout)
	}

	flushCtx, cancel := context.WithTimeout(ctx, c.drainTimeout)
	defer cancel()
	c.worker.Flush(flushCtx) // a DB, o queda en WAL si la DB no responde
	return timedOut
}

// Recover replica el WAL residual al Store en el boot, antes de aceptar tráfico.
func Recover(ctx context.Context, w *wal.WAL, enc audit.Encryptor, store audit.Store) error {
	events, err := w.Recover()
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}
	batch := make([]audit.EncryptedEvent, 0, len(events))
	for _, e := range events {
		sealed, err := enc.Seal(e)
		if err != nil {
			log.Printf("WARN shutdown: fallo al cifrar evento de recovery: %v", err)
			continue
		}
		batch = append(batch, sealed)
	}
	return store.Write(ctx, batch)
}
