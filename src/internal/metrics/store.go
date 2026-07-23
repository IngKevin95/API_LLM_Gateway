package metrics

import (
	"context"
	"sort"
	"sync"
	"time"
)

// RequestMetric registra una métrica de request individual.
type RequestMetric struct {
	Provider   string
	Model      string
	LatencyMs  int64
	Status     int
	Tokens     int
	Cost       float64
	Timestamp  time.Time
}

// InMemoryStore es un almacén de métricas en memoria con rolling window.
type InMemoryStore struct {
	mu       sync.RWMutex
	metrics  []RequestMetric
	maxSize  int
	startIdx int
}

// NewInMemoryStore crea un store con tamaño máximo (últimas N métricas).
func NewInMemoryStore(maxSize int) *InMemoryStore {
	return &InMemoryStore{
		metrics: make([]RequestMetric, 0, maxSize),
		maxSize: maxSize,
	}
}

// Record registra una métrica individual.
func (s *InMemoryStore) Record(m RequestMetric) {
	s.mu.Lock()
	defer s.mu.Unlock()

	m.Timestamp = time.Now()
	if len(s.metrics) < s.maxSize {
		s.metrics = append(s.metrics, m)
	} else {
		// Rolling window: sobrescribir el más antiguo
		s.metrics[s.startIdx] = m
		s.startIdx = (s.startIdx + 1) % s.maxSize
	}
}

// GetMetrics devuelve agregados de métricas por modelo.
func (s *InMemoryStore) GetMetrics(ctx context.Context, providerFilter string) ([]ModelMetric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.metrics) == 0 {
		return []ModelMetric{}, nil
	}

	// Agrupar por (provider, model)
	groups := make(map[string]*aggregator)

	for _, m := range s.metrics {
		// Filtrar por proveedor si se especifica
		if providerFilter != "" && m.Provider != providerFilter {
			continue
		}

		key := m.Provider + ":" + m.Model
		if _, ok := groups[key]; !ok {
			groups[key] = &aggregator{
				provider: m.Provider,
				model:    m.Model,
			}
		}

		agg := groups[key]
		agg.latencies = append(agg.latencies, m.LatencyMs)
		agg.totalTokens += m.Tokens
		agg.totalCost += m.Cost
		agg.count++

		if m.Status >= 200 && m.Status < 300 {
			agg.successCount++
		}
	}

	// Convertir a ModelMetric
	var result []ModelMetric
	for _, agg := range groups {
		mm := agg.toModelMetric()
		result = append(result, mm)
	}

	// Ordenar por provider+model para consistencia
	sort.Slice(result, func(i, j int) bool {
		if result[i].Provider != result[j].Provider {
			return result[i].Provider < result[j].Provider
		}
		return result[i].Model < result[j].Model
	})

	return result, nil
}

type aggregator struct {
	provider     string
	model        string
	latencies    []int64
	successCount int
	count        int
	totalTokens  int
	totalCost    float64
}

func (a *aggregator) toModelMetric() ModelMetric {
	mm := ModelMetric{
		Provider: a.provider,
		Model:    a.model,
		Tokens:   a.totalTokens,
		Cost:     a.totalCost,
	}

	// Calcular promedio de latencia
	if len(a.latencies) > 0 {
		sum := int64(0)
		for _, l := range a.latencies {
			sum += l
		}
		mm.AvgLatencyMs = float64(sum) / float64(len(a.latencies))

		// Calcular P95
		sorted := make([]int64, len(a.latencies))
		copy(sorted, a.latencies)
		sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
		idx := (len(sorted) * 95) / 100
		if idx >= len(sorted) {
			idx = len(sorted) - 1
		}
		mm.P95LatencyMs = float64(sorted[idx])
	}

	// Calcular success rate
	if a.count > 0 {
		mm.SuccessRate = float64(a.successCount) / float64(a.count) * 100.0
	}

	return mm
}
