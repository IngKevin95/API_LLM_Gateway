package cost

import "context"

type CostRecord struct {
	AgentID          string
	ProviderID       string
	Model            string
	PromptTokens     int
	CompletionTokens int
	Cost             float64
}

// ModelFinder abstracta la búsqueda de modelos en el Registry para obtener su costo.
type ModelFinder interface {
	CostPer1M(modelName string) (int, bool)
}

// Sink es una función a la que el Tracker enviará el registro finalizado.
// Usualmente apunta al WAL o un logger.
type Sink func(ctx context.Context, record CostRecord) error

type Tracker struct {
	finder ModelFinder
	sink   Sink
}

func NewTracker(finder ModelFinder, sink Sink) *Tracker {
	return &Tracker{
		finder: finder,
		sink:   sink,
	}
}

func (t *Tracker) Track(ctx context.Context, record CostRecord) error {
	costPer1M, exists := t.finder.CostPer1M(record.Model)
	if exists {
		totalTokens := record.PromptTokens + record.CompletionTokens
		record.Cost = float64(totalTokens) * float64(costPer1M) / 1_000_000.0
	} else {
		record.Cost = 0.0
	}

	return t.sink(ctx, record)
}
