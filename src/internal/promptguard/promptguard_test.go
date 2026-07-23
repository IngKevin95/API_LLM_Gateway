package promptguard_test

import (
	"strings"
	"testing"
	"time"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/promptguard"
)

func req(msgs ...adapter.Message) adapter.Request { return adapter.Request{Messages: msgs} }

// HU-027 AC1 — Happy: envuelve el último msg user preservando el original.
func TestApply_WrapsLastUser(t *testing.T) {
	g := promptguard.New("Piensa paso a paso.\n\n%s")
	out := g.Apply(req(
		adapter.Message{Role: "system", Content: "sos útil"},
		adapter.Message{Role: "user", Content: "resolvé 2+2"},
	), true)

	last := out.Messages[len(out.Messages)-1]
	if last.Role != "user" {
		t.Fatalf("el último mensaje debe seguir siendo user")
	}
	if !strings.Contains(last.Content, "Piensa paso a paso") {
		t.Errorf("esperaba el template inyectado: %q", last.Content)
	}
	if !strings.Contains(last.Content, "resolvé 2+2") {
		t.Errorf("el texto original debe preservarse íntegro: %q", last.Content)
	}
}

// HU-027 AC1(neg)/opt-in: deshabilitado → sin cambios.
func TestApply_Disabled(t *testing.T) {
	g := promptguard.New("X %s")
	in := req(adapter.Message{Role: "user", Content: "hola"})
	out := g.Apply(in, false)
	if out.Messages[0].Content != "hola" {
		t.Errorf("deshabilitado no debe alterar el prompt: %q", out.Messages[0].Content)
	}
}

// HU-027 AC2 — Error: prompt inválido (sin msg user) → bypass sin excepción.
func TestApply_InvalidBypass(t *testing.T) {
	g := promptguard.New("X %s")
	in := req(adapter.Message{Role: "system", Content: "solo system"})
	out := g.Apply(in, true)
	if out.Messages[0].Content != "solo system" {
		t.Errorf("sin msg user debe hacer bypass: %+v", out.Messages)
	}
}

// HU-027 AC3 — Edge: tool calling intacto.
func TestApply_PreservesTools(t *testing.T) {
	g := promptguard.New("X %s")
	in := adapter.Request{
		Messages: []adapter.Message{{Role: "user", Content: "clima?"}},
		Tools:    []adapter.Tool{{Name: "get_weather", Schema: `{"type":"object"}`}},
	}
	out := g.Apply(in, true)
	if len(out.Tools) != 1 || out.Tools[0].Name != "get_weather" {
		t.Errorf("las tools deben mantenerse intactas: %+v", out.Tools)
	}
}

// HU-027 AC4 — Edge: overhead >100ms → bypass (prompt original).
func TestApply_TimeoutBypass(t *testing.T) {
	g := promptguard.New("X %s")
	g.Timeout = 50 * time.Millisecond
	g.Transform = func(s string) string { time.Sleep(200 * time.Millisecond); return "OPTIMIZADO " + s }
	out := g.Apply(req(adapter.Message{Role: "user", Content: "hola"}), true)
	if out.Messages[0].Content != "hola" {
		t.Errorf("timeout debe hacer bypass al original, obtuve %q", out.Messages[0].Content)
	}
}

// HU-027 AC5 — Edge: streaming → opera sobre el request, preserva el flag stream.
func TestApply_StreamingRequestOnly(t *testing.T) {
	g := promptguard.New("X %s")
	in := adapter.Request{Stream: true, Messages: []adapter.Message{{Role: "user", Content: "hola"}}}
	out := g.Apply(in, true)
	if !out.Stream {
		t.Error("el flag stream debe preservarse")
	}
	if !strings.Contains(out.Messages[0].Content, "hola") {
		t.Error("el request se optimiza; la respuesta token a token no se toca aquí")
	}
}
