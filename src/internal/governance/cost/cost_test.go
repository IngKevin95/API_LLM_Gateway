package cost_test

import (
	"testing"

	"api-llm-gateway/internal/governance/cost"
)

// HU-007 AC1 — Costo atribuido: registra costo = tokens × tarifa, atribuido a (agente, proveedor, modelo).
func TestCostTracker_RecordsCost_WithAttribution(t *testing.T) {
	tracker := cost.NewLocalCostTracker()

	// Dado: modelo con cost_per_token definido y agente identificado
	agentID := "claude-code"
	providerID := "openai"
	modelID := "gpt-4o"
	costPerToken := 0.00003 // $0.00003 per token
	inputTokens := 1000
	outputTokens := 500

	// Cuando: se registra una petición completada
	err := tracker.RecordCost(cost.CostRecord{
		AgentID:       agentID,
		ProviderID:    providerID,
		ModelID:       modelID,
		InputTokens:   int64(inputTokens),
		OutputTokens:  int64(outputTokens),
		CostPerToken:  costPerToken,
	})

	// Entonces: se registra costo atribuido correctamente
	if err != nil {
		t.Fatalf("RecordCost failed: %v", err)
	}

	// Verificar que se puede consultar
	totalCost, err := tracker.GetTotalCost(agentID, providerID)
	if err != nil {
		t.Fatalf("GetTotalCost failed: %v", err)
	}

	expectedCost := float64((inputTokens + outputTokens)) * costPerToken
	if totalCost < expectedCost-0.0001 || totalCost > expectedCost+0.0001 {
		t.Errorf("expected cost ~%.6f, got %.6f", expectedCost, totalCost)
	}
}

// HU-007 AC2 — Tarifa faltante: registra con costo "desconocido" sin perder el registro.
func TestCostTracker_MissingTariff_RecordsAsUnknown(t *testing.T) {
	tracker := cost.NewLocalCostTracker()

	// Dado: modelo sin cost_per_token
	err := tracker.RecordCost(cost.CostRecord{
		AgentID:       "claude-code",
		ProviderID:    "openai",
		ModelID:       "gpt-4o",
		InputTokens:   1000,
		OutputTokens:  500,
		CostPerToken:  0, // Tarifa desconocida
	})

	// Entonces: se registra sin error (aunque costo es 0)
	if err != nil {
		t.Fatalf("RecordCost should handle missing tariff: %v", err)
	}

	// Verificar que la petición se registró (aunque sin costo)
	cost, err := tracker.GetTotalCost("claude-code", "openai")
	if err != nil {
		t.Fatalf("GetTotalCost failed: %v", err)
	}
	if cost != 0 {
		t.Errorf("expected cost 0 for unknown tariff, got %.6f", cost)
	}
}

// HU-007 AC3 — Modelo gratuito: costo 0, cuenta en volumen pero no en gasto.
func TestCostTracker_FreeModel_RegistersZeroCost(t *testing.T) {
	tracker := cost.NewLocalCostTracker()

	// Dado: modelo con costo 0
	err := tracker.RecordCost(cost.CostRecord{
		AgentID:       "claude-code",
		ProviderID:    "openai",
		ModelID:       "gpt-3.5-turbo", // gratuito (en test)
		InputTokens:   1000,
		OutputTokens:  500,
		CostPerToken:  0, // gratuito
	})

	// Entonces: se registra con costo 0
	if err != nil {
		t.Fatalf("RecordCost failed: %v", err)
	}

	cost, _ := tracker.GetTotalCost("claude-code", "openai")
	if cost != 0 {
		t.Errorf("expected cost 0 for free model, got %.6f", cost)
	}
}

// HU-007 AC4 — Petición con failover: costo atribuido solo al proveedor que respondió.
func TestCostTracker_Failover_AttributesToCorrectProvider(t *testing.T) {
	tracker := cost.NewLocalCostTracker()

	// Dado: petición resuelta tras fallar 2 proveedores
	agentID := "claude-code"

	// Provider 1 falló (no registra)
	// Provider 2 falló (no registra)
	// Provider 3 respondió
	err := tracker.RecordCost(cost.CostRecord{
		AgentID:       agentID,
		ProviderID:    "openrouter", // El que efectivamente respondió
		ModelID:       "gpt-4o",
		InputTokens:   1000,
		OutputTokens:  500,
		CostPerToken:  0.00003,
	})

	// Entonces: costo atribuido solo a openrouter
	if err != nil {
		t.Fatalf("RecordCost failed: %v", err)
	}

	openrouterCost, _ := tracker.GetTotalCost(agentID, "openrouter")
	if openrouterCost == 0 {
		t.Error("expected cost in openrouter")
	}
}

// HU-007 AC5 — Stream abortado: contabiliza tokens parciales y ajusta costo exacto.
func TestCostTracker_StreamAborted_RecordsPartialTokens(t *testing.T) {
	tracker := cost.NewLocalCostTracker()

	// Dado: stream abortado a mitad
	// Cliente envía N tokens, pero solo recibe M antes de abortar
	err := tracker.RecordCost(cost.CostRecord{
		AgentID:       "claude-code",
		ProviderID:    "openai",
		ModelID:       "gpt-4o",
		InputTokens:   1000,
		OutputTokens:  250, // Partial — solo 250 de los posibles 500 tokens se enviaron
		CostPerToken:  0.00003,
	})

	// Entonces: se registra el costo exacto de los tokens parciales
	if err != nil {
		t.Fatalf("RecordCost failed: %v", err)
	}

	totalCost, _ := tracker.GetTotalCost("claude-code", "openai")
	expectedCost := float64(1000+250) * 0.00003
	if totalCost < expectedCost-0.0001 || totalCost > expectedCost+0.0001 {
		t.Errorf("expected cost ~%.6f for partial tokens, got %.6f", expectedCost, totalCost)
	}
}
