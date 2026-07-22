---
id: HU-038
titulo: Implementar Sync Worker (persistencia asincronista vía channels)
epica: EP-009
prioridad: Must
complejidad: M
estado: lista
---

# Implementar Sync Worker (persistencia asincronista vía channels)

Como **arquitecto de infraestructura**, quiero **escribir eventos de auditoría a PostgreSQL de forma completamente asincronista mediante channels de Go**, para **mantener la ruta crítica (< 100ms) libre de I/O a base de datos y garantizar que ninguna caída de BD bloquee la API**.

Contexto: Handler envía eventos asincronista → Sync Worker → PostgreSQL (vía KMS Envelope). Implementa backpressure si el channel se satura.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — eventos persistidos | Dado que Handler encolapa evento de auditoría en channel | Cuando Sync Worker procesa el batch (cada 1s o 1000 eventos) | Entonces inserta el lote en AuditLog vía KMS Envelope sin bloquear Handler |
| 2 | Error — backpressure | Dado que el channel está lleno (buffer saturado) | Cuando Handler intenta enqueue | Entonces Handler implementa retry con jitter (no falla crítico) o dropa eventos de baja prioridad |
| 3 | Edge — pérdida ante caída | Dado que Sync Worker muere sin flushear | Cuando el Gateway reinicia | Entonces el WAL local tiene los eventos no persistidos para recuperación (ver HU-039) |
| 4 | Integración — KMS Envelope | Dado que Sync Worker procesa eventos | Cuando obtiene la llave maestra de KMS | Entonces cifra localmente con DEK antes de escribir a DB |
| 5 | Performance — throughput | Dado que ingresan 43M eventos/día | Cuando Sync Worker empieza | Entonces soporta batchs de 1000+ eventos/segundo sin lag |

## Checklist INVEST

- [x] Independent — middleware de persistencia, no bloquea router
- [x] Negotiable — batch size, buffer channel, retry policy configurables
- [x] Valuable — cumple Obj.4 (auditoría inmutable) sin penalizar latencia
- [x] Estimable — patrón channels estándar en Go
- [x] Small — acotada a writer asincronista + backpressure
- [x] Testable — mock de DB, canales con límites, inyección de fallas

## Notas técnicas

Usa `sync.Pool` para reutilizar buffers de eventos. Implementa graceful shutdown (flush antes de exit). Coordina con HU-039 (WAL) y HU-040 (Graceful Shutdown).

> **OpenSpec change**: `ep-009-persistencia-asincronista` (EP-009)
