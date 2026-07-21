---
id: HU-005
titulo: Health checks periódicos con retiro y reactivación de proveedor
epica: EP-002
prioridad: Should
complejidad: M
estado: lista
---

# Health checks periódicos con retiro y reactivación de proveedor

Como **operador de la plataforma**, quiero **que la Gateway pruebe periódicamente cada proveedor y retire de la rotación los caídos, reactivándolos al recuperarse**, para **que las peticiones no se enruten a proveedores que sabemos que fallan**.

Contexto: convierte el failover reactivo (HU-004) en prevención proactiva. Actividad 4 del journey.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — proveedor sano se mantiene | Dado que un proveedor que responde OK a los health checks | Cuando corre el ciclo de health check | Entonces el proveedor permanece marcado sano y en la rotación |
| 2 | Error — detección de caída y retiro | Dado que un proveedor que empieza a devolver errores en los health checks | Cuando corre el ciclo de health check | Entonces el proveedor se marca no-sano y deja de recibir peticiones nuevas hasta recuperarse |
| 3 | Happy — reactivación | Dado que un proveedor marcado no-sano que vuelve a responder OK | Cuando corre el siguiente health check | Entonces se marca sano y vuelve a la rotación automáticamente |
| 4 | Error — todos no-sanos | Dado que todos los proveedores de una capacidad fallan los checks | Cuando llega una petición de esa capacidad | Entonces la Gateway responde 503 y el estado refleja la capacidad como no disponible |
| 5 | Edge — proveedor intermitente | Dado que un proveedor que alterna OK/fallo entre checks | Cuando corren varios ciclos | Entonces se aplica histéresis (N fallos para retirar, M éxitos para reactivar) para evitar oscilación |

## Checklist INVEST

- [x] Independent — usa el Registry; entregable tras failover básico
- [x] Negotiable — intervalo y umbrales configurables
- [x] Valuable — mejora disponibilidad real
- [x] Estimable — scheduler + estado por proveedor
- [x] Small — un sprint
- [x] Testable — checks simulados

## Notas técnicas

Intervalo configurable (p.ej. 60 s). Histéresis para intermitencia. Estado consultable por el Router.
