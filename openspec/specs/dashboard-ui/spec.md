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

