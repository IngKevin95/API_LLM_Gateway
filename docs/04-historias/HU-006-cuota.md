---
id: HU-006
titulo: Contabilizar y limitar consumo por cuota
epica: EP-003
prioridad: Should
complejidad: M
estado: lista
---

# Contabilizar y limitar consumo por cuota

Como **operador de la plataforma**, quiero **que la Gateway lleve la cuenta de requests y tokens por proveedor y clave y corte al agotar la cuota configurada**, para **aprovechar cuotas gratuitas sin excederlas y respetar los ToS**.

Contexto: Quota Manager, gobernanza de consumo acumulado por ventana temporal (minuto/día/mes). Actividad 5.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — dentro de cuota | Dado que un proveedor con cuota diaria de 500 requests y 100 consumidos | Cuando llega una petición | Entonces se atiende, se incrementa el contador y queda cuota restante consultable |
| 2 | Error — cuota agotada | Dado que un proveedor cuya cuota diaria llegó al límite | Cuando llega una petición dirigida a ese proveedor | Entonces la Gateway lo excluye de la selección (o hace failover) y no lo consume; el evento queda registrado |
| 3 | Edge — reinicio de ventana | Dado que un proveedor con cuota diaria agotada | Cuando cruza la medianoche de la ventana | Entonces el contador se reinicia y el proveedor vuelve a estar disponible |
| 4 | Edge — límite por tokens | Dado que un proveedor con límite de 1M tokens/día casi alcanzado | Cuando llega una petición cuyo tamaño estimado excedería el límite | Entonces la Gateway la rechaza o enruta a otro proveedor antes de consumir |
| 5 | Edge — race conditions y overshoot post-generación | Dado que un proveedor con 1 token restante de cuota | Cuando llegan 50 peticiones simultáneas exigiendo tokens | Entonces el Gateway valida la cuota de forma atómica en RAM permitiendo 1 y rechazando 49; y si una respuesta ya en curso excede el saldo al finalizar el stream, actualiza la cuota a cero/negativo bloqueando peticiones futuras |

## Checklist INVEST

- [x] Independent — usa Registry; entregable tras Registry
- [x] Negotiable — granularidad de ventana configurable
- [x] Valuable — evita exceder cuotas y viola ToS
- [x] Estimable — contadores + ventanas
- [x] Small — un sprint
- [x] Testable — cuotas simuladas

## Notas técnicas

Contadores por (proveedor, clave, ventana). La validación y el descuento de cuota ocurren de manera síncrona y atómica en la Local RAM Cache, mientras que la persistencia en base de datos se realiza asíncronamente mediante background workers para sobrevivir reinicios sin penalizar la latencia del proxy (<50ms).
