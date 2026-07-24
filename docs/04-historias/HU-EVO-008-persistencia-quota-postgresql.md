---
id: HU-EVO-008
titulo: Persistencia asíncrona en PostgreSQL de learned quotas
epica: EP-EVO-002
prioridad: Should
complejidad: M
estado: draft
---

# Persistencia asíncrona en PostgreSQL de learned quotas

Como **operador del Gateway**, quiero **que Quota Manager persista los learned quotas en PostgreSQL asincronamente (sin bloquear path crítico)**, para **sobrevivir reinicios y auditar histórico de aprendizaje**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — persistencia async | Dado que OpenAI devuelve `Remaining: 9950` | Cuando LearnFromHeaders() actualiza RAM | Entonces enqueue un job async que persista en `provider_quotas_learned` table y respuesta vuelve en <5ms (no bloqueada) |
| 2 | Happy — reinicio restaura learned | Dado que antes del crash, learned remaining = 500K | Cuando Gateway reinicia y Quota Manager carga desde DB | Entonces restaura 500K (no vuelve a quota_hint del YAML) |
| 3 | Error — DB down no bloquea requests | Dado que PostgreSQL está caído | Cuando request llega y LearnFromHeaders() intenta persistir | Entonces falla gracefully (log warning, sigue usando RAM) sin abortar el request |
| 4 | Edge — competencia write async | Dado que 100 requests paralelos actualizan learned quotas | Cuando async workers compiten por escribir en DB | Entonces usa UPSERT `ON CONFLICT (provider_id, model_id) DO UPDATE` para idempotencia |
| 5 | Edge — histórico auditoria | Dado que Cerebras cambió remaining 3 veces en 1 minuto | Cuando se persiste en DB | Entonces `provider_quotas_learned` tiene 3 rows con timestamp, para auditoría |

## Checklist INVEST

- [x] Independent — depende de HU-EVO-007 (learning en RAM)
- [x] Negotiable — batch size de async writes configurable
- [x] Valuable — auditoría + crash recovery
- [x] Estimable — background worker + DB schema
- [x] Small — 2 días
- [x] Testable — mock DB, simula crashes

## Notas técnicas

Schema PostgreSQL nuevo:
```sql
CREATE TABLE provider_quotas_learned (
    id SERIAL PRIMARY KEY,
    provider_id VARCHAR(255) NOT NULL,
    model_id VARCHAR(255),
    limit INT64,
    remaining INT64,
    reset_at TIMESTAMP,
    learned_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (provider_id, model_id, learned_at)
);
```

Sync Worker en `src/internal/quota/persist.go`:
```go
func (m *Manager) persistAsync(ctx context.Context, job QuotaLearnJob) {
    go func() {
        // Batch writes cada 100ms o 1000 jobs
        // UPSERT en PostgreSQL
        // Log si falla
    }()
}
```

---

## Relación con existentes

- Extiende: `src/internal/quota/manager.go`, `src/internal/syncworker/` (HU-026)
- Usa: HU-EVO-007 (learning)
- Requisito para: auditoría y crash recovery

## Estado real de implementación (actualizado tras reapertura de EP-EVO-002)

Implementado en `src/internal/quota/persister_postgres.go` (`PostgresPersister`), cableado en
`cmd/gateway/main.go` detrás de la variable de entorno `GATEWAY_QUOTA_POSTGRES_DSN` (opt-in; sin
declararla, el Manager sigue usando `NoPersister` — comportamiento por defecto sin cambios).

Diferencias respecto a las notas técnicas originales de esta historia (documentadas aquí para no
perder trazabilidad, no bloqueantes):

- **Tabla**: `learned_quota` (no `provider_quotas_learned`), con **upsert por
  `(provider_id, model_id)`** — 1 fila vigente por par, no 1 fila por evento de aprendizaje. AC5
  (histórico de auditoría con 3 rows) **no está implementado**: el modelo actual solo guarda el
  último valor aprendido, no el historial completo. Si se necesita auditoría histórica real, es
  una extensión futura (tabla append-only separada), no cubierta por esta corrección.
- **Worker**: un único worker consumiendo un `channel` con buffer (no batch de 100ms/1000 jobs);
  `Enqueue` no bloquea vía `select` con `default` (retorna `ErrPersistQueueFull` si el buffer está
  lleno) en vez de goroutine-por-job.
- **AC1, AC2, AC3, AC4**: implementados y probados contra PostgreSQL real (contenedor `postgres:16`
  vía `docker run` en tests con build tag `integration`, sin agregar `testcontainers-go` como
  dependencia): `src/internal/quota/persister_postgres_integration_test.go`.
- **AC5** (auditoría con histórico multi-row): no implementado — ver nota de tabla arriba.
