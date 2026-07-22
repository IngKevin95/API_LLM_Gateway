// Package promptguard optimiza prompts opt-in envolviendo el último mensaje
// `user` en un template, con bypass pasivo ante prompt inválido o timeout.
// Opera sobre el request; no toca la respuesta en streaming.
package promptguard

import (
	"fmt"
	"strings"
	"time"

	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter"
)

const defaultTimeout = 100 * time.Millisecond

// Guardian aplica la optimización opt-in del prompt.
type Guardian struct {
	Template string        // debe contener un %s donde va el texto original
	Timeout  time.Duration // presupuesto de optimización; 0 usa el default
	// Transform (opcional) reemplaza la lógica de optimización; por defecto
	// envuelve el texto en Template. Útil para inyectar en tests.
	Transform func(string) string
}

// New crea un Guardian con el template dado.
func New(template string) *Guardian { return &Guardian{Template: template} }

// Apply devuelve el request optimizado si enabled; hace bypass pasivo (request
// original) si está deshabilitado, no hay mensaje user, o la optimización
// excede el presupuesto de tiempo.
func (g *Guardian) Apply(req adapter.Request, enabled bool) adapter.Request {
	if !enabled {
		return req
	}
	idx := lastUserIndex(req.Messages)
	if idx < 0 {
		return req // prompt inválido: sin mensaje user → bypass
	}

	original := req.Messages[idx].Content
	optimized, ok := g.optimizeWithBudget(original)
	if !ok {
		return req // timeout → bypass al original
	}

	// Copia defensiva de los mensajes; tools y demás campos se preservan.
	msgs := make([]adapter.Message, len(req.Messages))
	copy(msgs, req.Messages)
	msgs[idx].Content = optimized
	req.Messages = msgs
	return req
}

func (g *Guardian) optimizeWithBudget(text string) (string, bool) {
	budget := g.Timeout
	if budget <= 0 {
		budget = defaultTimeout
	}
	transform := g.Transform
	if transform == nil {
		transform = g.wrap
	}
	done := make(chan string, 1)
	go func() { done <- transform(text) }()
	select {
	case out := <-done:
		return out, true
	case <-time.After(budget):
		return "", false // overhead excesivo → bypass
	}
}

func (g *Guardian) wrap(text string) string {
	if strings.Contains(g.Template, "%s") {
		return fmt.Sprintf(g.Template, text)
	}
	return g.Template + text
}

func lastUserIndex(msgs []adapter.Message) int {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == "user" && msgs[i].Content != "" {
			return i
		}
	}
	return -1
}
