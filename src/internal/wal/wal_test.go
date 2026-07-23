package wal_test

import (
	"path/filepath"
	"testing"

	"api-llm-gateway/internal/audit"
	"api-llm-gateway/internal/wal"
)

func ev(id string) audit.Event {
	return audit.Event{ID: id, Timestamp: 0, Tokens: 10, Status: 200, User: "c1"}
}

// HU-039 AC1 — Happy: append de N eventos → recuperables desde disco (append-only).
func TestWAL_AppendAndRecover(t *testing.T) {
	w, err := wal.New(t.TempDir(), 1<<20)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	for i := 0; i < 1000; i++ {
		if err := w.Append(ev("e")); err != nil {
			t.Fatalf("Append %d: %v", i, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	got, err := wal.New(w.Dir(), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	defer got.Close()
	recovered, err := got.Recover()
	if err != nil {
		t.Fatalf("Recover: %v", err)
	}
	if len(recovered) != 1000 {
		t.Errorf("esperaba 1000 registros recuperados, obtuve %d", len(recovered))
	}
}

// HU-039 AC3 — Edge: rotación al superar el tamaño máximo.
func TestWAL_Rotation(t *testing.T) {
	dir := t.TempDir()
	w, err := wal.New(dir, 256) // umbral chico para forzar rotación
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 200; i++ {
		if err := w.Append(ev("evento-con-algo-de-contenido")); err != nil {
			t.Fatal(err)
		}
	}
	w.Close()

	files, _ := filepath.Glob(filepath.Join(dir, "wal-*.log"))
	if len(files) < 2 {
		t.Errorf("esperaba ≥2 archivos WAL tras rotación, obtuve %d (%v)", len(files), files)
	}
}

// journey_smoke SS1: escribir → rotar → recuperar end-to-end desde disco.
// HU-039 AC2 — Error: Recover lee activo + archivados para replay.
func TestWAL_RecoverAcrossRotations(t *testing.T) {
	dir := t.TempDir()
	w, _ := wal.New(dir, 256)
	for i := 0; i < 200; i++ {
		_ = w.Append(ev("x"))
	}
	w.Close()

	fresh, _ := wal.New(dir, 256)
	defer fresh.Close()
	recovered, err := fresh.Recover()
	if err != nil {
		t.Fatalf("Recover: %v", err)
	}
	if len(recovered) != 200 {
		t.Errorf("esperaba 200 eventos a través de rotaciones, obtuve %d", len(recovered))
	}
}
