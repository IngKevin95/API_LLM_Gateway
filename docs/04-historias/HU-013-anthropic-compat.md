---
id: HU-013
titulo: Endpoint Anthropic-compat para Free Claude Code
epica: EP-005
prioridad: Must
complejidad: M
estado: lista
---

# Endpoint Anthropic-compat para Free Claude Code

Como **desarrollador que usa Free Claude Code**, quiero **apuntar el cliente a la Gateway vía un endpoint compatible con la API Messages de Anthropic**, para **conservar la experiencia de Claude Code sin depender directamente de Anthropic**.

Contexto: contrato Anthropic-compat que habilita Free Claude Code. Actividad 2 del journey.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — messages | Dado que un cliente Anthropic apuntando a la Gateway | Cuando envía una petición al endpoint Messages-compat | Entonces la Gateway enruta por capacidad y responde en el formato de respuesta de Anthropic |
| 2 | Happy — streaming | Dado que una petición con streaming activado | Cuando se envía | Entonces transmite eventos incrementales en el formato de streaming de Anthropic |
| 3 | Error — contrato inválido | Dado que un payload que no cumple el esquema Messages | Cuando se envía | Entonces responde error en formato Anthropic con detalle, sin filtrar internals |
| 4 | Edge — tool use | Dado que una petición que declara herramientas (tool use) | Cuando se envía | Entonces la Gateway preserva los bloques `tool_use` y `tool_result` y los eventos de streaming asociados en el formato del protocolo Messages de Anthropic, sin alterar sus campos |

## Checklist INVEST

- [x] Independent — se apoya en routing; entregable junto al routing
- [x] Negotiable — alcance de compatibilidad acordable
- [x] Valuable — habilita Free Claude Code
- [x] Estimable — mapeo de contrato
- [x] Small — un sprint
- [x] Testable — requests Messages de referencia

## Notas técnicas

Cubrir el subconjunto de Messages que usa Claude Code (mensajes, streaming, tool use). Ver HU-016 para la config del cliente.
