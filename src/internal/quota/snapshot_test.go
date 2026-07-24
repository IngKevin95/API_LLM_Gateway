package quota

import (
	"testing"
	"time"

	"api-llm-gateway/internal/adapter"
)

// HU-EVO-011-AC1/AC3: proveedor sin quota learned expone remaining desde
// quota_hint (via InitFromRegistry) y LearnedAt nil.
func TestSnapshot_ProviderWithoutLearnedQuota_UsesHintAndLearnedAtNil(t *testing.T) {
	m := NewInMemoryManager()
	hint := 14400
	m.InitFromRegistry(map[string]*int{"groq": &hint})

	snap := m.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("esperaba 1 entrada, obtuve %d", len(snap))
	}
	got := snap[0]
	if got.Provider != "groq" || got.Model != "" {
		t.Errorf("provider/model inesperados: %+v", got)
	}
	if got.Remaining != 14400 || got.Limit != 14400 {
		t.Errorf("esperaba remaining=limit=14400, obtuve %+v", got)
	}
	if got.LearnedAt != nil {
		t.Errorf("esperaba LearnedAt nil (nunca aprendió de headers), obtuve %v", got.LearnedAt)
	}
	if !got.Healthy {
		t.Errorf("esperaba Healthy=true con remaining>0")
	}
}

// HU-EVO-011-AC2: múltiples modelos por proveedor se listan con remaining
// individual cada uno, tras aprender headers reales por modelo.
func TestSnapshot_MultipleModelsPerProvider_ListedIndividually(t *testing.T) {
	m := NewInMemoryManager()
	m.LearnFromHeaders("mistral", "mistral-small", adapter.QuotaInfo{Limit: 1000, Remaining: 900})
	m.LearnFromHeaders("mistral", "mistral-medium", adapter.QuotaInfo{Limit: 2000, Remaining: 100})
	m.LearnFromHeaders("mistral", "mistral-large", adapter.QuotaInfo{Limit: 500, Remaining: 0})

	snap := m.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("esperaba 3 entradas (una por modelo), obtuve %d: %+v", len(snap), snap)
	}
	byModel := map[string]Snapshot{}
	for _, s := range snap {
		if s.Provider != "mistral" {
			t.Errorf("provider inesperado: %s", s.Provider)
		}
		byModel[s.Model] = s
	}
	if byModel["mistral-small"].Remaining != 900 {
		t.Errorf("mistral-small remaining inesperado: %+v", byModel["mistral-small"])
	}
	if byModel["mistral-medium"].Remaining != 100 {
		t.Errorf("mistral-medium remaining inesperado: %+v", byModel["mistral-medium"])
	}
	if byModel["mistral-large"].Remaining != 0 || byModel["mistral-large"].Healthy {
		t.Errorf("mistral-large debería estar agotado y unhealthy: %+v", byModel["mistral-large"])
	}
	for _, s := range snap {
		if s.LearnedAt == nil {
			t.Errorf("modelo %s: esperaba LearnedAt no-nil tras LearnFromHeaders", s.Model)
		}
	}
}

// HU-EVO-011-AC4: Snapshot con 125 cuotas (25 proveedores x 5 modelos)
// responde en <100ms (lectura pura en RAM, sin I/O).
func TestSnapshot_125Quotas_UnderLatencyBudget(t *testing.T) {
	m := NewInMemoryManager()
	for p := 0; p < 25; p++ {
		provider := "provider-" + string(rune('a'+p))
		for mIdx := 0; mIdx < 5; mIdx++ {
			model := "model-" + string(rune('a'+mIdx))
			m.LearnFromHeaders(provider, model, adapter.QuotaInfo{Limit: 1000, Remaining: int64(500 + mIdx)})
		}
	}

	start := time.Now()
	snap := m.Snapshot()
	elapsed := time.Since(start)

	if len(snap) != 125 {
		t.Fatalf("esperaba 125 entradas, obtuve %d", len(snap))
	}
	if elapsed > 100*time.Millisecond {
		t.Errorf("Snapshot tardó %s, esperaba <100ms", elapsed)
	}
}

// AC (edge negativo, línea con HU-EVO-005 AC5): remaining negativo aprendido
// se clampa a 0 y Healthy=false, no rompe Snapshot.
func TestSnapshot_NegativeLearnedRemaining_ClampedToZero(t *testing.T) {
	m := NewInMemoryManager()
	m.LearnFromHeaders("cerebras", "cerebras-model", adapter.QuotaInfo{Limit: 500, Remaining: -10})

	snap := m.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("esperaba 1 entrada, obtuve %d", len(snap))
	}
	if snap[0].Remaining != 0 {
		t.Errorf("esperaba remaining clamp a 0, obtuve %d", snap[0].Remaining)
	}
	if snap[0].Healthy {
		t.Errorf("esperaba Healthy=false con remaining=0")
	}
}

// Snapshot() debe ser seguro para llamarse concurrentemente con
// LearnFromHeaders (mismo mutex, sin deadlock por RLock reentrante).
func TestSnapshot_ConcurrentWithLearnFromHeaders_NoDeadlock(t *testing.T) {
	m := NewInMemoryManager()
	m.InitFromRegistry(map[string]*int{"groq": nil})

	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			m.LearnFromHeaders("groq", "mixtral", adapter.QuotaInfo{Limit: 100, Remaining: int64(i)})
		}
		close(done)
	}()
	for i := 0; i < 100; i++ {
		_ = m.Snapshot()
	}
	<-done
}
