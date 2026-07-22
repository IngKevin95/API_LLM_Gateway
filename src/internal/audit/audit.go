// Package audit define los tipos de dominio de la persistencia asincronista:
// el evento inmutable de auditoría y las fronteras hacia DB (Store) y KMS
// (Encryptor). Las implementaciones reales (PostgreSQL, KMS) se inyectan.
package audit

import "context"

// Event es la metadata inmutable de auditoría (nunca contiene prompts/respuestas).
type Event struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"ts"` // unix nanos
	Tokens    int    `json:"tokens"`
	Status    int    `json:"status"`
	Cost      int    `json:"cost"`
	User      string `json:"user"`
	Agent     string `json:"agent"`
	Tools     string `json:"tools"`
	CacheHit  bool   `json:"cache_hit"`
	Priority  int    `json:"prio"` // 0 = normal; <0 = baja (dropeable ante backpressure)
}

// EncryptedEvent es un evento ya cifrado con KMS Envelope, listo para persistir.
type EncryptedEvent struct {
	Ciphertext []byte
	DEKID      string
}

// Store persiste lotes de eventos cifrados (lo implementa el writer a PostgreSQL).
type Store interface {
	Write(ctx context.Context, batch []EncryptedEvent) error
}

// Encryptor cifra un evento con KMS Envelope (DEK local, KEK en KMS).
type Encryptor interface {
	Seal(Event) (EncryptedEvent, error)
}
