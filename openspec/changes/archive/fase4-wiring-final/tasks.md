## Fase 4: Sprint de Wiring y Puesta en Producción

Este sprint final tiene como objetivo integrar todas las piezas modulares construidas en la Fase 3 dentro del ciclo de vida del servidor en `cmd/gateway/main.go`. No se requiere lógica de negocio nueva, sino un ensamble arquitectónico de los módulos.

### 1. Model Context Protocol (MCP)
- [x] 1.1 Importar el paquete `internal/api/mcp` en `main.go`.
- [x] 1.2 Reemplazar el handler stub de `POST /mcp` por la implementación real instanciada vía `mcp.NewHandler`.
- [x] 1.3 Inyectar en el Handler el token de administrador (`GATEWAY_ADMIN_TOKEN`) y el `registry.Registry` como `ModelSource`.

### 2. Capa de Identidad (Auth y RBAC)
- [x] 2.1 Verificar que `identityMiddleware` evalúa la cadena completa en orden: JWT Local (Sesiones) -> PostgreSQL (`userKeys`) -> Legacy `apiKeyStore`.
- [x] 2.2 Asegurar que el contexto devuelto (`auth.Identity` y `AdminContext`) fluye correctamente hacia `/metrics` y `/alerts`.

### 3. Persistencia Asíncrona (Quotas)
- [x] 3.1 Confirmar que el Worker de sincronización (`syncworker`) del `PostgresPersister` de quotas es invocado (o arrancado como goroutine) en caso de que `GATEWAY_QUOTA_POSTGRES_DSN` esté provisto.
- [x] 3.2 Verificar graceful shutdown: El persister debe hacer flush final cuando se captura el `SIGTERM` (junto al shutdown del HTTP Server).

### 4. Router y Failover Engine
- [x] 4.1 Validar que los adaptadores se envuelven correctamente en el `quota.Middleware` (que inyecta uso/consumo real).
- [x] 4.2 Revisar que el fallback hook `RetireOn429` está correctamente conectado entre el `Failover Engine` y el `Health Monitor` (para penalizaciones en routing).

### 5. Compatibilidad Universal 
- [x] 5.1 Enlazar el flujo de normalización de formatos (OpenAI / Anthropic) dentro del despachador central de peticiones del Processor o del middleware HTTP.
