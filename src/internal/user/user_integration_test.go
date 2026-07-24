//go:build integration
// +build integration

// Mismo patrón que internal/alert/manager_integration_test.go: levanta un
// contenedor postgres:16 real vía `docker run`. Correr con:
//
//	go test ./internal/user/... -tags=integration -run TestStore -v
//	go test ./internal/user/... -tags=integration -run TestKeyStore -v
//
// HU-EVO-017 / HU-EVO-018: valida Create/List/Patch y Generate/List/Revoke/
// Authenticate contra PostgreSQL real.

package user

import (
	"context"
	"database/sql"
	"fmt"
	"os/exec"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

const testPGContainer = "user-store-it-pg"

func startTestPostgres(t *testing.T) *sql.DB {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker no disponible en PATH, saltando integration test de user.Store")
	}

	_ = exec.Command("docker", "rm", "-f", testPGContainer).Run()

	port := "55434"
	dsn := fmt.Sprintf("postgres://postgres:postgres@127.0.0.1:%s/postgres?sslmode=disable", port)

	runCmd := exec.Command("docker", "run", "-d", "--rm",
		"--name", testPGContainer,
		"-e", "POSTGRES_PASSWORD=postgres",
		"-p", port+":5432",
		"postgres:16",
	)
	if out, err := runCmd.CombinedOutput(); err != nil {
		t.Skipf("no se pudo levantar postgres real (docker run): %v\n%s", err, out)
	}
	t.Cleanup(func() {
		_ = exec.Command("docker", "rm", "-f", testPGContainer).Run()
	})

	deadline := time.Now().Add(30 * time.Second)
	var db *sql.DB
	var lastErr error
	for time.Now().Before(deadline) {
		db, lastErr = sql.Open("postgres", dsn)
		if lastErr == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			lastErr = db.PingContext(ctx)
			cancel()
			if lastErr == nil {
				return db
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("postgres de test no respondió a tiempo: %v", lastErr)
	return nil
}

// HU-EVO-017 AC1: crear usuario deja status=invited.
func TestStore_Create_DefaultsToInvited(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()

	s, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	u, err := s.Create(context.Background(), "ana@example.com", RoleOperator, "t1", []string{"capability:coding"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if u.Status != StatusInvited {
		t.Fatalf("esperaba status=invited, obtuve %q", u.Status)
	}
	if u.ID == "" {
		t.Fatalf("esperaba ID asignado")
	}
}

// HU-EVO-017 AC2: List filtra por tenant salvo globalAdmin.
func TestStore_List_FiltersByTenant(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()

	s, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if _, err := s.Create(context.Background(), "t1user@example.com", RoleOperator, "t1", nil); err != nil {
		t.Fatalf("Create t1: %v", err)
	}
	if _, err := s.Create(context.Background(), "t2user@example.com", RoleOperator, "t2", nil); err != nil {
		t.Fatalf("Create t2: %v", err)
	}

	t1Users, err := s.List(context.Background(), "t1", false)
	if err != nil {
		t.Fatalf("List t1: %v", err)
	}
	if len(t1Users) != 1 || t1Users[0].Tenant != "t1" {
		t.Fatalf("esperaba 1 usuario de t1, obtuve %+v", t1Users)
	}

	allUsers, err := s.List(context.Background(), "", true)
	if err != nil {
		t.Fatalf("List global: %v", err)
	}
	if len(allUsers) != 2 {
		t.Fatalf("esperaba 2 usuarios (admin global), obtuve %d", len(allUsers))
	}
}

// HU-EVO-017 AC3: PATCH a suspended corta acceso (verificado end-to-end vía
// KeyStore.Authenticate en TestKeyStore_Authenticate_SuspendedUser_Fails).
func TestStore_Patch_ChangesStatus(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()

	s, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	u, err := s.Create(context.Background(), "bob@example.com", RoleOperator, "t1", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	suspended := StatusSuspended
	updated, err := s.Patch(context.Background(), u.ID, nil, &suspended)
	if err != nil {
		t.Fatalf("Patch: %v", err)
	}
	if updated.Status != StatusSuspended {
		t.Fatalf("esperaba status=suspended, obtuve %q", updated.Status)
	}
}

// HU-EVO-017 AC4/AC5 se validan en el handler HTTP (403 no-admin, 409 email
// duplicado) — ver internal/handler/users_test.go y este caso de dominio:
func TestStore_Create_DuplicateEmail_ReturnsErrEmailExists(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()

	s, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if _, err := s.Create(context.Background(), "dup@example.com", RoleOperator, "t1", nil); err != nil {
		t.Fatalf("Create 1a: %v", err)
	}
	_, err = s.Create(context.Background(), "dup@example.com", RoleOperator, "t1", nil)
	if err != ErrEmailExists {
		t.Fatalf("esperaba ErrEmailExists, obtuve %v", err)
	}
}

// HU-EVO-018 AC1/AC2/AC3: generar, listar enmascarado, revocar.
func TestKeyStore_GenerateListRevoke(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()

	us, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	ks, err := NewKeyStore(db, us)
	if err != nil {
		t.Fatalf("NewKeyStore: %v", err)
	}
	u, err := us.Create(context.Background(), "carla@example.com", RoleOperator, "t1", nil)
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}
	active := StatusActive
	if _, err := us.Patch(context.Background(), u.ID, nil, &active); err != nil {
		t.Fatalf("activar usuario: %v", err)
	}

	plain, rec, err := ks.Generate(context.Background(), u.ID, "laptop")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if plain == "" {
		t.Fatalf("esperaba key en claro no vacía")
	}

	keys, err := ks.List(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(keys) != 1 || keys[0].Prefix == "" {
		t.Fatalf("esperaba 1 key enmascarada, obtuve %+v", keys)
	}
	for _, k := range keys {
		if k.Prefix == plain {
			t.Fatalf("List NUNCA debe exponer la key completa")
		}
	}

	// AC1/AC3: autentica antes de revocar, deja de autenticar después.
	if _, ok := ks.Authenticate(context.Background(), plain); !ok {
		t.Fatalf("esperaba autenticar con la key recién generada")
	}
	if err := ks.Revoke(context.Background(), u.ID, rec.ID); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	if _, ok := ks.Authenticate(context.Background(), plain); ok {
		t.Fatalf("key revocada NO debería autenticar más")
	}
}

// HU-EVO-018 AC4 (a nivel dominio): revocar la key de otro usuario devuelve
// ErrKeyForbidden.
func TestKeyStore_Revoke_WrongOwner_ReturnsForbidden(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()

	us, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	ks, err := NewKeyStore(db, us)
	if err != nil {
		t.Fatalf("NewKeyStore: %v", err)
	}
	owner, _ := us.Create(context.Background(), "owner@example.com", RoleOperator, "t1", nil)
	other, _ := us.Create(context.Background(), "other@example.com", RoleOperator, "t1", nil)

	_, rec, err := ks.Generate(context.Background(), owner.ID, "key1")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	err = ks.Revoke(context.Background(), other.ID, rec.ID)
	if err != ErrKeyForbidden {
		t.Fatalf("esperaba ErrKeyForbidden, obtuve %v", err)
	}
}

// HU-EVO-018 AC5: last_used_at se actualiza al autenticar.
func TestKeyStore_Authenticate_UpdatesLastUsedAt(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()

	us, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	ks, err := NewKeyStore(db, us)
	if err != nil {
		t.Fatalf("NewKeyStore: %v", err)
	}
	u, _ := us.Create(context.Background(), "diana@example.com", RoleOperator, "t1", nil)
	active := StatusActive
	if _, err := us.Patch(context.Background(), u.ID, nil, &active); err != nil {
		t.Fatalf("activar: %v", err)
	}
	plain, rec, err := ks.Generate(context.Background(), u.ID, "cli")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	keysBefore, _ := ks.List(context.Background(), u.ID)
	if keysBefore[0].LastUsedAt != nil {
		t.Fatalf("esperaba last_used_at nil antes del primer uso")
	}

	if _, ok := ks.Authenticate(context.Background(), plain); !ok {
		t.Fatalf("esperaba autenticar")
	}

	keysAfter, err := ks.List(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	var found bool
	for _, k := range keysAfter {
		if k.ID == rec.ID {
			found = true
			if k.LastUsedAt == nil {
				t.Fatalf("esperaba last_used_at seteado tras autenticar")
			}
		}
	}
	if !found {
		t.Fatalf("no encontré la key generada en List tras autenticar")
	}
}

// HU-EVO-017 AC3 end-to-end: usuario suspendido pierde acceso inmediato aún
// con key vigente (no revocada).
func TestKeyStore_Authenticate_SuspendedUser_Fails(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()

	us, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	ks, err := NewKeyStore(db, us)
	if err != nil {
		t.Fatalf("NewKeyStore: %v", err)
	}
	u, _ := us.Create(context.Background(), "erik@example.com", RoleOperator, "t1", nil)
	active := StatusActive
	if _, err := us.Patch(context.Background(), u.ID, nil, &active); err != nil {
		t.Fatalf("activar: %v", err)
	}
	plain, _, err := ks.Generate(context.Background(), u.ID, "key")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, ok := ks.Authenticate(context.Background(), plain); !ok {
		t.Fatalf("esperaba autenticar mientras está activo")
	}

	suspended := StatusSuspended
	if _, err := us.Patch(context.Background(), u.ID, nil, &suspended); err != nil {
		t.Fatalf("suspender: %v", err)
	}
	if _, ok := ks.Authenticate(context.Background(), plain); ok {
		t.Fatalf("usuario suspendido NO debería poder autenticar, aunque la key siga sin revocar")
	}
}
