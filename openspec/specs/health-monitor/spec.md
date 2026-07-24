## ADDED Requirements

### Requirement: Detección de 429 con retiro temporal y backoff
El Health Monitor SHALL detectar respuestas HTTP 429 de cualquier proveedor, retirarlo
temporalmente de la selección del Router respetando `Retry-After` (o un default de 30s si el
header está ausente), reactivarlo automáticamente al vencer el retiro, abortar streams mid-stream
sin aplicar failover transparente, y aplicar backoff exponencial ante 429 repetidos.
(Traza: HU-EVO-004)

#### Scenario: Proveedor recupera tras 429
- **WHEN** Health Monitor lanza un health check contra un proveedor que respondió 429 hace 30s
- **THEN** ve respuesta 200, marca el proveedor como `healthy=true` y lo reactiva en el Router

#### Scenario: Retiro temporal respetando Retry-After
- **WHEN** un request a un proveedor devuelve 429 con header `Retry-After: 60`
- **THEN** Health Monitor retira ese proveedor de la selección por 60s y lo reactiva automáticamente después

#### Scenario: 429 sin Retry-After usa default
- **WHEN** un proveedor devuelve 429 sin header `Retry-After`
- **THEN** Health Monitor asume un default de 30s, retira el proveedor y lo reactiva después de ese tiempo

#### Scenario: 429 mid-stream aborta sin failover transparente
- **WHEN** un Stream en curso recibe 429 a mitad de la emisión
- **THEN** Health Monitor aborta el stream, retorna error al cliente y retira el proveedor; no hay failover transparente mid-stream

#### Scenario: Backoff exponencial ante 429 repetidos
- **WHEN** un proveedor recibe 3 respuestas 429 en 10s
- **THEN** Health Monitor incrementa la duración del retiro exponencialmente (30s → 60s → 120s)
