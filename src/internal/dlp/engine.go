package dlp

import (
	"context"
	"regexp"
)

// Engine representa el motor de Data Loss Prevention para redacción.
type Engine struct {
	emailRe *regexp.Regexp
	ccRe    *regexp.Regexp
}

// NewEngine inicializa el motor con las reglas de redacción.
func NewEngine() *Engine {
	return &Engine{
		// Regex simple para emails
		emailRe: regexp.MustCompile(`(?i)[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}`),
		// Regex simple para tarjetas (16 dígitos con o sin guiones/espacios)
		ccRe: regexp.MustCompile(`(?:\d[ -]*?){13,16}`),
	}
}

// Redact aplica las reglas de redacción al payload dado.
// Si el context expira antes o durante el proceso, devuelve error.
func (e *Engine) Redact(ctx context.Context, payload []byte) ([]byte, error) {
	// Verificación de timeout temprana
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	redacted := e.emailRe.ReplaceAll(payload, []byte("[REDACTED_EMAIL]"))
	redacted = e.ccRe.ReplaceAll(redacted, []byte("[REDACTED_CREDITCARD]"))

	// Verificación de timeout final
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return redacted, nil
}
