## ADDED Requirements

### Requirement: Tab Team con gestión RBAC (admin-only)
El dashboard SHALL renderizar una tab "Team", visible solo si `GET /users/me` devuelve
`role=admin`, con una tabla de usuarios (usuario, rol, scopes/tenant, estado, última actividad)
poblada desde `GET /users`, permitiendo invitar (`POST /users`) y cambiar rol/estado
(`PATCH /users/{id}`) sin recargar la página. (Traza: HU-EVO-021)

#### Scenario: Ver tabla de equipo
- **WHEN** un admin abre la tab "Team"
- **THEN** ve la tabla completa poblada desde `GET /users`, con badges de rol y estado

#### Scenario: Invitar miembro
- **WHEN** el admin completa email+rol+scopes en el modal "Invite member" y confirma
- **THEN** se llama `POST /users` y la nueva fila aparece en estado "invited" sin recargar

#### Scenario: Cambiar rol o suspender
- **WHEN** el admin cambia el rol o suspende a un usuario desde la fila
- **THEN** se llama `PATCH /users/{id}` y el badge de esa fila se actualiza in-place

#### Scenario: No-admin no ve la tab
- **WHEN** un usuario con rol `operator` carga el dashboard
- **THEN** la tab "Team" no aparece en la navegación (el frontend no expone controles que el
  backend igual rechazaría con 403)

#### Scenario: Invitar email duplicado
- **WHEN** `POST /users` responde `409`
- **THEN** la UI muestra un mensaje de error y conserva los datos ya tipeados en el formulario

### Requirement: Tab Profile & Security con autoservicio de credenciales
El dashboard SHALL renderizar una tab "Profile & Security", visible para cualquier usuario
autenticado, mostrando el perfil propio (`GET /users/me`), sus API keys
(`GET/POST/DELETE /users/{id}/api-keys`), sus sesiones activas (`GET/DELETE /sessions`) y el
estado de 2FA (`POST /auth/mfa/enroll` + `POST /auth/mfa/verify`). (Traza: HU-EVO-022)

#### Scenario: Ver perfil y API keys
- **WHEN** el usuario abre la tab "Profile & Security"
- **THEN** ve su tarjeta de perfil y la tabla de sus API keys con el prefijo enmascarado

#### Scenario: Generar nueva key
- **WHEN** el usuario genera una key nueva
- **THEN** se llama `POST /users/{id}/api-keys` y la key en texto plano se muestra una única vez
  en un modal, luego solo aparece enmascarada en la tabla

#### Scenario: Revocar key y cerrar sesión
- **WHEN** el usuario revoca una key o cierra una sesión
- **THEN** se llaman `DELETE /users/{id}/api-keys/{keyId}` y `DELETE /sessions/{id}` y la fila
  correspondiente desaparece sin recargar la página

#### Scenario: Activar 2FA
- **WHEN** el usuario activa el toggle de 2FA, escanea el QR y confirma el código
- **THEN** se llaman `POST /auth/mfa/enroll` y `POST /auth/mfa/verify`, y el toggle queda activo

#### Scenario: Código 2FA incorrecto
- **WHEN** el código de confirmación de 2FA es incorrecto
- **THEN** la UI muestra el error del backend sin activar el toggle, permitiendo reintentar
