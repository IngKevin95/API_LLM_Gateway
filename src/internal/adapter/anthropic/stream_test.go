package anthropic_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"api-llm-gateway/internal/adapter"
	"api-llm-gateway/internal/adapter/anthropic"
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

// HU-021b AC1 — Happy: eventos nativos → tokens transparentes.
func TestStream_Happy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fl := w.(http.Flusher)
		fmt.Fprint(w, "event: message_start\ndata: {\"type\":\"message_start\"}\n\n")
		fl.Flush()
		for _, tok := range []string{"hola", " ", "mundo"} {
			fmt.Fprintf(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":%q}}\n\n", tok)
			fl.Flush()
		}
		fmt.Fprint(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
		fl.Flush()
	}))
	defer srv.Close()

	ad := anthropic.New(srv.URL, "sk-ant")
	ts, err := ad.Stream(context.Background(), adapter.Request{Model: "claude-opus-4", Stream: true})
	if err != nil {
		t.Fatalf("Stream error: %v", err)
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

// HU-021b AC2 — Edge: fallo pre-primer-token → *ProviderError (failover).
func TestStream_FailoverBeforeFirstToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ad := anthropic.New(srv.URL, "sk-ant")
	_, err := ad.Stream(context.Background(), adapter.Request{Model: "claude-opus-4", Stream: true})
	var pe *adapter.ProviderError
	if !isProviderErr(err, &pe) || !pe.Retryable {
		t.Fatalf("esperaba *ProviderError retryable pre-primer-token, obtuve %v", err)
	}
}

// HU-021b AC3 — Edge: corte mid-stream → error por Stream Idle Timeout.
func TestStream_MidStreamIdleTimeout(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fl := w.(http.Flusher)
		fmt.Fprint(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\"hola\"}}\n\n")
		fl.Flush()
		<-release
	}))
	defer srv.Close()
	defer close(release)

	ad := anthropic.New(srv.URL, "sk-ant")
	ad.StreamIdle = int64(100 * time.Millisecond)
	ts, err := ad.Stream(context.Background(), adapter.Request{Model: "claude-opus-4", Stream: true})
	if err != nil {
		t.Fatalf("Stream error: %v", err)
	}
	defer ts.Close()

	tok, ok, err := ts.Next()
	if err != nil || !ok || tok != "hola" {
		t.Fatalf("primer token esperado 'hola', obtuve tok=%q ok=%v err=%v", tok, ok, err)
	}
	if _, _, err := ts.Next(); err == nil {
		t.Errorf("esperaba error por Stream Idle Timeout tras el corte")
	}
}
