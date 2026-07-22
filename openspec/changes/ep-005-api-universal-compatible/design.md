## Context
La API Gateway necesita exponer interfaces compatibles con los estándares de la industria (OpenAI y Anthropic) para habilitar herramientas que asumen estos contratos, como Free Claude Code u OpenCode. Esto nos permite actuar como un proxy inteligente y "LLM universal". Se debe implementar la traducción bidireccional (Request ↔ Internal ↔ Response) y soportar Server-Sent Events (SSE) para el streaming de completaciones.

## Goals / Non-Goals
**Goals:**
- Soportar los payloads exactos de OpenAI para chat y embeddings.
- Soportar el payload de Anthropic Messages con streaming y tool use.
- Habilitar MCP para el descubrimiento de agentes.
- Enrutar por capacidad cuando no se especifica un modelo.

**Non-Goals:**
- Soportar la API de Completions legacy (v1/completions). solo Chat.
- Implementar tool execution en la Gateway (solo hacer passthrough del tool_use).
- Mantener estado de las conversaciones (stateless gateway).

## Decisions
1. **Parsers y Serializadores Nativos:** En vez de hacer un simple proxy reverso HTTP puro, parseamos la request a una estructura Go fuertemente tipada (`api/openai/types.go` y `api/anthropic/types.go`). Rationale: nos permite inspeccionar los mensajes, inyectar el sistema KMS, cobrar uso y luego serializar al proveedor final.
2. **Streaming por `http.Flusher`:** El soporte SSE se realizará usando la interfaz `http.Flusher` de Go en el handler, escribiendo chunks a la conexión de forma continua e inyectando un timeout en caso de caída del backend.
3. **MCP Server por Stdio / SSE:** El protocolo MCP se integrará inicialmente vía SSE, exponiendo un endpoint `/mcp/sse` o usando Stdio si se llama como subprocess, delegando a un submódulo `internal/mcp`.

## Risks / Trade-offs
- [Riesgo] Dificultad para mantener sincronizados los tipos de datos con las APIs de OpenAI/Anthropic que evolucionan. → Mitigación: Usar un subconjunto estricto de campos necesarios y descartar/ignorar extras (strict unmarshaling solo de lo clave).
- [Riesgo] Timeout en streaming largo. → Mitigación: Ajustar tiempos de `ReadTimeout`/`WriteTimeout` del `http.Server` de Go para tolerar respuestas largas en SSE.
