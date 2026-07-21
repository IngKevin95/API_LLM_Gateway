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
}

// Monitor sondea proveedores y mantiene su estado de salud (seguro concurrente).
type Monitor struct {
	mu            sync.RWMutex
	state         map[string]*provState
	providers     []string
	probe         Prober
	failThreshold int
	okThreshold   int
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
	}
	for _, p := range providers {
		m.state[p] = &provState{healthy: true}
	}
	return m
}

// Healthy implementa router.HealthSource. Un proveedor desconocido se asume sano.
func (m *Monitor) Healthy(providerID, _ string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s := m.state[providerID]
	return s == nil || s.healthy
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
