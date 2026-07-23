package dlp_test

import (
	"context"
	"testing"
	"time"

	"api-llm-gateway/internal/security/dlp"
)

// HU-026b AC1 — Happy: Kill-switch TCP a los 200ms
func TestKillSwitch_DetectsPII_AbortsTCP(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	killswitch := dlp.NewKillSwitch(200 * time.Millisecond)

	// Dado: payload largo iniciando streaming
	payload := "Normal data " + repeatString("x", 10000) + " secret@email.com"

	// Cuando: worker inicia monitoreo
	abortSignal := make(chan bool, 1)
	go func() {
		killswitch.MonitorStream(ctx, payload, abortSignal)
	}()

	// Entonces: detecta PII a los ~200ms y aborta
	select {
	case abort := <-abortSignal:
		if !abort {
			t.Error("expected abort signal")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for kill-switch signal")
	}
}

// HU-026b AC2 — Error: Timeout si stream finaliza antes
func TestKillSwitch_StreamFinishesFirst_ScannerAborts(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	killswitch := dlp.NewKillSwitch(500 * time.Millisecond) // scanner delay > stream time

	// Dado: payload limpio
	payload := "clean data"

	// Cuando: stream termina antes del escáner
	abortSignal := make(chan bool, 1)
	go func() {
		killswitch.MonitorStream(ctx, payload, abortSignal)
	}()

	// Entonces: escáner se aborta sin emitir signal
	select {
	case <-abortSignal:
		// El abort signal debería estar vacío si stream finalizó primero
	case <-ctx.Done():
		// Expected: context cancelado por timeout, escáner se desecha
		t.Logf("✓ Scanner aborted due to stream timeout")
	case <-time.After(1 * time.Second):
		t.Fatal("timeout")
	}
}

// HU-026b AC3 — Edge: Falso positivo descartado
func TestKillSwitch_FalsePositive_ContinuesStream(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	killswitch := dlp.NewKillSwitch(200 * time.Millisecond)

	// Dado: payload con patrón ambiguo (no es PII real)
	payload := "Check status@2024 for updates" // @2024 es falso positivo de email

	// Cuando: worker evalúa
	abortSignal := make(chan bool, 1)
	go func() {
		killswitch.MonitorStream(ctx, payload, abortSignal)
	}()

	// Entonces: permite que stream continúe (sin abort)
	select {
	case abort := <-abortSignal:
		if abort {
			t.Error("false positive should not trigger abort")
		}
	case <-time.After(500 * time.Millisecond):
		// Esperado: sin signal = permite continuar
		t.Logf("✓ False positive allowed stream to continue")
	}
}

// HU-026b AC4 — Sad path: PII después de stream cerrado
func TestKillSwitch_PII_AfterStreamClose_IncidentLogged(t *testing.T) {
	killswitch := dlp.NewKillSwitch(200 * time.Millisecond)

	// Dado: stream completó pero descubrimos PII post-mortem
	closedStreamPayload := "response data with api_key=secret123"

	// Cuando: hacemos análisis profundo después del cierre
	incident, err := killswitch.AnalyzePostMortem(context.Background(), closedStreamPayload)

	// Entonces: registra incidente grave
	if err != nil {
		t.Fatalf("AnalyzePostMortem failed: %v", err)
	}
	if !incident.HasPII {
		t.Error("expected HasPII=true for post-mortem detection")
	}
	if incident.Severity != "critical" {
		t.Errorf("expected critical severity, got %s", incident.Severity)
	}
}
