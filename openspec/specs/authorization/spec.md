# authorization Specification

## Purpose
TBD - created by archiving change ep-004a-identidad-accesos. Update Purpose after archive.
## Requirements
### Requirement: Autorización por scope/RBAC y aislamiento multi-tenant
El sistema SHALL autorizar cada petición por scope de capacidad y tenant, aislando datos entre tenants y exigiendo scope explícito para capacidades/modelos restringidos; toda violación responde 403 y se registra. (Traza: HU-009)

#### Scenario: Scope permitido
- **WHEN** llega una petición `coding` del tenant T1 con una credencial de scope `capability:coding` y tenant T1
- **THEN** la Gateway autoriza y procesa

#### Scenario: Capacidad fuera de scope
- **WHEN** llega una petición `image` con una credencial sin scope `capability:image`
- **THEN** responde 403 sin procesar y registra el intento

#### Scenario: Cruce de tenant
- **WHEN** una credencial del tenant T1 intenta acceder a recursos/config del tenant T2
- **THEN** responde 403 y los datos de T2 no son visibles ni enumerables

#### Scenario: Modelo restringido
- **WHEN** una credencial autorizada a `coding` fuerza un `model` concreto vetado
- **THEN** responde 403 aunque la capacidad esté permitida

#### Scenario: Vision requiere scope de confianza
- **WHEN** llega una petición `vision` con una credencial sin el scope `capability:vision:trusted`
- **THEN** responde 403 sin procesar y registra el intento (vision deshabilita el DLP del contenido, por eso exige scope explícito)

