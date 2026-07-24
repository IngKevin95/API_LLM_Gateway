---
id: HU-EVO-017
titulo: Store de usuarios persistente en PostgreSQL con CRUD admin
epica: EP-EVO-004
prioridad: Must
complejidad: L
estado: lista
---

# Store de usuarios persistente en PostgreSQL con CRUD admin

Como **administrador del Gateway**, quiero **dar de alta, editar y suspender usuarios de mi equipo desde una API persistente**, para **no depender de editar variables de entorno y redeployar cada vez que cambia el equipo**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — invitar usuario | Dado que soy admin autenticado | Cuando hago `POST /users` con email, rol y scopes/tenants | Entonces se crea el registro en `users` con estado `invited`, respuesta 201 con el ID del usuario |
| 2 | Happy — listar usuarios filtrado por tenant | Dado que soy admin de tenant T1 (no admin global) | Cuando hago `GET /users` | Entonces solo veo usuarios con scope/tenant T1, nunca de otros tenants |
| 3 | Happy — cambiar rol/estado | Dado que existe un usuario `operator` activo | Cuando hago `PATCH /users/:id` con `{"status":"suspended"}` | Entonces el usuario pasa a `suspended` y pierde acceso inmediato (sus API keys dejan de autenticar) |
| 4 | Error — no-admin intenta invitar | Dado que soy un usuario con rol `operator` | Cuando hago `POST /users` | Entonces recibo 403 Forbidden, no se crea ningún registro |
| 5 | Edge — email duplicado | Dado que ya existe un usuario con `email=x@y.com` | Cuando invito de nuevo ese mismo email | Entonces recibo 409 Conflict, no se duplica el registro |

## Checklist INVEST

- [x] Independent — capa de datos nueva, no depende de HU-EVO-014/015 (dashboard) para existir, aunque las alimenta
- [x] Negotiable — elección de PostgreSQL vs otro store es detalle de implementación (ya hay PostgreSQL en el proyecto para `provider_alerts`, HU-EVO-012)
- [x] Valuable — elimina el cuello de botella operativo de editar env vars para dar de alta gente
- [x] Estimable — tabla `users` + CRUD REST + middleware de autorización admin-only
- [x] Small/Medium — 2-3 días
- [x] Testable — tests de integración contra PostgreSQL real (mismo patrón que `alert.Manager`, HU-EVO-012)

## Notas técnicas

Reemplaza gradualmente `apikey.Store` (in-memory, `src/internal/auth/apikey/apikey.go`) — este AC no elimina el store in-memory todavía (lo hace HU-EVO-018 al mover las API keys), solo agrega la tabla `users` como fuente de verdad de identidad/rol/estado. Nunca loguear ni persistir contraseñas en texto plano (hash con bcrypt/argon2 si se agrega login por password en una historia posterior).

---

## Relación con existentes

- Extiende: `internal/auth/apikey` (HU-009 RBAC legacy)
- Requisito para: HU-EVO-018 (API keys por usuario), HU-EVO-021 (UI Team & Roles)
