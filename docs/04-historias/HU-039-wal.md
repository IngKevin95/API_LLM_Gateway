---
id: HU-039
titulo: Write-Ahead Log (WAL) local con recuperación ante crash
epica: EP-009
prioridad: Must
complejidad: M
estado: lista
---

# Write-Ahead Log (WAL) local con recuperación ante crash

Como **arquitecto de confiabilidad**, quiero **mantener un Write-Ahead Log local (en disco) que registre eventos antes de ser persistidos en BD**, para **garantizar 0 pérdida de auditoría ante caída o OOM del Gateway**.

Contexto: antes que Sync Worker escriba a PostgreSQL, bufferea en WAL local. En restart, WAL se procesa primero.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — WAL registra eventos | Dado que Sync Worker encola 1000 eventos | Cuando los bufferea antes de flush a DB | Entonces WAL contiene los 1000 registros serializados en disco (append-only) |
| 2 | Error — recuperación ante crash | Dado que Gateway muere con 500 eventos en WAL no flusheados | Cuando reinicia | Entonces Recovery Worker (HU-040) procesa el WAL e inserta los 500 eventos en DB antes de aceptar tráfico nuevo |
| 3 | Edge — WAL rotación | Dado que WAL alcanza 100MB | Cuando Sync Worker flusheó a DB | Entonces archiva el WAL (ej. wal-20260720-001.log) y comienza uno nuevo |
| 4 | Integridad — durabilidad | Dado que OS crashea mid-write a WAL | Cuando recupera | Entonces el Journal de PostgreSQL + WAL recovery (fsync en ext4) permite recuperar hasta el último commit conocido |
| 5 | Performance — overhead | Dado que 43M eventos/día | Cuando se escriben en WAL | Entonces latencia adicional < 1ms por evento (append-only, sin sync forzado en camino crítico) |

## Checklist INVEST

- [x] Independent — storage local, no depende de Sync Worker detalle
- [x] Negotiable — rotate size, retention policy configurables
- [x] Valuable — RTO/RPO < 1h / < 15min (per PRD)
- [x] Estimable — append-only log pattern estándar
- [x] Small — integración con HU-038/040
- [x] Testable — simular crash, verificar recuperación

## Notas técnicas

WAL vive en `/var/lib/gateway/wal/`. Serialización: JSON línea por línea. Recover Worker (HU-040) lo procesa en boot.
