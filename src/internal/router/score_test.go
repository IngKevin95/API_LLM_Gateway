package router

import (
	"testing"

	"api-llm-gateway/internal/registry"
)

// Cierra hueco 1: HU-002a AC2 rama "deshabilitado en YAML".
func TestResolve_ExcludesDisabledTop(t *testing.T) {
	models := testModels()
	models[0].Disabled = true // "best" deshabilitado en config
	r := newRouter(models, nil, nil)
	chain, err := r.Resolve("chat", 1000)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(chain) != 2 || chain[0].Name != "mid" {
		t.Fatalf("con best deshabilitado esperaba [mid low], obtuve %v", names(chain))
	}
}

// Cierra hueco 2: HU-002b AC3 rama "menor costo" del desempate (función pura).
func TestTiebreak_CostThenAlphabetical(t *testing.T) {
	cheap := registry.Model{Name: "zeta", CostPer1M: 5}
	pricey := registry.Model{Name: "alpha", CostPer1M: 20}
	// A igual score, gana el más barato aunque su nombre sea posterior.
	if !tiebreak(cheap, pricey) {
		t.Errorf("desempate por costo: esperaba que 'zeta' (costo 5) preceda a 'alpha' (costo 20)")
	}
	// A igual costo, desempata alfabético.
	a := registry.Model{Name: "alpha", CostPer1M: 10}
	b := registry.Model{Name: "beta", CostPer1M: 10}
	if !tiebreak(a, b) {
		t.Errorf("desempate alfabético: esperaba 'alpha' antes de 'beta' a igual costo")
	}
	if tiebreak(b, a) {
		t.Errorf("desempate debe ser antisimétrico")
	}
}

// Cierra hueco 3 + determinismo (data-consistency #2): scoreAll es determinista
// y ordena por las 6 variables normalizadas.
func TestScoreAll_DeterministicAndRanks(t *testing.T) {
	r := newRouter(testModels(), nil, nil)
	models := r.source.ModelsFor("chat")

	s1 := r.scoreAll(models)
	s2 := r.scoreAll(models)
	for name, v := range s1 {
		if s2[name] != v {
			t.Fatalf("scoreAll no determinista para %q: %v != %v", name, v, s2[name])
		}
	}
	// "best" domina todos los ejes → mayor score; "low" el menor.
	if !(s1["best"] > s1["mid"] && s1["mid"] > s1["low"]) {
		t.Errorf("orden de score esperado best>mid>low, obtuve best=%.4f mid=%.4f low=%.4f", s1["best"], s1["mid"], s1["low"])
	}
}

// norm no divide por cero cuando min==max (data-consistency #1).
func TestNorm_EqualExtent(t *testing.T) {
	if got := norm(5, 5, 5); got != 1 {
		t.Errorf("norm con min==max esperaba 1, obtuve %v", got)
	}
}
