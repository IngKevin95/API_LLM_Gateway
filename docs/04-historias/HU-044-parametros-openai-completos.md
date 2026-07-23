---
id: HU-044
titulo: Parámetros OpenAI completos (temperature, top_p, etc.)
epica: EP-010
prioridad: Must
complejidad: M
estado: draft
---

# Parámetros OpenAI completos (temperature, top_p, etc.)

Como **desarrollador usando herramientas OpenAI-compatible (OpenWebUI, CrewAI, UI-TARS)**, quiero **enviar parámetros completos (temperature, top_p, frequency_penalty, etc.)**, para **control fino de la respuesta sin limitarme a lo que el Gateway soporta hoy**.

Contexto: Actualmente Handler OpenAI solo copia `model, messages, max_tokens, stream, tools`. Actividad 3 de EP-010.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — temperature traducida | Dado que cliente envía `temperature: 0.7` | Cuando Handler procesa | Entonces la temperatura se copia a internal Request y se pasa al adapter sin error |
| 2 | Happy — parámetros múltiples | Dado que cliente envía `temperature, top_p, frequency_penalty, presence_penalty` | Cuando se procesan | Entonces todos se copian a internal Request y se validan (rango 0-2 para frequency_penalty, etc.) |
| 3 | Happy — seed para reproducibilidad | Dado que cliente envía `seed: 12345` | Cuando se procesa | Entonces se pasa al adapter; adapters que lo soportan (OpenAI) lo traducen, adapters que no (Anthropic) lo ignoran con WARN en log |
| 4 | Happy — tool_choice específica | Dado que cliente envía `tools: [...]` y `tool_choice: "auto"` | Cuando se procesa | Entonces tool_choice se preserva en internal format y se traduce al schema del proveedor destino |
| 5 | Error — parámetro out of range | Dado que cliente envía `temperature: 3.5` (fuera de rango 0-2) | Cuando se valida | Entonces responde 400 Bad Request con "temperature out of range: 3.5" |
| 6 | Edge — parámetro desconocido recibido | Dado que cliente envía `unknown_param: true` | Cuando se procesa | Entonces se ignora sin error, se advierte en log (WARN), y la petición continúa |

## Checklist INVEST

- [x] Independent — se apoya en adapters (EP-002), handlers (EP-005); sin bloqueos
- [x] Negotiable — alcance: extension de tipos + mapeo a adapters
- [x] Valuable — fin del gate de "parámetros soportados"
- [x] Estimable — ampliar tipos + validación (~6 horas)
- [x] Small — un sprint
- [x] Testable — request con todos los parámetros, validación de ranges

## Notas técnicas

Parámetros a soportar (OpenAI standard):
- `temperature` (0-2, default 1.0) → adapter
- `top_p` (0-1, default 1.0) → adapter
- `frequency_penalty` (0-2, default 0) → adapter
- `presence_penalty` (0-2, default 0) → adapter
- `seed` (integer) → adapter si soporta (OpenAI sí, Anthropic no)
- `response_format` (object con type, opcional) → adapter si soporta
- `tool_choice` (string o object: auto, required, none, función específica) → adapter si soporta tools
- `top_k` (entero, para algunos modelos) → adapter
- `max_completion_tokens` (alias de max_tokens en algunos) → normalizar a max_tokens

Mapeo adapter:
- OpenAI adapter: pasa todos transparentemente
- Anthropic adapter: top_p soportado, temperature soportado, seed → WARN + ignora
- Google adapter: temperature soportado, top_p soportado, others → WARN + ignora

Validación obligatoria:
```
temperature: [0, 2]
top_p: [0, 1]
frequency_penalty: [0, 2]
presence_penalty: [0, 2]
seed: integer
```

Test: request con extremos (temperature: 0, 2, 0.5), con parámetros contradictorios (top_p + top_k simultáneamente en ciertos adapters)
