---
id: HU-009
titulo: Autorizar por scope/RBAC y tenant
epica: EP-004A
prioridad: Should
complejidad: M
estado: lista
---

# Autorizar por scope/RBAC y tenant

Como **operador de seguridad**, quiero **que cada credencial tenga scopes que limiten qué capacidades, modelos y tenant puede usar**, para **que un consumidor solo acceda a lo que le corresponde**.

Contexto: AuthZ sobre la AuthN de HU-008. Aislamiento multi-tenant. Actividad 1.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — scope permitido | Dado que una credencial con scope `capability:coding` y tenant T1 | Cuando llega una petición `coding` del tenant T1 | Entonces la Gateway autoriza y procesa |
| 2 | Error — capacidad fuera de scope | Dado que una credencial sin scope `capability:image` | Cuando llega una petición `image` con esa credencial | Entonces responde 403 sin procesar, y registra el intento |
| 3 | Error — cruce de tenant | Dado que una credencial del tenant T1 | Cuando intenta acceder a recursos/config del tenant T2 | Entonces responde 403; los datos de T2 no son visibles ni enumerables |
| 4 | Edge — modelo restringido | Dado que una credencial autorizada a `coding` pero con un modelo concreto vetado | Cuando fuerza `model` vetado | Entonces responde 403 aunque la capacidad esté permitida |
| 5 | Edge — Capacidad Vision requiere scope de confianza | Dado que la credencial NO tiene el scope `capability:vision:trusted` | Cuando llega una petición `vision` con esa credencial | Entonces responde 403 sin procesar y registra el intento (vision deshabilita el DLP del contenido, por lo que exige scope explícito) |

## Checklist INVEST

- [x] Independent — depende de HU-008 (AuthN) ya entregable
- [x] Negotiable — modelo de roles abierto
- [x] Valuable — control de acceso fino, multi-tenant
- [x] Estimable — evaluación de scopes
- [x] Small — un sprint
- [x] Testable — combinaciones scope/tenant

## Notas técnicas

Negación por defecto: sin scope explícito no hay acceso. Aislamiento estricto por tenant.

> **OpenSpec change**: `ep-004a-identidad-accesos` (EP-004A)
