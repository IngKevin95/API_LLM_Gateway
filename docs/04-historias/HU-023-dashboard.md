---
id: HU-023
titulo: Dashboard visual de métricas
epica: EP-007
prioridad: Should
complejidad: M
estado: lista
---

# API de Métricas para Dashboard

Como **operador de la plataforma**, quiero **un endpoint de API que exponga métricas agregadas**, para **alimentar un dashboard visual (SPA) independiente con datos de consumo, costos y latencias en tiempo real**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — JSON de métricas | Dado que el gateway ha procesado peticiones | Cuando un cliente consulta `GET /v1/metrics/dashboard` | Entonces la API devuelve un JSON con RPS, latencia y costo acumulado |
| 2 | Happy — filtro por proveedor | Dado que hay métricas de múltiples proveedores | Cuando un cliente incluye el query param `?provider=openai` | Entonces el JSON filtrado refleja solo datos de OpenAI |
| 3 | Edge — sin métricas (Empty state) | Dado que el gateway no ha recibido tráfico | Cuando un cliente consulta las métricas | Entonces devuelve un JSON con contadores en `0` sin errores `500` |
| 4 | Error — falla backend | Dado que base de datos (almacén de histórico) está caído | Cuando se consulta el endpoint | Entonces devuelve `500 Internal Server Error` indicando fallo de dependencia |

## Checklist INVEST

- [x] Independent — Solo requiere leer de base de datos, no interfiere con el enrutamiento.
- [x] Negotiable — La estructura del JSON se puede iterar en base a las necesidades del SPA.
- [x] Valuable — Desacopla la lógica de agregación del frontend, centralizando las métricas.
- [x] Estimable — Es un endpoint de lectura con queries de agregación SQL conocidas.
- [x] Small — Se restringe al backend, dejando la UI visual a otra historia fuera de este repo.
- [x] Testable — Verificable consultando el endpoint con curl y aserciones de JSON.

## Notas técnicas
- El Dashboard Frontend no forma parte de este ticket ni repo.
- Utilizar vistas materializadas o queries optimizadas si la latencia del endpoint excede los 500ms.
