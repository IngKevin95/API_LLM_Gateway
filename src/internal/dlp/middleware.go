package dlp

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"
)

type contextKey string

const scannerKey contextKey = "dlp_scanner"

// StreamScannerFunc define la firma esperada para la inspección asíncrona.
type StreamScannerFunc func(ctx context.Context, stream io.Reader, cancelFunc context.CancelFunc)

// ScannerFromContext extrae la función de inspección asíncrona inyectada por el middleware.
func ScannerFromContext(ctx context.Context) StreamScannerFunc {
	v := ctx.Value(scannerKey)
	if fn, ok := v.(StreamScannerFunc); ok {
		return fn
	}
	return nil
}

// Middleware devuelve un http.Handler interceptor que lee el body del request,
// lo pasa por el DLPEngine para su redacción, y luego lo reinyecta en el request.
// Además inyecta en el context la función ScanAsync para que los adaptadores puedan invocar el kill-switch.
func Middleware(engine *Engine, timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Inyectar el kill-switch (ScanAsync) en el context para que el adaptador (downstream) lo use
			r = r.WithContext(context.WithValue(r.Context(), scannerKey, StreamScannerFunc(engine.ScanAsync)))

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
