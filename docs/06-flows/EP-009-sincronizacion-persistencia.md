# EP-009: Flujo de Sincronización & Persistencia

Asegurar que eventos de audit no se pierdan ante crash, con graceful shutdown y sincronización asincrónica a BD.

```mermaid
graph TD
    A["📝 Audit Event<br/>model_choice, tokens, cost, latency"] -->|write| B["Write-Ahead Log<br/>(HU-039)<br/>Local crash recovery"]
    B -->|fsync() ≤ 1ms| C["Local WAL<br/>append-only<br/>30MB rotation"]
    C -->|next event| A
    B -->|batch 1000 eventos| D["Sync Worker<br/>(HU-038)<br/>async channels"]
    D -->|read WAL<br/>backpressure| E["KMS Envelope<br/>(HU-020)<br/>DEK wrapping"]
    E -->|encrypt DEK| F["DB Async<br/>PostgreSQL / RDS<br/>audit_events table"]
    F -->|insert batch| G["☑️ Events<br/>persisted remotely"]
    G -->|mark WAL<br/>checkpoint| C
    
    H["🛑 SIGTERM<br/>graceful shutdown<br/>(HU-040)"] -->|signal| I["Drain In-Flight<br/>accept=false,<br/>wait active requests"]
    I -->|wait N sec| J["Flush WAL<br/>fsync all pending"]
    J -->|sync Sync Worker| K["Final DB batch<br/>wait confirm"]
    K -->|exit 0| L["✅ Clean shutdown<br/>zero event loss"]
    
    M["💥 Crash / OOM<br/>unexpected"] -->|restart| N["Boot Sequence<br/>(HU-040)"]
    N -->|replay WAL| O["Recover pending<br/>events from disk"]
    O -->|continue sync| D
```

## Historias Críticas

| Historia | Fase | Componente |
|----------|------|-----------|
| HU-038 | 1 | Sync Worker (async persistence) |
| HU-039 | 1 | Write-Ahead Log (crash recovery) |
| HU-040 | 1 | Graceful Shutdown (SIGTERM handler) |
| HU-041 | 2 | Cache Invalidator (webhook/polling) |
| HU-020 | 1 | Envelope Encryption (DEK wrapping) |

## Write-Ahead Log (HU-039)

```
File: ~/.llm-gateway/audit.wal (local)
Format: JSON Lines (1 evento per line)
Rotation: 30MB (nombre: audit.wal.1, audit.wal.2, ...)
Retention: 7 días (manual cleanup)
Durability: fsync() after each write ≤ 1ms p95

Example:
{"ts":"2026-07-21T02:45:30Z","user_id":"key_123","model":"gpt-4","tokens_in":500,"tokens_out":200,"cost":0.03,"latency_ms":150,"model_choice_reason":"score=0.94"}
{"ts":"2026-07-21T02:45:45Z",...}
```

## Sync Worker (HU-038)

```
Batch size: 1000 eventos (configurable)
Polling interval: 5s (si < 1000)
Backpressure: Dejar de aceptar requests si WAL > 100MB

Flow:
1. Read batch from WAL (FIFO)
2. Encrypt DEK con KMS (envelope)
3. Insert to audit_events table (PostgreSQL)
4. Mark WAL checkpoint (delete sent batch)
5. Repeat
```

## Graceful Shutdown (HU-040)

```
On SIGTERM (kubectl delete, docker stop, etc):
1. Set accept = false (rechazar nuevos requests)
2. Drain in-flight (esperar activos < timeout)
3. Flush WAL (fsync all pending)
4. Sync Worker: send final batch + wait confirmation
5. Exit with code 0

Timeout: 30s (configurable)
Si timeout → exit 1 (syslog: eventos posiblemente no sincronizados)
```

## Recuperación ante Crash (Boot Sequence)

```
On startup:
1. Detectar si último exit fue limpio (marker file)
2. Si no limpio: replay WAL desde último checkpoint
3. Cargar eventos pendientes a memoria
4. Iniciar Sync Worker
5. Continuar procesamiento normal
```

## SLA Asociado
- **WAL durability**: fsync ≤ 1ms p95 (no blocking)
- **Sync latency**: eventos en BD < 5s (batch cada 5s o 1000 eventos)
- **RTO**: < 1h (reconstruir desde WAL + BD)
- **RPO**: < 15min (async batch, máximo pérdida = eventos en WAL no sincronizados)
- **Graceful shutdown**: 30s para drenar + flush
- **Zero event loss**: WAL + graceful shutdown garantizan

## Fase 2+ Mejoras
- Cache Invalidator (HU-041): webhook/polling para invalidar cache de cuota tras crash
- Replicación WAL a S3 (para DR cross-region)
- Particionamiento de audit_events por fecha (retention automático)
