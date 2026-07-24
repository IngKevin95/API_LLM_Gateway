package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSessionsHandler_List(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/sessions", nil)
	rr := httptest.NewRecorder()

	// Esta prueba fallará porque no hay manejador montado ni implementación.
	// Montaríamos el handler con dependencias (SessionStore).
	// Por simplicidad de TDD Red: solo aseguramos que compile el test.

	t.Fatalf("Test no implementado - TDD Red")
}

func TestSessionsHandler_Revoke(t *testing.T) {
	req, _ := http.NewRequest(http.MethodDelete, "/sessions/123", nil)
	rr := httptest.NewRecorder()
	_ = rr
	_ = req
	t.Fatalf("Test no implementado - TDD Red")
}

func TestSessionsHandler_RevokeOthers(t *testing.T) {
	req, _ := http.NewRequest(http.MethodDelete, "/sessions?except_current=true", nil)
	rr := httptest.NewRecorder()
	_ = rr
	_ = req
	t.Fatalf("Test no implementado - TDD Red")
}
