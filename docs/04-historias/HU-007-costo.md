---
id: HU-007
titulo: Registrar costo por petición, agente y proveedor
epica: EP-003
prioridad: Should
complejidad: S
estado: lista
---

# Registrar costo por petición, agente y proveedor

Como **owner/finanzas**, quiero **que cada petición registre su costo estimado atribuido a agente y proveedor**, para **controlar el gasto y comparar proveedores**.

Contexto: base del KPI "costo por 1k peticiones". Actividad 5.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — costo atribuido | Dado que un modelo con `cost_per_token` definido y un agente identificado | Cuando se completa una petición | Entonces se registra costo = tokens × tarifa, atribuido a (agente, proveedor, modelo) y consultable |
| 2 | Error — tarifa faltante | Dado que un modelo sin `cost_per_token` en el Registry | Cuando se completa una petición con ese modelo | Entonces se registra la petición con costo marcado "desconocido" y se advierte, sin perder el registro |
| 3 | Edge — modelo gratuito | Dado que un modelo con costo 0 | Cuando se completa la petición | Entonces se registra costo 0 y cuenta en volumen pero no en gasto |
| 4 | Edge — petición con failover | Dado que una petición resuelta tras fallar 2 proveedores | Cuando se completa vía el tercero | Entonces el costo se atribuye solo al proveedor que efectivamente respondió |
| 5 | Edge — stream abortado | Dado que el Gateway está sirviendo una respuesta vía streaming | Cuando el cliente aborta la conexión TCP repentinamente | Entonces el "Stream Telemetry" contabiliza los tokens parciales ya enviados y descuenta/atribuye el costo exacto |

## Checklist INVEST

- [x] Independent — se apoya en el resultado de HU-004 (failover) para atribuir costo al proveedor que respondió; entregable con telemetría mínima de request completada
- [x] Negotiable — modelo de costeo abierto
- [x] Valuable — control de gasto directo
- [x] Estimable — cálculo simple
- [x] Small — 1-2 días
- [x] Testable — tarifas simuladas

## Notas técnicas

El costo de tokens de entrada y salida puede diferir; permitir tarifas separadas.
