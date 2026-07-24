package quota

import (
	"sort"
	"sync"
	"time"

	"api-llm-gateway/internal/adapter"
)

type Consumption struct {
	Tokens   int
	Requests int
}

type Manager interface {
	Reserve(providerID string, estimate Consumption) bool
	Commit(providerID string, estimate Consumption, actual Consumption) error
	Remaining(providerID, model string) int
	LearnFromHeaders(providerID, modelID string, quota adapter.QuotaInfo)
	Snapshot() []Snapshot
}

// Snapshot describe la cuota remanente de un proveedor/modelo en un instante
// dado (lectura pura desde el mapa en RAM, sin I/O). La consumen
// metrics.Store (HU-EVO-011, expuesto en /metrics) y alert.Manager
// (HU-EVO-012, evaluado contra el umbral configurable).
type Snapshot struct {
	Provider  string
	Model     string // "" cuando el proveedor no tiene desglose por modelo aún (solo quota_hint)
	Limit     int64
	Remaining int64
	ResetAt   *time.Time
	Healthy   bool       // false cuando Remaining == 0
	LearnedAt *time.Time // nil si nunca aprendió de headers de respuesta (usa quota_hint inicial)
}

type providerState struct {
	limit            Consumption
	used             Consumption
	window           string // "YYYY-MM-DD"
	resetAt          *time.Time
	learnedRemaining *int64 // nil = calculate from limit-used; otherwise use this value
}

// modelState guarda el desglose de cuota aprendida por modelo individual
// (HU-EVO-011 AC2/AC5: un proveedor con varios modelos expone remaining
// independiente por cada uno, poblado al aprender headers reales).
type modelState struct {
	limit     int64
	remaining int64
	resetAt   *time.Time
	learnedAt *time.Time
}

type inMemoryManager struct {
	mu        sync.RWMutex
	states    map[string]*providerState
	perModel  map[string]map[string]*modelState // providerID -> modelID -> state
	clock     func() time.Time
	persister Persister // optional async persistence (HU-EVO-008)
}

func NewInMemoryManager() *inMemoryManager {
	return NewInMemoryManagerWithClock(time.Now)
}

func NewInMemoryManagerWithClock(clock func() time.Time) *inMemoryManager {
	return &inMemoryManager{
		states:    make(map[string]*providerState),
		perModel:  make(map[string]map[string]*modelState),
		clock:     clock,
		persister: &NoPersister{}, // default: no persistence
	}
}

func NewInMemoryManagerWithPersister(clock func() time.Time, persister Persister) *inMemoryManager {
	return &inMemoryManager{
		states:    make(map[string]*providerState),
		perModel:  make(map[string]map[string]*modelState),
		clock:     clock,
		persister: persister,
	}
}

// DefaultQuotaHint es la cuota inicial asumida cuando un proveedor no declara
// quota_hint en YAML (HU-EVO-005 AC4): 1M, un piso conservador de free tier.
const DefaultQuotaHint = 1_000_000

// InitFromRegistry inicializa `remaining` por proveedor desde quota_hint
// (HU-EVO-005 AC1/AC3/AC4). hints es providerID -> *quota_hint (nil = ausente
// en YAML -> aplica DefaultQuotaHint; <=0 -> se trata como agotado, remaining=0).
// Solo inicializa providers que aún no tengan estado (no pisa un valor ya
// aprendido en runtime o restaurado desde persistencia, AC2/AC5).
func (m *inMemoryManager) InitFromRegistry(hints map[string]*int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, hint := range hints {
		if _, exists := m.states[id]; exists {
			continue // ya inicializado (ej. restaurado desde PostgreSQL, AC5)
		}
		quota := DefaultQuotaHint
		if hint != nil {
			if *hint <= 0 {
				quota = 0 // AC3: quota_hint <=0 -> agotado
			} else {
				quota = *hint
			}
		}
		m.states[id] = &providerState{
			limit:  Consumption{Tokens: quota},
			window: currentWindow(m.clock()),
		}
	}
}

// RestoreRemaining fija `remaining` para providerID a un valor restaurado
// desde persistencia (ej. PostgreSQL en boot), con precedencia sobre
// quota_hint y sobre cualquier init previo (HU-EVO-005 AC5). Debe llamarse
// antes de InitFromRegistry para que este último respete el valor restaurado.
func (m *inMemoryManager) RestoreRemaining(providerID string, remaining int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.states[providerID] = &providerState{
		limit:  Consumption{Tokens: remaining},
		window: currentWindow(m.clock()),
	}
}

func (m *inMemoryManager) SetLimit(providerID string, limit Consumption) {
	m.mu.Lock()
	defer m.mu.Unlock()
	state, exists := m.states[providerID]
	if !exists {
		state = &providerState{
			window: currentWindow(m.clock()),
		}
		m.states[providerID] = state
	}
	state.limit = limit
}

func (m *inMemoryManager) Reserve(providerID string, estimate Consumption) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.states[providerID]
	if !exists {
		return false // No limit configured, or default? Let's assume false if no limit.
	}

	window := currentWindow(m.clock())
	if state.window != window {
		state.window = window
		state.used = Consumption{} // Reset
	}

	if state.used.Tokens+estimate.Tokens > state.limit.Tokens ||
		state.used.Requests+estimate.Requests > state.limit.Requests {
		return false
	}

	// Pre-deduct
	state.used.Tokens += estimate.Tokens
	state.used.Requests += estimate.Requests
	return true
}

func (m *inMemoryManager) Commit(providerID string, estimate Consumption, actual Consumption) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.states[providerID]
	if !exists {
		return nil // Nothing to commit if no state
	}

	window := currentWindow(m.clock())
	if state.window != window {
		// Window changed before commit. The new window just started.
		// We'll apply it to the new window to be safe against leaks.
		state.window = window
		state.used = Consumption{}
	}

	// Adjust the usage: we pre-deducted estimate, now we replace it with actual.
	state.used.Tokens += actual.Tokens - estimate.Tokens
	state.used.Requests += actual.Requests - estimate.Requests
	return nil
}

func (m *inMemoryManager) Remaining(providerID, model string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.states[providerID]
	if !exists {
		return 999999999
	}

	if state.learnedRemaining != nil {
		if *state.learnedRemaining < 0 {
			return 0
		}
		return int(*state.learnedRemaining)
	}

	window := currentWindow(m.clock())
	if state.window != window {
		return state.limit.Tokens
	}

	remaining := state.limit.Tokens - state.used.Tokens
	if remaining < 0 {
		return 0
	}
	return remaining
}

// LearnFromHeaders learns quota from provider response headers (HU-EVO-007)
func (m *inMemoryManager) LearnFromHeaders(providerID, modelID string, quota adapter.QuotaInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.states[providerID]
	if !exists {
		state = &providerState{
			window: currentWindow(m.clock()),
		}
		m.states[providerID] = state
	}

	// Detect reset: ResetAt changed from previous value
	if quota.ResetAt != nil && state.resetAt != nil {
		if quota.ResetAt.After(*state.resetAt) {
			// Window reset detected: update limit from learned value
			if quota.Limit > 0 {
				state.limit = Consumption{Tokens: int(quota.Limit)}
			}
		}
	}

	// Update limit if learned value is provided and higher than current
	if quota.Limit > int64(state.limit.Tokens) {
		state.limit = Consumption{Tokens: int(quota.Limit)}
	}

	// Clamp remaining to 0 if negative
	remaining := quota.Remaining
	if remaining < 0 {
		remaining = 0
	}

	// Learn remaining from header
	state.learnedRemaining = &remaining
	state.resetAt = quota.ResetAt

	// HU-EVO-011 AC2: desglose por modelo individual, además del agregado por
	// proveedor. modelID vacío no genera entrada de modelo (se sigue exponiendo
	// solo el agregado del proveedor en Snapshot()).
	if modelID != "" {
		if m.perModel == nil {
			m.perModel = make(map[string]map[string]*modelState)
		}
		if m.perModel[providerID] == nil {
			m.perModel[providerID] = make(map[string]*modelState)
		}
		limit := state.limit.Tokens
		if quota.Limit > 0 {
			limit = int(quota.Limit)
		}
		now := m.clock()
		m.perModel[providerID][modelID] = &modelState{
			limit:     int64(limit),
			remaining: remaining,
			resetAt:   quota.ResetAt,
			learnedAt: &now,
		}
	}

	// Enqueue async persistence (HU-EVO-008)
	if m.persister != nil {
		go func() {
			_ = m.persister.Enqueue(PersistJob{
				ProviderID: providerID,
				ModelID:    modelID,
				Quota:      quota,
			})
		}()
	}
}

// Snapshot devuelve el estado de cuota actual de todos los proveedores
// conocidos (HU-EVO-011). Lectura pura desde el mapa en RAM (sin I/O), por
// eso puede llamarse en cada request de /metrics sin impacto de latencia
// (AC4: <100ms con 125 cuotas). Cuando un proveedor tiene desglose por
// modelo (aprendido de headers reales, AC2), se listan sus modelos
// individuales; si no, se expone un único registro agregado con
// model:"" y remaining/limit del quota_hint inicial (AC3), con
// LearnedAt:nil.
func (m *inMemoryManager) Snapshot() []Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]Snapshot, 0, len(m.states))
	for providerID, st := range m.states {
		models := m.perModel[providerID]
		if len(models) == 0 {
			remaining := m.remainingLocked(st)
			out = append(out, Snapshot{
				Provider:  providerID,
				Model:     "",
				Limit:     int64(st.limit.Tokens),
				Remaining: int64(remaining),
				ResetAt:   st.resetAt,
				Healthy:   remaining > 0,
				LearnedAt: nil,
			})
			continue
		}
		for modelID, ms := range models {
			out = append(out, Snapshot{
				Provider:  providerID,
				Model:     modelID,
				Limit:     ms.limit,
				Remaining: ms.remaining,
				ResetAt:   ms.resetAt,
				Healthy:   ms.remaining > 0,
				LearnedAt: ms.learnedAt,
			})
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Provider != out[j].Provider {
			return out[i].Provider < out[j].Provider
		}
		return out[i].Model < out[j].Model
	})
	return out
}

// remainingLocked replica la lógica de Remaining() sin re-adquirir el lock
// (el llamador ya sostiene m.mu). Evita el deadlock de RLock reentrante.
func (m *inMemoryManager) remainingLocked(state *providerState) int {
	if state.learnedRemaining != nil {
		if *state.learnedRemaining < 0 {
			return 0
		}
		return int(*state.learnedRemaining)
	}
	window := currentWindow(m.clock())
	if state.window != window {
		return state.limit.Tokens
	}
	remaining := state.limit.Tokens - state.used.Tokens
	if remaining < 0 {
		return 0
	}
	return remaining
}

func currentWindow(t time.Time) string {
	return t.Format("2006-01-02")
}
