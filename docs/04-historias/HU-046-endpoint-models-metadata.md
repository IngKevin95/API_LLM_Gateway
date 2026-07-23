---
id: HU-046
titulo: Endpoint /v1/models mejorado con metadata (capabilities, provider, cost)
epica: EP-010
prioridad: Should
complejidad: S
estado: draft
---

# Endpoint /v1/models mejorado con metadata (capabilities, provider, cost)

Como **integrador o cliente conociendo qué modelos están disponibles**, quiero **consultar `/v1/models` y obtener metadata (capacidades, proveedor, disponibilidad, costo)**, para **elegir modelos informadamente o para debugging**.

Contexto: Endpoint `/v1/models` existe (HU-012c rough draft) pero devuelve hardcoded 2 modelos mock. Actividad 5 de EP-010.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — listar todos los modelos | Dado que Registry tiene 5 modelos configurados | Cuando cliente GET `/v1/models` | Entonces retorna JSON con array `data` contiene todos, con `id, object, capabilities, provider, available, cost_per_1m_tokens` |
| 2 | Happy — filtering por capability | Dado que cliente quiere solo modelos "reasoning" | Cuando GET `/v1/models?capability=reasoning` | Entonces retorna solo modelos que tienen "reasoning" en capabilities array |
| 3 | Happy — sorting por score | Dado que cliente envía `?sort_by=quality` | Cuando se procesan | Entonces modelos ordenados por quality_score descendente |
| 4 | Happy — metadata completa | Dado que modelo en Registry tiene todos los fields | Cuando se retorna | Entonces respuesta incluye: id, object, capabilities[], provider, available (true/false), latency_p50_ms, quality_score, cost_per_1m_tokens, max_context_tokens |
| 5 | Error — capability inválida en filter | Dado que cliente filtra por `?capability=xyz` (no existe) | Cuando se procesa | Entonces retorna 400 con "unknown capability" o empty array (depende de diseño) |
| 6 | Edge — modelo no disponible (down) | Dado que Health Monitor marcó proveedor como down | Cuando se lista | Entonces `available: false` para esos modelos, pero aún aparecen en respuesta (solo con flag) |

## Checklist INVEST

- [x] Independent — se apoya en Registry (EP-001), Health Monitor (EP-002), handler OpenAI (EP-005); sin bloqueos
- [x] Negotiable — alcance: ampliar handler existente + queries al Registry
- [x] Valuable — permite debugging, eligibility checks
- [x] Estimable — ampliar handler (~4 horas)
- [x] Small — un sprint
- [x] Testable — GET /v1/models con varios states, filtering

## Notas técnicas

Response esperada:
```json
{
  "object": "list",
  "data": [
    {
      "id": "gpt-4o",
      "object": "model",
      "owned_by": "openai",
      "created": 1234567890,
      "capabilities": ["chat", "vision", "reasoning"],
      "provider": "openai",
      "available": true,
      "latency_p50_ms": 800,
      "quality_score": 95,
      "cost_per_1m_tokens": 15,
      "max_context_tokens": 128000
    },
    {
      "id": "claude-opus-4",
      "object": "model",
      "owned_by": "anthropic",
      "created": 1234567891,
      "capabilities": ["chat", "reasoning", "vision"],
      "provider": "anthropic",
      "available": false,
      "latency_p50_ms": 900,
      "quality_score": 92,
      "cost_per_1m_tokens": 20,
      "max_context_tokens": 200000
    }
  ]
}
```

Query params opcionales:
- `?capability=coding` → filtrar por capability
- `?provider=openai` → filtrar por proveedor
- `?available=true` → solo disponibles
- `?sort_by=quality|cost|latency` → ordenar

Handler location: `internal/api/openai/handler.go` (reutilizar):
```go
func (h *Handler) HandleModels(w http.ResponseWriter, r *http.Request) { ... }
```

Data source: Registry (EP-001) ya tiene ModelsFor(capability), ModelNames(), FindModel()
