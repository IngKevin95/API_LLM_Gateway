---
id: HU-EVO-009
titulo: Router penaliza score cuando remaining < 20%
epica: EP-EVO-002
prioridad: Should
complejidad: S
estado: draft
---

# Router penaliza score cuando remaining < 20%

Como **componente del Router**, quiero **que al calcular score de un proveedor, si remaining < 20% del limite, reste puntos automáticamente**, para **deprioritizar proveedores con cuota baja sin breaking change al algoritmo existente**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — penalización aplicada | Dado que Groq tiene limit=100 remaining=15 (15%) | Cuando Router.Score() se ejecuta | Entonces decrementa `score -= 5` (penalización fija) o `score *= 0.5` (penalización %) |
| 2 | Happy — sin penalización si > 20% | Dado que OpenAI tiene remaining=25 limit=100 (25%) | Cuando Router.Score() se ejecuta | Entonces score sin penalización, max priority |
| 3 | Edge — remaining = 0 | Dado que Cerebras está agotado (remaining=0) | Cuando Router.Score() se ejecuta | Entonces score = 0 o muy bajo, excluido de selección automáticamente |
| 4 | Edge — competencia múltiples en cuota baja | Dado que 3 proveedores tienen <20% remaining | Cuando Router elige entre ellos | Entonces elige el que tenga mejor latencia entre los penalizados |
| 5 | Edge — failover respeta penalización | Dado que proveedor 1 (penalizado) falló | Cuando failover intenta siguiente | Entonces intenta proveedor 2 (no penalizado) primero, luego fallback a penalizados si es necesario |

## Checklist INVEST

- [x] Independent — integración con Router existente, sin refactor
- [x] Negotiable — penalización (fija vs %) configurable por proveedor
- [x] Valuable — evita exhausting cuotas de proveedores casi agotados
- [x] Estimable — 1 línea de código en Router.score()
- [x] Small — 0.5 días
- [x] Testable — test de score con remaining variables

## Notas técnicas

Router en `src/internal/router/router.go`:

```go
func (r *Router) Score(provider Provider, model Model) float64 {
    score := 0.0
    // existente: calidad + latencia + disponibilidad + costo
    
    // NUEVO:
    remaining := r.quota.Remaining(provider.ID, model.ID)
    limit := r.quota.Limit(provider.ID, model.ID)
    if remaining < limit*0.2 { // 20%
        score *= 0.5  // Penalización 50%
    }
    
    return score
}
```

---

## Relación con existentes

- Integra: `src/internal/router/router.go` (HU-002a)
- Usa: HU-EVO-007 (learned remaining), HU-EVO-009 (learning en RAM)
- Integra con: failover chain (HU-004a)
