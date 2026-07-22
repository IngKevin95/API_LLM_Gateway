## ADDED Requirements

### Requirement: Adapter OpenRouter (HU-031)
El sistema SHALL tener un adapter que implemente `adapter.Adapter` para OpenRouter, inyectando los headers requeridos por la plataforma en cada petición.

#### Scenario: Chat transparente con headers obligatorios
- **GIVEN** un request de chat con un model ID de OpenRouter (ej. `anthropic/claude-3-haiku`)
- **WHEN** el adapter envía la petición a OpenRouter
- **THEN** incluye los headers `HTTP-Referer` y `X-Title` y retorna respuesta normalizada exitosamente

#### Scenario: Modelo upstream no disponible (503)
- **GIVEN** que OpenRouter reporta que el modelo upstream está saturado (503)
- **WHEN** el adapter recibe el código de estado
- **THEN** emite `*adapter.ProviderError{Retryable: true}` para que el Gateway inicie failover

#### Scenario: Timeout TTFT excedido
- **GIVEN** que OpenRouter no responde dentro del límite de timeout configurado
- **WHEN** el contexto del request expira
- **THEN** el adapter retorna error de contexto cancelado que el Failover Engine interpreta como retryable
