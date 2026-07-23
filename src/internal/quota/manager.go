package quota

import (
	"sync"
	"time"
)

type Consumption struct {
	Tokens   int
	Requests int
}

type Manager interface {
	Reserve(providerID string, estimate Consumption) bool
	Commit(providerID string, estimate Consumption, actual Consumption) error
	Remaining(providerID, model string) int
}

type providerState struct {
	limit   Consumption
	used    Consumption
	window  string // "YYYY-MM-DD"
}

type inMemoryManager struct {
	mu     sync.RWMutex
	states map[string]*providerState
	clock  func() time.Time
}

func NewInMemoryManager() *inMemoryManager {
	return NewInMemoryManagerWithClock(time.Now)
}

func NewInMemoryManagerWithClock(clock func() time.Time) *inMemoryManager {
	return &inMemoryManager{
		states: make(map[string]*providerState),
		clock:  clock,
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
		// If no limit is configured, we can assume infinite or 0. 
		// The Router expects remaining > 0 to use it.
		// Let's assume math.MaxInt32 if no limit.
		return 999999999
	}
	
	window := currentWindow(m.clock())
	if state.window != window {
		// New window has full limit
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
