package audit

import (
	"context"
	"time"
)

// AuditRecord representa un evento inmutable de auditoría.
// HU-010: Campos obligatorios para trazabilidad.
type AuditRecord struct {
	RequestID        string    // unique request identifier
	TenantID         string    // tenant for multi-tenancy
	AgentID          string    // agent identifier
	Capability       string    // routing capability
	Provider         string    // LLM provider
	Model            string    // model name
	TokensPrompt     int       // input tokens
	TokensCompletion int       // output tokens
	Cost             float64   // computed cost
	LatencyMs        int       // request latency
	StatusCode       int       // HTTP status
	Redacted         bool      // marked by DLP
	Timestamp        time.Time // UTC timestamp
}

// Auditor interface para registrar eventos.
type Auditor interface {
	RecordAudit(ctx context.Context, rec AuditRecord)
	TTL() time.Duration
	IsAppendOnly() bool
}

// chanAuditor implementa Auditor con un canal (mock BD).
// ponytail: canal en lugar de Postgres. Persistencia real en EP-009.
type chanAuditor struct {
	events chan<- AuditRecord
	ttl    time.Duration
}

func NewAuditor(events chan<- AuditRecord) Auditor {
	return &chanAuditor{
		events: events,
		ttl:    30 * 24 * time.Hour, // 30 días
	}
}

func (a *chanAuditor) RecordAudit(ctx context.Context, rec AuditRecord) {
	select {
	case a.events <- rec:
	case <-ctx.Done():
	}
}

func (a *chanAuditor) TTL() time.Duration {
	return a.ttl
}

func (a *chanAuditor) IsAppendOnly() bool {
	return true // garantía de contrato: append-only
}
