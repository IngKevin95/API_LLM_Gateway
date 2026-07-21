---
id: HU-021b
titulo: Adapter Anthropic — streaming (transformación de eventos)
epica: EP-002
prioridad: Must
complejidad: S
estado: lista
---

# Adapter Anthropic — streaming (transformación de eventos)

Como **desarrollador de la plataforma**, quiero **que el adapter de Anthropic traduzca sus eventos de streaming nativos a chunks SSE compatibles con OpenAI**, para **que el consumidor reciba tokens en vivo con un formato uniforme independientemente del proveedor**.

Contexto: depende de HU-021a (chat base). Aísla la transformación de eventos de streaming.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — streaming | Dado que una petición con `stream: true` hacia Anthropic | Cuando el adapter llama a la API | Entonces transforma los eventos nativos (`message_start`, `content_block_delta`) a chunks compatibles SSE de OpenAI |
| 2 | Edge — failover pre-primer-token | Dado que Anthropic falla antes del primer `content_block_delta` en modo stream | Cuando el adapter detecta el fallo | Entonces reporta falla estandarizada y permite el failover transparente |
| 3 | Edge — corte mid-stream | Dado que Anthropic deja de emitir eventos tras iniciar el stream | Cuando se excede el Stream Idle Timeout | Entonces el adapter cierra el socket y penaliza el score (no hay failover mid-stream) |

## Checklist INVEST

- [x] Independent — depende de HU-021a entregable
- [x] Negotiable — manejo de corte abierto
- [x] Valuable — tokens en vivo con formato uniforme
- [x] Estimable — capa de transformación de eventos acotada
- [x] Small — solo streaming
- [x] Testable — se simula el stream de eventos de Anthropic

## Notas técnicas

Mapear el ciclo `message_start` → `content_block_delta` → `message_stop` al esquema de chunks OpenAI.
