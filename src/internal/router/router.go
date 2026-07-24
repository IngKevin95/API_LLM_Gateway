// Package router resuelve una capacidad al modelo óptimo por un score
// determinista de 6 variables, filtrando candidatos no aptos antes de puntuar.
package router

import (
	"log"
	"sort"
	"strings"

	"api-llm-gateway/internal/registry"
	"api-llm-gateway/internal/tokenizer"
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

	log.Printf("router: resolving capability=%q", capability)

	var candidates []registry.Model
	for _, m := range r.source.ModelsFor(capability) {
		if m.Disabled {
			log.Printf("  %s (provider=%s): disabled", m.Name, m.ProviderID)
			continue
		}
		if !r.health.Healthy(m.ProviderID, m.Name) {
			log.Printf("  %s (provider=%s): unhealthy", m.Name, m.ProviderID)
			continue
		}
		if r.quota.Remaining(m.ProviderID, m.Name) <= 0 {
			log.Printf("  %s (provider=%s): no quota", m.Name, m.ProviderID)
			continue
		}
		if !r.tok.FitsWindow(estimatedTokens, m.MaxContextToks) {
			log.Printf("  %s (provider=%s): context window too small", m.Name, m.ProviderID)
			continue
		}
		log.Printf("  %s (provider=%s): candidate (quality=%d, latency=%dms)", m.Name, m.ProviderID, m.QualityScore, m.LatencyP50ms)
		candidates = append(candidates, m)
	}
	if len(candidates) == 0 {
		log.Printf("router: NO models available for capability=%q", capability)
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

	// Log routing decision
	if len(candidates) > 0 {
		chosen := candidates[0]
		score := scores[chosen.Name]
		log.Printf("router: routing decision: capability=%q, chosen_provider=%s, model=%s, score=%.2f", capability, chosen.ProviderID, chosen.Name, score)
	}

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
	wPenalty = 0.20 // HU-002: Penalización por fallos recientes
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
		penalization := 0.0 // HU-XXX: P (fallos recientes) pendiente de alimentar desde CircuitBreaker

		out[m.Name] = wQuality*qN + wCost*costN + wLatency*latN +
			wSpeed*speedN + wAvail*availN + wQuota*quotaN - (wPenalty * penalization)
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

// IsCapabilityPrefix checks if a model string uses the "router:capability" prefix.
func IsCapabilityPrefix(model string) bool {
	return strings.HasPrefix(model, "router:")
}

// ExtractCapabilityPrefix parses "router:capability" format.
// Returns (prefix, capability) e.g. ("router", "chat") for "router:chat".
// For non-prefixed models, returns ("", "").
func ExtractCapabilityPrefix(model string) (string, string) {
	if !IsCapabilityPrefix(model) {
		return "", ""
	}
	parts := strings.SplitN(model, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "router", ""
}

// ResolveCapabilityPrefix resolves "router:capability" format or explicit model names.
// If model is "router:chat", resolves by capability.
// Otherwise delegates to ResolveExplicit for backward compatibility.
func (r *Router) ResolveCapabilityPrefix(model string, estimatedTokens int) ([]registry.Model, error) {
	if IsCapabilityPrefix(model) {
		_, capability := ExtractCapabilityPrefix(model)
		return r.Resolve(capability, estimatedTokens)
	}
	// Backward compatibility: explicit model name
	return r.ResolveExplicit("", model, true, estimatedTokens)
}

// InferCapability detects the capability from request context.
// Priority:
// 1. reasoning_effort field -> "reasoning"
// 2. input field (no messages) -> "embedding"
// 3. image content in messages -> "vision"
// 4. Default -> "chat"
func InferCapability(request map[string]interface{}) string {
	// Check for reasoning_effort (Responses API)
	if _, hasReasoning := request["reasoning_effort"]; hasReasoning {
		return "reasoning"
	}

	// Check for input field (embeddings)
	if _, hasInput := request["input"]; hasInput {
		if _, hasMessages := request["messages"]; !hasMessages {
			return "embedding"
		}
	}

	// Check for image content in messages
	if messages, ok := request["messages"].([]interface{}); ok {
		if hasImageContent(messages) {
			return "vision"
		}
	}

	// Default to chat
	return "chat"
}

// hasImageContent checks if any message contains image content
func hasImageContent(messages []interface{}) bool {
	for _, msg := range messages {
		if m, ok := msg.(map[string]interface{}); ok {
			if content, ok := m["content"]; ok {
				// Content can be string (text only) or []interface{} (mixed)
				if contentList, ok := content.([]interface{}); ok {
					for _, item := range contentList {
						if itemMap, ok := item.(map[string]interface{}); ok {
							if contentType, ok := itemMap["type"].(string); ok {
								if contentType == "image_url" || contentType == "image" {
									return true
								}
							}
						}
					}
				}
			}
		}
	}
	return false
}
