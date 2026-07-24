package quota

import "api-llm-gateway/internal/adapter"

// PersistJob representa un trabajo de persistencia de quota aprendida
type PersistJob struct {
	ProviderID string
	ModelID    string
	Quota      adapter.QuotaInfo
}

// Persister define el contrato para persistencia async de quotas aprendidas (HU-EVO-008)
type Persister interface {
	// Enqueue encola un job de persistencia y retorna inmediatamente (<5ms)
	// No bloquea en DB; DB indisponible no causa error
	Enqueue(job PersistJob) error

	// Close cierra el worker background (usado en tests y graceful shutdown)
	Close() error
}

// NoPersister es un stub que no persiste nada (usado en tests y cuando persistencia está deshabilitada)
type NoPersister struct{}

func (np *NoPersister) Enqueue(job PersistJob) error {
	return nil
}

func (np *NoPersister) Close() error {
	return nil
}
