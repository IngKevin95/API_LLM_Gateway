## ADDED Requirements

### Requirement: Endpoint `/v1/chat/completions` (OpenAI-compat)
El sistema SHALL exponer un endpoint compatible con la API de OpenAI para completación de chat. El endpoint MUST parsear la petición OpenAI, enrutarla al proveedor adecuado (según el router interno o `model` explícito) y devolver la respuesta en el formato exacto de OpenAI.

#### Scenario: Petición exitosa de chat
- **WHEN** un cliente hace una petición válida sin streaming a `/v1/chat/completions`
- **THEN** la Gateway enruta la petición, invoca al proveedor real, y emite una respuesta HTTP 200 con el payload estándar de OpenAI (`id`, `object="chat.completion"`, `choices`, `usage`).

### Requirement: Streaming SSE para Chat
El endpoint `/v1/chat/completions` SHALL soportar `stream: true`. La Gateway MUST mantener la conexión abierta y emitir eventos Server-Sent Events (SSE) en formato OpenAI (`data: {...}`) conforme el proveedor emita chunks.

#### Scenario: Petición con streaming
- **WHEN** un cliente solicita `/v1/chat/completions` con `stream: true`
- **THEN** la Gateway emite chunks asíncronos (`object="chat.completion.chunk"`) finalizando con `data: [DONE]`.

### Requirement: Endpoint `/v1/embeddings` (OpenAI-compat)
El sistema SHALL exponer un endpoint `/v1/embeddings` que acepte uno o múltiples inputs. Debe enrutar la petición a un modelo con la capacidad de `embedding` y devolver el JSON de OpenAI.

#### Scenario: Generación de embeddings
- **WHEN** el cliente envía textos a `/v1/embeddings`
- **THEN** el sistema devuelve un array de vectores (`object="embedding"`) respetando la estructura de uso y datos de OpenAI.

### Requirement: Endpoint `/v1/models` (OpenAI-compat)
El sistema SHALL proveer un endpoint `/v1/models` para descubrir capacidades disponibles en la Gateway de forma estandarizada.

#### Scenario: Lista de modelos
- **WHEN** el cliente consulta `/v1/models`
- **THEN** se devuelve la lista de los alias o capacidades expuestos (ej. `router.coding`, `gpt-4o`) imitando el formato nativo de modelos de OpenAI.
