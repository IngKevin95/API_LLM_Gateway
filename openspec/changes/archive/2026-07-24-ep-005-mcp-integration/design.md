## Decisiones de diseño

### Protocolo: JSON-RPC 2.0 sobre HTTP POST

MCP usa JSON-RPC 2.0 como capa de mensajería. Se eligió transport HTTP streamable (POST `/mcp`) en lugar de stdio porque el Gateway es un servidor HTTP (no un proceso de CLI). Esto permite que Claude Code y otros clientes se conecten apuntando al endpoint con `--mcp-url`.

### Métodos implementados

| Método | Comportamiento |
|---|---|
| `initialize` | Handshake. Negocia `protocolVersion`. Si la versión del cliente < `2024-11-05` → HTTP 426 + error JSON-RPC. |
| `tools/list` | Devuelve herramientas builtin (`list_capabilities`, `get_quota`, `route_chat`) + herramientas dinámicas por modelo (via `ModelSource`). |
| `tools/call` | ACK stub — responde correctamente pero el wiring real al Router se completa en el sprint de integración final. |
| Otros | `codeMethodNotFound` (-32601). |

### AuthN

Bearer token validado antes del dispatch JSON-RPC. Token vacío = sin auth (modo dev). Tokens incorrectos → HTTP 403 + error JSON-RPC. Integración con `auth.Identity` de EP-004A difiere al wiring final de `cmd/gateway`.

### ModelSource (interfaz)

`Handler` recibe una interfaz `ModelSource` con `ModelNames()` y `HasCapability()`, la misma que satisface `*registry.Registry`. Sin dependencia directa al registry; permite tests con stub.

### Sin dependencias externas

Implementación sobre `net/http` + `encoding/json` estándar. No se añade `mark3labs/mcp-go` ni el SDK oficial para mantener el `go.mod` mínimo y evitar riesgo de API inestable del SDK.

### Wiring en cmd/gateway

El endpoint `/mcp` no se monta en `cmd/gateway/main.go` en este slice — eso corresponde al sprint de integración final (Fase 4) donde se cablea el mux completo. El handler está listo para montarse con una línea: `mux.Handle("/mcp", mcpHandler)`.
