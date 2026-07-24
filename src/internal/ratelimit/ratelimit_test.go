package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"api-llm-gateway/internal/ratelimit"
)

func ok() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
}

func keyByHeader(r *http.Request) string { return r.Header.Get("X-Client") }

// HU-022 AC1/AC2 — dentro del límite pasa; excedido → 429.
func TestRateLimit_WindowLimit(t *testing.T) {
	lim := ratelimit.NewLimiter(3, time.Minute)
	h := ratelimit.Middleware(lim, keyByHeader)(ok())
	do := func() int {
		req := httptest.NewRequest(http.MethodPost, "/v1/chat", nil)
		req.Header.Set("X-Client", "c1")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}
	for i := 0; i < 3; i++ {
		if code := do(); code != http.StatusOK {
			t.Fatalf("petición %d dentro del límite esperaba 200, obtuve %d", i+1, code)
		}
	}
	if code := do(); code != http.StatusTooManyRequests {
		t.Errorf("la 4a debe ser 429, obtuve %d", code)
	}
}

// HU-022 AC4 — validación atómica: 10 concurrentes con 1 token → 1 pasa (correr con -race).
func TestRateLimit_ConcurrentAtomic(t *testing.T) {
	lim := ratelimit.NewLimiter(1, time.Minute)
	var allowed int
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if lim.Allow("c1") {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if allowed != 1 {
		t.Errorf("con 1 token y 10 concurrentes esperaba 1 permitido, obtuve %d", allowed)
	}
}

// HU-022 AC5 — payload de texto excede el límite → 413 por Content-Length, sin leer el cuerpo.
func TestPayload_ExceedsLimit(t *testing.T) {
	limitFor := func(_ *http.Request) int64 { return 100 } // límite pequeño para el test
	h := ratelimit.PayloadMiddleware(limitFor)(ok())

	big := strings.NewReader(strings.Repeat("x", 500))
	req := httptest.NewRequest(http.MethodPost, "/v1/chat", big)
	req.ContentLength = 500
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("payload sobre el límite esperaba 413, obtuve %d", rec.Code)
	}
}

// HU-022 AC3/AC5 — dentro del límite pasa.
func TestPayload_WithinLimit(t *testing.T) {
	limitFor := func(_ *http.Request) int64 { return 1000 }
	h := ratelimit.PayloadMiddleware(limitFor)(ok())
	req := httptest.NewRequest(http.MethodPost, "/v1/chat", strings.NewReader("hola"))
	req.ContentLength = 4
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("payload dentro del límite esperaba 200, obtuve %d", rec.Code)
	}
}

// HU-022b AC1/AC2 — 2 vision concurrentes OK, 3ra → 429 (correr con -race).
func TestVisionConcurrency(t *testing.T) {
	vl := ratelimit.NewVisionLimiter(2)
	isVision := func(r *http.Request) bool { return r.Header.Get("X-Capability") == "vision" }
	// handler que retiene el slot hasta liberar
	release := make(chan struct{})
	blocking := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		<-release
	})
	h := ratelimit.VisionMiddleware(vl, isVision)(blocking)

	visReq := func() *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/v1/chat", nil)
		req.Header.Set("X-Capability", "vision")
		return req
	}
	var wg sync.WaitGroup
	codes := make([]int, 2)
	for i := 0; i < 2; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, visReq())
			codes[i] = rec.Code
		}()
	}
	time.Sleep(30 * time.Millisecond) // deja que ocupen los 2 slots
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, visReq()) // 3ra
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("3ra vision concurrente esperaba 429, obtuve %d", rec.Code)
	}
	close(release)
	wg.Wait()
}
