package quota

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// createLearnedQuotaTableSQL es la migración mínima (idempotente, CREATE TABLE
// IF NOT EXISTS) del schema de cuota aprendida, ejecutada al abrir la
// conexión. HU-EVO-008: PostgreSQL 16+ es el store declarado por el PRD
// técnico (docs/13-tech-prd/api-llm-gateway.md) para Quotas.
const createLearnedQuotaTableSQL = `
CREATE TABLE IF NOT EXISTS learned_quota (
	provider_id TEXT NOT NULL,
	model_id    TEXT NOT NULL,
	quota_limit BIGINT NOT NULL,
	remaining   BIGINT NOT NULL,
	reset_at    TIMESTAMPTZ,
	updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
	PRIMARY KEY (provider_id, model_id)
);`

// ErrPersistQueueFull se devuelve cuando el buffer de Enqueue está lleno; el
// caller (Manager.LearnFromHeaders) ya trata cualquier error de Enqueue como
// no crítico (fire-and-forget, AC1/AC3 HU-EVO-008: nunca bloquea el path
// crítico ni aborta el request).
var ErrPersistQueueFull = errors.New("quota: cola de persistencia llena, job descartado")

// PostgresPersister implementa Persister persistiendo asíncronamente a
// PostgreSQL vía un worker en background que consume de un canal con buffer
// (HU-EVO-008 AC1: Enqueue no bloquea el path crítico; AC3: DB caída no
// aborta el request, solo loguea warning y sigue).
type PostgresPersister struct {
	db     *sql.DB
	jobs   chan PersistJob
	done   chan struct{}
	closed chan struct{}
}

// NewPostgresPersister abre la conexión, verifica con Ping y aplica la
// migración mínima idempotente. queueSize es la capacidad del buffer async;
// si se llena, Enqueue devuelve ErrPersistQueueFull en vez de bloquear.
func NewPostgresPersister(dsn string, queueSize int) (*PostgresPersister, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.ExecContext(ctx, createLearnedQuotaTableSQL); err != nil {
		_ = db.Close()
		return nil, err
	}

	if queueSize <= 0 {
		queueSize = 1000
	}
	p := &PostgresPersister{
		db:     db,
		jobs:   make(chan PersistJob, queueSize),
		done:   make(chan struct{}),
		closed: make(chan struct{}),
	}
	go p.worker()
	return p, nil
}

// Enqueue encola job de forma no bloqueante (<5ms garantizado por el select
// con default: si el buffer está lleno, descarta y reporta el error en vez de
// esperar).
func (p *PostgresPersister) Enqueue(job PersistJob) error {
	select {
	case p.jobs <- job:
		return nil
	default:
		return ErrPersistQueueFull
	}
}

func (p *PostgresPersister) worker() {
	defer close(p.closed)
	for {
		select {
		case job, ok := <-p.jobs:
			if !ok {
				return
			}
			p.write(job)
		case <-p.done:
			// Drena lo que quede en el buffer antes de salir (best-effort).
			for {
				select {
				case job := <-p.jobs:
					p.write(job)
				default:
					return
				}
			}
		}
	}
}

func (p *PostgresPersister) write(job PersistJob) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var resetAt any
	if job.Quota.ResetAt != nil {
		resetAt = *job.Quota.ResetAt
	}

	_, err := p.db.ExecContext(ctx, `
		INSERT INTO learned_quota (provider_id, model_id, quota_limit, remaining, reset_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, now())
		ON CONFLICT (provider_id, model_id) DO UPDATE SET
			quota_limit = EXCLUDED.quota_limit,
			remaining   = EXCLUDED.remaining,
			reset_at    = EXCLUDED.reset_at,
			updated_at  = now()
	`, job.ProviderID, job.ModelID, job.Quota.Limit, job.Quota.Remaining, resetAt)
	if err != nil {
		// AC3: DB caída/error no aborta el request; solo loguea warning.
		log.Printf("WARN quota: postgres persister write failed provider=%q model=%q: %v", job.ProviderID, job.ModelID, err)
	}
}

// LoadRemaining lee el remaining persistido por proveedor, para restaurar en
// boot (linaje HU-EVO-005 AC5: RestoreRemaining tiene precedencia sobre
// quota_hint). Si un proveedor tiene múltiples modelos persistidos, toma el
// valor más reciente por updated_at.
func (p *PostgresPersister) LoadRemaining(ctx context.Context) (map[string]int64, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT DISTINCT ON (provider_id) provider_id, remaining
		FROM learned_quota
		ORDER BY provider_id, updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]int64)
	for rows.Next() {
		var providerID string
		var remaining int64
		if err := rows.Scan(&providerID, &remaining); err != nil {
			return nil, err
		}
		out[providerID] = remaining
	}
	return out, rows.Err()
}

// Close detiene el worker (drenando el buffer pendiente, best-effort) y
// cierra la conexión a la DB.
func (p *PostgresPersister) Close() error {
	close(p.done)
	<-p.closed
	return p.db.Close()
}
