// Package ratelimit provee defensa de abuso en RAM: rate limiting por ventana,
// límites de payload por tipo y concurrencia de vision por nodo. Cero red.
package ratelimit

import (
	"net/http"
	"sync"
	"time"
)

// Limiter aplica un límite de N peticiones por ventana temporal por clave,
// con validación atómica en RAM. Seguro concurrente.
type Limiter struct {
	mu     sync.Mutex
	max    int
	window time.Duration
	state  map[string]*counter
	now    func() time.Time
}

type counter struct {
	windowStart time.Time
	count       int
}

// NewLimiter crea un limitador de `max` peticiones por `window`.
func NewLimiter(max int, window time.Duration) *Limiter {
	return &Limiter{max: max, window: window, state: make(map[string]*counter), now: time.Now}
}

// Allow registra una petición y devuelve si está dentro del límite (atómico).
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	c := l.state[key]
	if c == nil || now.Sub(c.windowStart) >= l.window {
		l.state[key] = &counter{windowStart: now, count: 1}
		return true
	}
	if c.count >= l.max {
		return false
	}
	c.count++
	return true
}

// Middleware responde 429 (sin llamar al handler) al exceder el límite.
func Middleware(l *Limiter, key func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !l.Allow(key(r)) {
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// PayloadMiddleware rechaza con 413 los payloads que exceden el límite del
// tipo de contenido (10MB texto / 50MB vision, resuelto por `limitFor`), sin
// cargar el cuerpo (chequea Content-Length y envuelve con MaxBytesReader).
func PayloadMiddleware(limitFor func(*http.Request) int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limit := limitFor(r)
			if r.ContentLength > limit {
				http.Error(w, `{"error":"payload too large"}`, http.StatusRequestEntityTooLarge)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, limit) // defensa si Content-Length miente
			next.ServeHTTP(w, r)
		})
	}
}

// VisionLimiter limita la concurrencia de peticiones vision por nodo.
type VisionLimiter struct {
	slots chan struct{}
}

// NewVisionLimiter crea un limitador de `max` peticiones vision concurrentes.
func NewVisionLimiter(max int) *VisionLimiter {
	return &VisionLimiter{slots: make(chan struct{}, max)}
}

// TryAcquire toma un slot sin bloquear; false si está lleno.
func (v *VisionLimiter) TryAcquire() bool {
	select {
	case v.slots <- struct{}{}:
		return true
	default:
		return false
	}
}

// Release libera un slot.
func (v *VisionLimiter) Release() { <-v.slots }

// VisionMiddleware limita la concurrencia de vision (429 sin encolar la 3ra).
func VisionMiddleware(v *VisionLimiter, isVision func(*http.Request) bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isVision(r) {
				next.ServeHTTP(w, r)
				return
			}
			if !v.TryAcquire() {
				http.Error(w, `{"error":"vision concurrency exceeded"}`, http.StatusTooManyRequests)
				return
			}
			defer v.Release()
			next.ServeHTTP(w, r)
		})
	}
}
