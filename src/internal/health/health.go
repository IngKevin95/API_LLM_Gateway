// Package health implementa el Health Monitor: sondea proveedores periódicamente
// con histéresis para evitar oscilación, y expone el estado de salud vivo como
// HealthSource del router (reemplaza el StaticHealth de EP-001).
package health

import (
	"context"
	"sync"
	"time"
)

// Prober sondea un proveedor y devuelve si respondió sano.
type Prober func(provider string) bool

type provState struct {
	healthy bool
	fails   int
	oks     int

	// Retiro temporal por 429 (HU-EVO-004): retiredUntil > now() => no-sano
	// independientemente del resultado del probe periódico.
	retiredUntil time.Time
	backoffStep  int // fallos consecutivos por 429 sin Retry-After, para backoff exponencial
}

const (
	default429Retire    = 30 * time.Second
	maxBackoff429       = 120 * time.Second
)

// Monitor sondea proveedores y mantiene su estado de salud (seguro concurrente).
type Monitor struct {
	mu            sync.RWMutex
	state         map[string]*provState
	providers     []string
	probe         Prober
	failThreshold int
	okThreshold   int
	now           func() time.Time
}

// New crea un Monitor. Los proveedores arrancan sanos; se necesitan
// failThreshold fallos consecutivos para retirar y okThreshold éxitos para reactivar.
func New(providers []string, probe Prober, failThreshold, okThreshold int) *Monitor {
	if failThreshold < 1 {
		failThreshold = 1
	}
	if okThreshold < 1 {
		okThreshold = 1
	}
	m := &Monitor{
		state:         make(map[string]*provState, len(providers)),
		providers:     providers,
		probe:         probe,
		failThreshold: failThreshold,
		okThreshold:   okThreshold,
		now:           time.Now,
	}
	for _, p := range providers {
		m.state[p] = &provState{healthy: true}
	}
	return m
}

// Healthy implementa router.HealthSource. Un proveedor desconocido se asume sano.
// Un proveedor retirado temporalmente por 429 (HU-EVO-004) es no-sano hasta que
// vence retiredUntil, sin esperar al próximo CheckOnce (AC1: reactivación al
// primer chequeo posterior a vencer el retiro).
func (m *Monitor) Healthy(providerID, _ string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s := m.state[providerID]
	if s == nil {
		return true
	}
	if !s.retiredUntil.IsZero() {
		if m.clock().Before(s.retiredUntil) {
			return false
		}
		return true // AC1/AC2/AC3: reactivación automática al vencer el retiro
	}
	return s.healthy
}

// RetireOn429 retira temporalmente providerID tras recibir un 429 (HU-EVO-004
// AC1/AC2/AC3). retryAfter es la duración indicada por el header Retry-After
// del proveedor; si es <=0 se usa el default de 30s (AC3). Fallos consecutivos
// sin Retry-After escalan con backoff exponencial 30s→60s→120s, con tope
// configurable (AC5); un retiro con Retry-After explícito no incrementa el backoff.
func (m *Monitor) RetireOn429(providerID string, retryAfter time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.state[providerID]
	if s == nil {
		s = &provState{healthy: true}
		m.state[providerID] = s
	}

	wait := retryAfter
	if wait <= 0 {
		s.backoffStep++
		wait = backoffDuration(s.backoffStep)
	}
	s.retiredUntil = m.clock().Add(wait)
	s.healthy = false
}

// backoffDuration calcula el retiro por defecto ante 429 repetidos sin
// Retry-After: 30s → 60s → 120s, con tope en maxBackoff429.
func backoffDuration(step int) time.Duration {
	if step < 1 {
		step = 1
	}
	d := default429Retire * time.Duration(1<<uint(step-1))
	if d > maxBackoff429 {
		d = maxBackoff429
	}
	return d
}

// clock devuelve el reloj del Monitor (time.Now por defecto; inyectable en tests).
func (m *Monitor) clock() time.Time {
	if m.now != nil {
		return m.now()
	}
	return time.Now()
}

// SetClock inyecta un reloj determinista (uso exclusivo de tests).
func (m *Monitor) SetClock(clock func() time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.now = clock
}

// CheckOnce sondea todos los proveedores una vez y aplica la histéresis.
func (m *Monitor) CheckOnce() {
	for _, p := range m.providers {
		ok := m.probe(p)
		m.mu.Lock()
		s := m.state[p]
		if ok {
			s.oks++
			s.fails = 0
			if s.oks >= m.okThreshold {
				s.healthy = true
				s.retiredUntil = time.Time{} // AC1: health check 200 limpia el retiro por 429
				s.backoffStep = 0
			}
		} else {
			s.fails++
			s.oks = 0
			if s.fails >= m.failThreshold {
				s.healthy = false
			}
		}
		m.mu.Unlock()
	}
}

// Run sondea en bucle cada `interval` hasta que el contexto se cancele.
func (m *Monitor) Run(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			m.CheckOnce()
		}
	}
}
