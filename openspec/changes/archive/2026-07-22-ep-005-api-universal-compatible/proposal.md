## Why

La Gateway requiere exponer endpoints compatibles con las APIs estándar de OpenAI y Anthropic para asegurar la adopción sin fricción. Esto permite que herramientas existentes (como Free Claude Code, OpenCode o aplicaciones de terceros) consuman la Gateway de forma transparente, logrando la meta de ser un "LLM universal".

## What Changes

- Implementación del endpoint `/v1/chat/completions` (OpenAI-compat) con y sin streaming SSE (HU-012a, HU-012b).
- Implementación del endpoint `/v1/embeddings` (OpenAI-compat) con soporte de enrutamiento y lotes (HU-012c).
- Implementación del endpoint `/v1/messages` (Anthropic-compat) soportando streaming y tool use para Free Claude Code (HU-013, HU-016).
- Implementación de un servidor MCP (Model Context Protocol) para descubrimiento y configuración por parte de agentes (HU-033).

## Capabilities

### New Capabilities

- `openai-compat-endpoints`: Endpoints compatibles con OpenAI (`chat/completions`, `embeddings`) que enrutan peticiones al modelo adecuado y manejan streaming SSE.
- `anthropic-compat-endpoints`: Endpoints compatibles con Anthropic (`messages`) diseñados específicamente para interactuar sin fricción con clientes como Free Claude Code, soportando streaming y tool use.
- `mcp-integration`: Servidor compatible con Model Context Protocol (MCP) para el descubrimiento de capacidades y configuración de agentes.

### Modified Capabilities

- Ninguna

## Impact

- **API HTTP**: Exposición pública de endpoints HTTP REST.
- **Rutas y Handlers**: Conexión entre los endpoints HTTP y la lógica core (Router, Retry, Circuit Breaker).
- **Integraciones**: Habilita el uso de Free Claude Code local apuntando a la Gateway.

## Trazabilidad

EP-005 · API universal compatible
- HU-012a
- HU-012b
- HU-012c
- HU-013
- HU-016
- HU-033
