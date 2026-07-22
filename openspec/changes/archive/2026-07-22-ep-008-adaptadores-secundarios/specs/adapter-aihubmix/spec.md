## ADDED Requirements

### Requirement: Adapter AIHubMix (HU-029)
El sistema SHALL tener un adapter que implemente `adapter.Adapter` para el proveedor AIHubMix usando su endpoint OpenAI-compatible.

#### Scenario: Chat básico exitoso
- **GIVEN** un request de chat válido y AIHubMix configurado como proveedor activo
- **WHEN** el Router selecciona AIHubMix y el adapter envía la petición
- **THEN** el adapter retorna la respuesta normalizada (`adapter.Response`) sin error

#### Scenario: Rate limit dispara failover
- **GIVEN** que AIHubMix devuelve HTTP 429
- **WHEN** el adapter procesa la respuesta
- **THEN** emite `*adapter.ProviderError{Retryable: true}` para que el Failover Engine active el siguiente proveedor

#### Scenario: Error upstream 500/503
- **GIVEN** que AIHubMix retorna 500 o 503
- **WHEN** el adapter recibe el código de estado
- **THEN** emite `*adapter.ProviderError{Retryable: true}` con el código HTTP incluido

#### Scenario: Parámetros no soportados ignorados
- **GIVEN** que el request incluye params en `Request.Params` no soportados por AIHubMix (ej. `logprobs`)
- **WHEN** el adapter construye el payload
- **THEN** omite el parámetro desconocido de forma segura y procesa el resto del request
