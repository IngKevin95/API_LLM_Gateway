---
id: HU-045
titulo: Parámetros Anthropic completos (temperature, top_k, tool_use, thinking)
epica: EP-010
prioridad: Must
complejidad: M
estado: draft
---

# Parámetros Anthropic completos (temperature, top_k, tool_use, thinking)

Como **desarrollador usando Free Claude Code o Claude Code**, quiero **enviar parámetros Anthropic nativos (temperature, top_k, top_p, thinking, tool_use)**, para **control fino sin perder funcionalidad de Claude**.

Contexto: Handler Anthropic `/v1/messages` actualmente copia solo `model, messages, max_tokens, stream, system`. Actividad 4 de EP-010.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — temperature y top_k | Dado que cliente Anthropic envía `temperature: 0.8, top_k: 40` | Cuando Handler procesa | Entonces ambos se copian a internal Request y se pasan al adapter Anthropic |
| 2 | Happy — thinking parameter | Dado que cliente envía `thinking: true` o `thinking: {budget_tokens: 5000}` | Cuando se procesa | Entonces thinking se configura en el adapter; respuesta incluye thinking blocks si aplica |
| 3 | Happy — tool_use completo | Dado que cliente envía `tools: [{name, description, input_schema}]` y `tool_choice: {type: "tool", name: "get_weather"}` | Cuando se procesa | Entonces tool_use blocks se preservan intactos, con tool_result handling incluido |
| 4 | Error — parámetro inválido | Dado que cliente envía `top_k: 0.5` (debe ser entero ≥ 1) | Cuando se valida | Entonces responde 400 Bad Request con "top_k must be integer ≥ 1" |
| 5 | Edge — tool_use traducida a OpenAI | Dado que cliente envía request Anthropic con tools, pero se enruta a OpenAI | Cuando se traduce | Entonces tool_use blocks se convierten a function call format compatible OpenAI, sin pérdida |
| 6 | Edge — thinking no soportado en fallback | Dado que cliente solicita thinking, pero enruta a OpenAI (no soporta thinking) | Cuando se procesa | Entonces el Gateway advierte en log y deshabilita thinking sin error, respondiendo normalmente |

## Checklist INVEST

- [x] Independent — se apoya en handler Anthropic (EP-005), adapters (EP-002); sin bloqueos
- [x] Negotiable — alcance: extension de tipos Anthropic + mapeo a otros adapters
- [x] Valuable — fin de la limitación "parámetros básicos solo"
- [x] Estimable — tipos + validación Anthropic (~6 horas)
- [x] Small — un sprint
- [x] Testable — request con thinking, tool_use, top_k

## Notas técnicas

Parámetros Anthropic a soportar:
- `temperature` (0-1, default 1.0) → adapter
- `top_k` (entero ≥ 1, default 0/disabled) → adapter
- `top_p` (0-1, default 1.0) → adapter
- `thinking` (boolean o {type, budget_tokens}) → adapter si soporta (Claude-opus/sonnet soportan)
- `tool_use` (array de tool definitions) → adapter
- `tool_choice` (auto, required, disabled, o específica {type, name}) → adapter

Mapeo adapter:
- Anthropic adapter: pasa todos transparentemente (nativo)
- OpenAI adapter: thinking → WARN + ignora, tool_use → convierte a function_call schema
- Google adapter: top_k/top_p soportados, thinking → WARN + ignora

Validación obligatoria:
```
temperature: [0, 1]
top_k: integer ≥ 1 (o 0 para disabled)
top_p: [0, 1]
thinking: boolean o {type: string, budget_tokens: integer}
tool_use: array of {name: string, description: string, input_schema: object}
tool_choice: "auto" | "required" | "disabled" | {type: "tool", name: string}
```

Thinking blocks handling:
- Si modelo soporta thinking, Gateway expone thinking blocks en respuesta
- Si no soporta, desactiva silenciosamente con WARN en log

Tool result handling:
- Cliente puede enviar `tool_result` en messages (Anthropic nativo)
- Debe preservarse en format Anthropic cuando enruta a Anthropic
- Si enruta a OpenAI, necesita conversión a ese schema

Test: request con thinking + tools, con fallback a OpenAI, con thinking deshabilitado en fallback
