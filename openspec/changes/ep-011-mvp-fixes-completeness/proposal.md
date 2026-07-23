# Proposal: MVP Fixes & Completeness (Observabilidad Operacional)

## Why

MVP actual tiene handlers rotos (`/v1/chat/completions` 500, `/v1/embeddings` 503, `/v1/messages` 400) y falta OmniRoute como proveedor. Sin logging estructurado ni métricas reales, el equipo no puede depurar ni operar en producción. Esta épica cierra el puente entre "discovery completado" (EP-001-010) y "MVP funcional end-to-end" que el equipo pueda desplegar y mantener.

## What Changes

- **Handlers OpenAI/Anthropic**: debug y fix de `/v1/chat/completions`, `/v1/embeddings`, `/v1/messages` (status codes correctos + error handling)
- **Validación de routing**: verificar que `Router.Route()` elige proveedor correcto según score
- **OmniRoute adapter**: crear adaptador para proveedor local gratuito (internal/adapter/omniroute/)
- **Config normalization**: alinear IDs de proveedores en config.yaml con buildAdapters() (google-gemini → google, local-ollama → ollama, etc.)
- **Logging estructurado JSON**: en todos los handlers (request ID, provider, latency, tokens, error)
- **Métricas operacionales**: exposición en `/metrics` (p50/p95/p99 latencies, success_rate por proveedor)

## Capabilities

### New Capabilities
- `omniroute-adapter`: Adaptador para OmniRoute (proveedor local gratuito, compatible OpenAI API)
- `structured-logging`: Logging JSON centralizado en handlers con request ID y traza
- `operational-metrics`: Endpoint `/metrics` con latencies reales (p50/p95/p99) y success_rate por provider

### Modified Capabilities
- `openai-handler`: Fix de ProcessChat() (status 200 + choice.content válido) + logging JSON
- `embeddings-handler`: Fix de ProcessEmbedding() + logging JSON
- `anthropic-handler`: Fix de ProcessMessages() (status 200 + content válido) + logging JSON
- `router`: Validación de score + logging de decisión de proveedor
- `provider-registry`: Normalización de IDs (config.yaml ↔ buildAdapters())

## Impact

**Code**:
- `internal/processor/gateway.go`: ProcessChat(), ProcessEmbedding(), ProcessMessages()
- `internal/router/`: routing logic + validation
- `internal/adapter/omniroute/`: nuevo adaptador
- `internal/config/`: normalization de provider IDs
- `cmd/gateway/main.go`: buildAdapters() + metrics handler

**APIs**:
- `POST /v1/chat/completions`: HTTP 200 (antes 500)
- `POST /v1/embeddings`: HTTP 200 (antes 503)
- `POST /v1/messages`: HTTP 200 (antes 400)
- `GET /metrics`: nuevas métricas operacionales

**Dependencies**:
- Ninguna nueva (logging vía slog, métricas vía math/stats built-in)

**External Services**:
- OmniRoute: agregado como proveedor local opcional

## Trazabilidad

**Épica**: EP-011 (MVP Fixes & Completeness)

**Historias**:
- HU-050: Añadir logging estructurado al OpenAI handler
- HU-051: Debuguear y fijar GatewayProcessor.ProcessChat()
- HU-052: Validar que Router.Route() elige proveedor correcto
- HU-053: Crear adaptador OmniRoute (internal/adapter/omniroute/)
- HU-054: Registrar adaptador OmniRoute en buildAdapters()
- HU-055: Test de conectividad OmniRoute → Gateway
- HU-056: Alinear IDs de proveedores en config.yaml con buildAdapters()
- HU-057: Implementar GatewayProcessor.ProcessEmbedding()
- HU-058: Debuguear y fijar handler Anthropic /v1/messages
- HU-059: Implementar logging estructurado en todos los handlers
- HU-060: Implementar /metrics con datos reales de operación
