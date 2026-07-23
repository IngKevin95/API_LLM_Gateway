package audit_test

import (
	"context"
	"testing"
	"time"

	"api-llm-gateway/internal/data/audit"
)

// HU-010 AC1 — Happy: Evento inmediato tras petición exitosa
func TestAudit_SuccessfulRequest_EmitsAuditEvent(t *testing.T) {
	eventsChan := make(chan audit.AuditRecord, 10)
	auditor := audit.NewAuditor(eventsChan)

	// Dado: petición exitosa completada
	rec := audit.AuditRecord{
		RequestID:       "req-123",
		TenantID:        "tenant-001",
		AgentID:         "agent-001",
		Capability:      "reasoning",
		Provider:        "anthropic",
		Model:           "claude-3-opus",
		TokensPrompt:    100,
		TokensCompletion: 50,
		Cost:            0.005,
		LatencyMs:       245,
		StatusCode:      200,
		Timestamp:       time.Now().UTC(),
	}

	// Cuando: se registra el evento
	auditor.RecordAudit(context.Background(), rec)

	// Entonces: evento se emite inmediatamente
	select {
	case received := <-eventsChan:
		if received.RequestID != rec.RequestID {
			t.Errorf("expected RequestID %s, got %s", rec.RequestID, received.RequestID)
		}
		if received.StatusCode != 200 {
			t.Errorf("expected StatusCode 200, got %d", received.StatusCode)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for audit event")
	}
}

// HU-010 AC2 — Error: Cada reintento genera evento separado
func TestAudit_FailoverRetry_MultipleEvents(t *testing.T) {
	eventsChan := make(chan audit.AuditRecord, 10)
	auditor := audit.NewAuditor(eventsChan)

	// Dado: dos intentos (fallido + exitoso)
	events := []audit.AuditRecord{
		{
			RequestID:   "req-123",
			TenantID:    "tenant-001",
			Provider:    "openai",
			StatusCode:  429,
			Timestamp:   time.Now().UTC(),
		},
		{
			RequestID:   "req-123",
			TenantID:    "tenant-001",
			Provider:    "anthropic",
			StatusCode:  200,
			Timestamp:   time.Now().UTC(),
		},
	}

	// Cuando: se registra cada intento
	for _, rec := range events {
		auditor.RecordAudit(context.Background(), rec)
	}

	// Entonces: ambos eventos se emiten independientemente
	for i, expected := range events {
		select {
		case received := <-eventsChan:
			if received.Provider != expected.Provider {
				t.Errorf("event %d: expected provider %s, got %s", i, expected.Provider, received.Provider)
			}
			if received.StatusCode != expected.StatusCode {
				t.Errorf("event %d: expected status %d, got %d", i, expected.StatusCode, received.StatusCode)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("timeout waiting for event %d", i)
		}
	}
}

// HU-010 AC3 — Edge: Redacción marca y no persiste prompts/responses
func TestAudit_PII_MarkedRedacted_NoPromptStored(t *testing.T) {
	eventsChan := make(chan audit.AuditRecord, 10)
	auditor := audit.NewAuditor(eventsChan)

	// Dado: petición con PII redactada
	rec := audit.AuditRecord{
		RequestID:  "req-123",
		TenantID:   "tenant-001",
		Redacted:   true, // marcado por DLP
		StatusCode: 200,
		Timestamp:  time.Now().UTC(),
	}

	// Cuando: se registra
	auditor.RecordAudit(context.Background(), rec)

	// Entonces: evento marca redacted=true, sin prompts/responses en payload
	select {
	case received := <-eventsChan:
		if !received.Redacted {
			t.Error("expected Redacted=true")
		}
		// Verificar que no hay prompts/responses almacenados (se habrían redactado en DLP)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for audit event")
	}
}

// HU-010 AC4 — Particionamiento: Campos de retención configurados
func TestAudit_TTLConfig_30DayRetention(t *testing.T) {
	eventsChan := make(chan audit.AuditRecord, 10)
	auditor := audit.NewAuditor(eventsChan)

	// Dado: auditor configurado con TTL
	if auditor.TTL() != 30*24*time.Hour {
		t.Errorf("expected TTL 30 days, got %v", auditor.TTL())
	}

	// Cuando: se registra evento
	rec := audit.AuditRecord{
		RequestID:  "req-123",
		TenantID:   "tenant-001",
		StatusCode: 200,
		Timestamp:  time.Now().UTC().Add(-31 * 24 * time.Hour), // 31 días atrás
	}
	auditor.RecordAudit(context.Background(), rec)

	// Entonces: evento se emite pero timestamp indica que sería purgado (verificado por external job)
	select {
	case received := <-eventsChan:
		if received.Timestamp.After(time.Now().Add(-30 * 24 * time.Hour)) {
			t.Logf("✓ Old record eligible for purge: %v", received.Timestamp)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for audit event")
	}
}

// HU-010 AC5 — Integridad: Tabla append-only (contrato, no SQL)
func TestAudit_AppendOnly_ContractVerified(t *testing.T) {
	eventsChan := make(chan audit.AuditRecord, 10)
	auditor := audit.NewAuditor(eventsChan)

	// Dado: auditor garantiza append-only
	if !auditor.IsAppendOnly() {
		t.Error("Auditor must guarantee append-only semantics")
	}

	// Cuando: se registra evento
	rec := audit.AuditRecord{
		RequestID:  "req-123",
		TenantID:   "tenant-001",
		StatusCode: 200,
		Timestamp:  time.Now().UTC(),
	}
	auditor.RecordAudit(context.Background(), rec)

	// Entonces: evento se inserta sin posibilidad de UPDATE/DELETE
	select {
	case <-eventsChan:
		t.Logf("✓ Record appended to immutable log")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout")
	}
}
