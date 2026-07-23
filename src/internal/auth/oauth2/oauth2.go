// Package oauth2 autentica peticiones por JWT de un IdP corporativo (OIDC),
// validando firma, audience, issuer y expiración.
package oauth2

import (
	"crypto/rsa"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"api-llm-gateway/internal/auth"
)

// Config parametriza la validación del JWT.
type Config struct {
	PublicKey *rsa.PublicKey // clave pública del IdP (JWKS en producción)
	Audience  string         // aud esperado
	Issuer    string         // iss esperado
}

// errWrongAudience distingue el fallo de audience (→403) del resto (→401).
var errWrongAudience = errors.New("oauth2: audience incorrecto")

// Middleware valida el JWT del header Authorization: Bearer.
func Middleware(cfg Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := bearer(r)
			if raw == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			id, err := cfg.verify(raw)
			if err != nil {
				if errors.Is(err, errWrongAudience) {
					http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
					return
				}
				log.Printf("WARN oauth2: JWT rechazado: %v", err) // AC3: loguea el intento
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(auth.WithIdentity(r.Context(), id)))
		})
	}
}

func (cfg Config) verify(raw string) (auth.Identity, error) {
	// Valida firma (RS256) + expiración + issuer. Audience se chequea aparte
	// para poder distinguir 403 de 401.
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("método de firma inesperado")
		}
		return cfg.PublicKey, nil
	}, jwt.WithValidMethods([]string{"RS256"}), jwt.WithIssuer(cfg.Issuer))
	if err != nil {
		return auth.Identity{}, err
	}
	if aud, _ := claims["aud"].(string); aud != cfg.Audience {
		return auth.Identity{}, errWrongAudience
	}

	id := auth.Identity{}
	id.Subject, _ = claims["sub"].(string)
	id.Tenant, _ = claims["tenant"].(string)
	for _, s := range toStrings(claims["scopes"]) {
		id.Scopes = append(id.Scopes, s)
	}
	return id, nil
}

func toStrings(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, e := range arr {
		if s, ok := e.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
	}
	return ""
}
