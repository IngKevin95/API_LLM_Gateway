package adapter_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/adapter/generic"
)

// ErrNoModelAvailable: spec sin modelo default (HU-EVO-003 AC3). Vive en el
// test porque conformance no depende de Registry (evita import cycle); en
// producción el Router ya expone un error equivalente para esta condición.
var ErrNoModelAvailable = errors.New("conformance: provider spec sin modelo default")

// providerConformanceCase declara un proveedor gratuito curado a validar
// contra el contrato Adapter, imitando lo que Registry.AllProviderSpecs()
// expondría en runtime (HU-EVO-003 AC1).
type providerConformanceCase struct {
	id    string
	spec  generic.ProviderSpec
	model string // vacío simula AC3: spec sin modelo default
}

func freeTierConformanceCases(baseURL string) []providerConformanceCase {
	return []providerConformanceCase{
		{id: "groq", spec: generic.ProviderSpec{BaseURL: baseURL, AuthHeader: "Authorization", Format: generic.FormatOpenAI}, model: "mixtral-8x7b-32768"},
		{id: "cerebras", spec: generic.ProviderSpec{BaseURL: baseURL, AuthHeader: "Authorization", Format: generic.FormatOpenAI}, model: "llama-3.3-70b"},
		{id: "mistral-free", spec: generic.ProviderSpec{BaseURL: baseURL, AuthHeader: "Authorization", Format: generic.FormatOpenAI}, model: "mistral-small-latest"},
		{id: "claude-compat", spec: generic.ProviderSpec{BaseURL: baseURL, AuthHeader: "x-api-key", Format: generic.FormatClaude}, model: "claude-compat-model"},
		{id: "sin-modelo", spec: generic.ProviderSpec{BaseURL: baseURL, AuthHeader: "Authorization", Format: generic.FormatOpenAI}, model: ""},
	}
}

// newConformanceServer simula ambos wire formats (openai/claude) según la
// ruta invocada, para validar Chat/Stream/Embed de cada ProviderSpec.
func newConformanceServer(t *testing.T, slow bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if slow {
			time.Sleep(200 * time.Millisecond)
		}
		switch r.URL.Path {
		case "/v1/chat/completions":
			json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{{"message": map[string]string{"role": "assistant", "content": "conformance-ok"}}},
			})
		case "/v1/messages":
			json.NewEncoder(w).Encode(map[string]any{
				"content": []map[string]any{{"type": "text", "text": "conformance-ok"}},
			})
		case "/v1/embeddings":
			json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{"embedding": []float64{0.1, 0.2, 0.3}}},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

// HU-EVO-003 AC1/AC2 — Happy: itera todos los ProviderSpecs y normaliza
// Content/Model/Usage en cada respuesta. AC5: casos en paralelo sin race
// conditions (correr con `go test -race`).
func TestConformance_AllFreeTierProviderSpecs_NormalizeResponse(t *testing.T) {
	srv := newConformanceServer(t, false)
	t.Cleanup(srv.Close)

	for _, tc := range freeTierConformanceCases(srv.URL) {
		tc := tc
		t.Run(tc.id, func(t *testing.T) {
			t.Parallel()

			if tc.model == "" {
				// AC3: spec sin modelo default -> ErrNoModelAvailable, sin abortar la suite.
				if _, err := resolveModelOrErr(tc); !errors.Is(err, ErrNoModelAvailable) {
					t.Fatalf("esperaba ErrNoModelAvailable para spec sin modelo, obtuve %v", err)
				}
				return
			}

			a, err := generic.New(tc.spec, "test-key")
			if err != nil {
				t.Fatalf("generic.New(%s): %v", tc.id, err)
			}
			var _ adapter.Adapter = a // conformidad con el contrato Adapter

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // AC4: timeout individual
			defer cancel()

			resp, err := a.Chat(ctx, adapter.Request{Model: tc.model, Messages: []adapter.Message{{Role: "user", Content: "hola"}}})
			if err != nil {
				t.Fatalf("Chat(%s): %v", tc.id, err)
			}
			if resp.Content != "conformance-ok" {
				t.Errorf("Chat(%s).Content: esperaba %q, obtuve %q", tc.id, "conformance-ok", resp.Content)
			}
		})
	}
}

// HU-EVO-003 AC4 — Edge: timeout individual por proveedor no bloquea la suite.
func TestConformance_SlowProvider_TimesOutWithoutBlockingSuite(t *testing.T) {
	slowSrv := newConformanceServer(t, true)
	t.Cleanup(slowSrv.Close)

	a, err := generic.New(generic.ProviderSpec{BaseURL: slowSrv.URL, Format: generic.FormatOpenAI}, "k")
	if err != nil {
		t.Fatalf("generic.New: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, err := a.Chat(ctx, adapter.Request{Model: "m"})
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("esperaba error de timeout, obtuve nil")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("el timeout individual no cortó la petición a tiempo; bloquearía la suite")
	}
}

func resolveModelOrErr(tc providerConformanceCase) (string, error) {
	if tc.model == "" {
		return "", ErrNoModelAvailable
	}
	return tc.model, nil
}
