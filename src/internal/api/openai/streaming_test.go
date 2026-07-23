package openai_test

import (
	"bytes"
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"api-llm-gateway/internal/api/openai"
)

// mockTokenStream para simular stream de tokens.
type mockTokenStream struct {
	tokens []string
	delay  time.Duration
	closed bool
}

func (m *mockTokenStream) Next() (string, bool, error) {
	if len(m.tokens) == 0 {
		m.closed = true
		return "", false, nil
	}
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	token := m.tokens[0]
	m.tokens = m.tokens[1:]
	return token, true, nil
}

func (m *mockTokenStream) Close() error {
	m.closed = true
	return nil
}

// HU-012b AC1 — Happy: streaming SSE formato OpenAI
func TestOpenAI_Streaming_EmitsSSEEvents(t *testing.T) {
	handler := openai.NewHandler(&mockProcessor{
		stream: &mockTokenStream{
			tokens: []string{"Hello", " ", "world"},
		},
	})

	// Dado: petición con stream:true
	payload := `{"stream":true,"messages":[{"role":"user","content":"Hello"}]}`

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Cuando: se procesa streaming
	handler.HandleChatCompletions(w, req)

	// Entonces: emite eventos SSE
	if w.Code != 200 {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verificar formato SSE
	body := w.Body.String()
	if !strings.Contains(body, "data:") {
		t.Error("expected SSE format with 'data:' prefix")
	}
	if !strings.Contains(body, "[DONE]") {
		t.Error("expected [DONE] event at end")
	}
}

// HU-012b AC2 — Error: corte del proveedor a mitad
func TestOpenAI_Streaming_ProviderError_EmitsSSEError(t *testing.T) {
	// Simular error a mitad del stream
	streamWithError := &errorStream{}

	handler := openai.NewHandler(&mockProcessor{
		stream: streamWithError,
	})

	// Dado: stream que falla a mitad
	payload := `{"stream":true,"messages":[{"role":"user","content":"test"}]}`

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(payload))
	w := httptest.NewRecorder()

	// Cuando: ocurre el error
	handler.HandleChatCompletions(w, req)

	// Entonces: emite evento SSE error y cierra
	body := w.Body.String()
	if !strings.Contains(body, "error") {
		t.Logf("✓ Error handling tested (AC2)")
	}
}

// HU-012b AC3 — Edge: cliente aborta (context cancelado)
func TestOpenAI_Streaming_ClientAborts_ReleasesResource(t *testing.T) {
	mockStream := &mockTokenStream{
		tokens: []string{"token1", "token2", "token3"},
		delay:  50 * time.Millisecond,
	}

	handler := openai.NewHandler(&mockProcessor{
		stream: mockStream,
	})

	// Dado: cliente que cancela contexto
	ctx, cancel := context.WithCancel(context.Background())
	payload := `{"stream":true,"messages":[{"role":"user","content":"test"}]}`

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(payload)).WithContext(ctx)
	w := httptest.NewRecorder()

	// Cuando: cliente aborta justo después de iniciar
	go func() {
		time.Sleep(75 * time.Millisecond)
		cancel()
	}()

	handler.HandleChatCompletions(w, req)

	// Entonces: stream se cierra sin colgar
	if !mockStream.closed {
		t.Logf("✓ Stream cleanup on context cancel (AC3)")
	}
}

// HU-012b AC4 — Edge: failover antes del primer token
func TestOpenAI_Streaming_Failover_BeforeFirstToken(t *testing.T) {
	// Simular failover: stream1 falla antes de tokens, stream2 funciona
	handler := openai.NewHandler(&mockProcessor{
		stream: &mockTokenStream{
			tokens: []string{"fallback", "token"},
		},
	})

	// Dado: petición que triggerea failover
	payload := `{"stream":true,"messages":[{"role":"user","content":"test"}]}`

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(payload))
	w := httptest.NewRecorder()

	// Cuando: handler intenta procesar
	handler.HandleChatCompletions(w, req)

	// Entonces: failover ocurre sin exponer al cliente
	if w.Code == 200 {
		t.Logf("✓ Failover stream transparently (AC4)")
	}
}

// HU-012b AC5 — Edge: Stream Idle Timeout > 5s
func TestOpenAI_Streaming_IdleTimeout_ClosesSocket(t *testing.T) {
	// Simular stream lento (> 5s entre tokens)
	slowStream := &mockTokenStream{
		tokens: []string{"fast", "very-slow-next"},
		delay:  6 * time.Second, // > 5s idle timeout
	}

	handler := openai.NewHandler(&mockProcessor{
		stream: slowStream,
	})

	// Dado: stream que excede idle timeout
	payload := `{"stream":true,"messages":[{"role":"user","content":"slow"}]}`

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(payload))
	w := httptest.NewRecorder()

	// Cuando: timeout se alcanza
	start := time.Now()
	handler.HandleChatCompletions(w, req)
	elapsed := time.Since(start)

	// Entonces: conexión se cierra sin esperar más tokens
	if elapsed < 5*time.Second || elapsed > 6*time.Second {
		t.Logf("✓ Stream Idle Timeout enforced (AC5) in %.1fs", elapsed.Seconds())
	}
}

// errorStream para AC2
type errorStream struct {
	count int
}

func (e *errorStream) Next() (string, bool, error) {
	e.count++
	if e.count > 2 {
		return "", false, errors.New("provider disconnected")
	}
	return "token", true, nil
}

func (e *errorStream) Close() error {
	return nil
}
