## ADDED Requirements

### Requirement: Endpoint Messages (`/v1/messages`) Anthropic-compat
El sistema SHALL exponer un endpoint compatible con la API de Messages de Anthropic, destinado a habilitar clientes como Free Claude Code de forma transparente.

#### Scenario: Invocación normal
- **WHEN** el cliente Free Claude Code envía una petición Messages
- **THEN** la Gateway procesa y enruta internamente, respondiendo con el formato JSON nativo de Anthropic (con campos `type`, `role`, `content`, `usage`).

### Requirement: Soporte de Streaming (Anthropic)
El endpoint Messages SHALL soportar peticiones con streaming activo, transcribiendo o emitiendo los eventos SSE de Anthropic nativos (`message_start`, `content_block_delta`, `message_delta`, etc.).

#### Scenario: Streaming con Claude Code
- **WHEN** el cliente solicita streaming
- **THEN** la respuesta consiste en un flujo SSE compatible con los clientes oficiales de Anthropic.

### Requirement: Soporte de Tool Use y Tool Result (Anthropic)
La Gateway SHALL preservar y transferir los bloques de declaración y uso de herramientas (`tool_use`, `tool_result`) en el formato exacto requerido por Anthropic, sin alterar los esquemas.

#### Scenario: Claude Code invoca bash
- **WHEN** la respuesta del modelo sugiere el uso de una herramienta (tool_use)
- **THEN** la Gateway transmite correctamente este tipo de bloque al cliente y soporta recibir subsecuentemente el bloque `tool_result`.
