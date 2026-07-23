// Package apikey autentica peticiones por API key en el header, con hash
// almacenado y comparación en tiempo constante. Nunca loguea la key.
package apikey

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"log"
	"net/http"
	"strings"

	"api-llm-gateway/internal/auth"
)

// Store mapea el hash de cada key a su identidad. No guarda la key en claro.
type Store struct {
	byHash map[string]auth.Identity
}

// NewStore crea un store vacío.
func NewStore() *Store { return &Store{byHash: make(map[string]auth.Identity)} }

func hash(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

// Add registra una key (se almacena solo su hash) con su identidad.
func (s *Store) Add(key string, id auth.Identity) { s.byHash[hash(key)] = id }

// lookup compara en tiempo constante contra los hashes conocidos.
func (s *Store) lookup(key string) (auth.Identity, bool) {
	want := hash(key)
	for h, id := range s.byHash {
		if subtle.ConstantTimeCompare([]byte(h), []byte(want)) == 1 {
			return id, true
		}
	}
	return auth.Identity{}, false
}

// extractKey toma la key solo del header (Authorization: Bearer o X-API-Key).
// Nunca la acepta desde query/body.
func extractKey(r *http.Request) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
	}
	return strings.TrimSpace(r.Header.Get("X-API-Key"))
}

// Middleware exige una API key válida en el header; inyecta la identidad en el ctx.
func Middleware(store *Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := extractKey(r)
			if key == "" {
				unauthorized(w)
				return
			}
			id, ok := store.lookup(key)
			if !ok {
				// Registra el intento sin exponer la key (solo un prefijo corto).
				log.Printf("WARN apikey: intento con key inválida (prefijo %s…) desde %s", mask(key), r.RemoteAddr)
				unauthorized(w)
				return
			}
			next.ServeHTTP(w, r.WithContext(auth.WithIdentity(r.Context(), id)))
		})
	}
}

func unauthorized(w http.ResponseWriter) {
	http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
}

// mask devuelve un prefijo corto para logs, nunca la key completa.
func mask(key string) string {
	if len(key) <= 6 {
		return "***"
	}
	return key[:6]
}
