// Package alert implementa el worker que vigila la cuota aprendida por
// quota.Manager y genera/persiste alertas en PostgreSQL cuando la cuota
// remanente cae bajo un umbral configurable (HU-EVO-012).
package alert

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// QuotaSnapshotter es el subconjunto de quota.Manager que Alert Manager
// necesita (evita acoplar el paquete alert al paquete quota más allá de esta
// única función).
type QuotaSnapshotter interface {
	Snapshot() []QuotaEntry
}

// QuotaEntry espeja quota.Snapshot; lo puebla el adaptador cableado en
// cmd/gateway/main.go (mismo patrón que metrics.QuotaEntry).
type QuotaEntry struct {
	Provider  string
	Model     string
	Limit     int64
	Remaining int64
}

// Alert es una fila persistida en la tabla provider_alerts.
type Alert struct {
	ID         int64
	ProviderID string
	ModelID    string
	Severity   string // "warning" | "critical"
	Message    string
	AlertTime  time.Time
	UpdatedAt  time.Time
	ResolvedAt *time.Time
}

// DefaultThreshold es el umbral por defecto (10%) cuando
// GATEWAY_ALERT_THRESHOLD no está configurado (HU-EVO-012 AC1/AC4).
const DefaultThreshold = 0.10

const createProviderAlertsTableSQL = `
CREATE TABLE IF NOT EXISTS provider_alerts (
	id SERIAL PRIMARY KEY,
	provider_id TEXT NOT NULL,
	model_id TEXT NOT NULL,
	severity TEXT NOT NULL,
	message TEXT NOT NULL,
	alert_time TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	resolved_at TIMESTAMPTZ
);`

// Manager corre el ciclo de revisión de cuota (Check) periódicamente (Run) y
// persiste alertas con dedup por (provider_id, model_id) mientras estén
// activas (sin resolved_at).
type Manager struct {
	db        *sql.DB
	quota     QuotaSnapshotter
	threshold float64
	clock     func() time.Time
}

// NewManager abre/valida la tabla provider_alerts (migración idempotente) y
// devuelve un Manager listo para Run/Check. threshold<=0 aplica
// DefaultThreshold.
func NewManager(db *sql.DB, quota QuotaSnapshotter, threshold float64) (*Manager, error) {
	if threshold <= 0 {
		threshold = DefaultThreshold
	}
	if db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := db.ExecContext(ctx, createProviderAlertsTableSQL); err != nil {
			return nil, fmt.Errorf("alert: migración provider_alerts: %w", err)
		}
	}
	return &Manager{
		db:        db,
		quota:     quota,
		threshold: threshold,
		clock:     time.Now,
	}, nil
}

// SetClock permite inyectar un reloj determinista en tests.
func (m *Manager) SetClock(clock func() time.Time) { m.clock = clock }

// Run arranca un ticker propio en la goroutine del caller; bloquea hasta que
// ctx se cancela. Diseñado para lanzarse en `go m.Run(ctx, interval)` desde
// cmd/gateway/main.go.
func (m *Manager) Run(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.Check(ctx); err != nil {
				log.Printf("WARN alert: Check falló: %v", err)
			}
		}
	}
}

// Check evalúa el snapshot de cuota actual contra el umbral configurado y
// genera/actualiza alertas (HU-EVO-012 AC1/AC3/AC5). No aborta ante error
// individual de fila: loguea y continúa con el resto.
func (m *Manager) Check(ctx context.Context) error {
	if m.quota == nil {
		return nil
	}
	entries := m.quota.Snapshot()
	for _, q := range entries {
		severity, message := m.classify(q)
		if severity == "" {
			continue
		}
		if err := m.generateAlert(ctx, q, severity, message); err != nil {
			log.Printf("WARN alert: generateAlert provider=%q model=%q: %v", q.Provider, q.Model, err)
		}
	}
	return nil
}

// classify decide si (y con qué severidad) una entrada de cuota debe
// generar alerta: "critical" si remaining==0, "warning" si remaining <
// limit*threshold. Sin alerta (severity=="") en cualquier otro caso, incluido
// limit<=0 (proveedor sin límite conocido, no evaluable).
func (m *Manager) classify(q QuotaEntry) (severity, message string) {
	if q.Limit <= 0 {
		return "", ""
	}
	label := q.Provider
	if q.Model != "" {
		label = q.Provider + "/" + q.Model
	}
	if q.Remaining == 0 {
		return "critical", fmt.Sprintf("%s EXHAUSTED", label)
	}
	if float64(q.Remaining) < float64(q.Limit)*m.threshold {
		return "warning", fmt.Sprintf("%s remaining < %.0f%%", label, m.threshold*100)
	}
	return "", ""
}

// generateAlert aplica dedup (HU-EVO-012 AC2): si ya existe una alerta activa
// (sin resolved_at) para (provider_id, model_id), solo actualiza updated_at;
// si no, inserta una nueva fila con alert_time=now().
func (m *Manager) generateAlert(ctx context.Context, q QuotaEntry, severity, message string) error {
	if m.db == nil {
		return nil // sin DB configurada: fail-soft, no-op (ver Migration Plan)
	}
	now := m.clock()

	res, err := m.db.ExecContext(ctx, `
		UPDATE provider_alerts
		SET updated_at = $1, severity = $2, message = $3
		WHERE provider_id = $4 AND model_id = $5 AND resolved_at IS NULL
	`, now, severity, message, q.Provider, q.Model)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n > 0 {
		return nil // alerta activa ya existía: solo se actualizó updated_at
	}

	_, err = m.db.ExecContext(ctx, `
		INSERT INTO provider_alerts (provider_id, model_id, severity, message, alert_time, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
	`, q.Provider, q.Model, severity, message, now)
	return err
}
