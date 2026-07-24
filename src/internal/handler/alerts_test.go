package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"api-llm-gateway/internal/auth"
)

// HU-EVO-013: sin identidad y sin admin -> 401.
func TestAlertsHandler_NoIdentity_Unauthorized(t *testing.T) {
	h := NewAlertsHandler(nil, nil)
	req := httptest.NewRequest("GET", "/alerts", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("esperaba 401, obtuve %d", w.Code)
	}
}

// Fail-soft: sin DB configurada, responde 200 con lista vacía (no 500), para
// no tumbar clientes de /alerts en despliegues sin PostgreSQL (ver
// design.md Migration Plan).
func TestAlertsHandler_NoDB_ReturnsEmptyList(t *testing.T) {
	h := NewAlertsHandler(nil, nil)
	req := httptest.NewRequest("GET", "/alerts", nil)
	req = req.WithContext(WithAdmin(req.Context()))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d: %s", w.Code, w.Body.String())
	}
	var resp AlertsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("respuesta no es JSON válido: %v", err)
	}
	if len(resp.Data) != 0 {
		t.Errorf("esperaba data vacía, obtuve %d filas", len(resp.Data))
	}
}

// HU-EVO-013 AC4: identidad con scope insuficiente no ve la alerta filtrada
// por allowed(), aunque el requester esté autenticado.
func TestAlertsHandler_allowed_FiltersInsufficientScope(t *testing.T) {
	h := NewAlertsHandler(nil, func(provider, model string) []string {
		if provider == "groq" {
			return []string{"vision"}
		}
		return nil
	})
	id := auth.Identity{Subject: "t1", Tenant: "t1", Scopes: []string{"capability:coding"}}
	row := AlertRow{Provider: "groq", Model: "mixtral"}
	if h.allowed(id, row) {
		t.Errorf("esperaba filtrar alerta de capability:vision para identidad con solo capability:coding")
	}
}

// Happy: scope suficiente sí ve la alerta.
func TestAlertsHandler_allowed_MatchingScope(t *testing.T) {
	h := NewAlertsHandler(nil, func(provider, model string) []string {
		return []string{"coding"}
	})
	id := auth.Identity{Subject: "t1", Tenant: "t1", Scopes: []string{"capability:coding"}}
	row := AlertRow{Provider: "groq", Model: "mixtral"}
	if !h.allowed(id, row) {
		t.Errorf("esperaba permitir alerta con scope coincidente")
	}
}

// Method not allowed.
func TestAlertsHandler_PostRejected(t *testing.T) {
	h := NewAlertsHandler(nil, nil)
	req := httptest.NewRequest("POST", "/alerts", nil)
	req = req.WithContext(WithAdmin(req.Context()))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("esperaba 405, obtuve %d", w.Code)
	}
}

func TestParsePagination_DefaultsAndOverrides(t *testing.T) {
	req := httptest.NewRequest("GET", "/alerts?page=2&limit=10", nil)
	page, limit := parsePagination(req)
	if page != 2 || limit != 10 {
		t.Fatalf("esperaba page=2 limit=10, obtuve page=%d limit=%d", page, limit)
	}

	req2 := httptest.NewRequest("GET", "/alerts", nil)
	page2, limit2 := parsePagination(req2)
	if page2 != 1 || limit2 != 50 {
		t.Fatalf("esperaba defaults page=1 limit=50, obtuve page=%d limit=%d", page2, limit2)
	}
}
