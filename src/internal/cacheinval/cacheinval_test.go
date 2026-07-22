package cacheinval_test

import (
	"errors"
	"testing"

	"github.com/IngKevin95/API_LLM_Gateway/internal/cacheinval"
)

// HU-041 AC4 — Fase 1: con el flag OFF, Invalidate es no-op.
func TestInvalidate_NoOpWhenDisabled(t *testing.T) {
	invalidated := 0
	inv := &cacheinval.Invalidator{Enabled: false, OnInvalidate: func(string) { invalidated++ }}
	if inv.Invalidate("tenant-X") {
		t.Error("con el flag OFF, Invalidate debe ser no-op (devolver false)")
	}
	if invalidated != 0 {
		t.Errorf("no-op no debe disparar hidratación, se disparó %d veces", invalidated)
	}
}

// Fase 2: con el flag ON, Invalidate encola la hidratación.
func TestInvalidate_EnabledTriggersHydration(t *testing.T) {
	invalidated := 0
	inv := &cacheinval.Invalidator{Enabled: true, OnInvalidate: func(string) { invalidated++ }}
	if !inv.Invalidate("tenant-X") {
		t.Error("con el flag ON, Invalidate debe actuar (devolver true)")
	}
	if invalidated != 1 {
		t.Errorf("esperaba 1 hidratación encolada, obtuve %d", invalidated)
	}
}

// HU-041 AC2 — patrón vigente en MVP: poll/retry ante miss (fail-fast + retry).
func TestOnMiss_RehydratesWithRetry(t *testing.T) {
	inv := &cacheinval.Invalidator{MaxRetries: 3}
	calls := 0
	err := inv.OnMiss("k", func() error {
		calls++
		if calls < 2 {
			return errors.New("transitorio")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("OnMiss debe recuperarse con retry: %v", err)
	}
	if calls != 2 {
		t.Errorf("esperaba 2 intentos (1 fallo + 1 éxito), obtuve %d", calls)
	}
}

// OnMiss agota los reintentos y devuelve el último error.
func TestOnMiss_ExhaustsRetries(t *testing.T) {
	inv := &cacheinval.Invalidator{MaxRetries: 2}
	calls := 0
	err := inv.OnMiss("k", func() error { calls++; return errors.New("siempre falla") })
	if err == nil {
		t.Error("esperaba error tras agotar los reintentos")
	}
	if calls != 2 {
		t.Errorf("esperaba 2 intentos, obtuve %d", calls)
	}
}
