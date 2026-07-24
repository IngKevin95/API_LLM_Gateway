## ADDED Requirements

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
