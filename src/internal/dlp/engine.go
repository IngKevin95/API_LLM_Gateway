package dlp

import (
	"context"
	"io"
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

// ScanAsync lee de manera continua (streaming) del stream. Si encuentra PII grave (emails o tarjetas)
// en vuelo, ejecuta cancelFunc para interrumpir abruptamente la conexión TCP o petición.
func (e *Engine) ScanAsync(ctx context.Context, stream io.Reader, cancelFunc context.CancelFunc) {
	buf := make([]byte, 4096)
	var window []byte

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, err := stream.Read(buf)
		if n > 0 {
			// Añadimos el nuevo chunk a la ventana actual
			window = append(window, buf[:n]...)

			// Verificamos si hay match
			if e.emailRe.Match(window) || e.ccRe.Match(window) {
				cancelFunc()
				return
			}

			// Mantener un solapamiento (ej. últimos 100 bytes) para cruce de fragmentos
			if len(window) > 100 {
				window = window[len(window)-100:]
			}
		}

		if err != nil {
			// EOF o error de lectura, terminamos el escaneo
			return
		}
	}
}
