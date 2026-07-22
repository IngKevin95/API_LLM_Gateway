## Context

EP-004B (auditoría) necesitará persistir eventos inmutables sin bloquear la ruta crítica ni perderlos ante crash. EP-009 provee esa infraestructura. ADR-001: Go idiomático, stdlib. PostgreSQL y KMS se abstraen tras interfaces (`Store`, `Encryptor`) con mocks en test; el WAL es real (archivos temporales en test). NFR de throughput (43M/día) y overhead OS (fsync/ext4) son diferibles a load-test/despliegue.

## Goals / Non-Goals

**Goals:**
- WAL append-only durable con rotación por tamaño y lectura de recuperación.
- Sync Worker asincronista (batching por tamaño/tiempo) que cifra con KMS Envelope y no bloquea el handler; backpressure controlado.
- Graceful Shutdown que drena requests y garantiza el flush de buffers/WAL; boot recovery del WAL residual.
- Cache Invalidator en modo Fase-1 (no-op tras flag + poll/retry-on-miss).

**Non-Goals:**
- Driver de PostgreSQL y SDK de KMS reales (detrás de interfaces; se cablean en EP-005/despliegue).
- El worker completo de polling/webhook del Cache Invalidator (Fase 2).
- Redacción de PII / rotación de secretos (EP-004B).

## Decisions

- **Tipos de dominio** (`internal/audit`): `Event` (metadata inmutable), interfaz `Store { Write(ctx, batch []EncryptedEvent) error }`, interfaz `Encryptor { Seal(Event) (EncryptedEvent, error) }` (KMS Envelope: DEK local, KEK en KMS). Mocks en test.
- **WAL** (`internal/wal`): archivo append-only; cada registro serializado (longitud-prefijada + JSON) para lectura robusta. `Append(Event)` escribe sin `fsync` forzado en el camino crítico (overhead <1ms objetivo; durabilidad vía journal del FS). Rotación: al superar `maxBytes`, renombra a `wal-<timestamp>-NNN.log` y abre uno nuevo. `Recover()` lee todos los registros (activo + archivados) para replay. Nombre de paquete `wal`.
- **Sync Worker** (`internal/syncworker`; no `sync`, choca con stdlib): consume un channel de `Event`; batchea por 1000 eventos o 1s (lo que ocurra primero); por cada batch: `Encryptor.Seal` cada evento → `Store.Write`. Antes del flush escribe al WAL (durabilidad); tras flush exitoso trunca/marca el WAL. Backpressure: `Enqueue` sobre channel lleno hace retry con jitter o dropea eventos de baja prioridad (nunca bloquea el handler). Seguro concurrente (test -race).
- **Graceful Shutdown** (`internal/shutdown`): `Shutdown(ctx)` — deja de aceptar, espera in-flight con timeout (<30s; DB timeout < shutdown timeout para evitar deadlock), ordena `Flush()` del Sync Worker (a DB, o WAL si DB no responde), y cierra. `Recover()` en boot: lee WAL residual, lo replica al Store, hidrata cachés, antes de aceptar tráfico. Señal real (`os/signal` SIGTERM) se conecta en `cmd/gateway`.
- **Cache Invalidator** (`internal/cacheinval`): en Fase 1, `Enabled=false` → `Invalidate` es no-op; el patrón vigente es poll/retry ante miss (fail-fast + re-hidratación en la próxima solicitud). El worker de polling/webhook (AC1/AC3/AC5) se difiere a Fase 2.

## Risks / Trade-offs

- **Durabilidad vs latencia**: no forzar `fsync` por evento (overhead <1ms) implica una ventana de pérdida ante crash de OS mid-write; mitigado por el journal del FS y el WAL append-only. La garantía absoluta (fsync) es configurable en despliegue.
- **Backpressure con drop**: dropear eventos de baja prioridad ante saturación protege la ruta crítica pero pierde auditoría de baja prioridad; es una decisión explícita del PRD (retry con jitter primero).
- **Concurrencia Sync Worker / Health Monitor por conexión DB** → deadlock; mitigado con timeout DB < shutdown timeout y tests -race.
- **NFR de throughput/overhead** no se miden en unit test → se verifican como smoke de proporcionalidad y se difieren a load-test (deferral honesto documentado).

## Migration Plan

Aditivo sobre lo ya en develop. El Sync Worker y el Graceful Shutdown se conectan al boot/shutdown de `cmd/gateway` (la señal SIGTERM y el drain del servidor HTTP se integran allí/EP-005). Sin migración de datos. Rollback = revertir el PR.

## Open Questions

- Formato exacto de serialización del WAL (JSON longitud-prefijado vs binario) — se fija en SS1 según simplicidad/robustez. No bloqueante.
