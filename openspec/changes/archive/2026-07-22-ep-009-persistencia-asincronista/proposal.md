## Why

La auditoría (EP-004B) y la resiliencia exigen que cada evento se persista de forma inmutable **sin bloquear la ruta crítica** y **sin perderse ante una caída**. Hoy no hay persistencia: si la ruta crítica escribiera a DB síncronamente, colapsaría el throughput; sin WAL, cualquier crash pierde auditoría; sin graceful shutdown, se pierden buffers. EP-009 es la infraestructura asincronista de fondo (Sync Worker + WAL + KMS Envelope + crash recovery) — última épica `foundational`, prerequisito silencioso de EP-004B. Objetivos: Obj. 4 (auditoría/seguridad), Obj. 2 (resiliencia; RTO<1h, RPO<15m).

## What Changes

- **Write-Ahead Log (WAL)**: log append-only en disco que registra eventos antes del flush a DB; rotación por tamaño; lectura para recuperación ante crash.
- **Sync Worker**: writer asincronista que consume eventos de un channel, los batchea (cada 1s o 1000 eventos) y los persiste vía un `Store`, cifrando con **KMS Envelope** (DEK local, KEK en KMS) antes de escribir; backpressure con retry/drop de baja prioridad.
- **Graceful Shutdown**: ante SIGTERM, drena las requests en vuelo (timeout <30s), flushea el buffer del Sync Worker a DB (o WAL si la DB no responde), y hidrata cachés en el boot desde el WAL residual, sin deadlocks.
- **Cache Invalidator (Fase 2)**: en MVP es no-op tras feature flag; solo aplica poll/retry ante miss. El worker completo de polling/webhook se difiere a post-MVP.
- PostgreSQL y KMS se abstraen tras interfaces (`Store`, `Encryptor`); en tests se usan mocks + WAL en archivos temporales. El cableado a PostgreSQL/KMS reales y al servidor HTTP (drain) se materializa en EP-005/despliegue.

## Capabilities

### New Capabilities
- `write-ahead-log`: log append-only durable con rotación por tamaño y lectura de recuperación; overhead mínimo en el camino crítico.
- `sync-worker`: persistencia asincronista por batching vía channel, con KMS Envelope (DEK/KEK) y manejo de backpressure, sin bloquear el handler.
- `graceful-shutdown`: drenado de requests en vuelo + flush garantizado de buffers/WAL ante SIGTERM, y secuencia de boot con recuperación del WAL residual, sin deadlock.
- `cache-invalidator`: en Fase 1 es no-op tras feature flag con fallback a poll/retry ante miss; el polling/webhook completo es Fase 2.

### Modified Capabilities
<!-- Ninguna spec previa cambia sus requisitos. -->

## Impact

- Código nuevo (Go): `src/internal/wal`, `src/internal/sync` (Sync Worker), `src/internal/shutdown`, `src/internal/cacheinval`. Interfaces `Store` (persistencia) y `Encryptor` (KMS Envelope) con implementaciones mock/no-op en tests.
- Integra en el boot/shutdown del `cmd/gateway` (secuencia de arranque/cierre); la conexión real a PostgreSQL/KMS y el drain del servidor HTTP se cablean en EP-005/despliegue.
- Dependencias: stdlib (`os`, `bufio`, `encoding/*`, `os/signal`); sin driver de DB ni SDK de KMS en esta épica (detrás de interfaces).
- Datos sensibles: los eventos de auditoría se cifran con KMS Envelope antes de tocar disco/DB; las llaves nunca se persisten en claro. Sin breaking changes.

## Trazabilidad

- **Épica**: EP-009 · Sincronización Asincronista y Persistencia (`layer: foundational`) — objetivos del PRD: Obj. 4 (auditoría inmutable, secretos), Obj. 2 (resiliencia; crash recovery).
- **Historias cubiertas** (por sub-slice):
  - SS1 — `write-ahead-log`: HU-039 (WAL append-only, rotación, lectura de recovery)
  - SS2 — `sync-worker`: HU-038 (Sync Worker batching + KMS Envelope + backpressure)
  - SS3 — `graceful-shutdown`: HU-040 (drain + flush garantizado + boot recovery)
  - SS4 — `cache-invalidator`: HU-041 (Fase 2: no-op tras flag + poll/retry-on-miss)
