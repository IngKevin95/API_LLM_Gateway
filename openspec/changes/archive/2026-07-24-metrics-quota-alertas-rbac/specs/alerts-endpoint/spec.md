## ADDED Requirements

### Requirement: GET /alerts filtrado por tenant y scopes autorizados
El endpoint `GET /alerts` SHALL devolver únicamente alertas pertenecientes al `tenant_id` del token
de autenticación del requester y a modelos/capabilities autorizados por sus scopes, salvo que el
token sea el admin (`GATEWAY_ADMIN_TOKEN`), en cuyo caso devuelve todas las alertas sin filtro de
tenant; SHALL soportar paginación vía `page`/`limit`.
(Traza: HU-EVO-013)

#### Scenario: Cliente ve sus propias alertas
- **WHEN** cliente T1 con scope `capability:coding` tiene una alerta en Groq/mixtral y ejecuta `GET /alerts` con su API key
- **THEN** la respuesta incluye esa alerta (pertenece a su tenant y scope)

#### Scenario: Cliente no ve alertas de otros tenants
- **WHEN** cliente T1 ejecuta `GET /alerts`
- **THEN** la respuesta no incluye alertas de T2 ni T3 (filtrado por `tenant_id` del token)

#### Scenario: Admin ve todas las alertas
- **WHEN** se usa el token admin (`GATEWAY_ADMIN_TOKEN`) para `GET /alerts` sin filtro de tenant
- **THEN** la respuesta incluye todas las alertas del Gateway, de todos los tenants

#### Scenario: Scope insuficiente filtra la alerta
- **WHEN** un usuario tiene scope `capability:chat` pero la alerta corresponde a un modelo con `capability:vision`
- **THEN** `GET /alerts` filtra esa alerta fuera de la respuesta (usuario no ve alertas de capabilities no autorizadas)

#### Scenario: Paginación respeta RBAC
- **WHEN** existen 1000 alertas y se ejecuta `GET /alerts?page=2&limit=50`
- **THEN** la respuesta devuelve 50 alertas de la página 2, ya filtradas por RBAC antes de paginar
