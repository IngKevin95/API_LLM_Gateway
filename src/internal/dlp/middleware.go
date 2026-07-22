package dlp

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"
)

// Middleware devuelve un http.Handler interceptor que lee el body del request,
// lo pasa por el DLPEngine para su redacción, y luego lo reinyecta en el request.
// Aplica un timeout máximo estricto para evitar degradación de latencia.
func Middleware(engine *Engine, timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body == nil {
				next.ServeHTTP(w, r)
				return
			}

			payload, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "bad request body", http.StatusBadRequest)
				return
			}
			_ = r.Body.Close()

			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			redacted, err := engine.Redact(ctx, payload)
			if err != nil {
				// PRD: aborta con HTTP 500 en redacción síncrona si falla/timeout
				http.Error(w, "DLP timeout: redaction took too long", http.StatusInternalServerError)
				return
			}

			// Reemplazar el body con el contenido redactado
			r.Body = io.NopCloser(bytes.NewBuffer(redacted))
			r.ContentLength = int64(len(redacted))

			next.ServeHTTP(w, r)
		})
	}
}
