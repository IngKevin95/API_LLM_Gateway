package quota_test

import (
	"testing"
	"time"

	"api-llm-gateway/internal/governance/quota"
)

// HU-006 AC1 — Dentro de cuota: contador incrementa, cuota restante consultable.
func TestQuotaManager_WithinQuota_AllowsRequest(t *testing.T) {
	// Dado: proveedor con cuota diaria de 500 requests, 100 ya consumidos
	mgr := quota.NewLocalQuotaManager()
	providerID := "openai"
	apiKey := "key1"
	quotaLimit := int64(500)
	consumed := int64(100)

	err := mgr.SetQuota(providerID, apiKey, quota.Window{
		Period:   quota.Daily,
		Limit:    quotaLimit,
		Consumed: consumed,
		ResetAt:  time.Now().AddDate(0, 0, 1),
	})
	if err != nil {
		t.Fatalf("SetQuota failed: %v", err)
	}

	// Cuando: llega una petición
	allowed, remaining, err := mgr.TryConsume(providerID, apiKey, 1)

	// Entonces: se atiende, contador incrementa, cuota restante consultable
	if err != nil {
		t.Fatalf("TryConsume failed: %v", err)
	}
	if !allowed {
		t.Error("request should be allowed within quota")
	}
	if remaining != quotaLimit-consumed-1 {
		t.Errorf("expected remaining %d, got %d", quotaLimit-consumed-1, remaining)
	}

	// Verificar que contador se incrementó
	q, err := mgr.GetQuota(providerID, apiKey)
	if err != nil {
		t.Fatalf("GetQuota failed: %v", err)
	}
	if q.Consumed != consumed+1 {
		t.Errorf("consumed should be %d, got %d", consumed+1, q.Consumed)
	}
}

// HU-006 AC2 — Cuota agotada: rechazada y se excluye.
func TestQuotaManager_QuotaExhausted_RejectsRequest(t *testing.T) {
	mgr := quota.NewLocalQuotaManager()
	providerID := "anthropic"
	apiKey := "key2"
	quotaLimit := int64(100)

	mgr.SetQuota(providerID, apiKey, quota.Window{
		Period:   quota.Daily,
		Limit:    quotaLimit,
		Consumed: quotaLimit, // Agotada
		ResetAt:  time.Now().AddDate(0, 0, 1),
	})

	// Cuando: llega una petición
	allowed, _, err := mgr.TryConsume(providerID, apiKey, 1)

	// Entonces: rechazada (allowed=false)
	if err != nil {
		t.Fatalf("TryConsume failed: %v", err)
	}
	if allowed {
		t.Error("request should be rejected when quota exhausted")
	}
}

// HU-006 AC3 — Reinicio de ventana: contador se reinicia.
func TestQuotaManager_WindowReset_ResetsCounter(t *testing.T) {
	mgr := quota.NewLocalQuotaManager()
	providerID := "google"
	apiKey := "key3"

	// Dado: cuota agotada
	pastReset := time.Now().Add(-time.Hour)
	mgr.SetQuota(providerID, apiKey, quota.Window{
		Period:   quota.Daily,
		Limit:    100,
		Consumed: 100,
		ResetAt:  pastReset, // Ya pasó la hora de reset
	})

	// Cuando: se valida después del reset
	allowed, _, _ := mgr.TryConsume(providerID, apiKey, 1)

	// Entonces: contador se reinicia y permite consumo
	if !allowed {
		t.Error("request should be allowed after window reset")
	}
}

// HU-006 AC4 — Límite por tokens: rechaza si excedería.
func TestQuotaManager_TokenLimit_RejectsIfExceeds(t *testing.T) {
	mgr := quota.NewLocalQuotaManager()
	providerID := "openai"
	apiKey := "key4"

	// Dado: 1M token/día casi alcanzado, 999k consumidos, 1k restante
	mgr.SetQuota(providerID, apiKey, quota.Window{
		Period:   quota.Daily,
		Limit:    1_000_000,
		Consumed: 999_000, // 1k restante
		ResetAt:  time.Now().AddDate(0, 0, 1),
	})

	// Cuando: llega petición que excedería (50k tokens)
	allowed, _, _ := mgr.TryConsume(providerID, apiKey, 50_000)

	// Entonces: rechazada
	if allowed {
		t.Error("request should be rejected when tokens would exceed limit")
	}
}

// HU-006 AC5 — Race conditions: validación atómica.
func TestQuotaManager_RaceConditions_AtomicValidation(t *testing.T) {
	mgr := quota.NewLocalQuotaManager()
	providerID := "test"
	apiKey := "key5"

	// Dado: 1 token restante
	mgr.SetQuota(providerID, apiKey, quota.Window{
		Period:   quota.Daily,
		Limit:    100,
		Consumed: 99,
		ResetAt:  time.Now().AddDate(0, 0, 1),
	})

	// Cuando: 50 goroutines intentan consumir simultáneamente
	results := make(chan bool, 50)

	for i := 0; i < 50; i++ {
		go func() {
			ok, _, _ := mgr.TryConsume(providerID, apiKey, 1)
			results <- ok
		}()
	}

	// Recolectar resultados
	allowed := 0
	for i := 0; i < 50; i++ {
		if <-results {
			allowed++
		}
	}

	// Entonces: solo 1 permitido (atómico), 49 rechazadas
	if allowed != 1 {
		t.Errorf("expected 1 allowed, got %d", allowed)
	}
}
