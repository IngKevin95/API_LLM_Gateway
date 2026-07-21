---
id: HU-021a
titulo: Adapter Anthropic — chat, traducción de roles y tool calling
epica: EP-002
prioridad: Must
complejidad: M
estado: lista
---

# Adapter Anthropic — chat, traducción de roles y tool calling

Como **desarrollador de la plataforma**, quiero **un adapter que traduzca el formato interno a la Messages API de Anthropic (roles, system, tools) con manejo de errores**, para **consumir la familia Claude en peticiones de chat**.

Contexto: la API de Anthropic difiere de OpenAI (headers propios, `system` separado, `max_tokens` obligatorio). El streaming se cubre en HU-021b.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — traducción de roles | Dado que un payload entrante con formato OpenAI (system prompt en array de messages) | Cuando el router selecciona Claude | Entonces el adapter extrae el `system` y mapea los mensajes al formato Messages API de Anthropic |
| 2 | Happy — tool calling | Dado que un payload compatible con OpenAI JSON Schema de tools | Cuando se envía al adapter Anthropic | Entonces transforma correctamente la estructura a `tool_use` de Anthropic Messages API |
| 3 | Error — parámetros no soportados | Dado que el cliente envía un parámetro de OpenAI no soportado por Anthropic (ej. `seed`) | Cuando el adapter lo procesa | Entonces ignora el parámetro de forma segura, lo advierte en log y permite la ejecución |
| 4 | Edge — max_tokens ausente | Dado que el request (formato OpenAI) omite el campo opcional `max_tokens` | Cuando se adapta para Anthropic (donde es obligatorio) | Entonces el adapter inyecta un valor por defecto seguro (ej. 4096) |
| 5 | Error — error de red | Dado que Anthropic devuelve un error 5xx/429 | Cuando el adapter intercepta la respuesta | Entonces se traduce al formato estándar de error del Gateway para activar el failover |

## Checklist INVEST

- [x] Independent — no depende del adapter de OpenAI
- [x] Negotiable — soporte inicial para texto; visión negociable para luego
- [x] Valuable — alternativa a OpenAI para razonamiento profundo (Claude), diferenciador clave
- [x] Estimable — API documentada
- [x] Small — solo chat + tool calling + errores
- [x] Testable — mocks con respuestas reales de la Messages API

## Notas técnicas

Cuidado con la conversión de roles y mensajes alternados (User/Assistant) que Anthropic exige de forma estricta.
