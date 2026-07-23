package authz_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"api-llm-gateway/internal/auth"
	"api-llm-gateway/internal/auth/apikey"
	"api-llm-gateway/internal/authz"
)

// journey_smoke SS1: cadena authN (API key) → authZ (scope/tenant) sobre un
// handler dummy. Verifica el camino feliz y los negativos end-to-end.
func TestChain_AuthNThenAuthZ(t *testing.T) {
	store := apikey.NewStore()
	store.Add("sk-live-xyz", auth.Identity{Subject: "c1", Tenant: "T1", Scopes: []string{"capability:coding"}})

	final := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	chain := apikey.Middleware(store)(authz.Middleware(headerExtractor)(final))

	cases := []struct {
		name    string
		key     string
		headers map[string]string
		want    int
	}{
		{"válida+scope", "sk-live-xyz", map[string]string{"X-Capability": "coding", "X-Tenant": "T1"}, http.StatusOK},
		{"sin key", "", nil, http.StatusUnauthorized},
		{"key inválida", "sk-nope", map[string]string{"X-Capability": "coding"}, http.StatusUnauthorized},
		{"fuera de scope", "sk-live-xyz", map[string]string{"X-Capability": "image", "X-Tenant": "T1"}, http.StatusForbidden},
		{"cruce tenant", "sk-live-xyz", map[string]string{"X-Capability": "coding", "X-Tenant": "T2"}, http.StatusForbidden},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/chat", nil)
			if c.key != "" {
				req.Header.Set("Authorization", "Bearer "+c.key)
			}
			for k, v := range c.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()
			chain.ServeHTTP(rec, req)
			if rec.Code != c.want {
				t.Errorf("%s: esperaba %d, obtuve %d", c.name, c.want, rec.Code)
			}
		})
	}
}
