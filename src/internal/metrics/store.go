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
	mu        sync.RWMutex
	metrics   []RequestMetric
	maxSize   int
	startIdx  int
	startTime time.Time

	// configuredProviders + quotaSnapshot (HU-EVO-005/EP-EVO-001): seteados en
	// boot desde el Registry/Quota Manager, para que /metrics reporte
	// proveedores y cuota aun sin tráfico previo (AC de journey_smoke).
	configuredProviders []string
	quotaSnapshot       map[string]int
}

// SetProviderSnapshot registra en boot los proveedores configurados y su cuota
// remanente inicial (desde Quota Manager.InitFromRegistry), para que
// GetGatewayMetrics los reporte incluso antes de la primera request.
func (s *InMemoryStore) SetProviderSnapshot(providerIDs []string, quota map[string]int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configuredProviders = append([]string(nil), providerIDs...)
	s.quotaSnapshot = quota
}

// NewInMemoryStore crea un store con tamaño máximo (últimas N métricas).
func NewInMemoryStore(maxSize int) *InMemoryStore {
	return &InMemoryStore{
		metrics:   make([]RequestMetric, 0, maxSize),
		maxSize:   maxSize,
		startTime: time.Now(),
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
	if len(s.metrics) == 0 {
		s.mu.RUnlock()
		return []ModelMetric{}, nil
	}
	metricsCopy := make([]RequestMetric, len(s.metrics))
	copy(metricsCopy, s.metrics)
	s.mu.RUnlock()

	// Agrupar por (provider, model)
	groups := make(map[string]*aggregator)

	for _, m := range metricsCopy {
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

// GetGatewayMetrics devuelve la estructura HU-060 (AC2+AC3) con uptime, requests breakdown, providers status, latency percentiles.
func (s *InMemoryStore) GetGatewayMetrics(ctx context.Context) (*GatewayMetrics, error) {
	s.mu.RLock()
	if len(s.metrics) == 0 {
		providers := s.snapshotProviderStatusLocked()
		quota := s.snapshotQuotaLocked()
		s.mu.RUnlock()
		return &GatewayMetrics{
			UptimeSeconds: int(time.Since(s.startTime).Seconds()),
			Requests:      RequestMetrics{ByHandler: make(map[string]int)},
			Providers:     providers,
			Quota:         quota,
			Models:        []ModelMetric{},
		}, nil
	}
	metricsCopy := make([]RequestMetric, len(s.metrics))
	copy(metricsCopy, s.metrics)
	startTime := s.startTime
	configuredProviders := s.configuredProviders
	quotaSnapshot := s.snapshotQuotaLocked()
	s.mu.RUnlock()

	// HU-060 AC2: uptime_seconds
	uptime := int(time.Since(startTime).Seconds())

	// HU-060 AC2: requests (total, by_handler, errors)
	var totalRequests int
	var errorCount int
	latencies := []int64{}
	byHandler := map[string]int{
		"/v1/chat/completions": 0,
		"/v1/embeddings":       0,
		"/v1/messages":         0,
		"/health":              0,
		"/metrics":             0,
		"/mcp":                 0,
	}
	providerSet := make(map[string]bool)
	for _, id := range configuredProviders {
		providerSet[id] = true
	}

	for _, m := range metricsCopy {
		totalRequests++
		latencies = append(latencies, m.LatencyMs)
		providerSet[m.Provider] = true
		if m.Status >= 400 {
			errorCount++
		}
	}

	// HU-060 AC3: providers (name, available, circuit_breaker_open)
	var providers []ProviderStatus
	for provider := range providerSet {
		providers = append(providers, ProviderStatus{
			Name:               provider,
			Available:          true,
			CircuitBreakerOpen: false,
		})
	}
	// Ordenar para consistencia
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Name < providers[j].Name
	})

	// HU-060 AC3: latency percentiles (p50, p95, p99)
	latMetrics := LatencyMetrics{}
	if len(latencies) > 0 {
		sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
		idx50 := (len(latencies) * 50) / 100
		idx95 := (len(latencies) * 95) / 100
		idx99 := (len(latencies) * 99) / 100
		if idx50 >= len(latencies) {
			idx50 = len(latencies) - 1
		}
		if idx95 >= len(latencies) {
			idx95 = len(latencies) - 1
		}
		if idx99 >= len(latencies) {
			idx99 = len(latencies) - 1
		}
		latMetrics.P50Ms = float64(latencies[idx50])
		latMetrics.P95Ms = float64(latencies[idx95])
		latMetrics.P99Ms = float64(latencies[idx99])
	}

	return &GatewayMetrics{
		UptimeSeconds: uptime,
		Requests: RequestMetrics{
			Total:     totalRequests,
			ByHandler: byHandler,
			Errors:    errorCount,
		},
		Providers: providers,
		Quota:     quotaSnapshot,
		Latency:   latMetrics,
		Models:    []ModelMetric{},
	}, nil
}

// snapshotProviderStatusLocked construye la lista de ProviderStatus a partir
// de configuredProviders; el llamador debe sostener s.mu (R o W).
func (s *InMemoryStore) snapshotProviderStatusLocked() []ProviderStatus {
	out := make([]ProviderStatus, 0, len(s.configuredProviders))
	for _, id := range s.configuredProviders {
		out = append(out, ProviderStatus{Name: id, Available: true})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// snapshotQuotaLocked copia el snapshot de cuota; el llamador debe sostener s.mu.
func (s *InMemoryStore) snapshotQuotaLocked() map[string]int {
	if s.quotaSnapshot == nil {
		return nil
	}
	out := make(map[string]int, len(s.quotaSnapshot))
	for k, v := range s.quotaSnapshot {
		out[k] = v
	}
	return out
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
