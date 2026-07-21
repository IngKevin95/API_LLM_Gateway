---
id: HU-010
titulo: Guardar traza inmutable (Auditoría)
epica: EP-004B
prioridad: Must
complejidad: M
estado: lista
---

# Guardar traza inmutable (Auditoría)

Como **oficial de auditoría**, quiero **registrar cada petición en una tabla inmutable con todos los datos de contexto (tenant, agente, proveedor, modelo, tokens, costo, latencia, estado HTTP, timestamp)**, para **garantizar trazabilidad completa y detección de anomalías sin poder borrar evidencia**.

Contexto: Cada petición genera un evento de auditoría asincronista. El contenido de prompts/responses se redacta antes de persistencia. La tabla es particionada por fecha (mes/semana) y tiene TTL de 30 días para purga automática.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — Auditoría registrada | Dado que se completa una petición exitosa al enrutador | Cuando la petición termina (antes de retornar al cliente) | Entonces se genera un evento inmediato hacia Sync Worker con: request_id, tenant_id, agent_id, capability, provider, model, tokens_prompt, tokens_completion, cost, latency_ms, status_code, timestamp (UTC) |
| 2 | Error — Fallo del proveedor auditado | Dado que el proveedor retorna 429/500 | Cuando el failover ejecuta | Entonces cada reintento genera un evento de auditoría independiente con el provider/model/status_code del intento fallido |
| 3 | Edge — Redacción antes de persistencia | Dado que el prompt contiene PII o secretos | Cuando se genera el evento de auditoría | Entonces prompts y responses están completamente auditados (no se guardan en AuditLog) y el campo `redacted: true` marca que fueron procesados por DLP |
| 4 | Particionamiento — Retención de 30 días | Dado que hace 31 días se registró una auditoría | Cuando corre el job nocturno de purga | Entonces los registros de >30 días se eliminan (drop table partition si es posible) |
| 5 | Integridad — Sin posibilidad de borrado manual | Dado que un actor con acceso a BD intenta `DELETE FROM AuditLog` | Cuando ejecuta la consulta | Entonces la tabla es append-only con trigger que rechaza UPDATEs/DELETEs (violación de integridad LOG) |

## Checklist INVEST

- [x] Independent — No depende más que de HU-001 (Registry para tenant_id) y Sync Worker (asincronía)
- [x] Negotiable — Campos de AuditLog y TTL configurables
- [x] Valuable — Cumple Obj.4 (Auditoría y detección de anomalías) + compliance regulatorio
- [x] Estimable — Schema relacional simple, triggers DB estándar, particionamiento nativo PostgreSQL
- [x] Small — Limitada a inserción asincronista + política de retención
- [x] Testable — Mock de Sync Worker; verificar que eventos se insertan en AuditLog; simular purga de 30+ días

## Notas técnicas

Schema AuditLog (PostgreSQL):
```sql
CREATE TABLE AuditLog (
  id BIGSERIAL PRIMARY KEY,
  request_id UUID NOT NULL,
  tenant_id VARCHAR(255) NOT NULL,
  agent_id VARCHAR(255),
  capability VARCHAR(50),
  provider VARCHAR(50),
  model VARCHAR(255),
  tokens_prompt INT,
  tokens_completion INT,
  cost DECIMAL(10, 6),
  latency_ms INT,
  status_code INT,
  redacted BOOLEAN DEFAULT false,
  timestamp TIMESTAMP DEFAULT NOW()
) PARTITION BY RANGE (DATE_TRUNC('month', timestamp));
```

Sync Worker encola eventos en canal; bulk-flushea cada 1s o 1000 eventos. TTL via `pg_partman` o cron job DROP TABLE.
