## ADDED Requirements

### Requirement: Circuit Breaker pasivo con Max In-Flight
El Circuit Breaker SHALL marcar un proveedor inalcanzable temporalmente ante fallo o exceso de Max In-Flight, hacer fast-fail 0-I/O cuando se supera el límite, y reactivar tras backoff con health check sano — para prevenir Failover Suicide. (Traza: HU-004b)

#### Scenario: Breaker pasivo tras fallo
- **WHEN** un proveedor devuelve 429/500/timeout o supera su Max In-Flight y ocurre failover al secundario
- **THEN** el proveedor se marca inalcanzable durante `cooldown_ms` para evitar Failover Suicide de peticiones concurrentes

#### Scenario: Max In-Flight excedido hace fast-fail
- **WHEN** las peticiones en curso superan el Max In-Flight del proveedor y llega una nueva
- **THEN** el breaker hace fast-fail (0 I/O) sin esperar el timeout

#### Scenario: Reactivación tras backoff
- **WHEN** transcurre el periodo de gracia (backoff fijo, ej. 30s) y el health check reporta al proveedor sano
- **THEN** el proveedor vuelve a la cadena de fallback
