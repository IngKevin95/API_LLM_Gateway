package openai_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter"
	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter/openai"
)

func drain(t *testing.T, ts adapter.TokenStream) (string, error) {
	t.Helper()
	var out string
	for {
		tok, ok, err := ts.Next()
		if err != nil {
			return out, err
		}
		if !ok {
			return out, nil
		}
		out += tok
	}
}

// HU-020b AC1 — Happy: streaming emite tokens transparentemente.
func TestStream_Happy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fl := w.(http.Flusher)
		for _, tok := range []string{"hola", " ", "mundo"} {
			fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":%q}}]}\n\n", tok)
			fl.Flush()
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
		fl.Flush()
	}))
	defer srv.Close()

	ad := openai.New(srv.URL, "sk-test")
	ts, err := ad.Stream(context.Background(), adapter.Request{Model: "gpt-4o", Stream: true, Messages: []adapter.Message{{Role: "user", Content: "hola"}}})
	if err != nil {
		t.Fatalf("Stream error inesperado: %v", err)
	}
	defer ts.Close()
	got, err := drain(t, ts)
	if err != nil {
		t.Fatalf("drain error: %v", err)
	}
	if got != "hola mundo" {
		t.Errorf("esperaba 'hola mundo', obtuve %q", got)
	}
}

// HU-020b AC2 — Edge: fallo pre-primer-token → *ProviderError (permite failover).
func TestStream_FailoverBeforeFirstToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ad := openai.New(srv.URL, "sk-test")
	_, err := ad.Stream(context.Background(), adapter.Request{Model: "gpt-4o", Stream: true})
	var pe *adapter.ProviderError
	if !isProviderError(err, &pe) || !pe.Retryable {
		t.Fatalf("esperaba *ProviderError retryable pre-primer-token, obtuve %v", err)
	}
}

// HU-020b AC3 — Edge: corte mid-stream tras el primer token → error por Stream Idle Timeout.
func TestStream_MidStreamIdleTimeout(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fl := w.(http.Flusher)
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"hola\"}}]}\n\n")
		fl.Flush()
		<-release // se cuelga tras el primer token
	}))
	defer srv.Close()
	defer close(release)

	ad := openai.New(srv.URL, "sk-test")
	ad.StreamIdle = 100 * time.Millisecond // idle corto para test determinista
	ts, err := ad.Stream(context.Background(), adapter.Request{Model: "gpt-4o", Stream: true})
	if err != nil {
		t.Fatalf("Stream error inesperado: %v", err)
	}
	defer ts.Close()

	tok, ok, err := ts.Next()
	if err != nil || !ok || tok != "hola" {
		t.Fatalf("primer token esperado 'hola', obtuve tok=%q ok=%v err=%v", tok, ok, err)
	}
	// El proveedor dejó de emitir → el idle timeout debe cortar con error.
	if _, _, err := ts.Next(); err == nil {
		t.Errorf("esperaba error por Stream Idle Timeout tras el corte mid-stream")
	}
}
