package registry

import (
	"errors"
	"path/filepath"
	"testing"
)

// HU-EVO-002 AC1 — Happy: MergeFreeTier carga proveedores adicionales y quedan disponibles para routing.
func TestMergeFreeTier_ValidYAML_ExposesProviders(t *testing.T) {
	reg, err := Load(writeConfig(t, validYAML), nil)
	if err != nil {
		t.Fatalf("Load base: %v", err)
	}

	freeTier := `
providers:
  - id: groq
    type: generic
    base_url: https://api.groq.com/openai/v1
    models:
      - name: mixtral-8x7b-32768
        capabilities: [chat]
    quota_hint: 14400
routing:
  capabilities: {}
`
	dir := t.TempDir()
	p := filepath.Join(dir, "free-tier.yaml")
	writeFile(t, p, freeTier)

	if err := reg.MergeFreeTier(p, nil); err != nil {
		t.Fatalf("MergeFreeTier: %v", err)
	}
	m, ok := reg.FindModel("mixtral-8x7b-32768")
	if !ok || m.ProviderID != "groq" {
		t.Fatalf("esperaba modelo mixtral-8x7b-32768 de groq cargado, obtuve %+v ok=%v", m, ok)
	}
}

// HU-EVO-002 AC2 — Happy: provider en free-tier.yaml sobrescribe el de config.yaml.
func TestMergeFreeTier_OverwritesExistingProvider(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-test-openai")
	t.Setenv("ANTHROPIC_API_KEY", "sk-test-anthropic")
	reg, err := Load(writeConfig(t, validYAML), nil)
	if err != nil {
		t.Fatalf("Load base: %v", err)
	}
	if reg.MaxInFlight("openai") != 300 {
		t.Fatalf("precondición: max_in_flight base incorrecto")
	}

	overwrite := `
providers:
  - id: openai
    type: generic
    base_url: https://free-tier-openai.example
    max_in_flight: 999
    models:
      - name: gpt-4o
        capabilities: [chat]
routing:
  capabilities: {}
`
	p := filepath.Join(t.TempDir(), "free-tier.yaml")
	writeFile(t, p, overwrite)
	if err := reg.MergeFreeTier(p, nil); err != nil {
		t.Fatalf("MergeFreeTier: %v", err)
	}
	if got := reg.MaxInFlight("openai"); got != 999 {
		t.Errorf("MaxInFlight(openai): esperaba 999 (sobrescrito por free-tier), obtuve %d", got)
	}
}

// HU-EVO-002 AC3 — Error: YAML malformado -> ErrInvalidConfig, fail-fast.
func TestMergeFreeTier_MalformedYAML_ReturnsErrInvalidConfig(t *testing.T) {
	reg, err := Load(writeConfig(t, validYAML), nil)
	if err != nil {
		t.Fatalf("Load base: %v", err)
	}
	p := filepath.Join(t.TempDir(), "free-tier.yaml")
	writeFile(t, p, `{"this is": "json, not yaml providers list"`)

	err = reg.MergeFreeTier(p, nil)
	if err == nil {
		t.Fatal("esperaba error por YAML malformado, obtuve nil")
	}
	if !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("esperaba errors.Is(err, ErrInvalidConfig), obtuve %v", err)
	}
}

// HU-EVO-002 AC4 — Edge: proveedor con models:[] vacío queda excluido del scoring sin crash.
func TestMergeFreeTier_EmptyModels_ExcludedFromScoring(t *testing.T) {
	reg, err := Load(writeConfig(t, validYAML), nil)
	if err != nil {
		t.Fatalf("Load base: %v", err)
	}
	p := filepath.Join(t.TempDir(), "free-tier.yaml")
	writeFile(t, p, `
providers:
  - id: cloudflare
    type: generic
    base_url: https://api.cloudflare.com
    models: []
routing:
  capabilities:
    chat:
      providers: [openai, anthropic, cloudflare]
`)
	if err := reg.MergeFreeTier(p, nil); err != nil {
		t.Fatalf("MergeFreeTier no debe fallar con models vacío: %v", err)
	}
	chat := reg.ModelsFor("chat")
	for _, m := range chat {
		if m.ProviderID == "cloudflare" {
			t.Fatalf("cloudflare no debe aportar modelos al scoring (models: [] vacío): %+v", chat)
		}
	}
}

// HU-EVO-002 AC5 — Edge: quota_hint negativo queda expuesto para que Quota Manager lo trate como agotado.
func TestMergeFreeTier_NegativeQuotaHint_ExposedForQuotaManager(t *testing.T) {
	reg, err := Load(writeConfig(t, validYAML), nil)
	if err != nil {
		t.Fatalf("Load base: %v", err)
	}
	p := filepath.Join(t.TempDir(), "free-tier.yaml")
	writeFile(t, p, `
providers:
  - id: cerebras
    type: generic
    base_url: https://api.cerebras.ai
    models:
      - name: llama-3-70b
        capabilities: [chat]
    quota_hint: -100
routing:
  capabilities: {}
`)
	if err := reg.MergeFreeTier(p, nil); err != nil {
		t.Fatalf("MergeFreeTier: %v", err)
	}
	hints := reg.QuotaHints()
	hint, ok := hints["cerebras"]
	if !ok || hint == nil {
		t.Fatalf("esperaba quota_hint expuesto para cerebras, obtuve %+v", hints)
	}
	if *hint != -100 {
		t.Errorf("quota_hint: esperaba -100, obtuve %d", *hint)
	}
}
