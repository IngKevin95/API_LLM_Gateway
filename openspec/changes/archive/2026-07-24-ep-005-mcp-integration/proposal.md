## Why

El Gateway implementa todos los endpoints REST (OpenAI-compat, Anthropic-compat) pero no tiene un servidor MCP (Model Context Protocol). Sin MCP, agentes como Claude Code no pueden descubrir las capacidades del Gateway de forma estandarizada ni integrarse en flujos multi-agente. MCP es el protocolo estándar de Anthropic para inter-comunicación de agentes con herramientas y contextos.

## Trazabilidad

- Épica: **EP-005** — API Universal Compatible (`docs/03-backlog/epicas.md`)
- Historia: **HU-033** — Integración MCP para multi-agentes

## What Changes

- `src/internal/api/mcp/handler.go`: reemplaza el stub ad-hoc por un servidor **JSON-RPC 2.0** completo sobre HTTP POST `/mcp`.
- `src/internal/api/mcp/types.go`: tipos legacy del stub conservados por compatibilidad; tipos de protocolo MCP viven en `handler.go`.
- `src/internal/api/mcp/handler_test.go`: 7 tests que cubren los 4 ACs de HU-033.
- Sin cambios en `cmd/gateway/main.go` (wiring del endpoint `/mcp` difiere al sprint de integración final).

## Capabilities

### New Capabilities

- `mcp-server`: Servidor MCP JSON-RPC 2.0 sobre HTTP con métodos `initialize` (handshake + negociación de versión), `tools/list` (descubrimiento de herramientas y modelos disponibles), `tools/call` (stub con ACK). AuthN Bearer token. Validación de versión de protocolo (426 ante versión incompatible).

### Modified Capabilities

_(ninguna)_

## Impact

- Código: `src/internal/api/mcp/` (3 archivos modificados).
- Sin dependencias externas nuevas — implementación pura sobre `net/http` y `encoding/json`.
- `cmd/gateway/main.go` no se toca en este slice; el endpoint `/mcp` se registra en el sprint de wiring final (Fase 4).
