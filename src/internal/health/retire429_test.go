package health_test

import (
	"testing"
	"time"

	"api-llm-gateway/internal/health"
)

// HU-EVO-004 AC1 — Happy: proveedor recupera tras 429 (health check 200).
func TestRetireOn429_RecoversAfterHealthCheckOK(t *testing.T) {
	ok := false
	m := health.New([]string{"groq"}, func(string) bool { return ok }, 1, 1)
	m.RetireOn429("groq", 0)
	if m.Healthy("groq", "") {
		t.Fatal("tras 429 debe quedar no-sano")
	}
	ok = true
	m.CheckOnce()
	if !m.Healthy("groq", "") {
		t.Error("tras health check 200 posterior al 429 debe recuperar salud")
	}
}

// HU-EVO-004 AC2 — Happy: retiro temporal respeta Retry-After.
func TestRetireOn429_RespectsRetryAfter(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	m := health.New([]string{"groq"}, func(string) bool { return true }, 1, 1)
	m.SetClock(func() time.Time { return now })

	m.RetireOn429("groq", 10*time.Second)
	if m.Healthy("groq", "") {
		t.Fatal("debe estar retirado dentro de la ventana de Retry-After")
	}
	now = now.Add(9 * time.Second)
	if m.Healthy("groq", "") {
		t.Error("no debe reactivar antes de vencer Retry-After (10s)")
	}
	now = now.Add(2 * time.Second) // total 11s > 10s
	if !m.Healthy("groq", "") {
		t.Error("debe reactivar automáticamente al vencer Retry-After")
	}
}

// HU-EVO-004 AC3 — Edge: 429 sin Retry-After -> default 30s.
func TestRetireOn429_NoRetryAfter_DefaultsTo30s(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	m := health.New([]string{"cerebras"}, func(string) bool { return true }, 1, 1)
	m.SetClock(func() time.Time { return now })

	m.RetireOn429("cerebras", 0)
	now = now.Add(29 * time.Second)
	if m.Healthy("cerebras", "") {
		t.Error("no debe reactivar antes de 30s por default")
	}
	now = now.Add(2 * time.Second)
	if !m.Healthy("cerebras", "") {
		t.Error("debe reactivar tras 30s por default")
	}
}

// HU-EVO-004 AC4 — Edge: 429 mid-stream aborta el retiro sin failover
// transparente: el estado de retiro sólo afecta futuras selecciones del
// Router (Healthy), nunca reintenta ni reabre la conexión ya en curso.
func TestRetireOn429_DoesNotAffectInFlightHealthyDecision(t *testing.T) {
	m := health.New([]string{"mistral"}, func(string) bool { return true }, 1, 1)
	// decisión tomada ANTES del 429 mid-stream (simula que el Router ya eligió mistral)
	decidedHealthyBeforeRetire := m.Healthy("mistral", "")
	if !decidedHealthyBeforeRetire {
		t.Fatal("precondición: mistral debe empezar sano")
	}
	m.RetireOn429("mistral", 0) // 429 llega mid-stream
	// la decisión ya tomada no se revierte retroactivamente (no hay estado por-stream que mutar);
	// sólo las decisiones FUTURAS ven el retiro.
	if m.Healthy("mistral", "") == decidedHealthyBeforeRetire {
		t.Error("una decisión Healthy tomada después del 429 debe reflejar el retiro (no debe seguir sano)")
	}
}

// HU-EVO-004 AC5 — Edge: backoff exponencial ante 429 repetidos (30s→60s→120s, tope).
func TestRetireOn429_ExponentialBackoff(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	m := health.New([]string{"gemini"}, func(string) bool { return true }, 1, 1)
	m.SetClock(func() time.Time { return now })

	m.RetireOn429("gemini", 0) // 1er 429 sin Retry-After -> 30s
	now = now.Add(31 * time.Second)
	if !m.Healthy("gemini", "") {
		t.Fatal("debe reactivar tras el 1er retiro (30s)")
	}

	m.RetireOn429("gemini", 0) // 2do 429 consecutivo -> 60s
	now = now.Add(31 * time.Second)
	if m.Healthy("gemini", "") {
		t.Error("2do retiro debe escalar a 60s, no reactivar a los 31s")
	}
	now = now.Add(30 * time.Second) // total 61s
	if !m.Healthy("gemini", "") {
		t.Fatal("debe reactivar tras 60s en el 2do retiro")
	}

	m.RetireOn429("gemini", 0) // 3er 429 consecutivo -> 120s
	now = now.Add(61 * time.Second)
	if m.Healthy("gemini", "") {
		t.Error("3er retiro debe escalar a 120s")
	}
	now = now.Add(60 * time.Second) // total 121s
	if !m.Healthy("gemini", "") {
		t.Fatal("debe reactivar tras 120s en el 3er retiro")
	}

	m.RetireOn429("gemini", 0) // 4to 429 consecutivo -> tope 120s (no sigue escalando)
	now = now.Add(121 * time.Second)
	if !m.Healthy("gemini", "") {
		t.Error("el backoff debe tener tope de 120s, no debe seguir escalando indefinidamente")
	}
}
