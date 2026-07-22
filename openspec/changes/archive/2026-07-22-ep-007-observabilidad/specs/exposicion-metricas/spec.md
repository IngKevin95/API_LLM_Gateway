## ADDED Requirements

### Requirement: Exposición de métricas agregadas por modelo y proveedor
El sistema SHALL exponer endpoints protegidos por scope de operador para consultar latencias, success rate, tokens, cuota y costo.

#### Scenario: Consulta de métricas actuales
- **WHEN** un operador consulta el endpoint de métricas
- **THEN** se devuelve latencia (avg/p95), success rate, tokens, quota y costo agregados por modelo y proveedor

#### Scenario: Petición sin permisos
- **WHEN** un consumidor sin permiso consulta el endpoint
- **THEN** el sistema devuelve un error 403 Forbidden

#### Scenario: Gateway sin tráfico (Empty State)
- **WHEN** se consultan las métricas pero no ha habido tráfico
- **THEN** devuelve un JSON con contadores en 0 sin dar error

### Requirement: Dashboard de Métricas y Ranking
El sistema SHALL permitir filtrar las métricas por proveedor para mostrar un ranking.

#### Scenario: Filtro por proveedor
- **WHEN** el operador filtra por provider en la query param
- **THEN** devuelve el ranking de modelos solo de ese proveedor

#### Scenario: Backend Histórico Caído
- **WHEN** el almacén de datos subyacente falla al consultar
- **THEN** devuelve 500 Internal Server Error
