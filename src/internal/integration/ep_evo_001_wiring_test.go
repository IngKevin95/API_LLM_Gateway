package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/adapter/generic"
	"api-llm-gateway/internal/health"
	"api-llm-gateway/internal/quota"
	"api-llm-gateway/internal/registry"
)

// INT-adapter-registry: el adapter genérico (HU-EVO-001) se construye a
// partir de un provider cargado por Registry.MergeFreeTier (HU-EVO-002),
// usando BaseURL/api_key resueltos del catálogo, no hardcodeados.
func TestWiring_AdapterGenerico_ConstruidoDesdeRegistry(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"choices": []map[string]any{{"message": map[string]string{"content": "wired-ok"}}}})
	}))
	defer srv.Close()

	dir := t.TempDir()
	baseCfg := filepath.Join(dir, "config.yaml")
	writeFile(t, baseCfg, "providers:\n  - id: placeholder\n    type: openai\n    models: []\nrouting:\n  capabilities: {}\n")
	reg, err := registry.Load(baseCfg, nil)
	if err != nil {
		t.Fatalf("registry.Load: %v", err)
	}

	freeTier := filepath.Join(dir, "free-tier.yaml")
	writeFile(t, freeTier, `
providers:
  - id: groq
    type: generic
    base_url: `+srv.URL+`
    models:
      - name: mixtral-8x7b-32768
        capabilities: [chat]
routing:
  capabilities: {}
`)
	if err := reg.MergeFreeTier(freeTier, nil); err != nil {
		t.Fatalf("MergeFreeTier: %v", err)
	}

	providers := reg.Providers()
	var groq registry.Provider
	for _, p := range providers {
		if p.ID == "groq" {
			groq = p
		}
	}
	if groq.ID == "" {
		t.Fatal("groq no cargado desde free-tier.yaml vía Registry")
	}

	a, err := generic.New(generic.ProviderSpec{BaseURL: groq.BaseURL, Format: generic.FormatOpenAI}, reg.APIKey("groq"))
	if err != nil {
		t.Fatalf("generic.New a partir del provider de Registry: %v", err)
	}
	resp, err := a.Chat(context.Background(), adapter.Request{Model: "mixtral-8x7b-32768"})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if resp.Content != "wired-ok" {
		t.Errorf("Content: esperaba %q, obtuve %q", "wired-ok", resp.Content)
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("escribir fixture %s: %v", path, err)
	}
}

// INT-registry-quota: Registry.QuotaHints() alimenta directamente
// quota.Manager.InitFromRegistry (HU-EVO-005), sin transformación manual.
func TestWiring_RegistryProviders_InicializanQuotaManager(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.yaml")
	writeFile(t, cfg, `
providers:
  - id: groq
    type: generic
    models:
      - name: mixtral-8x7b-32768
        capabilities: [chat]
    quota_hint: 14400
routing:
  capabilities: {}
`)
	reg, err := registry.Load(cfg, nil)
	if err != nil {
		t.Fatalf("registry.Load: %v", err)
	}

	m := quota.NewInMemoryManager()
	m.InitFromRegistry(reg.QuotaHints())

	if got := m.Remaining("groq", ""); got != 14400 {
		t.Errorf("Remaining(groq) tras wiring Registry->QuotaManager: esperaba 14400, obtuve %d", got)
	}
}

// INT-registry-health: los providerIDs de Registry son la clave que Health
// Monitor usa para retirar/reactivar (HU-EVO-004); ambos módulos hablan el
// mismo espacio de nombres providerID sin traducción intermedia.
func TestWiring_RegistryProviders_ClaveHealthMonitor(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.yaml")
	writeFile(t, cfg, `
providers:
  - id: cerebras
    type: generic
    models:
      - name: llama-3.3-70b
        capabilities: [chat]
routing:
  capabilities:
    chat:
      providers: [cerebras]
`)
	reg, err := registry.Load(cfg, nil)
	if err != nil {
		t.Fatalf("registry.Load: %v", err)
	}

	var providerIDs []string
	for _, p := range reg.Providers() {
		providerIDs = append(providerIDs, p.ID)
	}

	mon := health.New(providerIDs, func(string) bool { return true }, 1, 1)
	if !mon.Healthy("cerebras", "") {
		t.Fatal("cerebras debe iniciar sano")
	}
	mon.RetireOn429("cerebras", 0)
	if mon.Healthy("cerebras", "") {
		t.Error("providerID de Registry debe retirarse en Health Monitor tras 429")
	}
}
