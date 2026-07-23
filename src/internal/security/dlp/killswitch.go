package dlp

import (
	"context"
	"regexp"
	"time"
)

// PostMortemIncident registra PII detectada después de stream cerrado.
type PostMortemIncident struct {
	HasPII    bool
	Severity  string
	Message   string
	Timestamp time.Time
}

// KillSwitch monitorea streams en paralelo para PII (AC desacoplado).
// ponytail: regex simples. NLP ligero diferido a EP-009.
type KillSwitch struct {
	timeout  time.Duration
	patterns []*regexp.Regexp
}

func NewKillSwitch(timeout time.Duration) *KillSwitch {
	return &KillSwitch{
		timeout: timeout,
		patterns: []*regexp.Regexp{
			// Reutiliza patrones de DLP
			regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
			regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`),
			regexp.MustCompile(`\bapi[_-]?key=\S+\b`),
		},
	}
}

// MonitorStream observa stream en background, aborta si detecta PII.
// HU-026b AC1: detecta PII a los ~200ms y aborta TCP.
// HU-026b AC2: timeout si stream finaliza antes.
// HU-026b AC3: falso positivo permite continuar.
func (k *KillSwitch) MonitorStream(ctx context.Context, payload string, abortSignal chan<- bool) {
	done := make(chan bool, 1)
	go func() {
		// Ejecuta escaneo
		hasPII := k.detectPII(payload)
		done <- hasPII
	}()

	select {
	case hasPII := <-done:
		// AC1: PII detectada, aborta
		if hasPII {
			select {
			case abortSignal <- true:
			default:
			}
		}
	case <-time.After(k.timeout):
		// AC2: Timeout, stream finalizó antes, descarta escáner
		return
	case <-ctx.Done():
		// AC2: Stream cancelado externamente, descarta escáner
		return
	}
}

// AnalyzePostMortem analiza payload después de stream cerrado.
// HU-026b AC4: registra incidente grave si hay PII.
func (k *KillSwitch) AnalyzePostMortem(ctx context.Context, payload string) (*PostMortemIncident, error) {
	hasPII := k.detectPII(payload)

	incident := &PostMortemIncident{
		HasPII:    hasPII,
		Timestamp: time.Now().UTC(),
	}

	if hasPII {
		incident.Severity = "critical"
		incident.Message = "PII detected in closed stream: potential leak"
	}

	return incident, nil
}

// detectPII verifica si payload contiene PII según patrones.
func (k *KillSwitch) detectPII(payload string) bool {
	for _, pattern := range k.patterns {
		if pattern.MatchString(payload) {
			return true
		}
	}
	return false
}
