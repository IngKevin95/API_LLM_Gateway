## ADDED Requirements

### Requirement: Servidor MCP Integrado
El sistema SHALL implementar soporte pasivo como Servidor bajo el Model Context Protocol (MCP) para habilitar el descubrimiento de recursos y herramientas por parte de arquitecturas multi-agente.

#### Scenario: Descubrimiento de capacidades
- **WHEN** un cliente MCP realiza un handshake y consulta capacidades
- **THEN** la Gateway responde enumerando las herramientas (rutas, endpoints) e información de contexto soportadas.

### Requirement: Seguridad y Permisos MCP
El servidor MCP SHALL estar protegido o exigir los tokens de autenticación internos para cualquier tool MCP invocada, rechazando accesos anónimos.

#### Scenario: Invocación denegada
- **WHEN** un cliente sin credencial intenta invocar MCP
- **THEN** se deniega el acceso (401/403).
