package tokenizer

import "testing"

// AC1 (HU-035) — Happy: payload dentro de ventana (crudo + buffer 20%) → cabe.
func TestFitsWindow_WithinWindow(t *testing.T) {
	tk := NewHeuristic()
	// 80k tokens estimados, buffer 20% = 96k, modelo soporta 100k → cabe.
	if !tk.FitsWindow(80_000, 100_000) {
		t.Errorf("esperaba que 80k (+20%%=96k) quepa en ventana de 100k")
	}
}

// AC2 (HU-035) — Error: payload excede ventana → no cabe.
func TestFitsWindow_ExceedsWindow(t *testing.T) {
	tk := NewHeuristic()
	// 120k tokens estimados > 100k → no cabe (ni con buffer).
	if tk.FitsWindow(120_000, 100_000) {
		t.Errorf("esperaba que 120k NO quepa en ventana de 100k")
	}
}

// AC3 (HU-035) — Edge: buffer 20% empuja fuera aunque el crudo cabría.
func TestFitsWindow_BufferPushesOut(t *testing.T) {
	tk := NewHeuristic()
	// 85k crudo cabría en 100k, pero 85k*1.2=102k > 100k → no cabe.
	if tk.FitsWindow(85_000, 100_000) {
		t.Errorf("esperaba que el buffer 20%% descarte 85k en ventana de 100k")
	}
}

// La estimación heurística es determinista y > 0 para texto no vacío.
func TestEstimate_Deterministic(t *testing.T) {
	tk := NewHeuristic()
	a := tk.Estimate("hola mundo esto es una prueba")
	b := tk.Estimate("hola mundo esto es una prueba")
	if a != b {
		t.Errorf("estimación no determinista: %d != %d", a, b)
	}
	if a == 0 {
		t.Errorf("esperaba estimación > 0 para texto no vacío")
	}
}
