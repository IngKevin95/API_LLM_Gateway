## 1. Implementación Base del Protocolo MCP
- [x] 1.1 Eliminar stub ad-hoc en `src/internal/api/mcp/`
- [x] 1.2 Implementar servidor HTTP POST `/mcp` para JSON-RPC 2.0
- [x] 1.3 Estructurar tipos MCP y JSON-RPC en memoria
- [x] 1.4 Validar Bearer token antes de cualquier dispatch JSON-RPC
- [x] 1.5 Implementar manejo de errores estándar JSON-RPC (-32600, -32601, etc)

## 2. Handlers Específicos
- [x] 2.1 Método `initialize`: Handshake y validación de versión (`2024-11-05`)
- [x] 2.2 Método `initialize`: Responder con 426 Upgrade Required ante versiones incompatibles
- [x] 2.3 Método `tools/list`: Devolver catálogo de herramientas built-in (`get_quota`, `list_capabilities`, `route_chat`)
- [x] 2.4 Método `tools/list`: Integrar con `ModelSource` para exponer dinámicamente un tool `route_<model>` por modelo
- [x] 2.5 Método `tools/call`: Proveer stub/ACK para preparar el wiring final en la Fase 4

## 3. Calidad y Pruebas
- [x] 3.1 Pruebas unitarias para HTTP methods (solo POST) y payload JSON inválido
- [x] 3.2 Pruebas de autenticación (Token válido, inválido, sin token)
- [x] 3.3 Cobertura para compatibilidad de versión en `initialize`
- [x] 3.4 Verificar catálogo emitido por `tools/list`
- [x] 3.5 Todos los tests pasan (`go test ./src/internal/api/mcp/...`)
