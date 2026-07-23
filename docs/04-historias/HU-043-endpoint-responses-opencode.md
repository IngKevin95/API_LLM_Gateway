---
id: HU-043
titulo: Endpoint /responses para OpenCode (OpenAI Responses API)
epica: EP-010
prioridad: Must
complejidad: M
estado: draft
---

# Endpoint /responses para OpenCode (OpenAI Responses API)

Como **desarrollador usando OpenCode**, quiero **apuntar el cliente a un endpoint `/responses` compatible con OpenAI Responses API**, para **acceder al Gateway sin depender de OpenAI directamente**.

Contexto: OpenCode usa Responses API (extensión de OpenAI con parámetros de reasoning). Actividad 2 de EP-010.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — request básica Responses API | Dado que un cliente OpenCode envía petición a `/responses` con `model`, `input`, `stream: false` | Cuando el Gateway recibe la petición | Entonces normaliza a formato interno (adapter.Request), enruta por capability implícita (ej: reasoning), y retorna respuesta en formato Responses compatible |
| 2 | Happy — streaming con Responses | Dado que cliente envía `/responses` con `stream: true` | Cuando se abre el stream | Entonces el Gateway emite eventos SSE en formato Responses (no OpenAI chunks) |
| 3 | Happy — reasoning_effort traducido | Dado que cliente envía `reasoning_effort: "medium"` | Cuando se procesa | Entonces el parámetro se pasa al adapter (ej: Anthropic) traducido, sin error "unsupported parameter" |
| 4 | Error — input inválido | Dado que cliente envía malformado `input` (no array de message objects) | Cuando se valida | Entonces responde 400 Bad Request con detalle claro del error |
| 5 | Edge — fallback si reasoning falla | Dado que reasoning model (ej: o1) está colgado | Cuando se llama `/responses` | Entonces failover a siguiente de cadena (ej: claude-opus-4 con reasoning), sin fallar la cadena completa |

## Checklist INVEST

- [x] Independent — usa handlers existentes (EP-005), routing (EP-001), adapters (EP-002); entregable sin bloqueos
- [x] Negotiable — alcance: nuevo handler + conversión Responses ↔ internal format
- [x] Valuable — desbloquea OpenCode como cliente oficial
- [x] Estimable — handler + conversión (~8 horas)
- [x] Small — un sprint
- [x] Testable — requests Responses API de referencia

## Notas técnicas

Request Responses API típica:
```json
POST /responses
{
  "model": "gpt-5",
  "input": [
    {
      "type": "message",
      "role": "user",
      "content": [{"type": "input_text", "text": "Explain quantum computing"}]
    }
  ],
  "instructions": "You are...",
  "stream": false,
  "max_output_tokens": 2048,
  "reasoning_effort": "medium"
}
```

Conversión esperada:
- `input` → `messages` (parseando structure)
- `instructions` → `system`
- `reasoning_effort` → parámetro del adapter (Anthropic lo entiende, OpenAI no)
- `max_output_tokens` → `max_tokens`

Response esperada (non-stream):
```json
{
  "output": {
    "type": "message",
    "role": "assistant",
    "content": [{"type": "output_text", "text": "..."}]
  },
  "usage": {...}
}
```

Validación: todas las referencias a "Responses API" en código usan formato estandarizado (parser reusable)
