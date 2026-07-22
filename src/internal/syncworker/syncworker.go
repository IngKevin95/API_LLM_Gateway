// Package syncworker persiste eventos de auditoría de forma asincronista:
// consume un channel, batchea por tamaño/tiempo, cifra con KMS Envelope y
// escribe al Store, con manejo de backpressure para no bloquear al handler.
package syncworker

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/IngKevin95/API_LLM_Gateway/internal/audit"
	"github.com/IngKevin95/API_LLM_Gateway/internal/wal"
)

// ErrDropped indica que un evento de baja prioridad se descartó por saturación.
var ErrDropped = errors.New("syncworker: evento de baja prioridad dropeado por backpressure")

// ErrBackpressure indica saturación sin poder encolar un evento normal.
var ErrBackpressure = errors.New("syncworker: backpressure, no se pudo encolar")

// Config parametriza el batching y el buffer.
type Config struct {
	BatchSize     int
	FlushInterval time.Duration
	ChanBuffer    int
}

// Worker es el writer asincronista.
type Worker struct {
	store audit.Store
	enc   audit.Encryptor
	wal   *wal.WAL
	cfg   Config
	ch    chan audit.Event
}

// New construye un Worker con sus dependencias inyectadas.
func New(store audit.Store, enc audit.Encryptor, w *wal.WAL, cfg Config) *Worker {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 1000
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = time.Second
	}
	if cfg.ChanBuffer <= 0 {
		cfg.ChanBuffer = 10000
	}
	return &Worker{store: store, enc: enc, wal: w, cfg: cfg, ch: make(chan audit.Event, cfg.ChanBuffer)}
}

// Enqueue encola un evento sin bloquear al productor. Ante saturación: dropea
// los de baja prioridad (Priority<0) y devuelve backpressure para los normales.
func (wk *Worker) Enqueue(e audit.Event) error {
	select {
	case wk.ch <- e:
		return nil
	default:
	}
	// Channel lleno: baja prioridad se dropea; normal reintenta brevemente.
	if e.Priority < 0 {
		return ErrDropped
	}
	for i := 0; i < 3; i++ {
		time.Sleep(time.Duration(i+1) * time.Millisecond) // jitter creciente, acotado
		select {
		case wk.ch <- e:
			return nil
		default:
		}
	}
	return ErrBackpressure
}

// Run consume el channel y flushea por tamaño o por intervalo hasta ctx.Done.
func (wk *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(wk.cfg.FlushInterval)
	defer ticker.Stop()
	buf := make([]audit.Event, 0, wk.cfg.BatchSize)

	flush := func() {
		if len(buf) == 0 {
			return
		}
		wk.flush(ctx, buf)
		buf = buf[:0]
	}
	for {
		select {
		case <-ctx.Done():
			flush() // flush final del buffer en memoria
			return
		case e := <-wk.ch:
			buf = append(buf, e)
			if len(buf) >= wk.cfg.BatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

// flush escribe el batch al WAL (durabilidad), lo cifra y lo persiste al Store.
func (wk *Worker) flush(ctx context.Context, batch []audit.Event) {
	enc := make([]audit.EncryptedEvent, 0, len(batch))
	for _, e := range batch {
		if err := wk.wal.Append(e); err != nil {
			log.Printf("WARN syncworker: fallo al escribir WAL: %v", err)
		}
		sealed, err := wk.enc.Seal(e) // KMS Envelope antes de tocar el Store
		if err != nil {
			log.Printf("WARN syncworker: fallo al cifrar evento: %v", err)
			continue
		}
		enc = append(enc, sealed)
	}
	if len(enc) == 0 {
		return
	}
	if err := wk.store.Write(ctx, enc); err != nil {
		// La DB no respondió: los eventos ya están en el WAL para recovery.
		log.Printf("WARN syncworker: Store.Write falló (%v); eventos quedan en WAL", err)
	}
}

// Flush drena el channel pendiente y persiste todo (usado por el graceful shutdown).
func (wk *Worker) Flush(ctx context.Context) {
	buf := make([]audit.Event, 0, wk.cfg.BatchSize)
	for {
		select {
		case e := <-wk.ch:
			buf = append(buf, e)
		default:
			wk.flush(ctx, buf)
			return
		}
	}
}
