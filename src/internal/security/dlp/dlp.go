package dlp

import (
	"context"
	"errors"
	"regexp"
	"time"
)

var (
	ErrTimeout       = errors.New("redaction timeout")
	ErrBase64Skipped = errors.New("base64 content skipped to async")
)

// SyncRedactor redacta PII síncrono (< 10ms) antes de enviar al modelo.
// ponytail: regex duros (email/tarjeta/SSN). ML NLP diferido a EP-009.
type SyncRedactor struct {
	timeout time.Duration
	patterns []*regexp.Regexp
}

func NewSyncRedactor(timeout time.Duration) *SyncRedactor {
	return &SyncRedactor{
		timeout: timeout,
		patterns: []*regexp.Regexp{
			// Email pattern
			regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
			// Credit card (luhn-like, basic)
			regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`),
			// SSN (XXX-XX-XXXX)
			regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
		},
	}
}

// Redact redacta PII del texto con timeout.
// HU-026a AC1: clean payload pasa sin cambios.
// HU-026a AC2: email/tarjeta/SSN enmascarado con ***.
// HU-026a AC3: timeout > 50ms falla con error.
// HU-026a AC4: Base64 largo omitido, devuelve ErrBase64Skipped.
func (r *SyncRedactor) Redact(ctx context.Context, text string) (string, error) {
	// AC4: Detectar Base64 largo (> 1KB) y omitir
	if len(text) > 1024 && isLikelyBase64(text) {
		return text, ErrBase64Skipped
	}

	done := make(chan string, 1)
	go func() {
		result := text
		for _, pattern := range r.patterns {
			result = pattern.ReplaceAllString(result, "***")
		}
		done <- result
	}()

	// AC3: Timeout contra regex catastrophic backtracking
	select {
	case result := <-done:
		return result, nil
	case <-time.After(r.timeout):
		return "", ErrTimeout
	}
}

// isLikelyBase64 verifica si texto parece Base64 (simple heurística)
func isLikelyBase64(text string) bool {
	// Si el texto es muy largo y solo contiene A-Za-z0-9+/=, probable Base64
	// ponytail: heurística simple. Caso real chequearía magic bytes.
	if len(text) < 100 {
		return false
	}

	base64Chars := 0
	for _, c := range text {
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '+' || c == '/' || c == '=' {
			base64Chars++
		}
	}

	// Si > 90% de caracteres son base64-válidos, probable base64
	return float64(base64Chars)/float64(len(text)) > 0.9
}
