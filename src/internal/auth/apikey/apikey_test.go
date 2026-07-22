package apikey_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IngKevin95/API_LLM_Gateway/internal/auth"
	"github.com/IngKevin95/API_LLM_Gateway/internal/auth/apikey"
)

// handler dummy que reporta la identidad inyectada en el context.
func idHandler(seen *auth.Identity) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id, ok := auth.FromContext(r.Context()); ok {
			*seen = id
		}
		w.WriteHeader(http.StatusOK)
	})
}

func newStore() *apikey.Store {
	s := apikey.NewStore()
	s.Add("sk-live-abc123", auth.Identity{Subject: "cliente-1", Tenant: "T1", Scopes: []string{"capability:coding"}})
	return s
}

// HU-008 AC1 — Happy: key válida en header → 200 + identidad en ctx.
func TestMiddleware_ValidKey(t *testing.T) {
	var seen auth.Identity
	h := apikey.Middleware(newStore())(idHandler(&seen))

	req := httptest.NewRequest(http.MethodPost, "/v1/chat", nil)
	req.Header.Set("Authorization", "Bearer sk-live-abc123")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d", rec.Code)
	}
	if seen.Subject != "cliente-1" || seen.Tenant != "T1" {
		t.Errorf("identidad no inyectada en ctx: %+v", seen)
	}
}

// HU-008 AC2 — Error: sin key → 401.
func TestMiddleware_NoKey(t *testing.T) {
	h := apikey.Middleware(newStore())(idHandler(new(auth.Identity)))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/chat", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("esperaba 401 sin key, obtuve %d", rec.Code)
	}
}

// HU-008 AC3 — Error: key inválida/revocada → 401, sin loguear la key completa.
func TestMiddleware_InvalidKey(t *testing.T) {
	h := apikey.Middleware(newStore())(idHandler(new(auth.Identity)))
	req := httptest.NewRequest(http.MethodPost, "/v1/chat", nil)
	req.Header.Set("Authorization", "Bearer sk-live-INVALIDA")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("esperaba 401 con key inválida, obtuve %d", rec.Code)
	}
}

// HU-008 AC4 — Edge: key en query string → 401 (canal inseguro).
func TestMiddleware_KeyInQueryRejected(t *testing.T) {
	h := apikey.Middleware(newStore())(idHandler(new(auth.Identity)))
	req := httptest.NewRequest(http.MethodPost, "/v1/chat?api_key=sk-live-abc123", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("esperaba 401 por key en query, obtuve %d", rec.Code)
	}
}
