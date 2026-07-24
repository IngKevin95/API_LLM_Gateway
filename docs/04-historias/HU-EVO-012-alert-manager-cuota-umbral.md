---
id: HU-EVO-012
titulo: Implementar Alert Manager que genera alertas cuando remaining < umbral
epica: EP-EVO-003
prioridad: Should
complejidad: M
estado: draft
---

# Implementar Alert Manager que genera alertas cuando remaining < umbral

Como **operador del Gateway**, quiero **que un Alert Manager revise cuota cada 1 minuto y genere alertas en PostgreSQL cuando remaining < umbral (default 10%)**, para **notificar operadores y usuarios de cuota baja proactivamente**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — alerta generada | Dado que Groq tiene remaining=1200 limit=14400 (8.3%) | Cuando Alert Manager corre | Entonces inserta en `provider_alerts` tabla con `severity: "warning"`, `message: "Groq remaining < 10%"`, `alert_time: now` |
| 2 | Happy — no duplica alertas | Dado que Groq ya tiene alerta activa de hace 5 minutos | Cuando Alert Manager corre de nuevo | Entonces no genera otra alerta, solo actualiza `updated_at` de la existente |
| 3 | Error — threshold 0 | Dado que Cerebras está agotado (remaining=0) | Cuando Alert Manager procesa | Entonces genera alerta con `severity: "critical"` y `message: "Cerebras EXHAUSTED"` |
| 4 | Edge — umbral configurable | Dado que operador quiere alertar en 30% en lugar de 10% | Cuando edita config | Entonces Alert Manager respeta nuevo umbral (configurable sin redeploy) |
| 5 | Edge — múltiples modelos alertan | Dado que Mistral tiene 3 modelos, 2 bajo umbral | Cuando Alert Manager procesa | Entonces genera 2 alertas (una por modelo), no una por proveedor |

## Checklist INVEST

- [x] Independent — depende de HU-EVO-007 (learned quota)
- [x] Negotiable — umbral y severidad configurables
- [x] Valuable — detección proactiva de cuota baja
- [x] Estimable — background worker + dedup logic
- [x] Small — 2 días
- [x] Testable — mock quota values, verifica inserts

## Notas técnicas

Alert Manager en `src/internal/alert/manager.go`:

```go
func (a *Manager) Check(ctx context.Context) error {
    quotas := a.quota.Snapshot()
    for _, q := range quotas {
        if q.Remaining < q.Limit * 0.1 { // 10%
            a.generateAlert(ctx, q, "warning")
        } else if q.Remaining == 0 {
            a.generateAlert(ctx, q, "critical")
        }
    }
}

func (a *Manager) generateAlert(ctx context.Context, quota Quota, severity string) {
    // Check if exists, upsert if not
    // Otherwise update updated_at
}
```

Schema:
```sql
CREATE TABLE provider_alerts (
    id SERIAL PRIMARY KEY,
    provider_id VARCHAR(255),
    model_id VARCHAR(255),
    severity VARCHAR(50), -- "warning" | "critical"
    message TEXT,
    alert_time TIMESTAMP,
    updated_at TIMESTAMP,
    resolved_at TIMESTAMP,
    UNIQUE (provider_id, model_id, alert_time)
);
```

---

## Relación con existentes

- Implementa: nuevo `src/internal/alert/manager.go`
- Usa: HU-EVO-007 (learned quota)
- Alimenta: HU-EVO-013 (filtrado RBAC), HU-EVO-015 (notificaciones)
