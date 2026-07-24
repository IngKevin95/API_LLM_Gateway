---
id: HU-EVO-010
titulo: Manejo de 429 con reset timeout y Retry-After
epica: EP-EVO-002
prioridad: Should
complejidad: M
estado: draft
---

# Manejo de 429 con reset timeout y Retry-After

Como **manejador de failover**, quiero **que cuando un proveedor devuelva 429, se extraiga `Retry-After` header y se haga blacklist del proveedor hasta ese tiempo**, para **respetar límites de rate-limit sin violar ToS de proveedores**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — Retry-After en segundos | Dado que Groq devuelve 429 con `Retry-After: 60` | Cuando adapter.Chat() devuelve error 429 | Entonces extrae 60, retira Groq de selección por 60s, y failover al siguiente |
| 2 | Happy — Retry-After en fecha | Dado que Mistral devuelve 429 con `Retry-After: Wed, 23 Jul 2026 19:00:00 GMT` | Cuando adapter procesa | Entonces parsea fecha, calcula delta a ahora, usa ese tiempo |
| 3 | Happy — sin Retry-After | Dado que un provider devuelve 429 sin header `Retry-After` | Cuando adapter procesa | Entonces asume default 30s y retira por 30s |
| 4 | Error — 429 mid-stream | Dado que un stream comenzó correctamente | Cuando a mitad OpenAI devuelve 429 en chunk | Entonces aborta el stream, retorna error (no failover mid-stream), y retira OpenAI |
| 5 | Edge — múltiples 429 consecutivos | Dado que Cerebras recibe 5 × 429 en 2 minutos | Cuando Health Monitor acumula | Entonces incrementa blacklist exponencialmente (30s → 60s → 120s → 240s) |

## Checklist INVEST

- [x] Independent — integración con failover existente (HU-004a)
- [x] Negotiable — default retry delay configurable
- [x] Valuable — respeta ToS, evita ban
- [x] Estimable — parsing Retry-After + blacklist duration
- [x] Small — 1-2 días
- [x] Testable — mock 429s con/sin Retry-After

## Notas técnicas

Adapter response extendido (en HU-EVO-006):
```go
type Error struct {
    StatusCode int
    Message string
    RetryAfter *time.Duration  // Extraído de header
}
```

Failover en `src/internal/failover/failover.go`:

```go
if err.StatusCode == 429 {
    retryAfter := err.RetryAfter
    if retryAfter == nil {
        d := 30 * time.Second
        retryAfter = &d
    }
    h.health.Blacklist(provider.ID, *retryAfter)
    return h.failoverNext(capability) // Siguiente en cadena
}
```

---

## Relación con existentes

- Integra: `src/internal/failover/failover.go` (HU-004a/c)
- Usa: HU-EVO-006 (headers parseados con RetryAfter)
- Coordina con: HU-EVO-004 (health monitor blacklist)
