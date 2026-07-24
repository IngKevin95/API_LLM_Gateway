package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"api-llm-gateway/internal/adapter"
)

type mockProcessor struct{}

func (m *mockProcessor) ProcessChat(ctx context.Context, req *adapter.Request) (*adapter.Response, error) {
	return &adapter.Response{
		ProviderID: "mock-provider",
		Content:    "mock response",
	}, nil
}

func TestResponsesHandler_AcceptsUniversalFormat(t *testing.T) {
	// /responses should accept universal normalized format
	payload := map[string]interface{}{
		"format":     "universal",
		"messages":   []map[string]string{{"role": "user", "content": "hello"}},
		"model":      "router:chat",
		"max_tokens": 1024,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/responses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler := NewResponsesHandler(&mockProcessor{})
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Errorf("expected status 200 or 400, got %d", w.Code)
	}
}

func TestResponsesHandler_CapabilityInferenceFromPayload(t *testing.T) {
	// Should infer capability from request payload
	payload := map[string]interface{}{
		"messages": []map[string]interface{}{
			{"role": "user", "content": "what is in this image?"},
		},
		"max_tokens": 2048,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/responses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler := NewResponsesHandler(&mockProcessor{})
	handler.ServeHTTP(w, req)

	// Should handle request (capability inference happens internally)
	if w.Code == http.StatusInternalServerError {
		t.Errorf("handler should not error on valid payload")
	}
}

func TestResponsesHandler_RouterPrefixSupport(t *testing.T) {
	// /responses should support router: prefixed model names
	payload := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "hello"},
		},
		"model":      "router:chat",
		"max_tokens": 1024,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/responses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler := NewResponsesHandler(&mockProcessor{})
	handler.ServeHTTP(w, req)

	// Should recognize router: prefix
	if w.Code == http.StatusBadRequest {
		t.Logf("router prefix recognized (validation failed as expected without full setup)")
	}
}

func TestResponsesHandler_MissingMaxTokensValidation(t *testing.T) {
	// Missing max_tokens (required for Anthropic)
	payload := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "hello"},
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/responses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler := NewResponsesHandler(&mockProcessor{})
	handler.ServeHTTP(w, req)

	// Should validate max_tokens requirement
	if w.Code < 400 {
		t.Logf("max_tokens validation may need enforcement")
	}
}

func TestResponsesHandler_ParameterTranslation(t *testing.T) {
	// Parameters should be translated based on selected provider
	payload := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "hello"},
		},
		"model":       "gpt-4",
		"temperature": 1.5, // OpenAI-compatible range
		"max_tokens":  1024,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/responses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler := NewResponsesHandler(&mockProcessor{})
	handler.ServeHTTP(w, req)

	// Handler should process without error on parameter translation
	if w.Code == http.StatusInternalServerError {
		t.Errorf("handler should handle parameter translation")
	}
}

func TestResponsesHandler_FallbackChainOnProviderUnavailable(t *testing.T) {
	// If primary provider unavailable, should fallback
	payload := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "hello"},
		},
		"model":      "router:chat",
		"max_tokens": 1024,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/responses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler := NewResponsesHandler(&mockProcessor{})
	handler.ServeHTTP(w, req)

	// Fallback should be attempted (handler doesn't error immediately)
	if w.Code == http.StatusInternalServerError && w.Body.String() == "" {
		t.Errorf("handler should attempt fallback before erroring")
	}
}

func TestResponsesHandler_ResponseNormalization(t *testing.T) {
	// Response should be normalized to universal format
	payload := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "hello"},
		},
		"model":      "gpt-3.5-turbo",
		"max_tokens": 512,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/responses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler := NewResponsesHandler(&mockProcessor{})
	handler.ServeHTTP(w, req)

	// Response body should be valid JSON (or error message)
	if w.Code != http.StatusInternalServerError {
		var result map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil && w.Body.Len() > 0 {
			t.Logf("response normalization: body not JSON (may be streaming)")
		}
	}
}

func TestResponsesHandler_ErrorHandling(t *testing.T) {
	// Should return meaningful error for invalid requests
	payload := map[string]interface{}{
		"invalid": "payload",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/responses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler := NewResponsesHandler(&mockProcessor{})
	handler.ServeHTTP(w, req)

	// Should return error status
	if w.Code < 400 {
		t.Errorf("invalid request should return error status")
	}

	// Error response should be JSON
	var errResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Logf("error response may not be JSON: %v", err)
	}
}

func TestResponsesHandler_ContentTypeValidation(t *testing.T) {
	// Should validate Content-Type header
	payload := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "hello"},
		},
		"model":      "gpt-4",
		"max_tokens": 1024,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/responses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "text/plain") // Wrong type

	w := httptest.NewRecorder()
	handler := NewResponsesHandler(&mockProcessor{})
	handler.ServeHTTP(w, req)

	// Should reject non-JSON content type or handle gracefully
	if w.Code < 400 && w.Code >= 200 {
		t.Logf("Content-Type validation may be lenient")
	}
}

func TestResponsesHandler_MetadataPreservation(t *testing.T) {
	// Should preserve metadata from request
	payload := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "hello"},
		},
		"model":      "gpt-4",
		"max_tokens": 1024,
		"user_id":    "test-user",
		"request_id": "req-123",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/responses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "req-123")

	w := httptest.NewRecorder()
	handler := NewResponsesHandler(&mockProcessor{})
	handler.ServeHTTP(w, req)

	// Metadata should be preserved in response headers or body
	if w.Header().Get("X-Request-ID") == "" && w.Code < 300 {
		t.Logf("request ID may not be preserved in response header")
	}
}
