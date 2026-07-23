---
id: HU-047
titulo: Middleware de normalización automática de formatos (format auto-detect)
epica: EP-010
prioridad: Should
complejidad: M
estado: draft
---

# Middleware de normalización automática de formatos (format auto-detect)

Como **arquitecto del Gateway**, quiero **un middleware que detecte y normalice automáticamente formatos de entrada (OpenAI vs. Anthropic vs. Responses)**, para **ser tolerante con pequeñas variaciones y traducir parámetros sin errores**.

Contexto: Herramientas pueden enviar formato casi-compatible pero con pequeñas diferencias. Middleware absorbe esas diferencias. Actividad 6 de EP-010.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — auto-detect OpenAI format | Dado que petición tiene `messages: [{role, content}]` y NO tiene `max_tokens` (Anthropic obligatorio) | Cuando middleware procesa | Entonces detecta formato OpenAI y normaliza: inyecta default max_tokens, omite `system` si está en messages |
| 2 | Happy — auto-detect Anthropic format | Dado que petición tiene `messages` + `max_tokens` (obligatorio) | Cuando middleware procesa | Entonces detecta Anthropic, preserva intacto |
| 3 | Happy — auto-detect Responses format | Dado que petición tiene `input: [{type: "message", role, content}]` en lugar de `messages` | Cuando middleware procesa | Entonces detecta Responses API, convierte `input` → `messages` |
| 4 | Happy — parámetro desconocido mapeado | Dado que cliente OpenWebUI envía parámetro no standard `stop_sequences` | Cuando middleware procesa | Entonces intenta mapear a equivalente conocido (`stop` en OpenAI, `stop_sequences` en Anthropic); si no existe, WARN en log e ignora |
| 5 | Error — ambigua o malformada | Dado que petición es JSON inválido o carece de campo `messages` completamente | Cuando middleware valida | Entonces pasa error al Handler (mismo comportamiento que hoy) |
| 6 | Edge — múltiples formatos combinados | Dado que cliente mezcla `system` + `messages` (Anthropic style) con `tools` + `tool_choice` (OpenAI style) | Cuando middleware normaliza | Entonces combina ambos correctamente: preserva system, preserva tools/tool_choice |

## Checklist INVEST

- [x] Independent — sitúa se entre router.go y handlers; sin bloqueos de otros componentes
- [x] Negotiable — alcance: detector + mapeador; parámetros soportados definibles
- [x] Valuable — tolerancia universal, reduce errores de integración
- [x] Estimable — detector + mapper (~8 horas)
- [x] Small — un sprint
- [x] Testable — inputs con múltiples formatos, edge cases de mezcla

## Notas técnicas

Middleware flow:
```
HTTP Request
    ↓
JSON Decode → raw JSON object
    ↓
Format Detector (heurística)
    ↓
Format Normalizer (mapper)
    ↓
Adapter.Request (internal)
    ↓
Handler (existing logic)
```

Format detector heurística:
```
if input contains "input" field && input[0] has "type": "message"
  → Responses API
else if max_tokens present && max_tokens is integer && messages present
  → Anthropic (likely)
else if messages present && no max_tokens
  → OpenAI (likely)
else
  → Unknown/Error
```

Parámetro mapper (conocidos → internos):
| Origen | Destino | Condición |
|--------|---------|-----------|
| OpenAI `stop_sequences` | internal `stop` | any |
| OpenAI `tools` | internal `tools` | preserve as-is |
| Anthropic `tools` | internal `tools` | convert type from tool_use to function |
| Responses `instructions` | internal `system` | merge with existing system |
| Responses `input` | internal `messages` | convert structure |

Logging:
- WARN: parámetro desconocido, parámetro no soportado por adapter, format inference ambigua
- DEBUG: format detected, normalization applied
- ERROR: malformed input, validation failed

Test cases:
- OpenAI basic → Anthropic adapter
- Anthropic with thinking → OpenAI adapter (thinking stripped with WARN)
- Responses API full → each adapter
- Mixed format (system + messages + tools) → internal format
- Unknown parameter → WARN + proceed
