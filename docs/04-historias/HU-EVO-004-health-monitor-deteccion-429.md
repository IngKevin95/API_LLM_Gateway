---
id: HU-EVO-004
titulo: Health Monitor detecta y retira temporalmente proveedores con 429
epica: EP-EVO-001
prioridad: Should
complejidad: M
estado: lista
---

# Health Monitor detecta y retira temporalmente proveedores con 429

Como **operador del Gateway**, quiero **que Health Monitor detecte HTTP 429 de cualquier proveedor y lo retire automáticamente de la selección hasta que se recupere**, para **evitar requests fallidos consecutivos y mejorar failover**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — proveedor recupera | Dado que Groq respondió 429 hace 30s | Cuando Health Monitor lanza health check | Entonces ve respuesta 200, marca proveedor como `healthy=true` y lo reactiva en Router |
| 2 | Happy — retiro temporal | Dado que un request a Cerebras devolvió 429 con `Retry-After: 60` | Cuando Health Monitor lo procesa | Entonces retira Cerebras de selección por 60s y lo reactiva después |
| 3 | Error — 429 sin Retry-After | Dado que Mistral devuelve 429 sin header `Retry-After` | Cuando Health Monitor lo procesa | Entonces asume default 30s, retira, y reactiva después |
| 4 | Edge — 429 mid-stream | Dado que un Stream comenzó y a mitad devuelve 429 | Cuando se detecta el 429 | Entonces aborta stream, retorna error al cliente, y retira proveedor; no hay failover transparente mid-stream |
| 5 | Edge — múltiples 429 rápido | Dado que un proveedor recibe 3 × 429 en 10s | Cuando Health Monitor acumula fallos | Entonces incrementa blacklist duration exponencialmente (30s → 60s → 120s) |

## Checklist INVEST

- [x] Independent — integración con Health Monitor existente (HU-005)
- [x] Negotiable — duración de blacklist configurable por proveedor
- [x] Valuable — reduce retry storm, mejora resiliencia
- [x] Estimable — actualización de Health Monitor + Circuit Breaker
- [x] Small — 2 días
- [x] Testable — mock 429s, verifica reactivación

## Notas técnicas

Health Monitor en `src/internal/health/health.go` extendido:

```go
func (m *Monitor) RecordError(providerID, modelID string, status int, retryAfter int) {
    if status == 429 {
        m.blacklist[providerID] = time.Now().Add(time.Duration(retryAfter) * time.Second)
        m.exponentialBackoff(providerID) // Incrementa duración si hay fallos recientes
    }
}

func (m *Monitor) IsHealthy(providerID, modelID string) bool {
    if blacklistedUntil, ok := m.blacklist[providerID]; ok {
        if time.Now().Before(blacklistedUntil) {
            return false // Aún está en blacklist
        }
        delete(m.blacklist, providerID) // Expira blacklist
    }
    return true
}
```

---

## Relación con existentes

- Extiende: `src/internal/health/health.go` (HU-005, ya existe)
- Integra con: `src/internal/breaker/breaker.go` (Circuit Breaker)
- Usa: HU-EVO-001/002 (nuevos proveedores)
- Requiere: HU-EVO-007 (aprendizaje de headers con 429 detection)

## Change

Implementado por el openspec change `adapters-proveedores-gratuitos-curados` (EP-EVO-001, branch `feature/ep-evo-001-adapters-gratuitos`). Ver `openspec/changes/adapters-proveedores-gratuitos-curados/`.
