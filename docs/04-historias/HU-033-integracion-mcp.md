---
id: HU-033
titulo: Integración MCP para multi-agentes
epica: EP-005
prioridad: Could
complejidad: M
estado: lista
---

# Integración MCP para multi-agentes

Como **arquitecto de agentes**, quiero **que el Gateway actúe como un servidor MCP (Model Context Protocol)**, para **que los agentes descubran e interactúen con herramientas y configuraciones del Gateway de forma estandarizada**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Descubrimiento MCP | Dado que un agente compatible con MCP se conecta al Gateway | Cuando solicita las capacidades disponibles | Entonces el Gateway responde con la lista de modelos, cuotas y herramientas soportadas |
| 2 | Error — Permisos MCP | Dado que el Agent intenta invocar una tool MCP sin permisos | Cuando el Gateway evalúa la llamada | Entonces bloquea la invocación retornando 403 Forbidden |
| 3 | Payload Malformado MCP | Dado que un cliente envía un handshake MCP con un esquema inválido | Cuando se recibe en el Gateway | Entonces el sistema retorna 400 Bad Request especificando el error de esquema |
| 4 | Incompatibilidad de Versiones | Dado un cliente con un protocolo MCP obsoleto | Cuando inicia el discovery | Entonces el Gateway notifica el error de versión (426 Upgrade Required) y cierra la conexión |

## Checklist INVEST

- [x] Independent — Endpoint o sub-protocolo independiente.
- [x] Negotiable — Soportar solo herramientas básicas vs full file-system access.
- [x] Valuable — Habilita Claude Code y multi-agentes para usar el Gateway como cerebro central.
- [x] Estimable — El protocolo MCP está tipado y documentado.
- [x] Small — Solo implementa el passthrough o router de recursos.
- [x] Testable — Conectar cliente oficial MCP y listar tools.

## Notas técnicas
- Asegurar alineación con NFRs de latencia y uso de caché si aplica.
