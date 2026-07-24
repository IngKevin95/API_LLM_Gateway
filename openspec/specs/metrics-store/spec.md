# metrics-store Specification

## Purpose
TBD - created by archiving change metrics-quota-alertas-rbac. Update Purpose after archive.
## Requirements
### Requirement: Snapshot de cuota por proveedor/modelo en /metrics
`metrics.Store` SHALL invocar `quota.Manager.Snapshot()` en cada request a `GET /metrics` y
exponer un bloque `quota: [{provider, model, limit, remaining, reset_at, healthy}]`, filtrado por
los scopes/capabilities del requester cuando no es admin, y responder en menos de 100ms por ser
lectura en RAM sin I/O adicional.
(Traza: HU-EVO-011)

#### Scenario: Snapshot devuelto con quota learned
- **WHEN** Quota Manager tiene learned quota para Groq y se ejecuta `GET /metrics`
- **THEN** la respuesta incluye `quota: [{provider: "groq", model: "mixtral", limit: 14400, remaining: 14200, reset_at: "2026-07-24T19:00:00Z", healthy: true}]`

#### Scenario: Múltiples modelos por proveedor
- **WHEN** Mistral tiene 3 modelos y se ejecuta `GET /metrics`
- **THEN** la respuesta lista los 3 modelos con `remaining` individual para cada uno

#### Scenario: Proveedor sin quota learned aún
- **WHEN** un proveedor recién agregado nunca respondió y se ejecuta `GET /metrics`
- **THEN** la respuesta lista ese proveedor con `remaining: <quota_hint>` (valor inicial del YAML) y `learned_at: null`

#### Scenario: Respuesta rápida con volumen alto
- **WHEN** hay 25 proveedores x 5 modelos (125 cuotas) y se ejecuta `GET /metrics`
- **THEN** la respuesta se entrega en menos de 100ms

#### Scenario: Filtrado por scope del requester
- **WHEN** un usuario no-admin con scope `capability:coding` ejecuta `GET /metrics`
- **THEN** la respuesta filtra el bloque `quota` mostrando solo los modelos que el usuario puede usar

