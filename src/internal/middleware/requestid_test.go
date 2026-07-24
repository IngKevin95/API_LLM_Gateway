package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRequestID_HeaderPresent verifies that X-Request-ID header is extracted and propagated
func TestRequestID_HeaderPresent(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract from context
		reqID := FromContext(r.Context())
		if reqID != "req-12345" {
			t.Errorf("expected req-12345, got %s", reqID)
		}
		// Verify response header is set
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "req-12345")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Verify response header
	if resp := w.Header().Get("X-Request-ID"); resp != "req-12345" {
		t.Errorf("expected X-Request-ID=req-12345 in response, got %s", resp)
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// TestRequestID_HeaderMissing verifies that request ID is generated if header is absent
func TestRequestID_HeaderMissing(t *testing.T) {
	var capturedID string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = FromContext(r.Context())
		if capturedID == "" {
			t.Error("expected generated request ID, got empty")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	// No X-Request-ID header
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Verify response header has generated ID
	respID := w.Header().Get("X-Request-ID")
	if respID == "" {
		t.Error("expected X-Request-ID in response header, got empty")
	}
	if respID != capturedID {
		t.Errorf("response header ID %s != context ID %s", respID, capturedID)
	}
}

// TestFromContext_MissingKey verifies FromContext handles missing key gracefully
func TestFromContext_MissingKey(t *testing.T) {
	ctx := context.Background()
	reqID := FromContext(ctx)
	if reqID != "" {
		t.Errorf("expected empty string for missing key, got %s", reqID)
	}
}

// TestFromContext_WrongType verifies FromContext handles non-string values gracefully
func TestFromContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), RequestIDKey, 12345) // int, not string
	reqID := FromContext(ctx)
	if reqID != "" {
		t.Errorf("expected empty string for wrong type, got %s", reqID)
	}
}
