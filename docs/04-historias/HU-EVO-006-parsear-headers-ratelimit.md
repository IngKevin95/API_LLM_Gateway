---
id: HU-EVO-006
titulo: Parsear headers estándar X-RateLimit-* por adapter
epica: EP-EVO-002
prioridad: Must
complejidad: M
estado: draft
---

# Parsear headers estándar X-RateLimit-* por adapter

Como **arquitecto del Gateway**, quiero **que cada adapter extraiga automáticamente X-RateLimit-Limit, Remaining, Reset del response HTTP**, para **alimentar el aprendizaje de cuota en tiempo real sin cambios de API en adapters**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — OpenAI headers | Dado que OpenAI devuelve `X-RateLimit-Limit-Requests: 10000, X-RateLimit-Remaining-Requests: 9950` | Cuando adapter.Chat() devuelve respuesta | Entonces extrae y devuelve como `{Limit: 10000, Remaining: 9950, ResetAt: ...}` |
| 2 | Happy — Anthropic headers | Dado que Anthropic devuelve `anthropic-ratelimit-requests-limit: 60000, anthropic-ratelimit-requests-remaining: 59999` | Cuando adapter.Chat() devuelve | Entonces normaliza a schema estándar y devuelve |
| 3 | Happy — Groq headers | Dado que Groq devuelve `x-ratelimit-limit-requests: 30, x-ratelimit-remaining-requests: 29` | Cuando adapter.Chat() devuelve | Entonces normaliza y devuelve (case-insensitive) |
| 4 | Error — provider sin headers de cuota | Dado que un provider no devuelve ningún header X-RateLimit-* | Cuando adapter.Chat() devuelve | Entonces devuelve `{Limit: 0, Remaining: 0, ResetAt: nil}` sin crash (graceful fallback) |
| 5 | Edge — múltiples formatos de reset | Dado que algunos headers usan `X-RateLimit-Reset: <unix-timestamp>` y otros `Retry-After: <seconds>` | Cuando adapter parsea | Entonces detecta formato y convierte a `time.Time` normalizado |

## Checklist INVEST

- [x] Independent — no toca lógica de adapters existentes, solo adición
- [x] Negotiable — fallback graceful si headers ausentes
- [x] Valuable — habilita learning sin cambios de API
- [x] Estimable — parseo de headers + normalización
- [x] Small — 2 días
- [x] Testable — mocks de respuestas HTTP con/sin headers

## Notas técnicas

Struct nuevo en `src/internal/adapter/types.go`:
```go
type QuotaInfo struct {
    Limit     int64
    Remaining int64
    ResetAt   *time.Time
}

type Response struct {
    Content string
    Quota   QuotaInfo  // Nuevo
    // ... existente
}
```

Header mapping en cada adapter (ej: `openai/adapter.go`):
```go
func (a *Adapter) extractQuota(headers http.Header) QuotaInfo {
    // Parsea X-RateLimit-Limit-Requests, etc.
    // Devuelve QuotaInfo normalizado
}
```

---

## Relación con existentes

- Extiende: `src/internal/adapter/adapter.go` interface (sin breaking change)
- Integra: `src/internal/adapter/{openai,anthropic,google,etc}/adapter.go`
- Alimenta: HU-EVO-007 (aprendizaje)
- Requiere: HU-EVO-001 (adapters genéricos soportan esto)
