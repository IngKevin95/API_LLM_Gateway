package quota

import (
	"fmt"
	"sync"
	"time"
)

type Window struct {
	Period   string    // "daily", "monthly", "hourly"
	Limit    int64     // límite de requests/tokens
	Consumed int64     // consumo actual
	ResetAt  time.Time // cuándo se reinicia la ventana
}

// Manager define la interfaz para gestión de cuotas.
type Manager interface {
	TryConsume(providerID, apiKey string, amount int64) (allowed bool, remaining int64, err error)
	GetQuota(providerID, apiKey string) (*Window, error)
	SetQuota(providerID, apiKey string, w Window) error
}

// LocalQuotaManager implementa gestión de cuota en RAM con lock atómico.
// ponytail: contadores en RAM, persistencia diferida a EP-009 (sync-worker).
type LocalQuotaManager struct {
	mu     sync.RWMutex
	quotas map[string]map[string]*Window // [providerID][apiKey]
}

const (
	Daily   = "daily"
	Monthly = "monthly"
	Hourly  = "hourly"
)

func NewLocalQuotaManager() *LocalQuotaManager {
	return &LocalQuotaManager{
		quotas: make(map[string]map[string]*Window),
	}
}

func (m *LocalQuotaManager) SetQuota(providerID, apiKey string, w Window) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.quotas[providerID] == nil {
		m.quotas[providerID] = make(map[string]*Window)
	}

	m.quotas[providerID][apiKey] = &w
	return nil
}

func (m *LocalQuotaManager) GetQuota(providerID, apiKey string) (*Window, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.quotas[providerID] == nil {
		return nil, fmt.Errorf("provider %s not found", providerID)
	}

	w, ok := m.quotas[providerID][apiKey]
	if !ok {
		return nil, fmt.Errorf("quota for %s:%s not found", providerID, apiKey)
	}

	// Retornar copia para evitar modificaciones externas
	copy := *w
	return &copy, nil
}

// TryConsume valida y decrementa cuota de forma atómica.
// AC1: Dentro de cuota → allowed=true, remaining actualizado, consumed incrementado.
// AC2: Cuota agotada → allowed=false.
// AC3: Reinicio de ventana — si ResetAt pasó, reinicia Consumed.
// AC4: Límite por tokens — rechaza si amount excedería.
// AC5: Race conditions — validación atómica via mutex.
func (m *LocalQuotaManager) TryConsume(providerID, apiKey string, amount int64) (bool, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.quotas[providerID] == nil {
		return false, 0, fmt.Errorf("provider %s not found", providerID)
	}

	w, ok := m.quotas[providerID][apiKey]
	if !ok {
		return false, 0, fmt.Errorf("quota for %s:%s not found", providerID, apiKey)
	}

	// AC3: Reinicio de ventana — si ResetAt pasó, reiniciar contador
	if time.Now().After(w.ResetAt) {
		w.Consumed = 0
		w.ResetAt = computeNextReset(w.Period)
	}

	// AC4 + AC1-AC2: Validar si se puede consumir
	remaining := w.Limit - w.Consumed

	// AC2: Cuota agotada
	if remaining <= 0 {
		return false, 0, nil
	}

	// AC4: Límite por tokens — rechaza si amount excedería
	if amount > remaining {
		return false, remaining, nil
	}

	// AC1: Dentro de cuota — consumir y retornar remaining
	w.Consumed += amount
	newRemaining := w.Limit - w.Consumed
	if newRemaining < 0 {
		newRemaining = 0
	}

	return true, newRemaining, nil
}

// computeNextReset calcula la próxima hora de reset según el período.
func computeNextReset(period string) time.Time {
	now := time.Now()
	switch period {
	case Daily:
		return now.AddDate(0, 0, 1)
	case Monthly:
		return now.AddDate(0, 1, 0)
	case Hourly:
		return now.Add(time.Hour)
	default:
		return now.AddDate(0, 0, 1) // default: daily
	}
}
