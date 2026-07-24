package generic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"api-llm-gateway/internal/adapter"
)

// HU-EVO-0010 AC2 (reapertura) — Retry-After como fecha HTTP RFC1123 debe
// parsearse en el path de producción real (generic.parseRetryAfter), no solo
// en el código muerto de adapter/quota.go.
func TestParseRetryAfter_RFC1123Date(t *testing.T) {
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	future := now.Add(90 * time.Second).Truncate(time.Second)
	header := future.Format(time.RFC1123)

	got := parseRetryAfterAt(header, func() time.Time { return now })

	want := future.Sub(now)
	if got != want {
		t.Errorf("parseRetryAfterAt(%q): esperaba %v, obtuve %v", header, want, got)
	}
}

func TestParseRetryAfter_RFC1123DateInPast_ClampsToZero(t *testing.T) {
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	past := now.Add(-30 * time.Second)
	header := past.Format(time.RFC1123)

	got := parseRetryAfterAt(header, func() time.Time { return now })
	if got != 0 {
		t.Errorf("fecha RFC1123 en el pasado: esperaba 0, obtuve %v", got)
	}
}

func TestParseRetryAfter_DeltaSecondsStillWorks(t *testing.T) {
	got := parseRetryAfterAt("45", time.Now)
	if got != 45*time.Second {
		t.Errorf("delta-seconds: esperaba 45s, obtuve %v", got)
	}
}

func TestParseRetryAfter_Invalid_ReturnsZero(t *testing.T) {
	got := parseRetryAfterAt("not-a-date-or-int", time.Now)
	if got != 0 {
		t.Errorf("header inválido: esperaba 0, obtuve %v", got)
	}
}

// Chat_OpenAI 429 en la respuesta inicial usando fecha RFC1123 en Retry-After
// -- extremo a extremo vía checkStatus (no solo la función pura).
func TestChat_OpenAI_429_RetryAfterRFC1123_PropagaDuracion(t *testing.T) {
	retryAt := time.Now().Add(2 * time.Minute).Truncate(time.Second)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", retryAt.Format(time.RFC1123))
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"rate limited"}`))
	}))
	defer srv.Close()

	a, err := New(ProviderSpec{BaseURL: srv.URL, Format: FormatOpenAI}, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = a.Chat(context.Background(), adapter.Request{Model: "m", Messages: []adapter.Message{{Role: "user", Content: "hi"}}})
	provErr, ok := err.(*adapter.ProviderError)
	if !ok {
		t.Fatalf("esperaba *adapter.ProviderError, obtuve %T (%v)", err, err)
	}
	if provErr.Status != 429 {
		t.Fatalf("Status: esperaba 429, obtuve %d", provErr.Status)
	}
	// Tolerancia de unos segundos por el tiempo de ejecución del test.
	if provErr.RetryAfter < 118*time.Second || provErr.RetryAfter > 120*time.Second {
		t.Errorf("RetryAfter: esperaba ~120s, obtuve %v", provErr.RetryAfter)
	}
}

// INT-adapter-quotainfo (reapertura) — generic.Chat debe poblar
// Response.QuotaInfo desde los headers reales de la respuesta HTTP, en
// producción, no solo en tests unitarios de ExtractQuota.
func TestChat_OpenAIFormat_PopulatesQuotaInfoFromHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-ratelimit-limit-requests", "100")
		w.Header().Set("x-ratelimit-remaining-requests", "42")
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]string{"role": "assistant", "content": "ok"}}},
		})
	}))
	defer srv.Close()

	a, err := New(ProviderSpec{BaseURL: srv.URL, Format: FormatOpenAI}, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	resp, err := a.Chat(context.Background(), adapter.Request{Model: "m", Messages: []adapter.Message{{Role: "user", Content: "hi"}}})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if resp.QuotaInfo.Limit != 100 || resp.QuotaInfo.Remaining != 42 {
		t.Errorf("QuotaInfo: esperaba Limit=100 Remaining=42, obtuve %+v", resp.QuotaInfo)
	}
}

func TestChat_ClaudeFormat_PopulatesQuotaInfoFromHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("anthropic-ratelimit-requests-limit", "50")
		w.Header().Set("anthropic-ratelimit-requests-remaining", "5")
		json.NewEncoder(w).Encode(map[string]any{
			"content": []map[string]string{{"type": "text", "text": "hola"}},
		})
	}))
	defer srv.Close()

	a, err := New(ProviderSpec{BaseURL: srv.URL, Format: FormatClaude}, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	resp, err := a.Chat(context.Background(), adapter.Request{Model: "m", Messages: []adapter.Message{{Role: "user", Content: "hi"}}})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if resp.QuotaInfo.Limit != 50 || resp.QuotaInfo.Remaining != 5 {
		t.Errorf("QuotaInfo: esperaba Limit=50 Remaining=5, obtuve %+v", resp.QuotaInfo)
	}
}

// HU-EVO-0010 AC4 (reapertura) — 429 (rate_limit_error) llegado a mitad del
// stream SSE debe abortar el stream (Next() devuelve *adapter.ProviderError
// Status=429) sin intentar failover transparente ni seguir emitiendo tokens.
func TestStream_OpenAI_MidStreamRateLimitError_Aborts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		fmt.Fprintf(w, "data: %s\n\n", `{"choices":[{"delta":{"content":"hola"}}]}`)
		flusher.Flush()
		fmt.Fprintf(w, "data: %s\n\n", `{"error":{"message":"rate limited mid-stream","type":"rate_limit_exceeded"}}`)
		flusher.Flush()
	}))
	defer srv.Close()

	a, err := New(ProviderSpec{BaseURL: srv.URL, Format: FormatOpenAI}, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ts, err := a.Stream(context.Background(), adapter.Request{Model: "m", Messages: []adapter.Message{{Role: "user", Content: "hi"}}})
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	defer ts.Close()

	tok, ok, err := ts.Next()
	if err != nil || !ok || tok != "hola" {
		t.Fatalf("primer token: esperaba (\"hola\", true, nil), obtuve (%q, %v, %v)", tok, ok, err)
	}

	_, ok, err = ts.Next()
	if ok {
		t.Fatalf("esperaba abort tras error mid-stream, pero ok=true")
	}
	provErr, isProvErr := err.(*adapter.ProviderError)
	if !isProvErr {
		t.Fatalf("esperaba *adapter.ProviderError, obtuve %T (%v)", err, err)
	}
	if provErr.Status != 429 {
		t.Errorf("Status: esperaba 429, obtuve %d", provErr.Status)
	}
	if provErr.Retryable {
		t.Errorf("Retryable: esperaba false (no failover transparente mid-stream), obtuve true")
	}
}

func TestStream_Claude_MidStreamRateLimitError_Aborts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		fmt.Fprintf(w, "data: %s\n\n", `{"type":"content_block_delta","delta":{"text":"hola"}}`)
		flusher.Flush()
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", `{"type":"error","error":{"type":"rate_limit_error","message":"rate limited"}}`)
		flusher.Flush()
	}))
	defer srv.Close()

	a, err := New(ProviderSpec{BaseURL: srv.URL, Format: FormatClaude}, "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ts, err := a.Stream(context.Background(), adapter.Request{Model: "m", Messages: []adapter.Message{{Role: "user", Content: "hi"}}})
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	defer ts.Close()

	tok, ok, err := ts.Next()
	if err != nil || !ok || tok != "hola" {
		t.Fatalf("primer token: esperaba (\"hola\", true, nil), obtuve (%q, %v, %v)", tok, ok, err)
	}

	_, ok, err = ts.Next()
	if ok {
		t.Fatalf("esperaba abort tras error mid-stream, pero ok=true")
	}
	provErr, isProvErr := err.(*adapter.ProviderError)
	if !isProvErr {
		t.Fatalf("esperaba *adapter.ProviderError, obtuve %T (%v)", err, err)
	}
	if provErr.Status != 429 {
		t.Errorf("Status: esperaba 429, obtuve %d", provErr.Status)
	}
}
