package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/IngKevin95/API_LLM_Gateway/internal/auth"
	"github.com/IngKevin95/API_LLM_Gateway/internal/auth/apikey"
	"github.com/IngKevin95/API_LLM_Gateway/internal/auth/oauth2"
)

// journey_smoke SS3: distintos métodos AuthN (API key, OAuth2) convergen en el
// mismo contrato: inyectan auth.Identity en el context para las capas de abajo.
// (mTLS produce la misma auth.Identity, verificado en mtls_test.go.)
func TestAuthN_MethodsShareIdentityContract(t *testing.T) {
	// API key
	store := apikey.NewStore()
	store.Add("sk-1", auth.Identity{Subject: "via-apikey", Tenant: "T1"})
	apikeyMW := apikey.Middleware(store)

	// OAuth2 (RS256)
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": "via-oauth2", "aud": "gateway", "iss": "idp", "tenant": "T1",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	signed, _ := tok.SignedString(key)
	oauthMW := oauth2.Middleware(oauth2.Config{PublicKey: &key.PublicKey, Audience: "gateway", Issuer: "idp"})

	cases := []struct {
		name    string
		mw      func(http.Handler) http.Handler
		setAuth func(*http.Request)
		want    string
	}{
		{"apikey", apikeyMW, func(r *http.Request) { r.Header.Set("Authorization", "Bearer sk-1") }, "via-apikey"},
		{"oauth2", oauthMW, func(r *http.Request) { r.Header.Set("Authorization", "Bearer "+signed) }, "via-oauth2"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var seen auth.Identity
			h := c.mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				seen, _ = auth.FromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest(http.MethodPost, "/v1/chat", nil)
			c.setAuth(req)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("%s: esperaba 200, obtuve %d", c.name, rec.Code)
			}
			if seen.Subject != c.want {
				t.Errorf("%s: esperaba identidad %q, obtuve %q", c.name, c.want, seen.Subject)
			}
		})
	}
}
