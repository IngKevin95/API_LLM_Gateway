package health_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"api-llm-gateway/internal/health"
)

// HU-005 AC1 — Happy: proveedor sano se mantiene sano.
func TestMonitor_HealthyStaysHealthy(t *testing.T) {
	m := health.New([]string{"A"}, func(string) bool { return true }, 3, 2)
	for i := 0; i < 5; i++ {
		m.CheckOnce()
	}
	if !m.Healthy("A", "") {
		t.Error("un proveedor que responde OK debe permanecer sano")
	}
}

// HU-005 AC2 — Error: caída sostenida → retiro tras N fallos.
func TestMonitor_FailureRetires(t *testing.T) {
	ok := true
	m := health.New([]string{"A"}, func(string) bool { return ok }, 3, 2)
	m.CheckOnce() // sano
	ok = false
	m.CheckOnce()
	m.CheckOnce()
	if !m.Healthy("A", "") {
		t.Error("no debe retirarse antes de alcanzar el umbral de fallos")
	}
	m.CheckOnce() // 3er fallo → retiro
	if m.Healthy("A", "") {
		t.Error("tras N fallos debe marcarse no-sano")
	}
}

// HU-005 AC3 — Happy: reactivación tras M éxitos.
func TestMonitor_Reactivates(t *testing.T) {
	ok := false
	m := health.New([]string{"A"}, func(string) bool { return ok }, 2, 2)
	m.CheckOnce()
	m.CheckOnce() // no-sano
	if m.Healthy("A", "") {
		t.Fatal("debe estar no-sano")
	}
	ok = true
	m.CheckOnce()
	if m.Healthy("A", "") {
		t.Error("no debe reactivarse antes del umbral de éxitos")
	}
	m.CheckOnce() // 2do éxito → reactiva
	if !m.Healthy("A", "") {
		t.Error("tras M éxitos debe reactivarse")
	}
}

// HU-005 AC4 — Error: todos no-sanos.
func TestMonitor_AllUnhealthy(t *testing.T) {
	m := health.New([]string{"A", "B"}, func(string) bool { return false }, 1, 1)
	m.CheckOnce()
	if m.Healthy("A", "") || m.Healthy("B", "") {
		t.Error("todos deben quedar no-sanos")
	}
}

// HU-005 AC5 — Edge: intermitente no oscila (histéresis).
func TestMonitor_IntermittentHysteresis(t *testing.T) {
	i := 0
	// alterna OK/fallo
	m := health.New([]string{"A"}, func(string) bool { i++; return i%2 == 0 }, 3, 3)
	for k := 0; k < 10; k++ {
		m.CheckOnce()
	}
	// Nunca alcanza 3 fallos consecutivos → se mantiene sano, sin oscilar.
	if !m.Healthy("A", "") {
		t.Error("con histéresis N=3 un proveedor intermitente no debe retirarse")
	}
}

// Concurrencia: Run() escribe mientras Healthy() lee (correr con -race).
func TestMonitor_ConcurrentRunAndRead(t *testing.T) {
	m := health.New([]string{"A"}, func(string) bool { return true }, 2, 2)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); m.Run(ctx, time.Millisecond) }()
	for k := 0; k < 200; k++ {
		_ = m.Healthy("A", "")
	}
	cancel()
	wg.Wait()
}
