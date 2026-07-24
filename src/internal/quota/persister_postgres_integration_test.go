//go:build integration
// +build integration

// Este archivo requiere el build tag "integration" y Docker disponible en el
// host (levanta un contenedor postgres:16 real vía `docker run`, sin agregar
// una dependencia nueva de tipo testcontainers-go al módulo Go — usa
// os/exec, ya cubierto por la stdlib). Correr con:
//
//	go test ./internal/quota/... -tags=integration -run TestPostgresPersister -v
//
// HU-EVO-008 / INT-quota-persist: valida el persister real contra una
// PostgreSQL real, no un mock — Enqueue -> worker async -> INSERT/UPSERT ->
// LoadRemaining lee lo escrito.

package quota

import (
	"context"
	"database/sql"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"api-llm-gateway/internal/adapter"
	_ "github.com/lib/pq"
)

const testPGContainer = "quota-persister-it-pg"

func startTestPostgres(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker no disponible en PATH, saltando integration test de PostgresPersister")
	}

	_ = exec.Command("docker", "rm", "-f", testPGContainer).Run() // limpieza de corridas previas

	port := "55432"
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
	var lastErr error
	for time.Now().Before(deadline) {
		db, err := sql.Open("postgres", dsn)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			lastErr = db.PingContext(ctx)
			cancel()
			_ = db.Close()
			if lastErr == nil {
				return dsn
			}
		} else {
			lastErr = err
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("postgres de test no respondió a tiempo: %v", lastErr)
	return ""
}

// HU-EVO-008 AC1/AC2 + INT-quota-persist — Enqueue real llega a PostgreSQL:
// escribe, el UPSERT actualiza en conflicto, y LoadRemaining lee el último
// valor persistido.
func TestPostgresPersister_EnqueueWritesAndLoadRemainingReadsBack(t *testing.T) {
	dsn := startTestPostgres(t)

	p, err := NewPostgresPersister(dsn, 100)
	if err != nil {
		t.Fatalf("NewPostgresPersister: %v", err)
	}
	defer p.Close()

	if err := p.Enqueue(PersistJob{
		ProviderID: "groq",
		ModelID:    "mixtral-8x7b",
		Quota:      adapter.QuotaInfo{Limit: 1000, Remaining: 750},
	}); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	// Segundo job para el mismo (provider, model) -> ejercita el UPSERT.
	if err := p.Enqueue(PersistJob{
		ProviderID: "groq",
		ModelID:    "mixtral-8x7b",
		Quota:      adapter.QuotaInfo{Limit: 1000, Remaining: 600},
	}); err != nil {
		t.Fatalf("Enqueue (update): %v", err)
	}

	// El worker es async: esperar a que ambos writes se apliquen.
	var remaining map[string]int64
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		remaining, err = p.LoadRemaining(context.Background())
		if err != nil {
			t.Fatalf("LoadRemaining: %v", err)
		}
		if remaining["groq"] == 600 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if remaining["groq"] != 600 {
		t.Fatalf("esperaba remaining=600 tras UPSERT (último write gana), obtuve %d", remaining["groq"])
	}
}

// Reproduce la secuencia exacta que cmd/gateway/main.go ejecuta en boot
// cuando GATEWAY_QUOTA_POSTGRES_DSN está declarado: escribir vía Enqueue (una
// corrida anterior "aprendiendo" cuota), abrir un Manager nuevo con el mismo
// persister, restaurar con LoadRemaining+RestoreRemaining, y verificar que
// Remaining() refleja el valor persistido (AC5 linaje HU-EVO-005: precedencia
// sobre quota_hint).
func TestPostgresPersister_BootRestoreSequence_MatchesMainGo(t *testing.T) {
	dsn := startTestPostgres(t)

	// "Corrida anterior": un Manager aprende cuota y la persiste.
	p1, err := NewPostgresPersister(dsn, 100)
	if err != nil {
		t.Fatalf("NewPostgresPersister (writer): %v", err)
	}
	if err := p1.Enqueue(PersistJob{ProviderID: "cerebras", ModelID: "llama-70b", Quota: adapter.QuotaInfo{Limit: 500, Remaining: 123}}); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if restored, _ := p1.LoadRemaining(context.Background()); restored["cerebras"] == 123 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	p1.Close()

	// "Reinicio del gateway": Manager nuevo + persister nuevo sobre la misma DB.
	p2, err := NewPostgresPersister(dsn, 100)
	if err != nil {
		t.Fatalf("NewPostgresPersister (reader): %v", err)
	}
	defer p2.Close()

	qm := NewInMemoryManagerWithPersister(time.Now, p2)
	restored, err := p2.LoadRemaining(context.Background())
	if err != nil {
		t.Fatalf("LoadRemaining: %v", err)
	}
	for providerID, remaining := range restored {
		qm.RestoreRemaining(providerID, int(remaining))
	}
	// InitFromRegistry no debe pisar lo restaurado (AC5).
	qm.InitFromRegistry(map[string]*int{"cerebras": nil})

	if got := qm.Remaining("cerebras", ""); got != 123 {
		t.Fatalf("Remaining tras boot-restore: esperaba 123 (persistido), obtuve %d", got)
	}
}

// HU-EVO-008 AC3 — DSN inválido/DB inalcanzable no debe colgar el proceso:
// NewPostgresPersister falla rápido (ping con timeout) en vez de bloquear.
func TestPostgresPersister_UnreachableDSN_FailsFastAtConstruction(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker no disponible en PATH")
	}
	start := time.Now()
	_, err := NewPostgresPersister("postgres://postgres:postgres@127.0.0.1:1/nope?sslmode=disable&connect_timeout=1", 10)
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("esperaba error con DSN inalcanzable")
	}
	if elapsed > 10*time.Second {
		t.Errorf("NewPostgresPersister tardó %v con DB inalcanzable, esperaba fallo rápido", elapsed)
	}
}
