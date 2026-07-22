package oauth2_test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/IngKevin95/API_LLM_Gateway/internal/auth"
	"github.com/IngKevin95/API_LLM_Gateway/internal/auth/oauth2"
)

func mustKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func sign(t *testing.T, key *rsa.PrivateKey, claims jwt.MapClaims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	s, err := tok.SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func run(t *testing.T, key *rsa.PrivateKey, token string) (int, auth.Identity) {
	var seen auth.Identity
	h := oauth2.Middleware(oauth2.Config{PublicKey: &key.PublicKey, Audience: "gateway", Issuer: "corp-idp"})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if id, ok := auth.FromContext(r.Context()); ok {
				seen = id
			}
			w.WriteHeader(http.StatusOK)
		}))
	req := httptest.NewRequest(http.MethodPost, "/v1/chat", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Code, seen
}

// HU-025a AC1 — Happy: JWT válido → autoriza + identidad.
func TestOAuth2_ValidJWT(t *testing.T) {
	key := mustKey(t)
	tok := sign(t, key, jwt.MapClaims{
		"sub": "svc-1", "aud": "gateway", "iss": "corp-idp",
		"exp": time.Now().Add(time.Hour).Unix(), "tenant": "T1",
	})
	code, id := run(t, key, tok)
	if code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d", code)
	}
	if id.Subject != "svc-1" || id.Tenant != "T1" {
		t.Errorf("identidad mal mapeada: %+v", id)
	}
}

// HU-025a AC2 — Error: token expirado → 401.
func TestOAuth2_Expired(t *testing.T) {
	key := mustKey(t)
	tok := sign(t, key, jwt.MapClaims{"sub": "x", "aud": "gateway", "iss": "corp-idp", "exp": time.Now().Add(-time.Hour).Unix()})
	if code, _ := run(t, key, tok); code != http.StatusUnauthorized {
		t.Errorf("token expirado esperaba 401, obtuve %d", code)
	}
}

// HU-025a AC3 — Error: firma inválida → 401.
func TestOAuth2_BadSignature(t *testing.T) {
	key := mustKey(t)
	other := mustKey(t) // firmado con otra clave
	tok := sign(t, other, jwt.MapClaims{"sub": "x", "aud": "gateway", "iss": "corp-idp", "exp": time.Now().Add(time.Hour).Unix()})
	if code, _ := run(t, key, tok); code != http.StatusUnauthorized {
		t.Errorf("firma inválida esperaba 401, obtuve %d", code)
	}
}

// HU-025a AC4 — Sad: JWT válido pero aud incorrecto → 403.
func TestOAuth2_WrongAudience(t *testing.T) {
	key := mustKey(t)
	tok := sign(t, key, jwt.MapClaims{"sub": "x", "aud": "otra-app", "iss": "corp-idp", "exp": time.Now().Add(time.Hour).Unix()})
	if code, _ := run(t, key, tok); code != http.StatusForbidden {
		t.Errorf("aud incorrecto esperaba 403, obtuve %d", code)
	}
}
