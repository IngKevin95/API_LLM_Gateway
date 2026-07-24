package quota

import "testing"

func intPtr(v int) *int { return &v }

// HU-EVO-005 AC1 — Happy: init desde YAML sin hacer requests.
func TestInitFromRegistry_UsesQuotaHint(t *testing.T) {
	m := NewInMemoryManager()
	m.InitFromRegistry(map[string]*int{"groq": intPtr(14400)})

	if got := m.Remaining("groq", ""); got != 14400 {
		t.Errorf("Remaining(groq): esperaba 14400, obtuve %d", got)
	}
}

// HU-EVO-005 AC2 — Happy: valor aprendido en runtime sobrescribe quota_hint inicial.
func TestInitFromRegistry_LearnedValueOverridesHint(t *testing.T) {
	m := NewInMemoryManager()
	m.InitFromRegistry(map[string]*int{"groq": intPtr(14400)})

	// simula aprendizaje desde header X-RateLimit-Remaining: 14300
	m.SetLimit("groq", Consumption{Tokens: 14300})

	if got := m.Remaining("groq", ""); got != 14300 {
		t.Errorf("Remaining(groq) tras aprendizaje: esperaba 14300, obtuve %d", got)
	}
}

// HU-EVO-005 AC3 — Error: quota_hint 0 o negativo -> agotado (remaining=0).
func TestInitFromRegistry_ZeroOrNegativeHint_TreatedAsExhausted(t *testing.T) {
	m := NewInMemoryManager()
	m.InitFromRegistry(map[string]*int{
		"cerebras-zero": intPtr(0),
		"cerebras-neg":  intPtr(-100),
	})

	if got := m.Remaining("cerebras-zero", ""); got != 0 {
		t.Errorf("Remaining(cerebras-zero): esperaba 0, obtuve %d", got)
	}
	if got := m.Remaining("cerebras-neg", ""); got != 0 {
		t.Errorf("Remaining(cerebras-neg): esperaba 0, obtuve %d", got)
	}
}

// HU-EVO-005 AC4 — Edge: proveedor sin quota_hint -> default 1M.
func TestInitFromRegistry_MissingHint_DefaultsTo1M(t *testing.T) {
	m := NewInMemoryManager()
	m.InitFromRegistry(map[string]*int{"nuevo-provider": nil})

	if got := m.Remaining("nuevo-provider", ""); got != DefaultQuotaHint {
		t.Errorf("Remaining(nuevo-provider): esperaba default %d, obtuve %d", DefaultQuotaHint, got)
	}
}

// HU-EVO-005 AC5 — Edge: valor restaurado desde persistencia (PostgreSQL) tiene
// precedencia sobre quota_hint del YAML; InitFromRegistry no lo pisa.
func TestRestoreRemaining_TakesPrecedenceOverQuotaHint(t *testing.T) {
	m := NewInMemoryManager()
	m.RestoreRemaining("mistral", 500_000_000)                    // restaurado desde PostgreSQL en boot
	m.InitFromRegistry(map[string]*int{"mistral": intPtr(14400)}) // quota_hint del YAML, no debe pisar

	if got := m.Remaining("mistral", ""); got != 500_000_000 {
		t.Errorf("Remaining(mistral): esperaba valor restaurado 500000000, obtuve %d", got)
	}
}
