# dashboard-ui Specification

## Purpose
TBD - created by archiving change ui-react-dashboard-notificaciones. Update Purpose after archive.
## Requirements
### Requirement: Dashboard con 4 tabs interactivos
El dashboard SHALL renderizar 4 tabs (Overview, Quotas, Alerts, Providers) al cargar `/dashboard`,
todos interactivos e independientemente navegables sin reload de página.
(Traza: HU-EVO-014)

#### Scenario: Dashboard carga
- **WHEN** el usuario entra en `/dashboard`
- **THEN** la página renderiza 4 tabs (Overview, Quotas, Alerts, Providers) y son clicables sin recargar

### Requirement: Tab Overview con métricas agregadas
El tab Overview SHALL mostrar uptime, total de requests, errores, y latencia p50/p95/p99, leídos
del mismo payload que `GET /metrics` (HU-060).
(Traza: HU-EVO-014)

#### Scenario: Overview muestra métricas del Gateway
- **WHEN** el Gateway lleva 10h online y el tab Overview se abre
- **THEN** muestra uptime, total requests, errors, latencia p50/p95/p99 consistentes con el JSON de `/metrics`

### Requirement: Tab Quotas con tabla de cuota por proveedor/modelo
El tab Quotas SHALL mostrar una tabla `[Provider, Model, Limit, Remaining, ResetAt, HealthStatus]`
poblada desde el bloque `quota` de `GET /metrics`, refrescada automáticamente cada 5 segundos sin
intervención del usuario.
(Traza: HU-EVO-014)

#### Scenario: Quotas muestra tabla completa
- **WHEN** existen 5 proveedores conocidos y el tab Quotas se abre
- **THEN** la tabla muestra las columnas Provider, Model, Limit, Remaining, ResetAt, HealthStatus para cada uno

#### Scenario: Refresco automático sin reload manual
- **WHEN** el usuario permanece en el tab Quotas y pasan 5 segundos
- **THEN** los datos se actualizan automáticamente sin que el usuario recargue la página

### Requirement: Tab Alerts con lista filtrada por RBAC
El tab Alerts SHALL mostrar la lista de alertas devuelta por `GET /alerts` (ya filtrada
server-side por tenant/scope), con columnas `[Severity, Provider, Model, Message, AlertTime]`,
resaltando en rojo las filas con `severity=critical`.
(Traza: HU-EVO-014)

#### Scenario: Alerts muestra alertas activas
- **WHEN** hay 3 alertas activas devueltas por `GET /alerts` y el tab Alerts se abre
- **THEN** lista las 3 alertas con Severity, Provider, Model, Message, AlertTime, mostrando en rojo las `critical`

#### Scenario: Filtrado por tenant respetado en UI
- **WHEN** un usuario T1 accede al dashboard con su API key
- **THEN** el tab Alerts solo muestra alertas que `GET /alerts` devolvió para T1 (el cliente no filtra nada adicional ni expone datos de otros tenants)

### Requirement: Tab Providers con estado por proveedor
El tab Providers SHALL mostrar, para cada proveedor conocido, su estado (healthy/unhealthy), última
respuesta y estado de circuit breaker, derivados de `GET /metrics`.
(Traza: HU-EVO-014)

#### Scenario: Providers muestra estado de cada proveedor
- **WHEN** existen 5 proveedores y el tab Providers se abre
- **THEN** muestra para cada uno su estado healthy/unhealthy, última respuesta y estado de circuit breaker

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

