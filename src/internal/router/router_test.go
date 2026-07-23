package router

import (
	"testing"

	"api-llm-gateway/internal/registry"
	"api-llm-gateway/internal/tokenizer"
)

// --- fuentes stub para los tests ---

type allHealthy struct{ down map[string]bool }

func (h allHealthy) Healthy(_ , model string) bool { return !h.down[model] }

type fixedQuota struct{ zero map[string]bool }

func (q fixedQuota) Remaining(_, model string) int {
	if q.zero[model] {
		return 0
	}
	return 1000
}

// modelos de prueba: "best" domina todos los ejes, "mid" domina a "low".
func testModels() []registry.Model {
	return []registry.Model{
		{Name: "best", ProviderID: "p", Capabilities: []string{"chat"}, QualityScore: 95, LatencyP50ms: 100, CostPer1M: 0, MaxContextToks: 100_000},
		{Name: "mid", ProviderID: "p", Capabilities: []string{"chat"}, QualityScore: 90, LatencyP50ms: 500, CostPer1M: 10, MaxContextToks: 100_000},
		{Name: "low", ProviderID: "p", Capabilities: []string{"chat"}, QualityScore: 70, LatencyP50ms: 900, CostPer1M: 20, MaxContextToks: 100_000},
	}
}

func names(ms []registry.Model) []string {
	out := make([]string, len(ms))
	for i, m := range ms {
		out[i] = m.Name
	}
	return out
}

// stubSource satisface ModelSource devolviendo los modelos dados para la capacidad.
type stubSource struct{ models []registry.Model }

func (s stubSource) ModelsFor(capability string) []registry.Model {
	var out []registry.Model
	for _, m := range s.models {
		for _, c := range m.Capabilities {
			if c == capability {
				out = append(out, m)
			}
		}
	}
	return out
}

func (s stubSource) HasCapability(capability string) bool {
	for _, m := range s.models {
		for _, c := range m.Capabilities {
			if c == capability {
				return true
			}
		}
	}
	return false
}

func (s stubSource) FindModel(name string) (registry.Model, bool) {
	for _, m := range s.models {
		if m.Name == name {
			return m, true
		}
	}
	return registry.Model{}, false
}

func (s stubSource) ModelNames() []string {
	out := make([]string, len(s.models))
	for i, m := range s.models {
		out[i] = m.Name
	}
	return out
}

func newRouter(models []registry.Model, down, zeroQuota map[string]bool) *Router {
	return New(stubSource{models: models},
		allHealthy{down: down},
		fixedQuota{zero: zeroQuota},
		tokenizer.NewHeuristic(),
	)
}

// HU-002a AC1 — Happy: 3 modelos → cadena ordenada por score desc (best domina).
func TestResolve_OrdersByScoreDesc(t *testing.T) {
	r := newRouter(testModels(), nil, nil)
	chain, err := r.Resolve("chat", 1000)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	got := names(chain)
	want := []string{"best", "mid", "low"}
	for i := range want {
		if i >= len(got) || got[i] != want[i] {
			t.Fatalf("orden esperado %v, obtuve %v", want, got)
		}
	}
}

// HU-002a AC2 — Edge: top deshabilitado/unhealthy → excluido, segundo toma primer lugar.
func TestResolve_ExcludesUnhealthyTop(t *testing.T) {
	r := newRouter(testModels(), map[string]bool{"best": true}, nil)
	chain, err := r.Resolve("chat", 1000)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	got := names(chain)
	if len(got) != 2 || got[0] != "mid" {
		t.Fatalf("con best unhealthy esperaba [mid low], obtuve %v", got)
	}
}

// HU-002a AC2 — Edge: modelo sin cuota se excluye.
func TestResolve_ExcludesNoQuota(t *testing.T) {
	r := newRouter(testModels(), nil, map[string]bool{"best": true})
	chain, err := r.Resolve("chat", 1000)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	got := names(chain)
	if len(got) != 2 || got[0] != "mid" {
		t.Fatalf("con best sin cuota esperaba [mid low], obtuve %v", got)
	}
}

// HU-002a AC3 — Edge: modelo cuya ventana es superada (con buffer) se descarta pre-score.
func TestResolve_DiscardsByContextWindow(t *testing.T) {
	models := testModels()
	models[0].MaxContextToks = 50_000 // "best" no aguanta el request grande
	r := newRouter(models, nil, nil)
	// 60k tokens estimados: best (50k) se descarta; mid/low (100k) quedan.
	chain, err := r.Resolve("chat", 60_000)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	got := names(chain)
	if len(got) != 2 || got[0] != "mid" {
		t.Fatalf("con best fuera de ventana esperaba [mid low], obtuve %v", got)
	}
}
