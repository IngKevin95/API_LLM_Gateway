package cost

import (
	"fmt"
	"sync"
)

type CostRecord struct {
	AgentID      string
	ProviderID   string
	ModelID      string
	InputTokens  int64
	OutputTokens int64
	CostPerToken float64
}

// Tracker registra el costo de peticiones por agente y proveedor.
type Tracker interface {
	RecordCost(record CostRecord) error
	GetTotalCost(agentID, providerID string) (float64, error)
}

// LocalCostTracker almacena costos en memoria.
// ponytail: costos en RAM, persistencia diferida a EP-009 (sync-worker).
type LocalCostTracker struct {
	mu    sync.RWMutex
	costs map[string]map[string]float64 // [agentID][providerID] = totalCost
}

func NewLocalCostTracker() *LocalCostTracker {
	return &LocalCostTracker{
		costs: make(map[string]map[string]float64),
	}
}

// RecordCost registra el costo de una petición completada.
// AC1: Registra costo = (inputTokens + outputTokens) × costPerToken, atribuido a (agente, proveedor, modelo).
func (t *LocalCostTracker) RecordCost(record CostRecord) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.costs[record.AgentID] == nil {
		t.costs[record.AgentID] = make(map[string]float64)
	}

	// AC1: Calcular costo = tokens × tarifa
	totalTokens := float64(record.InputTokens + record.OutputTokens)
	recordCost := totalTokens * record.CostPerToken

	// Acumular al total
	t.costs[record.AgentID][record.ProviderID] += recordCost

	return nil
}

// GetTotalCost retorna el costo acumulado para un agente y proveedor.
func (t *LocalCostTracker) GetTotalCost(agentID, providerID string) (float64, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.costs[agentID] == nil {
		return 0, fmt.Errorf("no cost records for agent %s", agentID)
	}

	cost, ok := t.costs[agentID][providerID]
	if !ok {
		return 0, fmt.Errorf("no cost records for agent %s and provider %s", agentID, providerID)
	}

	return cost, nil
}
