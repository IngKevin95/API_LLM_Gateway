// Package router resuelve una capacidad al modelo óptimo por un score
// determinista de 6 variables, filtrando candidatos no aptos antes de puntuar.
package router

import (
	"log"
	"sort"

	"github.com/IngKevin95/API_LLM_Gateway/internal/registry"
	"github.com/IngKevin95/API_LLM_Gateway/internal/tokenizer"
)

// ModelSource entrega el catálogo de modelos (lo satisface *registry.Registry).
type ModelSource interface {
	ModelsFor(capability string) []registry.Model
	HasCapability(capability string) bool
	FindModel(name string) (registry.Model, bool)
	ModelNames() []string
}

// HealthSource indica si un modelo está sano (fuente viva la inyecta EP-002).
type HealthSource interface {
	Healthy(providerID, model string) bool
}

// QuotaSource indica la cuota restante de un modelo (fuente viva la inyecta EP-003).
type QuotaSource interface {
	Remaining(providerID, model string) int
}

// Router resuelve capacidades a cadenas de fallback ordenadas por score.
type Router struct {
	source ModelSource
	health HealthSource
	quota  QuotaSource
	tok    tokenizer.Tokenizer
}

// New construye un Router con sus fuentes inyectadas.
func New(source ModelSource, health HealthSource, quota QuotaSource, tok tokenizer.Tokenizer) *Router {
	return &Router{source: source, health: health, quota: quota, tok: tok}
}

// Resolve devuelve los modelos aptos para la capacidad, ordenados por score
// descendente. Filtra deshabilitados, no-sanos, sin cuota y fuera de ventana
// antes de puntuar. Devuelve ErrUnknownCapability si la capacidad no existe y
// ErrNoModelsAvailable si no queda ningún candidato apto.
func (r *Router) Resolve(capability string, estimatedTokens int) ([]registry.Model, error) {
	if !r.source.HasCapability(capability) {
		return nil, ErrUnknownCapability
	}

	var candidates []registry.Model
	for _, m := range r.source.ModelsFor(capability) {
		if m.Disabled {
			continue
		}
		if !r.health.Healthy(m.ProviderID, m.Name) {
			continue
		}
		if r.quota.Remaining(m.ProviderID, m.Name) <= 0 {
			continue
		}
		if !r.tok.FitsWindow(estimatedTokens, m.MaxContextToks) {
			continue
		}
		candidates = append(candidates, m)
	}
	if len(candidates) == 0 {
		return nil, ErrNoModelsAvailable
	}

	scores := r.scoreAll(candidates)
	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if scores[a.Name] != scores[b.Name] {
			return scores[a.Name] > scores[b.Name] // mayor score primero
		}
		return tiebreak(a, b) // empate de score: menor costo, luego alfabético
	})
	return candidates, nil
}

// tiebreak desempata dos modelos de igual score: primero menor costo, en su
// defecto orden alfabético del ID (HU-002b AC3). Total y determinista.
func tiebreak(a, b registry.Model) bool {
	if a.CostPer1M != b.CostPer1M {
		return a.CostPer1M < b.CostPer1M
	}
	return a.Name < b.Name
}

// ResolveExplicit usa un modelo explícito. Si está sano lo devuelve sin scoring.
// Si no existe devuelve ModelNotFoundError; si está no-sano aplica la cadena de
// fallback (anotando la sustitución) cuando allowFallback, o ErrModelUnavailable
// si la política lo prohíbe.
func (r *Router) ResolveExplicit(capability, model string, allowFallback bool, estimatedTokens int) ([]registry.Model, error) {
	m, ok := r.source.FindModel(model)
	if !ok {
		return nil, &ModelNotFoundError{Requested: model, Valid: r.source.ModelNames()}
	}
	if r.health.Healthy(m.ProviderID, m.Name) {
		return []registry.Model{m}, nil // usa exactamente ese modelo, sin scoring
	}
	if !allowFallback {
		return nil, ErrModelUnavailable
	}
	log.Printf("router: modelo explícito %q no-sano; aplicando cadena de fallback de %q", model, capability)
	return r.Resolve(capability, estimatedTokens)
}

// pesos del score (suman 1.0).
const (
	wQuality = 0.35
	wCost    = 0.20
	wLatency = 0.15
	wSpeed   = 0.15
	wAvail   = 0.10
	wQuota   = 0.05
)

// scoreAll normaliza min-max cada eje sobre el conjunto de candidatos y
// devuelve el score total por modelo (6 variables).
func (r *Router) scoreAll(models []registry.Model) map[string]float64 {
	out := make(map[string]float64, len(models))
	if len(models) == 0 {
		return out
	}

	qMin, qMax := extent(models, func(m registry.Model) float64 { return float64(m.QualityScore) })
	lMin, lMax := extent(models, func(m registry.Model) float64 { return float64(m.LatencyP50ms) })
	cMin, cMax := extent(models, func(m registry.Model) float64 { return float64(m.CostPer1M) })
	quMin, quMax := extent(models, func(m registry.Model) float64 { return float64(r.quota.Remaining(m.ProviderID, m.Name)) })

	for _, m := range models {
		qN := norm(float64(m.QualityScore), qMin, qMax)           // mayor calidad, mejor
		latN := 1 - norm(float64(m.LatencyP50ms), lMin, lMax)     // menor latencia, mejor
		speedN := latN                                            // velocidad ≈ inverso de latencia (hasta tener throughput real)
		costN := 1 - norm(float64(m.CostPer1M), cMin, cMax)       // menor costo, mejor
		quotaN := norm(float64(r.quota.Remaining(m.ProviderID, m.Name)), quMin, quMax)
		availN := 1.0 // candidatos ya pasaron el filtro de salud

		out[m.Name] = wQuality*qN + wCost*costN + wLatency*latN +
			wSpeed*speedN + wAvail*availN + wQuota*quotaN
	}
	return out
}

func extent(models []registry.Model, f func(registry.Model) float64) (min, max float64) {
	min, max = f(models[0]), f(models[0])
	for _, m := range models[1:] {
		v := f(m)
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

// norm escala v a [0,1]; si min==max devuelve 1 (eje no diferencia).
func norm(v, min, max float64) float64 {
	if max == min {
		return 1
	}
	return (v - min) / (max - min)
}
