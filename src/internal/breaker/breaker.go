// Package breaker implementa un Circuit Breaker pasivo con Max In-Flight por
// proveedor, para prevenir Failover Suicide de peticiones concurrentes.
package breaker

import (
	"sync"
	"time"
)

// Config parametriza el breaker.
type Config struct {
	MaxInFlight int           // tope de peticiones concurrentes por proveedor
	Cooldown    time.Duration // tiempo inalcanzable tras un Trip
}

type provState struct {
	inFlight     int
	trippedUntil time.Time
}

// Breaker es seguro para uso concurrente.
type Breaker struct {
	mu    sync.Mutex
	cfg   Config
	state map[string]*provState
	now   func() time.Time
}

// New crea un breaker con la configuración dada.
func New(cfg Config) *Breaker {
	if cfg.MaxInFlight <= 0 {
		cfg.MaxInFlight = 1
	}
	return &Breaker{cfg: cfg, state: make(map[string]*provState), now: time.Now}
}

func (b *Breaker) get(provider string) *provState {
	s := b.state[provider]
	if s == nil {
		s = &provState{}
		b.state[provider] = s
	}
	return s
}

// Allow intenta reservar un slot in-flight. Devuelve false (fast-fail, 0 I/O) si
// el proveedor está tripped o alcanzó el Max In-Flight.
func (b *Breaker) Allow(provider string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	s := b.get(provider)
	if b.now().Before(s.trippedUntil) {
		return false
	}
	if s.inFlight >= b.cfg.MaxInFlight {
		return false
	}
	s.inFlight++
	return true
}

// Release libera un slot in-flight.
func (b *Breaker) Release(provider string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	s := b.get(provider)
	if s.inFlight > 0 {
		s.inFlight--
	}
}

// Trip marca al proveedor inalcanzable durante el cooldown.
func (b *Breaker) Trip(provider string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.get(provider).trippedUntil = b.now().Add(b.cfg.Cooldown)
}

// Available indica si el proveedor no está tripped (fuera del cooldown).
func (b *Breaker) Available(provider string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return !b.now().Before(b.get(provider).trippedUntil)
}
