package learning

import "log"

type ModelStats struct {
	Model       string
	Samples     int
	AvgLatency  float64
	SuccessRate float64
}

type MetricProvider interface {
	GetRecentStats() ([]ModelStats, error)
}

type WeightUpdater interface {
	UpdateWeights(weights map[string]float64) error
}

type Engine struct {
	provider MetricProvider
	updater  WeightUpdater
	prev     map[string]float64
}

func NewEngine(provider MetricProvider, updater WeightUpdater) *Engine {
	return &Engine{
		provider: provider,
		updater:  updater,
		prev:     make(map[string]float64),
	}
}

func (e *Engine) SetPreviousWeights(w map[string]float64) {
	e.prev = w
}

func (e *Engine) Evaluate() error {
	stats, err := e.provider.GetRecentStats()
	if err != nil {
		return err
	}

	if len(stats) == 0 {
		return nil
	}

	// 1. Check for rollback (degradation)
	// If any model has < 0.90 success rate and we have previous weights, rollback
	degraded := false
	for _, s := range stats {
		if s.Samples >= 100 && s.SuccessRate < 0.90 {
			degraded = true
			break
		}
	}

	if degraded && len(e.prev) > 0 {
		log.Println("Learning Engine: Degradation detected. Rolling back weights.")
		return e.updater.UpdateWeights(e.prev)
	}

	// 2. Validate sample size
	// If the total sample or any sample is too low, we skip adjusting
	validSamples := 0
	for _, s := range stats {
		if s.Samples >= 100 {
			validSamples++
		}
	}

	// Require at least one model to have sufficient data to adjust anything
	if validSamples == 0 {
		return nil
	}

	// 3. Heuristic weight calculation based on latency
	// Inverse proportion to latency (faster gets more weight)
	weights := make(map[string]float64)
	var totalInvLat float64

	for _, s := range stats {
		if s.Samples < 100 {
			// fallback weight if no data
			weights[s.Model] = 0.1
			totalInvLat += 0.1
			continue
		}
		invLat := 1000.0 / s.AvgLatency // Ex: 100ms -> 10, 500ms -> 2
		weights[s.Model] = invLat
		totalInvLat += invLat
	}

	// Normalize and apply guardrails
	maxWeight := 0.8 // Max 80% to a single model to preserve diversity
	for m, w := range weights {
		norm := w / totalInvLat
		if norm > maxWeight {
			norm = maxWeight
		}
		weights[m] = norm
	}

	// Save prev before updating
	// (In a real system, we'd query the current weights first)

	return e.updater.UpdateWeights(weights)
}
