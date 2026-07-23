## Why

EP-005 estableció compatibilidad universal con 2 clientes (OpenWebUI, Free Claude Code) y parámetros básicos. La adopción amplia requiere soporte para 8 herramientas diferentes con routing automático por capacidad, parámetros avanzados (temperature, top_p, seed, thinking, etc.) y formato auto-detect. Sin estas capacidades, cada nueva herramienta que se quiera integrar exige cambios de código manual en el Gateway.

## What Changes

- Agregar routing automático por capability (modelo implícito) sin requerir `model` explícito en la request
- Nuevo endpoint `/responses` para soportar Responses API (formato OpenCode)
- Ampliar parámetros OpenAI soportados: temperature, top_p, seed, tool_choice, response_format (traducción desde cualquier formato)
- Ampliar parámetros Anthropic soportados: temperature, top_k, thinking, tool_use (traducción desde cualquier formato)
- Mejorar endpoint `/v1/models` con metadata por modelo (capability, latency, cost, availability)
- Middleware de normalización automática de formatos (OpenAI ↔ Anthropic ↔ Responses) sin intervención manual
- 8 guías de configuración reproducibles (una por herramienta: OpenWebUI, OpenCode, Claude Code, Free Claude Code, OpenHands, OpenClaw, CrewAI, UI-TARS)

## Capabilities

### New Capabilities

- `automatic-routing-capability`: Routing por capacidad implícita sin nombre de modelo. Acepta `model: "router:coding"` o ausencia de `model` para que el Gateway resuelva automáticamente el mejor modelo disponible según capability. Soporta fallback transparente.

- `responses-api-endpoint`: Nuevo endpoint `/responses` que acepta Responses API format (con `reasoning_effort`, `input`, etc.) y traduce internamente al formato normalizado del Gateway. Soporta streaming. Específico para OpenCode.

- `parameter-translation-openai`: Traducción completa de parámetros OpenAI (temperature, top_p, seed, tool_choice, response_format) desde cualquier formato de entrada (OpenAI, Anthropic, Responses) al formato interno normalizado, con mapeo a proveedores backend.

- `parameter-translation-anthropic`: Traducción completa de parámetros Anthropic (temperature, top_k, thinking, tool_use, max_tokens) desde cualquier formato de entrada al formato interno normalizado, con mapeo a proveedores backend.

- `format-auto-detection-middleware`: Middleware que detecta automáticamente el formato de entrada (OpenAI-compatible, Anthropic Messages, Responses API) sin necesidad de endpoints distintos, normaliza internamente, y enruta al adapter correcto. Tolerancia con variaciones menores de formato.

- `model-metadata-discovery`: Ampliación de `/v1/models` para retornar metadata por modelo (capability soportada, latencia p95, costo relativo, disponibilidad actual). Habilita debugging y elegibilidad checks en cliente.

- `client-setup-guides`: Documentación reproducible (8 archivos markdown con ejemplos curl/Python funcionales, env vars exactas, troubleshooting) para cada herramienta soportada. Cada guía es ejecutable sin cambios de código del cliente.

### Modified Capabilities

- `openai-compatible-api`: Ampliación de `/v1/chat/completions` + `/v1/embeddings` (EP-005) para soportar todos los parámetros OpenAI, enrutamiento automático por capability, y normalización de formatos variados. Sin breaking changes — requests sin `model` ahora se enrutan automáticamente en lugar de fallar.

- `anthropic-compatible-api`: Ampliación de `/v1/messages` (EP-005) para soportar todos los parámetros Anthropic (incluyendo `thinking` para extended thinking), enrutamiento automático, y normalización. Sin breaking changes — requests existentes siguen funcionando como antes.

## Impact

**Endpoints HTTP**: `/v1/chat/completions`, `/v1/messages`, `/responses`, `/v1/models` (todos ampliados)

**Componentes core**: Router (nuevo branch de routing automático), Middleware (nuevo pipeline de normalización), Adapters (traducción de parámetros para OpenAI/Anthropic/Google/OpenRouter/AIHubMix/local)

**Clientes soportados**: OpenWebUI, OpenCode, Claude Code, Free Claude Code, OpenHands, OpenClaw, CrewAI, UI-TARS (8 herramientas)

**Dependencias**:
- EP-001 (Router básico) — DONE
- EP-002 (Adapters OpenAI/Anthropic) — DONE
- EP-005 (Endpoints HTTP base) — DONE

**Breaking changes**: Ninguno. Las nuevas capacidades son aditivas. Requests existentes sin cambios siguen funcionando.

## Trazabilidad

**Épica**: EP-010  
**Historias cubiertas**: HU-042 (routing automático), HU-043 (endpoint /responses), HU-044 (parámetros OpenAI), HU-045 (parámetros Anthropic), HU-046 (metadata /v1/models), HU-047 (middleware normalización), HU-048 (documentación)  
**OpenSpec change**: compatibilidad-universal-clientes
