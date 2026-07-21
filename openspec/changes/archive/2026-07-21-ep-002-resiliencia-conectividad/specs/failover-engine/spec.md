## ADDED Requirements

### Requirement: Failover en cadena con degradación local
El Failover Engine SHALL recorrer la cadena de fallback del router probando cada eslabón vía su Adapter, degradar a modelo local como último eslabón, retornar 502 al agotar el pool, y NO hacer failover ante 400 del cliente. (Traza: HU-004a)

#### Scenario: Failover simple
- **WHEN** el `providerA` devuelve 503 y el engine intercepta el error
- **THEN** enruta la petición al `providerB` de forma transparente para el cliente

#### Scenario: Pool agotado
- **WHEN** todos los providers de la capacidad fallan y no hay más alternativas
- **THEN** retorna 502 Bad Gateway con el detalle del último error

#### Scenario: Retry condicional (429)
- **WHEN** el `providerA` devuelve 429 Too Many Requests
- **THEN** el engine aplica failover en lugar de retry, para proteger la cuota global del proveedor A

#### Scenario: Degradación a modelo local
- **WHEN** todos los providers remotos fallan y hay un modelo local configurado
- **THEN** el engine envía la petición al modelo local sin exponer el fallo al cliente

#### Scenario: Payload mal formado (400) no hace failover
- **WHEN** el cliente envía un prompt inválido y el primer proveedor devuelve 400
- **THEN** el Gateway NO intenta failover y retorna 400 al cliente inmediatamente

### Requirement: Timeouts dinámicos por capacidad y Stream Idle Timeout
El Failover Engine SHALL aplicar un TTFT por capacidad (estricto para chat/código, relajado para reasoning) y un Stream Idle Timeout mid-stream, abortando y penalizando según corresponda. (Traza: HU-004c)

#### Scenario: TTFT excedido en capacidad estándar
- **WHEN** el proveedor primario tarda más del umbral estricto pre-stream (ej. 2.0s) para chat/código
- **THEN** el Gateway aborta la conexión primaria y ejecuta el failover silenciosamente

#### Scenario: Timeout dinámico para reasoning
- **WHEN** llega una petición de capacidad `reasoning`
- **THEN** el Gateway aplica el timeout de reasoning configurado (ej. < 30s) sin disparar failover durante el pensamiento prolongado

#### Scenario: Stream Idle Timeout
- **WHEN** el proveedor deja de emitir tokens por más del Stream Idle Timeout configurado
- **THEN** el Gateway corta el socket unilateralmente y penaliza el score del proveedor
