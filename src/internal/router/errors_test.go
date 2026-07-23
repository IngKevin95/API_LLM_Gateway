package router

import (
	"bytes"
	"errors"
	"log"
	"os"
	"strings"
	"testing"

	"api-llm-gateway/internal/registry"
)

// HU-002b AC1 — Sad: capacidad desconocida → ErrUnknownCapability (→400).
func TestResolve_UnknownCapability(t *testing.T) {
	r := newRouter(testModels(), nil, nil)
	_, err := r.Resolve("quantum_computing", 1000)
	if !errors.Is(err, ErrUnknownCapability) {
		t.Fatalf("esperaba ErrUnknownCapability, obtuve %v", err)
	}
}

// HU-002b AC2 — Sad: ningún modelo apto → ErrNoModelsAvailable (→503).
func TestResolve_NoModelsAvailable(t *testing.T) {
	// Todos los modelos de chat caídos.
	r := newRouter(testModels(), map[string]bool{"best": true, "mid": true, "low": true}, nil)
	_, err := r.Resolve("chat", 1000)
	if !errors.Is(err, ErrNoModelsAvailable) {
		t.Fatalf("esperaba ErrNoModelsAvailable, obtuve %v", err)
	}
}

// HU-002b AC3 — Edge: empate de score desempata determinista (orden alfabético del ID).
func TestResolve_TiebreakAlphabetical(t *testing.T) {
	// Dos modelos idénticos en todos los ejes → score igual → desempata por nombre.
	models := []registry.Model{
		{Name: "beta", ProviderID: "p", Capabilities: []string{"chat"}, QualityScore: 90, LatencyP50ms: 500, CostPer1M: 10, MaxContextToks: 100_000},
		{Name: "alpha", ProviderID: "p", Capabilities: []string{"chat"}, QualityScore: 90, LatencyP50ms: 500, CostPer1M: 10, MaxContextToks: 100_000},
	}
	r := newRouter(models, nil, nil)
	got, err := r.Resolve("chat", 1000)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(got) != 2 || got[0].Name != "alpha" {
		t.Fatalf("esperaba [alpha beta] por desempate alfabético, obtuve %v", names(got))
	}
}

// HU-003 AC1 — Happy: modelo explícito sano → usa ese modelo, sin scoring.
func TestResolveExplicit_HealthyModel(t *testing.T) {
	r := newRouter(testModels(), nil, nil)
	got, err := r.ResolveExplicit("chat", "mid", false, 1000)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(got) != 1 || got[0].Name != "mid" {
		t.Fatalf("esperaba exactamente [mid], obtuve %v", names(got))
	}
}

// HU-003 AC2 — Error: modelo inexistente → ModelNotFoundError listando válidos.
func TestResolveExplicit_NotFound(t *testing.T) {
	r := newRouter(testModels(), nil, nil)
	_, err := r.ResolveExplicit("chat", "gpt-9000", false, 1000)
	var nf *ModelNotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("esperaba ModelNotFoundError, obtuve %v", err)
	}
	if len(nf.Valid) == 0 {
		t.Errorf("el error debe listar modelos válidos")
	}
}

// HU-003 AC3 — Edge: modelo caído con fallback permitido → cadena de fallback + anota sustitución.
func TestResolveExplicit_UnhealthyWithFallback(t *testing.T) {
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	r := newRouter(testModels(), map[string]bool{"best": true}, nil)
	got, err := r.ResolveExplicit("chat", "best", true, 1000)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(got) == 0 || got[0].Name == "best" {
		t.Fatalf("esperaba cadena de fallback sin 'best' al frente, obtuve %v", names(got))
	}
	if !strings.Contains(logBuf.String(), "best") {
		t.Errorf("esperaba log de sustitución nombrando 'best': %q", logBuf.String())
	}
}

// HU-003 AC4 — Edge: modelo caído sin fallback → ErrModelUnavailable (→503), sin sustituir.
func TestResolveExplicit_UnhealthyNoFallback(t *testing.T) {
	r := newRouter(testModels(), map[string]bool{"best": true}, nil)
	_, err := r.ResolveExplicit("chat", "best", false, 1000)
	if !errors.Is(err, ErrModelUnavailable) {
		t.Fatalf("esperaba ErrModelUnavailable, obtuve %v", err)
	}
}
