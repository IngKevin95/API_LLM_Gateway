package authz_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IngKevin95/API_LLM_Gateway/internal/auth"
	"github.com/IngKevin95/API_LLM_Gateway/internal/authz"
)

// extractor de prueba: lee la intención de la petición desde headers.
func headerExtractor(r *http.Request) authz.Access {
	return authz.Access{
		Capability: r.Header.Get("X-Capability"),
		Model:      r.Header.Get("X-Model"),
		Tenant:     r.Header.Get("X-Tenant"),
	}
}

func ok200() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
}

// arma una petición ya autenticada con la identidad dada.
func authedReq(id auth.Identity, headers map[string]string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/v1/chat", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req.WithContext(auth.WithIdentity(req.Context(), id))
}

func run(id auth.Identity, headers map[string]string) int {
	h := authz.Middleware(headerExtractor)(ok200())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, authedReq(id, headers))
	return rec.Code
}

// HU-009 AC1 — Happy: scope permitido + tenant correcto → autoriza.
func TestAuthz_ScopeAllowed(t *testing.T) {
	id := auth.Identity{Tenant: "T1", Scopes: []string{"capability:coding"}}
	if code := run(id, map[string]string{"X-Capability": "coding", "X-Tenant": "T1"}); code != http.StatusOK {
		t.Errorf("esperaba 200, obtuve %d", code)
	}
}

// HU-009 AC2 — Error: capacidad fuera de scope → 403.
func TestAuthz_CapabilityOutOfScope(t *testing.T) {
	id := auth.Identity{Tenant: "T1", Scopes: []string{"capability:coding"}}
	if code := run(id, map[string]string{"X-Capability": "image", "X-Tenant": "T1"}); code != http.StatusForbidden {
		t.Errorf("esperaba 403, obtuve %d", code)
	}
}

// HU-009 AC3 — Error: cruce de tenant → 403.
func TestAuthz_CrossTenant(t *testing.T) {
	id := auth.Identity{Tenant: "T1", Scopes: []string{"capability:coding"}}
	if code := run(id, map[string]string{"X-Capability": "coding", "X-Tenant": "T2"}); code != http.StatusForbidden {
		t.Errorf("esperaba 403 por cruce de tenant, obtuve %d", code)
	}
}

// HU-009 AC4 — Edge: modelo vetado → 403 aunque la capacidad esté permitida.
func TestAuthz_VettedModel(t *testing.T) {
	id := auth.Identity{Tenant: "T1", Scopes: []string{"capability:coding"}, VettedModels: []string{"gpt-4o"}}
	if code := run(id, map[string]string{"X-Capability": "coding", "X-Tenant": "T1", "X-Model": "gpt-4o"}); code != http.StatusForbidden {
		t.Errorf("esperaba 403 por modelo vetado, obtuve %d", code)
	}
}

// HU-009 AC5 — Edge: vision sin scope trusted → 403.
func TestAuthz_VisionRequiresTrustedScope(t *testing.T) {
	id := auth.Identity{Tenant: "T1", Scopes: []string{"capability:vision"}} // no :trusted
	if code := run(id, map[string]string{"X-Capability": "vision", "X-Tenant": "T1"}); code != http.StatusForbidden {
		t.Errorf("esperaba 403 (vision exige scope trusted), obtuve %d", code)
	}
	// con el scope trusted, autoriza
	id2 := auth.Identity{Tenant: "T1", Scopes: []string{"capability:vision:trusted"}}
	if code := run(id2, map[string]string{"X-Capability": "vision", "X-Tenant": "T1"}); code != http.StatusOK {
		t.Errorf("con vision:trusted esperaba 200, obtuve %d", code)
	}
}
