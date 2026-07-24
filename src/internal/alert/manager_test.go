package alert

import (
	"context"
	"testing"
)

type stubSnapshotter struct {
	entries []QuotaEntry
}

func (s stubSnapshotter) Snapshot() []QuotaEntry { return s.entries }

// HU-EVO-012 AC1: remaining < limit*threshold genera severity "warning".
func TestClassify_BelowThreshold_Warning(t *testing.T) {
	m := &Manager{threshold: DefaultThreshold}
	sev, msg := m.classify(QuotaEntry{Provider: "groq", Model: "mixtral", Limit: 14400, Remaining: 1200})
	if sev != "warning" {
		t.Fatalf("esperaba severity=warning, obtuve %q", sev)
	}
	if msg == "" {
		t.Errorf("esperaba mensaje no vacío")
	}
}

// HU-EVO-012 AC3: remaining==0 genera severity "critical" con mensaje EXHAUSTED.
func TestClassify_Exhausted_Critical(t *testing.T) {
	m := &Manager{threshold: DefaultThreshold}
	sev, msg := m.classify(QuotaEntry{Provider: "cerebras", Limit: 500, Remaining: 0})
	if sev != "critical" {
		t.Fatalf("esperaba severity=critical, obtuve %q", sev)
	}
	if msg != "cerebras EXHAUSTED" {
		t.Errorf("mensaje inesperado: %q", msg)
	}
}

// Happy path: remaining por encima del umbral no genera alerta.
func TestClassify_AboveThreshold_NoAlert(t *testing.T) {
	m := &Manager{threshold: DefaultThreshold}
	sev, _ := m.classify(QuotaEntry{Provider: "groq", Limit: 14400, Remaining: 14000})
	if sev != "" {
		t.Fatalf("esperaba sin alerta, obtuve severity=%q", sev)
	}
}

// HU-EVO-012 AC4: umbral configurable cambia la clasificación sin redeploy
// (se prueba pasando un threshold distinto al Manager, equivalente a leer
// GATEWAY_ALERT_THRESHOLD en boot).
func TestClassify_ConfigurableThreshold(t *testing.T) {
	m30 := &Manager{threshold: 0.30}
	sev, _ := m30.classify(QuotaEntry{Provider: "groq", Limit: 1000, Remaining: 250}) // 25% < 30%
	if sev != "warning" {
		t.Fatalf("con umbral 30%%, 25%% remaining debería alertar; obtuve %q", sev)
	}

	m10 := &Manager{threshold: 0.10}
	sev, _ = m10.classify(QuotaEntry{Provider: "groq", Limit: 1000, Remaining: 250}) // 25% > 10%
	if sev != "" {
		t.Fatalf("con umbral 10%%, 25%% remaining NO debería alertar; obtuve %q", sev)
	}
}

// Edge: limit<=0 (proveedor sin límite conocido aún) no es evaluable, no genera alerta.
func TestClassify_NoLimit_NotEvaluable(t *testing.T) {
	m := &Manager{threshold: DefaultThreshold}
	sev, _ := m.classify(QuotaEntry{Provider: "groq", Limit: 0, Remaining: 0})
	if sev != "" {
		t.Fatalf("esperaba sin alerta con limit<=0, obtuve %q", sev)
	}
}

// Check sin DB configurada (fail-soft): no debe entrar en pánico ni error.
func TestCheck_WithoutDB_NoOp(t *testing.T) {
	m := &Manager{
		threshold: DefaultThreshold,
		quota:     stubSnapshotter{entries: []QuotaEntry{{Provider: "groq", Limit: 1000, Remaining: 0}}},
	}
	if err := m.Check(context.Background()); err != nil {
		t.Fatalf("Check sin DB no debería devolver error, obtuvo %v", err)
	}
}

// HU-EVO-012 AC5: Mistral con 3 modelos, 2 bajo umbral -> 2 clasificaciones
// distintas (una por modelo), no una agregada por proveedor.
func TestClassify_MultipleModelsPerProvider_IndependentPerModel(t *testing.T) {
	m := &Manager{threshold: DefaultThreshold}
	entries := []QuotaEntry{
		{Provider: "mistral", Model: "small", Limit: 1000, Remaining: 900},  // ok
		{Provider: "mistral", Model: "medium", Limit: 1000, Remaining: 50}, // warning
		{Provider: "mistral", Model: "large", Limit: 1000, Remaining: 0},   // critical
	}
	var warnings, criticals, none int
	for _, e := range entries {
		sev, _ := m.classify(e)
		switch sev {
		case "warning":
			warnings++
		case "critical":
			criticals++
		default:
			none++
		}
	}
	if warnings != 1 || criticals != 1 || none != 1 {
		t.Fatalf("esperaba 1 warning, 1 critical, 1 sin alerta; obtuve warnings=%d criticals=%d none=%d", warnings, criticals, none)
	}
}
