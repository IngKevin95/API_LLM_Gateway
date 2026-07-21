# provider-adapters Specification

## Purpose
TBD - created by archiving change ep-002-resiliencia-conectividad. Update Purpose after archive.
## Requirements
### Requirement: Adapter OpenAI de chat y tool calling
El adapter OpenAI SHALL traducir el request interno al formato `/v1/chat/completions`, preservar el schema de tools, y normalizar respuesta y errores. (Traza: HU-020a)

#### Scenario: Chat básico
- **WHEN** el router selecciona un modelo OpenAI para un payload de chat normal
- **THEN** el adapter transforma la request a `/v1/chat/completions` y devuelve la respuesta normalizada

#### Scenario: Tool calling preservado
- **WHEN** el request incluye definición de tools y se enruta a OpenAI
- **THEN** el adapter preserva intacto el schema de tools y el formato de function call en la respuesta

#### Scenario: Falla externa normalizada
- **WHEN** OpenAI responde 500 o hay timeout
- **THEN** el adapter captura el error y retorna el formato estandarizado de falla para que la Gateway inicie failover

### Requirement: Adapter OpenAI de streaming SSE
El adapter OpenAI SHALL emitir tokens vía SSE transparentemente, permitir failover solo pre-primer-token, y cortar el socket ante Stream Idle Timeout sin failover mid-stream. (Traza: HU-020b)

#### Scenario: Streaming feliz
- **WHEN** llega un payload de chat con `stream: true` enrutado a OpenAI
- **THEN** el adapter establece un pipe SSE y emite los tokens transparentemente

#### Scenario: Failover pre-primer-token
- **WHEN** OpenAI falla antes de emitir el primer token en modo stream
- **THEN** el adapter reporta falla estandarizada y permite el failover transparente (nada emitido al cliente)

#### Scenario: Corte mid-stream
- **WHEN** OpenAI deja de emitir tokens tras comenzar y se excede el Stream Idle Timeout
- **THEN** el adapter cierra el socket y penaliza el score, sin failover mid-stream

### Requirement: Adapter OpenAI de embeddings
El adapter OpenAI SHALL redirigir a `/v1/embeddings`, respetar el límite de batch del proveedor sin truncar en silencio, y normalizar errores. (Traza: HU-020c)

#### Scenario: Embeddings feliz
- **WHEN** el router selecciona un modelo text-embedding de OpenAI para un payload a vectorizar
- **THEN** el adapter redirige a `/v1/embeddings` y retorna los vectores normalizados

#### Scenario: Lote grande
- **WHEN** el payload trae un lote grande de textos
- **THEN** el adapter respeta el límite de batch (particionando o rechazando con error claro) sin truncar silenciosamente

#### Scenario: Modelo de embedding no soportado
- **WHEN** se solicita un modelo de embedding inexistente en OpenAI
- **THEN** el adapter retorna falla estandarizada para que la Gateway aplique fallback

### Requirement: Adapter Anthropic de chat, roles y tool calling
El adapter Anthropic SHALL traducir el formato OpenAI al Messages API (extrae `system`, mapea roles), transformar tools a `tool_use`, inyectar `max_tokens` por defecto si falta, ignorar parámetros no soportados con WARN, y normalizar errores. (Traza: HU-021a)

#### Scenario: Traducción de roles
- **WHEN** llega un payload formato OpenAI (system en el array de messages) y se selecciona Claude
- **THEN** el adapter extrae el `system` y mapea los mensajes al formato Messages API de Anthropic

#### Scenario: Tool calling traducido
- **WHEN** el payload trae tools en JSON Schema OpenAI
- **THEN** el adapter transforma la estructura a `tool_use` de Anthropic Messages API

#### Scenario: Parámetro no soportado
- **WHEN** el cliente envía un parámetro OpenAI no soportado por Anthropic (ej. `seed`)
- **THEN** el adapter lo ignora de forma segura, lo advierte en log y permite la ejecución

#### Scenario: max_tokens ausente
- **WHEN** el request omite `max_tokens` (opcional en OpenAI, obligatorio en Anthropic)
- **THEN** el adapter inyecta un valor por defecto seguro (ej. 4096)

#### Scenario: Error de red normalizado
- **WHEN** Anthropic responde 5xx/429
- **THEN** el adapter traduce al formato estándar de error del Gateway para activar el failover

### Requirement: Adapter Anthropic de streaming
El adapter Anthropic SHALL transformar los eventos nativos a chunks SSE compatibles OpenAI, permitir failover solo pre-primer-token, y cortar ante Stream Idle Timeout sin failover mid-stream. (Traza: HU-021b)

#### Scenario: Streaming traducido
- **WHEN** una petición con `stream: true` va a Anthropic
- **THEN** el adapter transforma los eventos nativos (`message_start`, `content_block_delta`) a chunks SSE OpenAI

#### Scenario: Failover pre-primer-token
- **WHEN** Anthropic falla antes del primer `content_block_delta`
- **THEN** el adapter reporta falla estandarizada y permite el failover transparente

#### Scenario: Corte mid-stream
- **WHEN** Anthropic deja de emitir eventos tras iniciar y se excede el Stream Idle Timeout
- **THEN** el adapter cierra el socket y penaliza el score, sin failover mid-stream

### Requirement: Adapter local OpenAI-compatible
El adapter local SHALL reutilizar la traducción OpenAI-compat contra un `base_url` local, aplicar el timeout dinámico, y fallar con error estandarizado ante respuestas no compatibles sin crashear. (Traza: HU-024)

#### Scenario: Petición local feliz
- **WHEN** un servidor local OpenAI-compatible responde 200 a una petición de chat enrutada al modelo local
- **THEN** el adapter se comunica y devuelve la respuesta en el formato estándar del Gateway

#### Scenario: Timeout local
- **WHEN** el servidor local está colgado y no responde antes del TTFT/timeout dinámico
- **THEN** el adapter aplica el timeout y retorna el formato estandarizado de falla para iniciar failover

#### Scenario: Respuesta no OpenAI-compatible
- **WHEN** el servidor local responde 200 con un cuerpo que no respeta el esquema OpenAI (JSON mal formado o campos faltantes)
- **THEN** el adapter falla con error estandarizado sin crashear y la Gateway lo trata como falla del proveedor

