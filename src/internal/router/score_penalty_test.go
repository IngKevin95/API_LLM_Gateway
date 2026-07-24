package router

import (
	"testing"

	"api-llm-gateway/internal/registry"
	"api-llm-gateway/internal/tokenizer"
)

// perModelQuota satisface QuotaSource devolviendo un remaining distinto por
// modelo, para poder ejercitar scoreAll() (el código real de producción,
// router.go:161-167) en vez de reimplementar Score() en el test.
type perModelQuota struct{ remaining map[string]int }

func (q perModelQuota) Remaining(_, model string) int { return q.remaining[model] }

// modelos idénticos en calidad/latencia/costo: la única variable que puede
// explicar una diferencia de score es la penalización por cuota baja.
func penaltyTestModels() []registry.Model {
	return []registry.Model{
		{Name: "low-quota", ProviderID: "openai", Capabilities: []string{"chat"}, QualityScore: 80, LatencyP50ms: 300, CostPer1M: 10},
		{Name: "high-quota", ProviderID: "groq", Capabilities: []string{"chat"}, QualityScore: 80, LatencyP50ms: 300, CostPer1M: 10},
	}
}

// HU-EVO-009 AC1 (reapertura) — remaining < 20% del máximo entre candidatos
// penaliza el score real de scoreAll(); el test anterior (score_penalty_test.go
// original) reimplementaba Score() en un stub que nunca tocaba router.go.
func TestScoreAll_PenalizesLowRemainingQuota(t *testing.T) {
	models := penaltyTestModels()
	r := New(stubSource{models: models}, allHealthy{}, perModelQuota{remaining: map[string]int{
		"low-quota":  15,  // 15% del máximo (100) -> por debajo del umbral 20%
		"high-quota": 100, // máximo -> sin penalización
	}}, tokenizer.NewHeuristic())

	scores := r.scoreAll(models)

	if scores["low-quota"] >= scores["high-quota"] {
		t.Fatalf("esperaba low-quota penalizado por debajo de high-quota, obtuve low=%.4f high=%.4f",
			scores["low-quota"], scores["high-quota"])
	}
	// La penalización es -wPenalty*1.0 = -0.20 exacto sobre el score de
	// low-quota respecto de si no hubiera penalización (remaining=quMax).
	unpenalized := r.scoreAll([]registry.Model{
		{Name: "low-quota", ProviderID: "openai", Capabilities: []string{"chat"}, QualityScore: 80, LatencyP50ms: 300, CostPer1M: 10},
	})
	// Con un solo candidato, quMax==quMin==remaining -> norm=1, sin penalización
	// (remaining >= threshold porque threshold = quMax*0.2 = remaining*0.2 < remaining).
	if unpenalized["low-quota"]-scores["low-quota"] < 0.15 {
		t.Errorf("delta de penalización esperado >= wPenalty(0.20) aprox, obtuve %.4f", unpenalized["low-quota"]-scores["low-quota"])
	}
}

// HU-EVO-009 AC2 (reapertura) — remaining exactamente en el umbral (20% del
// máximo) NO se penaliza (comparación estricta "<" en router.go:165).
func TestScoreAll_NoPenalization_AtExactThreshold(t *testing.T) {
	models := penaltyTestModels()
	r := New(stubSource{models: models}, allHealthy{}, perModelQuota{remaining: map[string]int{
		"low-quota":  20, // exactamente 20% de 100 -> NO penalizado (threshold = 20, 20 < 20 es falso)
		"high-quota": 100,
	}}, tokenizer.NewHeuristic())

	scores := r.scoreAll(models)
	// Sin penalización, low-quota solo pierde por el eje quota (quotaN menor),
	// nunca por el bloque completo de wPenalty=0.20.
	quotaOnlyDelta := wQuota * 1.0 // peso máximo del eje quota
	if scores["high-quota"]-scores["low-quota"] > quotaOnlyDelta+0.01 {
		t.Errorf("diferencia de score excede el eje quota solo (sin penalty): delta=%.4f, esperaba <= %.4f",
			scores["high-quota"]-scores["low-quota"], quotaOnlyDelta+0.01)
	}
}

// HU-EVO-009 AC3 (reapertura) — proveedor con remaining=0 recibe la máxima
// penalización disponible.
func TestScoreAll_ExhaustedProvider_MaxPenalty(t *testing.T) {
	models := penaltyTestModels()
	r := New(stubSource{models: models}, allHealthy{}, perModelQuota{remaining: map[string]int{
		"low-quota":  0,
		"high-quota": 100,
	}}, tokenizer.NewHeuristic())

	scores := r.scoreAll(models)
	if scores["low-quota"] >= scores["high-quota"] {
		t.Fatalf("esperaba proveedor agotado (remaining=0) penalizado por debajo, obtuve low=%.4f high=%.4f",
			scores["low-quota"], scores["high-quota"])
	}
}
