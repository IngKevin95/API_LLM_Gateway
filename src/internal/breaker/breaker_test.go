package breaker_test

import (
	"sync"
	"testing"
	"time"

	"github.com/IngKevin95/API_LLM_Gateway/internal/breaker"
)

// HU-004b AC2 — Edge: Max In-Flight excedido → fast-fail (Allow=false, 0 I/O).
func TestBreaker_MaxInFlight(t *testing.T) {
	b := breaker.New(breaker.Config{MaxInFlight: 2, Cooldown: time.Second})
	if !b.Allow("A") || !b.Allow("A") {
		t.Fatal("las primeras 2 adquisiciones deben permitirse")
	}
	if b.Allow("A") {
		t.Error("la 3a debe fast-fail por Max In-Flight")
	}
	b.Release("A")
	if !b.Allow("A") {
		t.Error("tras Release debe permitirse de nuevo")
	}
}

// HU-004b AC1 — Passive: tras Trip el proveedor queda inalcanzable durante el cooldown.
func TestBreaker_TripMarksUnreachable(t *testing.T) {
	b := breaker.New(breaker.Config{MaxInFlight: 10, Cooldown: time.Second})
	if !b.Available("A") {
		t.Fatal("un proveedor nuevo debe estar disponible")
	}
	b.Trip("A")
	if b.Available("A") {
		t.Error("tras Trip debe quedar inalcanzable durante el cooldown")
	}
	if b.Allow("A") {
		t.Error("Allow debe fast-fail mientras está tripped")
	}
}

// HU-004b AC3 — Edge: reactivación tras el backoff.
func TestBreaker_ReactivatesAfterCooldown(t *testing.T) {
	b := breaker.New(breaker.Config{MaxInFlight: 10, Cooldown: 50 * time.Millisecond})
	b.Trip("A")
	if b.Available("A") {
		t.Fatal("inalcanzable durante el cooldown")
	}
	time.Sleep(70 * time.Millisecond)
	if !b.Available("A") {
		t.Error("tras el cooldown debe reactivarse")
	}
}

// Concurrencia: in-flight nunca supera el máximo bajo carga (correr con -race).
func TestBreaker_ConcurrentInFlight(t *testing.T) {
	const max = 5
	b := breaker.New(breaker.Config{MaxInFlight: max, Cooldown: time.Second})
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if b.Allow("A") {
				b.Release("A")
			}
		}()
	}
	wg.Wait()
	// Tras drenar, deben poder adquirirse exactamente `max` slots.
	got := 0
	for b.Allow("A") {
		got++
	}
	if got != max {
		t.Errorf("esperaba %d slots libres tras drenar, obtuve %d", max, got)
	}
}
