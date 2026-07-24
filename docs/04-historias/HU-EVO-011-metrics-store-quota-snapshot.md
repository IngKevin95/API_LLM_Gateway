---
id: HU-EVO-011
titulo: Extender metrics.Store con snapshot de cuota por proveedor/modelo
epica: EP-EVO-003
prioridad: Must
complejidad: M
estado: draft
---

# Extender metrics.Store con snapshot de cuota por proveedor/modelo

Como **desarrollador del Gateway**, quiero **que metrics.Store llame `quota.Manager.Snapshot()` cada vez que `/metrics` es consultado**, para **exponer desglose de cuota actual por proveedor y modelo en el endpoint**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — snapshot devuelto | Dado que Quota Manager tiene learned quota para Groq | Cuando GET `/metrics` se ejecuta | Entonces respuesta incluye bloque `quota: [{provider: "groq", model: "mixtral", limit: 14400, remaining: 14200, reset_at: "2026-07-24T19:00:00Z", healthy: true}]` |
| 2 | Happy — múltiples modelos | Dado que Mistral tiene 3 modelos | Cuando GET `/metrics` se ejecuta | Entonces lista los 3 con remaining individual para cada modelo |
| 3 | Error — proveedor sin quota learned | Dado que un proveedor recién agregado nunca respondió | Cuando GET `/metrics` se ejecuta | Entonces lista con `remaining: <quota_hint>` (valor inicial del YAML) y `learned_at: null` |
| 4 | Edge — respuesta rápida | Dado que hay 25 proveedores x 5 modelos = 125 cuotas | Cuando GET `/metrics` se ejecuta | Entonces respuesta < 100ms (snapshot es lectura en RAM, sin I/O) |
| 5 | Edge — auth respetado | Dado que usuario no-admin consulta GET `/metrics` | Cuando request llega con scope `capability:coding` | Entonces respuesta filtra cuota solo de modelos que usuario puede usar |

## Checklist INVEST

- [x] Independent — depende de HU-EVO-007 (learning en RAM)
- [x] Negotiable — schema de respuesta extensible
- [x] Valuable — visibilidad de cuota en operaciones
- [x] Estimable — snapshot method + serialización
- [x] Small — 1 día
- [x] Testable — mock quota.Manager, verifica structure

## Notas técnicas

Metrics Store en `src/internal/metrics/store.go`:

```go
type QuotaSnapshot struct {
    Provider string `json:"provider"`
    Model string `json:"model"`
    Limit int64 `json:"limit"`
    Remaining int64 `json:"remaining"`
    ResetAt *time.Time `json:"reset_at,omitempty"`
    Healthy bool `json:"healthy"`
}

func (s *Store) GetMetrics(ctx context.Context) Metrics {
    // ... existente
    
    // NUEVO:
    quotas := s.quota.Snapshot()  // Lee desde Manager en RAM
    m.Quota = quotas
    
    return m
}
```

---

## Relación con existentes

- Extiende: `src/internal/metrics/store.go` (HU-060)
- Usa: `src/internal/quota/manager.go` (HU-EVO-007)
- Alimenta: HU-EVO-014 (UI React)
