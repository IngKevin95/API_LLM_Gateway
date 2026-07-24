//go:build integration
// +build integration

package user

import (
	"context"
	"testing"
)

func TestSessionStore_CreateListRevoke(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()

	us, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	ss, err := NewSessionStore(db, us)
	if err != nil {
		t.Fatalf("NewSessionStore: %v", err)
	}

	u, _ := us.Create(context.Background(), "session1@example.com", RoleOperator, "t1", nil)

	// Crear sesión
	sessionID, err := ss.Create(context.Background(), u.ID, "Mozilla/5.0", "192.168.1.10")
	if err != nil {
		t.Fatalf("Create session: %v", err)
	}
	if sessionID == "" {
		t.Fatalf("esperaba un sessionID no vacío")
	}

	// Listar sesiones
	sessions, err := ss.List(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("List sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("esperaba 1 sesión activa, obtuve %d", len(sessions))
	}
	if sessions[0].ID != sessionID || sessions[0].UserAgent != "Mozilla/5.0" || sessions[0].IP != "192.168.1.10" {
		t.Errorf("datos de sesión incorrectos: %+v", sessions[0])
	}

	// Revocar sesión
	if err := ss.Revoke(context.Background(), u.ID, sessionID); err != nil {
		t.Fatalf("Revoke session: %v", err)
	}

	// Verificar que ya no está activa
	sessionsAfter, _ := ss.List(context.Background(), u.ID)
	if len(sessionsAfter) != 0 {
		t.Fatalf("esperaba 0 sesiones tras revocar, obtuve %d", len(sessionsAfter))
	}

	// Verificar validez (middleware check)
	valid, err := ss.IsValid(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("IsValid error: %v", err)
	}
	if valid {
		t.Fatalf("sesión revocada no debería ser válida")
	}
}

func TestSessionStore_RevokeOthers(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()

	us, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore error: %v", err)
	}
	ss, err := NewSessionStore(db, us)
	if err != nil {
		t.Fatalf("NewSessionStore error: %v", err)
	}

	u, _ := us.Create(context.Background(), "session2@example.com", RoleOperator, "t1", nil)

	_, _ = ss.Create(context.Background(), u.ID, "UA1", "IP1")
	s2, _ := ss.Create(context.Background(), u.ID, "UA2", "IP2")
	_, _ = ss.Create(context.Background(), u.ID, "UA3", "IP3")

	// Revocar todas excepto s2
	if err := ss.RevokeOthers(context.Background(), u.ID, s2); err != nil {
		t.Fatalf("RevokeOthers: %v", err)
	}

	sessions, _ := ss.List(context.Background(), u.ID)
	if len(sessions) != 1 || sessions[0].ID != s2 {
		t.Fatalf("esperaba solo la sesión s2 activa, obtuve: %+v", sessions)
	}
}
