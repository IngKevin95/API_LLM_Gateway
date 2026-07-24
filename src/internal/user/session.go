package user

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// Session representa una sesión activa de un usuario
type Session struct {
	ID         string
	UserID     string
	UserAgent  string
	IP         string
	LastUsedAt time.Time
	CreatedAt  time.Time
}

// SessionStore gestiona la persistencia de las sesiones
type SessionStore struct {
	db    *sql.DB
	users *Store
}

// NewSessionStore crea un nuevo SessionStore y asegura que la tabla exista.
func NewSessionStore(db *sql.DB, users *Store) (*SessionStore, error) {
	s := &SessionStore{db: db, users: users}
	if err := s.initDB(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *SessionStore) initDB() error {
	query := `
	CREATE TABLE IF NOT EXISTS sessions (
		id VARCHAR(128) PRIMARY KEY,
		user_id INTEGER NOT NULL,
		user_agent TEXT,
		ip VARCHAR(45),
		last_used_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`
	_, err := s.db.Exec(query)
	return err
}

func generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Create registra una nueva sesión.
func (s *SessionStore) Create(ctx context.Context, userID, userAgent, ip string) (string, error) {
	id := generateSessionID()
	query := `INSERT INTO sessions (id, user_id, user_agent, ip) VALUES ($1, $2, $3, $4)`
	_, err := s.db.ExecContext(ctx, query, id, userID, userAgent, ip)
	if err != nil {
		return "", fmt.Errorf("creando sesion: %w", err)
	}
	return id, nil
}

// List devuelve las sesiones activas de un usuario.
func (s *SessionStore) List(ctx context.Context, userID string) ([]Session, error) {
	query := `SELECT id, user_id, user_agent, ip, last_used_at, created_at FROM sessions WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var ses Session
		var ua, ip sql.NullString
		if err := rows.Scan(&ses.ID, &ses.UserID, &ua, &ip, &ses.LastUsedAt, &ses.CreatedAt); err != nil {
			return nil, err
		}
		if ua.Valid {
			ses.UserAgent = ua.String
		}
		if ip.Valid {
			ses.IP = ip.String
		}
		sessions = append(sessions, ses)
	}
	return sessions, rows.Err()
}

// Revoke elimina una sesión específica.
func (s *SessionStore) Revoke(ctx context.Context, userID, sessionID string) error {
	query := `DELETE FROM sessions WHERE id = $1 AND user_id = $2`
	_, err := s.db.ExecContext(ctx, query, sessionID, userID)
	return err
}

// RevokeOthers elimina todas las sesiones de un usuario excepto la indicada.
func (s *SessionStore) RevokeOthers(ctx context.Context, userID, currentSessionID string) error {
	query := `DELETE FROM sessions WHERE user_id = $1 AND id != $2`
	_, err := s.db.ExecContext(ctx, query, userID, currentSessionID)
	return err
}

// IsValid verifica si una sesión existe (es válida). También actualiza last_used_at.
func (s *SessionStore) IsValid(ctx context.Context, sessionID string) (bool, error) {
	query := `UPDATE sessions SET last_used_at = CURRENT_TIMESTAMP WHERE id = $1 RETURNING id`
	var id string
	err := s.db.QueryRowContext(ctx, query, sessionID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
