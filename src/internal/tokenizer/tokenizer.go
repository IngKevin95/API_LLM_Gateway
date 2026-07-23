// Package tokenizer estima tokens de un request y valida la ventana de
// contexto de un modelo aplicando un buffer de seguridad.
package tokenizer

import "strings"

// bufferPercent es el margen de seguridad sobre la estimación (PRD: 20%).
const bufferPercent = 20

// Tokenizer estima tokens y valida contra la ventana de contexto de un modelo.
type Tokenizer interface {
	// Estimate devuelve los tokens estimados de un texto.
	Estimate(text string) int
	// FitsWindow indica si estimatedTokens, aumentado por el buffer de
	// seguridad, cabe en una ventana de maxContextTokens.
	FitsWindow(estimatedTokens, maxContextTokens int) bool
}

// Heuristic es una implementación por defecto basada en conteo de palabras.
// ponytail: heurística palabra→token; sustituible por tiktoken-go (HU-035 Negotiable).
type Heuristic struct{}

// NewHeuristic devuelve el tokenizador heurístico por defecto.
func NewHeuristic() Heuristic { return Heuristic{} }

// Estimate cuenta palabras y aproxima ~1.3 tokens por palabra.
func (Heuristic) Estimate(text string) int {
	words := len(strings.Fields(text))
	return words + words*3/10
}

// FitsWindow aplica el buffer del 20% y compara contra la ventana.
func (Heuristic) FitsWindow(estimatedTokens, maxContextTokens int) bool {
	withBuffer := estimatedTokens + estimatedTokens*bufferPercent/100
	return withBuffer <= maxContextTokens
}
