package middleware

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"regexp"
)

var reqIDRegex = regexp.MustCompile(`[^a-zA-Z0-9-]`)

func sanitizeRequestID(id string) string {
	id = reqIDRegex.ReplaceAllString(id, "")
	if len(id) > 64 {
		id = id[:64]
	}
	return id
}

type contextKey string

const RequestIDKey contextKey = "request_id"

// RequestID es un middleware que extrae o genera un request ID y lo añade al contexto.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Intentar extraer de header
		requestID := r.Header.Get("X-Request-ID")
		if requestID != "" {
			// Sanitize: allow only alphanumeric and hyphens, truncate to 64 chars
			requestID = sanitizeRequestID(requestID)
		}
		if requestID == "" {
			// Generate a secure random ID if missing or became empty after sanitization
			b := make([]byte, 16)
			rand.Read(b)
			requestID = fmt.Sprintf("req-%x", b)
		}

		// Inyectar en contexto con key tipada
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		r = r.WithContext(ctx)

		// Propagar en response header
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r)
	})
}

// FromContext extrae el request ID del contexto.
func FromContext(ctx context.Context) string {
	val := ctx.Value(RequestIDKey)
	if val == nil {
		return ""
	}
	if id, ok := val.(string); ok {
		return id
	}
	return ""
}
