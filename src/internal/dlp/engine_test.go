package dlp

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

// AC1: Redacta emails reemplazándolos con [REDACTED_EMAIL]
func TestEngine_RedactEmails(t *testing.T) {
	eng := NewEngine()
	input := "Hola, contacta a admin@empresa.com o soporte.tecnico@otra-empresa.co para ayuda."
	ctx := context.Background()

	gotBytes, err := eng.Redact(ctx, []byte(input))
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if strings.Contains(string(gotBytes), "admin@empresa.com") || strings.Contains(string(gotBytes), "soporte.tecnico@otra-empresa.co") {
		t.Errorf("emails no fueron redactados: %q", string(gotBytes))
	}
	if !strings.Contains(string(gotBytes), "[REDACTED_EMAIL]") {
		t.Errorf("no se encontró marcador de redacción: %q", string(gotBytes))
	}
}

// AC2: Redacta tarjetas de crédito (16 dígitos) con [REDACTED_CREDITCARD]
func TestEngine_RedactCreditCards(t *testing.T) {
	eng := NewEngine()
	input := "Mi tarjeta es 1234-5678-9012-3456 y expira pronto."
	ctx := context.Background()

	gotBytes, err := eng.Redact(ctx, []byte(input))
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if strings.Contains(string(gotBytes), "1234-5678-9012-3456") {
		t.Errorf("tarjeta no fue redactada: %q", string(gotBytes))
	}
	if !strings.Contains(string(gotBytes), "[REDACTED_CREDITCARD]") {
		t.Errorf("no se encontró marcador de redacción de tarjeta: %q", string(gotBytes))
	}
}

// AC4: ScanAsync detecta PII en stream y cancela el contexto
func TestEngine_ScanAsync_KillSwitch(t *testing.T) {
	eng := NewEngine()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Stream que emite fragmentos lentamente
	pr, pw := io.Pipe()
	go func() {
		pw.Write([]byte("Todo normal hasta aqui... "))
		time.Sleep(10 * time.Millisecond)
		pw.Write([]byte("y de repente... 1234-"))
		time.Sleep(10 * time.Millisecond)
		pw.Write([]byte("5678-9012-3456 ups"))
		pw.Close()
	}()

	// Llamada asíncrona a ScanAsync
	go eng.ScanAsync(ctx, pr, cancel)

	select {
	case <-ctx.Done():
		// Éxito, el kill-switch funcionó
	case <-time.After(1 * time.Second):
		t.Fatal("el kill switch no fue gatillado")
	}
}

// AC3: Aborta si excede el límite de tiempo del contexto
func TestEngine_RedactContextTimeout(t *testing.T) {
	eng := NewEngine()
	// Un input gigante para intentar causar algo de carga
	input := strings.Repeat("Texto de relleno admin@empresa.com con tarjeta 1234-5678-1234-5678 ", 10000)

	// Contexto ya expirado
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	// damos tiempo para que expire
	time.Sleep(10 * time.Millisecond)

	_, err := eng.Redact(ctx, []byte(input))
	if err == nil {
		t.Fatal("esperaba error por timeout, obtuve nil")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("esperaba context.DeadlineExceeded, obtuve %v", err)
	}
}
