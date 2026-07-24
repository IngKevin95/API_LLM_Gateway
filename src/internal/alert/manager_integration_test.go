//go:build integration
// +build integration

// Mismo patrón que internal/quota/persister_postgres_integration_test.go:
// levanta un contenedor postgres:16 real vía `docker run` (sin dependencia
// nueva de testcontainers-go). Correr con:
//
//	go test ./internal/alert/... -tags=integration -run TestManager -v
//
// HU-EVO-012 / INT-alertmanager-postgres: valida Check()+generateAlert()
// contra PostgreSQL real -- inserta, dedup por (provider_id, model_id), y
// distingue severidad.

package alert

import (
	"context"
	"database/sql"
	"fmt"
	"os/exec"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

const testPGContainer = "alert-manager-it-pg"

func startTestPostgres(t *testing.T) *sql.DB {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker no disponible en PATH, saltando integration test de alert.Manager")
	}

	_ = exec.Command("docker", "rm", "-f", testPGContainer).Run()

	port := "55433"
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

// HU-EVO-012 AC1+AC2: primera corrida inserta la alerta; la segunda corrida
// (misma condición sostenida) NO duplica la fila, solo actualiza updated_at.
func TestManager_Check_InsertsThenDedupsOnSecondRun(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()

	snap := stubSnapshotter{entries: []QuotaEntry{{Provider: "groq", Model: "mixtral", Limit: 14400, Remaining: 1200}}}
	m, err := NewManager(db, snap, DefaultThreshold)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	if err := m.Check(context.Background()); err != nil {
		t.Fatalf("Check (1a corrida): %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT count(*) FROM provider_alerts WHERE provider_id='groq' AND model_id='mixtral'`).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	if count != 1 {
		t.Fatalf("esperaba 1 fila tras primera corrida, obtuve %d", count)
	}

	if err := m.Check(context.Background()); err != nil {
		t.Fatalf("Check (2a corrida): %v", err)
	}
	if err := db.QueryRow(`SELECT count(*) FROM provider_alerts WHERE provider_id='groq' AND model_id='mixtral'`).Scan(&count); err != nil {
		t.Fatalf("query count tras 2a corrida: %v", err)
	}
	if count != 1 {
		t.Fatalf("esperaba seguir en 1 fila (dedup, no duplica), obtuve %d", count)
	}
}

// HU-EVO-012 AC3: proveedor agotado (remaining=0) genera severity=critical
// persistido en PostgreSQL.
func TestManager_Check_ExhaustedProvider_PersistsCritical(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()

	snap := stubSnapshotter{entries: []QuotaEntry{{Provider: "cerebras", Model: "", Limit: 500, Remaining: 0}}}
	m, err := NewManager(db, snap, DefaultThreshold)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	if err := m.Check(context.Background()); err != nil {
		t.Fatalf("Check: %v", err)
	}

	var severity, message string
	if err := db.QueryRow(`SELECT severity, message FROM provider_alerts WHERE provider_id='cerebras'`).Scan(&severity, &message); err != nil {
		t.Fatalf("query: %v", err)
	}
	if severity != "critical" {
		t.Fatalf("esperaba severity=critical, obtuve %q", severity)
	}
	if message != "cerebras EXHAUSTED" {
		t.Fatalf("mensaje inesperado: %q", message)
	}
}
