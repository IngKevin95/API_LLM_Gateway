package dlp

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// AC4: Middleware redacta cuerpo del request con límite de 50ms
func TestMiddleware_RedactSuccess(t *testing.T) {
	eng := NewEngine()
	mw := Middleware(eng, 50*time.Millisecond)

	called := false
	var finalBody string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		b, _ := io.ReadAll(r.Body)
		finalBody = string(b)
	})

	input := `{"message": "Mi correo es admin@empresa.com"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(input))
	rec := httptest.NewRecorder()

	mw(next).ServeHTTP(rec, req)

	if !called {
		t.Fatal("next no fue llamado")
	}
	if !strings.Contains(finalBody, "[REDACTED_EMAIL]") {
		t.Errorf("cuerpo no fue redactado: %q", finalBody)
	}
}

// AC5: Middleware devuelve HTTP 500 si redacción excede el timeout
func TestMiddleware_RedactTimeout(t *testing.T) {
	eng := NewEngine()
	// Configuramos un timeout imposible de 1ns
	mw := Middleware(eng, 1*time.Nanosecond)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next no debió ser llamado")
	})

	input := `{"message": "Mi correo es admin@empresa.com"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(input))

	// Cancelamos el contexto base para garantizar que el motor vea el timeout
	ctx, cancel := context.WithCancel(req.Context())
	cancel()
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	mw(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("esperaba HTTP 500, obtuve %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "DLP timeout") {
		t.Errorf("esperaba mensaje de error de timeout, obtuve %q", rec.Body.String())
	}
}
