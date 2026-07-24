//go:build integration
// +build integration

// Mismo patrón docker que internal/quota/persister_postgres_integration_test.go
// y internal/alert/manager_integration_test.go. Correr con:
//
//	go test ./internal/handler/... -tags=integration -run TestAlertsHandler_Integration -v
//
// HU-EVO-013 / INT-alertsendpoint-postgres: valida GetAlerts contra
// PostgreSQL real -- admin ve todo, identidad con scope ve solo lo suyo,
// paginación respeta el filtrado RBAC (no pagina antes de filtrar).

package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"api-llm-gateway/internal/auth"
)

const testPGContainer = "alerts-handler-it-pg"

func startTestPostgres(t *testing.T) *sql.DB {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker no disponible en PATH, saltando integration test de AlertsHandler")
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
		t.Skipf("no se pudo levantar postgres real: %v\n%s", err, out)
	}
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-f", testPGContainer).Run() })

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
				if _, err := db.Exec(`
					CREATE TABLE provider_alerts (
						id SERIAL PRIMARY KEY,
						provider_id TEXT NOT NULL,
						model_id TEXT NOT NULL,
						severity TEXT NOT NULL,
						message TEXT NOT NULL,
						alert_time TIMESTAMPTZ NOT NULL DEFAULT now(),
						updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
						resolved_at TIMESTAMPTZ
					)`); err != nil {
					t.Fatalf("crear tabla: %v", err)
				}
				return db
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("postgres de test no respondió a tiempo: %v", lastErr)
	return nil
}

// HU-EVO-013 AC1/AC2/AC3/AC4: admin ve todas las alertas; una identidad con
// scope capability:coding ve solo la alerta de Groq (modelo con capability
// coding), no la de vision (capability distinta, AC4).
func TestAlertsHandler_Integration_RBACFiltering(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()

	if _, err := db.Exec(`
		INSERT INTO provider_alerts (provider_id, model_id, severity, message)
		VALUES
			('groq', 'mixtral', 'warning', 'groq remaining < 10%'),
			('openai', 'gpt4-vision', 'critical', 'openai EXHAUSTED')
	`); err != nil {
		t.Fatalf("seed: %v", err)
	}

	capLookup := func(provider, model string) []string {
		if provider == "groq" {
			return []string{"coding"}
		}
		return []string{"vision"}
	}
	h := NewAlertsHandler(db, capLookup)

	// Admin: ve las 2 alertas.
	adminReq := httptest.NewRequest("GET", "/alerts", nil)
	adminReq = adminReq.WithContext(WithAdmin(adminReq.Context()))
	adminW := httptest.NewRecorder()
	h.ServeHTTP(adminW, adminReq)
	if adminW.Code != http.StatusOK {
		t.Fatalf("admin: esperaba 200, obtuve %d: %s", adminW.Code, adminW.Body.String())
	}
	var adminResp AlertsResponse
	if err := json.Unmarshal(adminW.Body.Bytes(), &adminResp); err != nil {
		t.Fatalf("admin: json inválido: %v", err)
	}
	if adminResp.Total != 2 {
		t.Fatalf("admin: esperaba total=2, obtuve %d", adminResp.Total)
	}

	// Cliente con scope capability:coding: solo ve la alerta de groq.
	id := auth.Identity{Subject: "t1", Tenant: "t1", Scopes: []string{"capability:coding"}}
	userReq := httptest.NewRequest("GET", "/alerts", nil)
	userReq = userReq.WithContext(auth.WithIdentity(userReq.Context(), id))
	userW := httptest.NewRecorder()
	h.ServeHTTP(userW, userReq)
	if userW.Code != http.StatusOK {
		t.Fatalf("user: esperaba 200, obtuve %d: %s", userW.Code, userW.Body.String())
	}
	var userResp AlertsResponse
	if err := json.Unmarshal(userW.Body.Bytes(), &userResp); err != nil {
		t.Fatalf("user: json inválido: %v", err)
	}
	if userResp.Total != 1 {
		t.Fatalf("user: esperaba total=1 (solo groq), obtuve %d: %+v", userResp.Total, userResp.Data)
	}
	if userResp.Data[0].Provider != "groq" {
		t.Fatalf("user: esperaba provider=groq, obtuve %q", userResp.Data[0].Provider)
	}
}

// HU-EVO-013 AC5: paginación aplica DESPUÉS del filtrado RBAC (no expone
// conteos que incluyan filas de otro scope).
func TestAlertsHandler_Integration_PaginationRespectsRBAC(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()

	for i := 0; i < 5; i++ {
		if _, err := db.Exec(`
			INSERT INTO provider_alerts (provider_id, model_id, severity, message)
			VALUES ('groq', $1, 'warning', 'test')
		`, fmt.Sprintf("model-%d", i)); err != nil {
			t.Fatalf("seed fila %d: %v", i, err)
		}
	}
	// Alerta de otro scope, no debe contar en la paginación del usuario.
	if _, err := db.Exec(`
		INSERT INTO provider_alerts (provider_id, model_id, severity, message)
		VALUES ('openai', 'gpt4-vision', 'critical', 'other scope')
	`); err != nil {
		t.Fatalf("seed vision: %v", err)
	}

	capLookup := func(provider, model string) []string {
		if provider == "groq" {
			return []string{"coding"}
		}
		return []string{"vision"}
	}
	h := NewAlertsHandler(db, capLookup)
	id := auth.Identity{Subject: "t1", Tenant: "t1", Scopes: []string{"capability:coding"}}

	req := httptest.NewRequest("GET", "/alerts?page=1&limit=2", nil)
	req = req.WithContext(auth.WithIdentity(req.Context(), id))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	var resp AlertsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json inválido: %v", err)
	}
	if resp.Total != 5 {
		t.Fatalf("esperaba total=5 (solo groq, tras filtro RBAC), obtuve %d", resp.Total)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("esperaba 2 filas en la página, obtuve %d", len(resp.Data))
	}
	for _, row := range resp.Data {
		if row.Provider != "groq" {
			t.Fatalf("página contiene fila fuera de scope: %+v", row)
		}
	}
}
