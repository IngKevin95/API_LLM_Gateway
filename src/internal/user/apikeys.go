package user

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"api-llm-gateway/internal/auth"
)

// APIKey es la proyección pública (nunca incluye la key en claro) de una fila
// de `api_keys` (HU-EVO-018).
type APIKey struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"` // ej. "sk-***4f2a"
	Scopes     []string   `json:"scopes"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

var (
	ErrKeyNotFound  = errors.New("apikeys: key no encontrada")
	ErrKeyForbidden = errors.New("apikeys: la key no pertenece a este usuario")
)

const createAPIKeysTableSQL = `
CREATE TABLE IF NOT EXISTS api_keys (
	id SERIAL PRIMARY KEY,
	user_id INTEGER NOT NULL,
	name TEXT NOT NULL,
	key_hash TEXT NOT NULL UNIQUE,
	key_prefix TEXT NOT NULL,
	scopes TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	last_used_at TIMESTAMPTZ,
	revoked_at TIMESTAMPTZ
);`

// KeyStore persiste API keys por usuario en PostgreSQL, hasheadas
// (sha256, comparación en tiempo constante -- mismo criterio que
// internal/auth/apikey/apikey.go).
type KeyStore struct {
	db    *sql.DB
	users *Store
}

// NewKeyStore aplica la migración idempotente. users se usa para resolver el
// Identity.Tenant/Scopes efectivos al autenticar.
func NewKeyStore(db *sql.DB, users *Store) (*KeyStore, error) {
	if db == nil {
		return nil, fmt.Errorf("apikeys: NewKeyStore requiere db no nil")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := db.ExecContext(ctx, createAPIKeysTableSQL); err != nil {
		return nil, fmt.Errorf("apikeys: migración api_keys: %w", err)
	}
	return &KeyStore{db: db, users: users}, nil
}

func hashKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

func randomKey() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "sk-" + hex.EncodeToString(buf), nil
}

func maskPrefix(key string) string {
	if len(key) <= 4 {
		return "***"
	}
	return "sk-***" + key[len(key)-4:]
}

// Generate crea una key nueva para userID (HU-EVO-018 AC1). Devuelve la key
// en texto plano UNA sola vez; solo su hash se persiste.
func (k *KeyStore) Generate(ctx context.Context, userID, name string) (plainKey string, rec APIKey, err error) {
	plainKey, err = randomKey()
	if err != nil {
		return "", APIKey{}, err
	}
	rec = APIKey{UserID: userID, Name: name, Prefix: maskPrefix(plainKey)}
	err = k.db.QueryRowContext(ctx, `
		INSERT INTO api_keys (user_id, name, key_hash, key_prefix, created_at)
		VALUES ($1, $2, $3, $4, now())
		RETURNING id, created_at
	`, userID, name, hashKey(plainKey), rec.Prefix).Scan(&rec.ID, &rec.CreatedAt)
	if err != nil {
		return "", APIKey{}, err
	}
	return plainKey, rec, nil
}

// List devuelve las keys de userID enmascaradas (HU-EVO-018 AC2), nunca la
// key completa.
func (k *KeyStore) List(ctx context.Context, userID string) ([]APIKey, error) {
	rows, err := k.db.QueryContext(ctx, `
		SELECT id, user_id, name, key_prefix, created_at, last_used_at, revoked_at
		FROM api_keys WHERE user_id=$1 ORDER BY id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []APIKey
	for rows.Next() {
		var a APIKey
		var lastUsed, revoked sql.NullTime
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.Prefix, &a.CreatedAt, &lastUsed, &revoked); err != nil {
			return nil, err
		}
		if lastUsed.Valid {
			t := lastUsed.Time
			a.LastUsedAt = &t
		}
		if revoked.Valid {
			t := revoked.Time
			a.RevokedAt = &t
		}
		out = append(out, a)
	}
	if out == nil {
		out = []APIKey{}
	}
	return out, rows.Err()
}

// Revoke marca revoked_at=now() (HU-EVO-018 AC3). Solo el dueño de la key
// (ownerUserID==userID) puede revocarla, salvo que el caller ya haya validado
// admin antes de llamar (ver handler).
func (k *KeyStore) Revoke(ctx context.Context, userID, keyID string) error {
	res, err := k.db.ExecContext(ctx, `
		UPDATE api_keys SET revoked_at = now()
		WHERE id=$1 AND user_id=$2 AND revoked_at IS NULL
	`, keyID, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		return nil
	}
	// Distinguir 404 (no existe/ya revocada) de 403 (pertenece a otro usuario):
	var owner string
	if err := k.db.QueryRowContext(ctx, `SELECT user_id::text FROM api_keys WHERE id=$1`, keyID).Scan(&owner); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrKeyNotFound
		}
		return err
	}
	if owner != userID {
		return ErrKeyForbidden
	}
	return ErrKeyNotFound // existía, pertenecía al usuario, pero ya estaba revocada
}

// Authenticate resuelve una key en texto plano a una auth.Identity real
// (HU-EVO-018 reemplaza el seed GATEWAY_API_KEYS por PostgreSQL). Compara en
// tiempo constante, actualiza last_used_at (AC5), y falla si la key está
// revocada o el usuario dueño no está activo (HU-EVO-017 AC3: suspensión
// corta el acceso inmediatamente, incluso con key vigente).
func (k *KeyStore) Authenticate(ctx context.Context, plainKey string) (auth.Identity, bool) {
	want := hashKey(plainKey)

	rows, err := k.db.QueryContext(ctx, `
		SELECT id, user_id, key_hash FROM api_keys WHERE revoked_at IS NULL
	`)
	if err != nil {
		return auth.Identity{}, false
	}
	defer rows.Close()

	var matchID, matchUserID string
	found := false
	for rows.Next() {
		var id, userID, h string
		if err := rows.Scan(&id, &userID, &h); err != nil {
			continue
		}
		if subtle.ConstantTimeCompare([]byte(h), []byte(want)) == 1 {
			matchID, matchUserID, found = id, userID, true
			break
		}
	}
	if !found {
		return auth.Identity{}, false
	}

	u, err := k.users.Get(ctx, matchUserID)
	if err != nil || !u.IsActive() {
		return auth.Identity{}, false // usuario suspendido/inexistente: acceso inmediato cortado
	}

	_, _ = k.db.ExecContext(ctx, `UPDATE api_keys SET last_used_at = now() WHERE id=$1`, matchID)

	// Subject = ID de usuario (no el email) para que el handler HTTP pueda
	// resolver ownership de recursos (/users/:id/api-keys) por comparación
	// directa contra el :id de la URL, sin una consulta extra.
	return auth.Identity{Subject: u.ID, Tenant: u.Tenant, Scopes: u.Scopes}, true
}

// IsOwnerOrAdmin separa "sos admin" de "sos el dueño del recurso" (usado por
// el handler HTTP de /users/:id/api-keys, HU-EVO-018 AC4).
func IsOwnerOrAdmin(id auth.Identity, ownerUserID string, admin bool) bool {
	if admin {
		return true
	}
	return strings.EqualFold(id.Subject, ownerUserID)
}
