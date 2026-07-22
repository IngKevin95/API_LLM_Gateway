## 1. Setup y Tipos Base
- [ ] 1.1 Definir estructuras base de Request/Response de OpenAI en `src/internal/api/openai/types.go` (ChatRequest, ChatResponse, Chunk).
- [ ] 1.2 Definir estructuras base de Request/Response de Anthropic en `src/internal/api/anthropic/types.go` (MessageRequest, MessageResponse, Events).

## 2. OpenAI Compat (SS1 & SS2 & SS3)
- [ ] 2.1 Implementar handler de completación (sin streaming) en `src/internal/api/openai/handler.go` (`/v1/chat/completions`).
- [ ] 2.2 Integrar routing en el handler de completación según `model` del payload o inferir la capacidad de routing.
- [ ] 2.3 Implementar soporte SSE (streaming) con la interfaz `http.Flusher` para OpenAI.
- [ ] 2.4 Implementar handler de embeddings en `/v1/embeddings` para procesar inputs simples y múltiples.
- [ ] 2.5 Implementar endpoint simulado en `/v1/models` devolviendo un listado harcodeado con los alias habilitados de la Gateway.

## 3. Anthropic Compat (SS4)
- [ ] 3.1 Implementar handler Messages (`/v1/messages`) en `src/internal/api/anthropic/handler.go`.
- [ ] 3.2 Implementar soporte SSE (streaming) para Anthropic, transcribiendo los eventos `message_start`, `content_block_delta`, etc.
- [ ] 3.3 Habilitar soporte en las estructuras de Anthropic para parsear transparentemente `tool_use` y `tool_result`.

## 4. MCP Integration (SS5)
- [ ] 4.1 Definir estructuras del protocolo MCP (Model Context Protocol).
- [ ] 4.2 Implementar handler `/mcp/sse` o mecanismo básico para listar `tools` simuladas y verificar seguridad (401/403).

## 5. Pruebas y Enrutado
- [ ] 5.1 Conectar todos los nuevos endpoints al router principal (`src/cmd/gateway/main.go` o un mux central).
- [ ] 5.2 Desarrollar tests unitarios (TDD) para el manejo de JSONs y aserciones de `Flusher` mock.
