package middleware

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

// RequestID es un middleware que extrae o genera un request ID y lo añade al contexto.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Intentar extraer de header
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			// Generar un ID aleatorio simple si no existe
			requestID = fmt.Sprintf("req-%d-%d", time.Now().UnixNano(), rand.Intn(100000))
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
