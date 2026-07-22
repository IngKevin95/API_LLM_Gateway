---
id: HU-040
titulo: Graceful Shutdown con flush obligatorio de WAL/buffers
epica: EP-009
prioridad: Must
complejidad: S
estado: lista
---

# Graceful Shutdown con flush obligatorio de WAL/buffers

Como **operador de producción**, quiero **que el Gateway drene y persista todos los eventos pendientes antes de salir**, para **evitar pérdida de auditoría durante rolling deployments o cambios de configuración**.

Contexto: Gateway recibe SIGTERM → flush WAL → flush Sync Worker buffers → cierra conexiones → exit.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — graceful drain | Dado que Gateway recibe SIGTERM | Cuando cierra nuevas conexiones HTTP | Entonces espera (timeout < 30s) a que se drenen las en-flight requests y se flushee el WAL |
| 2 | Error — timeout en drain | Dado que una petición tarda > 25s (timeout configurado) | Cuando graceful shutdown lo espera | Entonces cierra la conexión y registra el evento de timeout |
| 3 | Edge — WAL flush garantizado | Dado que hay 1000 eventos en buffer de Sync Worker | Cuando SIGTERM llega | Entonces Sync Worker flusheó todo a DB (o WAL si DB no responde) antes de exit |
| 4 | Recovery — boot sequence | Dado que Gateway bootea tras graceful shutdown | Cuando Recovery Worker arranca | Entonces procesa WAL residual (si quedó) e hidranta caché de cuota/auth |
| 5 | Deadlock prevention | Dado que Health Monitor y Sync Worker compiten por conexión DB | Cuando Graceful Shutdown ordena flush | Entonces no hay deadlock; timeout DB < shutdown timeout |

## Checklist INVEST

- [x] Independent — coordinación local de Sync Worker + WAL
- [x] Negotiable — timeouts configurables
- [x] Valuable — garantiza 0 pérdida durante deploys
- [x] Estimable — signal handler + drain pattern estándar
- [x] Small — manejador de SIGTERM/SIGINT
- [x] Testable — enviar SIGTERM, verificar WAL persistido

## Notas técnicas

Coordinación: main() registra signal handler → notifica channels de Sync Worker / Health Monitor / LLM Handler → cada uno drena → exit. MaxDrainWait: 30s. Recovery vive en HU-040 boot.

> **OpenSpec change**: `ep-009-persistencia-asincronista` (EP-009)
