---
id: HU-017
titulo: Exponer métricas por modelo y proveedor
epica: EP-007
prioridad: Could
complejidad: M
estado: lista
---

# Exponer métricas por modelo y proveedor

Como **operador de la plataforma**, quiero **consultar métricas por modelo y proveedor (latencia, success_rate, tokens, quota_remaining, availability, costo)**, para **entender el estado real de la plataforma y comparar proveedores**.

Contexto: observabilidad que alimenta decisiones y el KPI de éxito. Actividad 5, release v2.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — métricas actuales | Dado que tráfico procesado por varios modelos | Cuando el operador consulta el endpoint de métricas | Entonces devuelve por modelo/proveedor: latencia (avg/p95), success_rate, tokens, quota_remaining, availability y costo |
| 2 | Happy — ranking | Dado que métricas acumuladas de varios proveedores | Cuando el operador pide el ranking | Entonces devuelve proveedores/modelos ordenados por un criterio seleccionable (calidad, costo, latencia) |
| 3 | Error — sin datos aún | Dado que la Gateway recién iniciada sin tráfico | Cuando se consultan métricas | Entonces devuelve estructura vacía o "sin datos" explícito, no un error |
| 4 | Edge — acceso restringido | Dado que un consumidor sin permiso de operador | Cuando consulta el endpoint de métricas | Entonces responde 403; las métricas no se exponen a cualquier credencial |

## Checklist INVEST

- [x] Independent — usa telemetría de HU-007/010; entregable en v2
- [x] Negotiable — formato de exposición abierto
- [x] Valuable — visibilidad operativa
- [x] Estimable — agregación + endpoint
- [x] Small — un sprint
- [x] Testable — métricas simuladas

## Notas técnicas

Endpoint protegido por scope de operador. Compatible con scrapers (formato de métricas estándar) opcional.
