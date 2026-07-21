package registry

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeConfig escribe un config.yaml temporal y devuelve su ruta.
func writeConfig(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatalf("escribir fixture: %v", err)
	}
	return p
}

const validYAML = `
providers:
  - id: openai
    type: openai
    base_url: https://api.openai.com/v1
    api_key: ${OPENAI_API_KEY}
    max_in_flight: 300
    models:
      - name: gpt-4o
        capabilities: [chat, vision, reasoning]
        quality_score: 95
        latency_p50_ms: 800
        cost_per_1m_tokens: 15
  - id: anthropic
    type: anthropic
    base_url: https://api.anthropic.com
    api_key: ${ANTHROPIC_API_KEY}
    max_in_flight: 250
    models:
      - name: claude-opus-4
        capabilities: [chat, reasoning]
        quality_score: 92
        latency_p50_ms: 900
        cost_per_1m_tokens: 20
routing:
  capabilities:
    chat:
      providers: [openai, anthropic]
      stream_idle_timeout_ms: 5000
    vision:
      providers: [openai]
      stream_idle_timeout_ms: 10000
`

// AC1 — Happy: carga válida expone modelos habilitados por capacidad.
func TestLoad_ValidConfig_ExposesEnabledModels(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-test-openai")
	t.Setenv("ANTHROPIC_API_KEY", "sk-test-anthropic")

	reg, err := Load(writeConfig(t, validYAML))
	if err != nil {
		t.Fatalf("Load devolvió error inesperado: %v", err)
	}

	chat := reg.ModelsFor("chat")
	if len(chat) != 2 {
		t.Fatalf("chat: esperaba 2 modelos habilitados, obtuve %d (%v)", len(chat), chat)
	}
	vision := reg.ModelsFor("vision")
	if len(vision) != 1 || vision[0].Name != "gpt-4o" {
		t.Fatalf("vision: esperaba [gpt-4o], obtuve %v", vision)
	}
	if reg.ModelCount() != 2 {
		t.Fatalf("ModelCount: esperaba 2, obtuve %d", reg.ModelCount())
	}
}

// AC2 — Error: sintaxis YAML inválida falla el arranque nombrando el archivo.
func TestLoad_InvalidYAML_FailsFast(t *testing.T) {
	// Tab en la indentación: YAML lo rechaza como sintaxis inválida.
	path := writeConfig(t, "providers:\n\t- id: openai\n")
	reg, err := Load(path)
	if err == nil {
		t.Fatal("esperaba error por YAML inválido, obtuve nil")
	}
	if reg != nil {
		t.Fatal("no debe devolver Registry parcial ante error")
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("el error debe nombrar el archivo %q: %v", path, err)
	}
	// HU-001 AC2: el error de sintaxis debe surfacear la línea (yaml.v3 emite "line N").
	if !strings.Contains(err.Error(), "line") {
		t.Errorf("el error de sintaxis debe nombrar la línea: %v", err)
	}
}

// AC3 — Edge: API key literal se rechaza sin imprimir el valor.
func TestLoad_LiteralSecret_Rejected(t *testing.T) {
	const secret = "sk-super-secret-literal-123"
	body := `
providers:
  - id: openai
    type: openai
    api_key: ` + secret + `
    models:
      - name: gpt-4o
        capabilities: [chat]
routing:
  capabilities:
    chat:
      providers: [openai]
`
	reg, err := Load(writeConfig(t, body))
	if err == nil {
		t.Fatal("esperaba error por secreto literal, obtuve nil")
	}
	if reg != nil {
		t.Fatal("no debe devolver Registry parcial ante secreto literal")
	}
	if strings.Contains(err.Error(), secret) {
		t.Errorf("el error NO debe contener el valor del secreto: %v", err)
	}
	if !strings.Contains(err.Error(), "${") {
		t.Errorf("el error debe exigir referencia ${VAR}: %v", err)
	}
}

// AC3 — Happy: ${VAR} se resuelve desde el entorno.
func TestLoad_EnvVarSecret_Resolved(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-resolved-value")
	body := `
providers:
  - id: openai
    type: openai
    api_key: ${OPENAI_API_KEY}
    models:
      - name: gpt-4o
        capabilities: [chat]
routing:
  capabilities:
    chat:
      providers: [openai]
`
	reg, err := Load(writeConfig(t, body))
	if err != nil {
		t.Fatalf("Load error inesperado: %v", err)
	}
	if got := reg.APIKey("openai"); got != "sk-resolved-value" {
		t.Errorf("APIKey(openai): esperaba valor resuelto, obtuve %q", got)
	}
}

// AC5 — Happy: parámetros de red físicos (max_in_flight, stream_idle_timeout) expuestos.
func TestLoad_NetworkParams_Exposed(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-x")
	reg, err := Load(writeConfig(t, validYAML))
	if err != nil {
		t.Fatalf("Load error inesperado: %v", err)
	}
	if got := reg.MaxInFlight("openai"); got != 300 {
		t.Errorf("MaxInFlight(openai): esperaba 300, obtuve %d", got)
	}
	if got := reg.MaxInFlight("anthropic"); got != 250 {
		t.Errorf("MaxInFlight(anthropic): esperaba 250, obtuve %d", got)
	}
	if got := reg.StreamIdleTimeoutMs("chat"); got != 5000 {
		t.Errorf("StreamIdleTimeoutMs(chat): esperaba 5000, obtuve %d", got)
	}
	if got := reg.StreamIdleTimeoutMs("vision"); got != 10000 {
		t.Errorf("StreamIdleTimeoutMs(vision): esperaba 10000, obtuve %d", got)
	}
}

// AC4 — Edge: capacidad sin modelos habilitados se marca no disponible con WARN, sin abortar.
func TestLoad_CapabilityWithoutModels_WarnsNotAvailable(t *testing.T) {
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	const body = `
providers:
  - id: openai
    type: openai
    models:
      - name: gpt-4o
        capabilities: [chat]
routing:
  capabilities:
    chat:
      providers: [openai]
    embedding:
      providers: [openai]
`
	reg, err := Load(writeConfig(t, body))
	if err != nil {
		t.Fatalf("no debe abortar por capacidad vacía: %v", err)
	}
	if reg.Available("embedding") {
		t.Error("embedding debe marcarse NO disponible (sin modelos habilitados)")
	}
	if !reg.Available("chat") {
		t.Error("chat debe marcarse disponible")
	}
	out := logBuf.String()
	if !strings.Contains(out, "WARN") || !strings.Contains(out, "embedding") {
		t.Errorf("esperaba WARN nombrando 'embedding' en el log, obtuve: %q", out)
	}
}

// AC2 — Error: campo obligatorio faltante (provider sin id) falla nombrando el campo.
func TestLoad_MissingProviderID_FailsFast(t *testing.T) {
	const body = `
providers:
  - type: openai
    models:
      - name: gpt-4o
        capabilities: [chat]
routing:
  capabilities:
    chat:
      providers: [openai]
`
	reg, err := Load(writeConfig(t, body))
	if err == nil {
		t.Fatal("esperaba error por provider sin id, obtuve nil")
	}
	if reg != nil {
		t.Fatal("no debe devolver Registry parcial ante error")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Errorf("el error debe nombrar el campo faltante 'id': %v", err)
	}
}
