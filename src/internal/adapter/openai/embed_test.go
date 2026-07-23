package openai_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/adapter/openai"
)

// HU-020c AC1 — Happy: embeddings redirige a /v1/embeddings y normaliza vectores.
func TestEmbed_Happy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/embeddings" {
			t.Errorf("path esperado /v1/embeddings, obtuve %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"embedding":[0.1,0.2]},{"embedding":[0.3,0.4]}]}`))
	}))
	defer srv.Close()

	ad := openai.New(srv.URL, "sk-test")
	emb, err := ad.Embed(context.Background(), adapter.Request{Model: "text-embedding-3-small", Input: []string{"a", "b"}})
	if err != nil {
		t.Fatalf("Embed error inesperado: %v", err)
	}
	if len(emb.Vectors) != 2 || emb.Vectors[0][0] != 0.1 || emb.Vectors[1][1] != 0.4 {
		t.Errorf("vectores mal normalizados: %+v", emb.Vectors)
	}
}

// HU-020c AC2 — Edge: lote mayor al límite → error claro, sin truncar en silencio.
func TestEmbed_LargeBatchRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("no debe llegar al proveedor si excede el batch")
	}))
	defer srv.Close()

	ad := openai.New(srv.URL, "sk-test")
	ad.MaxBatch = 2
	_, err := ad.Embed(context.Background(), adapter.Request{Model: "text-embedding-3-small", Input: []string{"a", "b", "c"}})
	if err == nil {
		t.Fatal("esperaba error por exceder el batch")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "batch") {
		t.Errorf("el error debe mencionar el límite de batch: %v", err)
	}
}

// HU-020c AC3 — Error: modelo no soportado → *ProviderError normalizado.
func TestEmbed_UnsupportedModel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"message":"model not found"}}`))
	}))
	defer srv.Close()

	ad := openai.New(srv.URL, "sk-test")
	_, err := ad.Embed(context.Background(), adapter.Request{Model: "text-embedding-nope", Input: []string{"a"}})
	var pe *adapter.ProviderError
	if !isProviderError(err, &pe) || pe.Status != 404 {
		t.Fatalf("esperaba *ProviderError status 404, obtuve %v", err)
	}
}
