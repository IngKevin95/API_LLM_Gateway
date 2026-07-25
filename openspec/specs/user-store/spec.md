# user-store Specification

## Purpose
Store persistente de usuarios (identidad, rol, estado, tenant, scopes, API keys) en PostgreSQL,
con CRUD admin-only y autoservicio de perfil propio para cualquier usuario autenticado.

## Requirements

### Requirement: Store de usuarios persistente con CRUD admin
El Gateway SHALL persistir usuarios (email, rol, estado, tenant, scopes) en PostgreSQL, exponiendo
`POST /users`, `GET /users` y `PATCH /users/{id}` restringidos a administradores. (Traza:
HU-EVO-017)

#### Scenario: Invitar usuario
- **GIVEN** un admin autenticado
- **WHEN** hace `POST /users` con email, rol y scopes/tenant
- **THEN** se crea el registro en `users` con `status=invited` y responde `201` con el ID

#### Scenario: Listar usuarios filtrado por tenant
- **GIVEN** un admin de tenant T1 (no admin global)
- **WHEN** hace `GET /users`
- **THEN** solo ve usuarios con `tenant=T1`, nunca de otros tenants

#### Scenario: Cambiar rol/estado
- **GIVEN** un usuario `operator` activo
- **WHEN** un admin hace `PATCH /users/{id}` con `{"status":"suspended"}`
- **THEN** el usuario pasa a `suspended` y sus API keys dejan de autenticar de inmediato

#### Scenario: No-admin intenta invitar
- **GIVEN** un usuario con rol `operator`
- **WHEN** intenta `POST /users`
- **THEN** recibe `403 Forbidden` y no se crea ningún registro

#### Scenario: Email duplicado
- **GIVEN** ya existe un usuario con `email=x@y.com`
- **WHEN** se invita de nuevo ese mismo email
- **THEN** recibe `409 Conflict`, sin duplicar el registro

### Requirement: GET /users/me expone el perfil propio
El Gateway SHALL exponer `GET /users/me`, resolviendo el usuario autenticado desde `auth.Identity`
(no desde `AdminContext`), de forma que cualquier usuario autenticado -- admin u operator -- vea su
propio perfil sin necesitar permisos de administrador. (Traza: HU-EVO-022)

#### Scenario: Usuario autenticado ve su propio perfil
- **GIVEN** un usuario autenticado con un JWT o API key válida
- **WHEN** hace `GET /users/me`
- **THEN** recibe `200` con su propio registro (`id`, `email`, `role`, `status`, `tenant`, `scopes`)

#### Scenario: Sin identidad resuelta
- **GIVEN** una petición sin token válido
- **WHEN** hace `GET /users/me`
- **THEN** recibe `401 Unauthorized`, sin exponer ningún dato de otro usuario
