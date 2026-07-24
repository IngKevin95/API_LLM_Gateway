---
id: HU-EVO-007
titulo: Implementar LearnFromHeaders() en Quota Manager con actualización RAM
epica: EP-EVO-002
prioridad: Must
complejidad: M
estado: draft
---

# Implementar LearnFromHeaders() en Quota Manager con actualización RAM

Como **componente del Quota Manager**, quiero **que después de cada request, se llame `LearnFromHeaders(providerID, modelID, quotaInfo)` para actualizar remaining en RAM**, para **mantener cuota fresca sin I/O a PostgreSQL en el path crítico**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — update remaining | Dado que OpenAI devuelve `QuotaInfo{Remaining: 9950}` | Cuando LearnFromHeaders() se ejecuta | Entonces `Remaining("openai")` es 9950 inmediatamente (atomic update en RAM) |
| 2 | Happy — learned > quota_hint | Dado que servidor dice remaining = 1M pero quota_hint = 500K | Cuando Learn actualiza | Entonces toma el valor aprendido más alto (trusted hacia servidor) |
| 3 | Error — learned < 0 (overshoot) | Dado que servidor devuelve `Remaining: -100` (consumo excedió limite) | Cuando Learn procesa | Entonces clampea a 0 y marca como `exhausted` |
| 4 | Edge — race condition múltiples requests | Dado que 10 requests paralelos a OpenAI llegan simultáneamente | Cuando todos llaman LearnFromHeaders() | Entonces actualización en RAM es atómica (mutex), final state = último valor recibido |
| 5 | Edge — reset de ventana | Dado que un provider tenía remaining = 0 ayer | Cuando hoy llega un request y server devuelve `ResetAt: yesterday + 24h` y `Remaining: <nuevo>` | Entonces detecta reset (ResetAt cruzó ahora), actualiza remaining, y reactiva en Router |

## Checklist INVEST

- [x] Independent — depende de HU-EVO-006 (headers parseados)
- [x] Negotiable — threshold configurable para agotamiento
- [x] Valuable — cuota real en RAM, sin latencia de DB
- [x] Estimable — atomic updates + time.Now() para reset detection
- [x] Small — 1-2 días
- [x] Testable — race condition tests + mock servers

## Notas técnicas

Quota Manager en `src/internal/quota/manager.go`:

```go
func (m *Manager) LearnFromHeaders(providerID, modelID string, quota QuotaInfo) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    key := providerID // o (providerID, modelID)
    current := m.quotas[key]
    
    // Detecta reset (ResetAt cruzó ahora)
    if quota.ResetAt != nil && quota.ResetAt.After(current.ResetAt) {
        m.quotas[key] = Quota{
            Limit: quota.Limit,
            Remaining: quota.Remaining,
            ResetAt: quota.ResetAt,
            LearnedAt: time.Now(),
        }
    } else if quota.Remaining > current.Remaining {
        // Puede ser reset o error del servidor — actualiza conservador
        m.quotas[key].Remaining = quota.Remaining
    }
    // Si Remaining < current, ignora (puede ser race de otro request paralelo)
    
    return nil
}
```

---

## Relación con existentes

- Extiende: `src/internal/quota/manager.go` (HU-006)
- Usa: HU-EVO-006 (QuotaInfo extraído de headers)
- Alimenta: HU-EVO-008 (persistencia async en DB)
- Integra con: HU-EVO-009 (Router penaliza < 20%)
