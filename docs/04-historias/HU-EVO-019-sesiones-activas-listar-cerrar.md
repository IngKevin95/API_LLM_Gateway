---
id: HU-EVO-019
titulo: Sesiones activas - listar y cerrar
epica: EP-EVO-004
prioridad: Should
complejidad: M
estado: lista
---

# Sesiones activas: listar y cerrar

Como **usuario del Gateway**, quiero **ver desde dónde tengo sesión iniciada y poder cerrarlas**, para **revocar acceso si sospecho que un dispositivo fue comprometido**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — listar sesiones | Dado que tengo 2 sesiones activas (dashboard abierto en 2 navegadores distintos) | Cuando hago `GET /sessions` | Entonces veo ambas con dispositivo/user-agent resumido, IP/ubicación aproximada y última actividad |
| 2 | Happy — cerrar una sesión | Dado que tengo una sesión activa que ya no reconozco | Cuando hago `DELETE /sessions/:id` | Entonces esa sesión específica queda invalidada, la próxima request con ese token devuelve 401 |
| 3 | Happy — cerrar todas menos la actual | Dado que tengo 3 sesiones activas incluyendo la que estoy usando ahora | Cuando hago `DELETE /sessions` con el flag `except_current=true` | Entonces todas se invalidan excepto la que estoy usando en ese momento |
| 4 | Error — cerrar sesión de otro usuario | Dado que soy un usuario no-admin | Cuando intento cerrar el ID de sesión de otra persona | Entonces recibo 403 Forbidden |
| 5 | Edge — sesión ya expirada | Dado que una sesión ya expiró por inactividad | Cuando aparece en `GET /sessions` | Entonces se marca como `expired` en vez de listarse como activa, y no aparece en el conteo de "sesiones activas" |

## Checklist INVEST

- [x] Independent — depende de que exista un mecanismo de sesión/token asociado a `users` (HU-EVO-017), módulo propio
- [x] Negotiable — TTL de sesión y granularidad de "ubicación aproximada" (solo país/ciudad por IP, sin GPS) son detalle de implementación
- [x] Valuable — control de seguridad esperado en cualquier producto empresarial, reduce superficie de ataque ante robo de credenciales
- [x] Estimable — tabla `sessions` + endpoints + resolución de user-agent/IP a texto legible
- [x] Small/Medium — 2 días
- [x] Testable — test de integración: crear sesión, listarla, cerrarla, verificar invalidación

## Notas técnicas

No requiere geolocalización precisa: usar un lookup de IP a país/ciudad de bajo costo (o omitir el campo si no hay proveedor de geo-IP disponible, mostrando solo la IP — no inventar ubicación falsa).

---

## Relación con existentes

- Depende de: HU-EVO-017 (tabla `users`)
- Requisito para: HU-EVO-022 (UI Profile & Security, sección Security/sesiones)
