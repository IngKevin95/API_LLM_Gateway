---
id: HU-020b
titulo: Adapter OpenAI — streaming SSE
epica: EP-002
prioridad: Must
complejidad: S
estado: lista
---

# Adapter OpenAI — streaming SSE

Como **desarrollador de la plataforma**, quiero **que el adapter de OpenAI soporte streaming Server-Sent Events**, para **emitir tokens en vivo hacia el consumidor sin bufferizar la respuesta completa**.

Contexto: depende de HU-020a (chat base). Aísla la mecánica de streaming del adapter.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — streaming | Dado que un payload de chat con `stream: true` | Cuando el router selecciona OpenAI | Entonces el adapter establece un pipe de SSE y emite los tokens transparentemente |
| 2 | Edge — failover pre-primer-token | Dado que OpenAI falla antes de emitir el primer token en modo stream | Cuando el adapter detecta el fallo | Entonces reporta falla estandarizada y permite el failover transparente (aún no se emitió nada al cliente) |
| 3 | Edge — corte mid-stream | Dado que OpenAI deja de emitir tokens tras haber comenzado el stream | Cuando se excede el Stream Idle Timeout | Entonces el adapter cierra el socket y penaliza el score (no hay failover mid-stream) |

## Checklist INVEST

- [x] Independent — depende de HU-020a entregable
- [x] Negotiable — manejo de corte de conexión abierto
- [x] Valuable — UX de tokens en vivo para el integrador
- [x] Estimable — capa de streaming acotada
- [x] Small — solo streaming
- [x] Testable — se simula un stream SSE mockeado

## Notas técnicas

Reutiliza la traducción de HU-020a; solo añade el manejo del canal SSE y del idle timeout.

> **OpenSpec change**: `ep-002-resiliencia-conectividad` (EP-002)
