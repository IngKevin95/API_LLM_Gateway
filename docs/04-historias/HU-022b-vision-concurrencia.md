---
id: HU-022b
titulo: Límite de concurrencia y enrutamiento de red para capacidad vision
epica: EP-004A
prioridad: Should
complejidad: S
estado: lista
---

# Límite de concurrencia y enrutamiento de red para capacidad vision

Como **ingeniero de seguridad**, quiero **un límite de concurrencia estricto y una política de balanceo dedicada para la capacidad `vision`**, para **evitar el bufferbloat y el agotamiento de VRAM/red que provocan los payloads densos de imagen**.

Contexto: se extrae de HU-022 (rate limiting genérico) porque la política de `vision` es específica de esa capacidad, no de la defensa perimetral HTTP general. Los Then se expresan en términos observables por el cliente.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — dentro del límite vision | Dado que un nodo con 1 petición `vision` activa (límite de 2 por nodo) | Cuando llega una segunda petición `vision` | Entonces devuelve un HTTP 200 sin emitir timeout |
| 2 | Error — concurrencia vision excedida | Dado que un nodo ya atiende 2 peticiones `vision` activas | Cuando llega una tercera petición `vision` | Entonces el Gateway responde `429 Too Many Requests` sin encolarla, evitando bufferbloat |
| 3 | Edge — enrutamiento dedicado de vision | Dado que llega una petición de capacidad `vision` al balanceador | Cuando se decide el nodo destino | Entonces se enruta por política de menor carga (no por Hash L7 de API Key), aceptando un leve drift de cuota en RAM local a favor de la estabilidad física del cluster |

## Checklist INVEST

- [x] Independent — depende de HU-022 (rate limiting) entregable
- [x] Negotiable — umbral (2/nodo) y política de balanceo configurables
- [x] Valuable — evita colapsos de VRAM/red por payloads de imagen densos
- [x] Estimable — contador atómico por nodo + regla de balanceo
- [x] Small — alcance mínimo a la capacidad vision
- [x] Testable — se fuerza la concurrencia de peticiones vision

## Notas técnicas

El detalle interno de balanceo (Least Connections vs Hash L7) vive aquí como nota, no como AC: el AC solo compromete el comportamiento observable (429 al exceder 2 concurrentes y enrutamiento por carga).
