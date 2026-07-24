// Package user implementa el store persistente de usuarios (HU-EVO-017):
// identidad, rol y estado, con CRUD admin-only sobre PostgreSQL. Reemplaza
// gradualmente la fuente de verdad de identidad que hoy vive solo en
// apikey.Store (in-memory); este store agrega la tabla `users`, HU-EVO-018
// mueve las API keys a depender de ella.
package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Role es el rol RBAC de un usuario. admin puede administrar usuarios de
// cualquier tenant (o del propio, si no es admin global); operator solo
// consume el Gateway.
type Role string

const (
	RoleAdmin    Role = "admin"
	RoleOperator Role = "operator"
)

func (r Role) valid() bool {
	return r == RoleAdmin || r == RoleOperator
}

// Status es el ciclo de vida del usuario.
type Status string

const (
	StatusInvited   Status = "invited"
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
)

func (s Status) valid() bool {
	return s == StatusInvited || s == StatusActive || s == StatusSuspended
}

// User es una fila persistida en la tabla `users`.
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      Role      `json:"role"`
	Status    Status    `json:"status"`
	Tenant    string    `json:"tenant"`
	Scopes    []string  `json:"scopes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Errores de dominio expuestos al handler HTTP para mapear a códigos.
var (
	ErrEmailExists  = errors.New("user: email ya registrado")
	ErrNotFound     = errors.New("user: no encontrado")
	ErrInvalidInput = errors.New("user: entrada inválida")
)

const createUsersTableSQL = `
CREATE TABLE IF NOT EXISTS users (
	id SERIAL PRIMARY KEY,
	email TEXT NOT NULL UNIQUE,
	role TEXT NOT NULL,
	status TEXT NOT NULL,
	tenant TEXT NOT NULL DEFAULT '',
	scopes TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);`

// Store persiste usuarios en PostgreSQL (HU-EVO-017).
type Store struct {
	db *sql.DB
}

// NewStore aplica la migración idempotente y devuelve un Store listo.
func NewStore(db *sql.DB) (*Store, error) {
	if db == nil {
		return nil, fmt.Errorf("user: NewStore requiere db no nil")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := db.ExecContext(ctx, createUsersTableSQL); err != nil {
		return nil, fmt.Errorf("user: migración users: %w", err)
	}
	return &Store{db: db}, nil
}

func joinScopes(scopes []string) string { return strings.Join(scopes, ",") }

func splitScopes(raw string) []string {
	if raw == "" {
		return nil
	}
	return strings.Split(raw, ",")
}

// Create inserta un usuario nuevo en estado `invited` (HU-EVO-017 AC1).
// Devuelve ErrEmailExists si el email ya está en uso (AC5).
func (s *Store) Create(ctx context.Context, email string, role Role, tenant string, scopes []string) (*User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil, fmt.Errorf("%w: email vacío", ErrInvalidInput)
	}
	if !role.valid() {
		return nil, fmt.Errorf("%w: role %q inválido", ErrInvalidInput, role)
	}

	var exists bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)`, email).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailExists
	}

	u := &User{Email: email, Role: role, Status: StatusInvited, Tenant: tenant, Scopes: scopes}
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO users (email, role, status, tenant, scopes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, now(), now())
		RETURNING id, created_at, updated_at
	`, u.Email, string(u.Role), string(u.Status), u.Tenant, joinScopes(u.Scopes)).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		// Carrera entre el SELECT EXISTS y el INSERT: el constraint UNIQUE es la
		// fuente de verdad final contra duplicados concurrentes.
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, ErrEmailExists
		}
		return nil, err
	}
	return u, nil
}

// List devuelve usuarios, filtrados por tenant salvo que globalAdmin sea true
// (HU-EVO-017 AC2: aislamiento multi-tenant para admins no-globales).
func (s *Store) List(ctx context.Context, tenant string, globalAdmin bool) ([]User, error) {
	var rows *sql.Rows
	var err error
	if globalAdmin {
		rows, err = s.db.QueryContext(ctx, `SELECT id, email, role, status, tenant, scopes, created_at, updated_at FROM users ORDER BY id`)
	} else {
		rows, err = s.db.QueryContext(ctx, `SELECT id, email, role, status, tenant, scopes, created_at, updated_at FROM users WHERE tenant=$1 ORDER BY id`, tenant)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []User
	for rows.Next() {
		var u User
		var role, status, scopes string
		if err := rows.Scan(&u.ID, &u.Email, &role, &status, &u.Tenant, &scopes, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		u.Role, u.Status, u.Scopes = Role(role), Status(status), splitScopes(scopes)
		out = append(out, u)
	}
	if out == nil {
		out = []User{}
	}
	return out, rows.Err()
}

// Get busca un usuario por ID.
func (s *Store) Get(ctx context.Context, id string) (*User, error) {
	var u User
	var role, status, scopes string
	err := s.db.QueryRowContext(ctx, `SELECT id, email, role, status, tenant, scopes, created_at, updated_at FROM users WHERE id=$1`, id).
		Scan(&u.ID, &u.Email, &role, &status, &u.Tenant, &scopes, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	u.Role, u.Status, u.Scopes = Role(role), Status(status), splitScopes(scopes)
	return &u, nil
}

// Patch aplica cambios parciales de rol/estado (HU-EVO-017 AC3). Campos nil
// no se tocan.
func (s *Store) Patch(ctx context.Context, id string, role *Role, status *Status) (*User, error) {
	if role != nil && !role.valid() {
		return nil, fmt.Errorf("%w: role %q inválido", ErrInvalidInput, *role)
	}
	if status != nil && !status.valid() {
		return nil, fmt.Errorf("%w: status %q inválido", ErrInvalidInput, *status)
	}
	current, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if role != nil {
		current.Role = *role
	}
	if status != nil {
		current.Status = *status
	}
	res, err := s.db.ExecContext(ctx, `UPDATE users SET role=$1, status=$2, updated_at=now() WHERE id=$3`,
		string(current.Role), string(current.Status), id)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrNotFound
	}
	return s.Get(ctx, id)
}

// IsActive indica si el usuario puede autenticar (HU-EVO-017 AC3: suspended
// pierde acceso inmediato; invited tampoco autentica hasta activarse).
func (u *User) IsActive() bool { return u.Status == StatusActive }
