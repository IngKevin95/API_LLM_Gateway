---
id: HU-042
titulo: Routing automático por capability (modelo implícito)
epica: EP-010
prioridad: Must
complejidad: M
estado: draft
---

# Routing automático por capability (modelo implícito)

Como **desarrollador integrando el Gateway en una herramienta**, quiero **omitir el parámetro `model` y usar `model: "router:coding"` o capability implícita**, para **que el Gateway elija automáticamente el mejor modelo sin cambiar mi cliente**.

Contexto: Habilita el modo "desacoplado" donde agentes no conocen modelos concretos, solo capacidades. Actividad 1 de EP-010.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — routing por prefijo `router:` | Dado que un cliente OpenAI-compat enva `model: "router:coding"` | Cuando la petición llega al Handler de chat | Entonces el Gateway extrae "coding", llama Router.Resolve("coding", tokens), ordena por score y usa el primero de la cadena de fallback |
| 2 | Happy — auto-routing si `auto_route_enabled=true` | Dado que config.yaml tiene `gateway.auto_route_enabled: true` y el cliente NO envía `model` | Cuando la petición llega sin field `model` | Entonces el Gateway usa `gateway.default_capability` (ej: "chat") para llamar Router.Resolve, sin error 400 |
| 3 | Happy — streaming con routing automático | Dado que un cliente envía `stream: true` y `model: "router:reasoning"` | Cuando se abre el stream | Entonces la respuesta usa el primer modelo de la cadena de reasoning, con TTFT relajado (30s) y tokens se emiten correctamente |
| 4 | Error — capability desconocida | Dado que cliente envía `model: "router:unknown"` | Cuando se procesa | Entonces el Gateway responde 400 Bad Request con mensaje "unknown capability: unknown" sin procesar la cadena de fallback |
| 5 | Edge — sin modelos disponibles para capability | Dado que capability "reasoning" solo tiene Anthropic, pero Anthropic está down | Cuando se resuelve capability | Entonces el Gateway responde 503 Service Unavailable con "no models available for reasoning" |

## Checklist INVEST

- [x] Independent — se apoya en Router existente (EP-001), Handlers (EP-005); entregable en un sprint
- [x] Negotiable — alcance: nuevos parámetros en config.yaml + lógica de detección en Handlers
- [x] Valuable — habilita desacoplamiento completo agente-modelo
- [x] Estimable — parsing + llamada Router.Resolve (~6 horas)
- [x] Small — un sprint
- [x] Testable — requests con `router:*` + test de fallback

## Notas técnicas

- Router.Resolve ya está implementado (EP-001); aquí solo se agranda el Handler para detectar prefijo `router:` y extraer capability
- Si `auto_route_enabled=false` (default), comportamiento actual: solo acepta modelo explícito
- Si `auto_route_enabled=true`, parámetro `model` es **opcional** — si falta, usa `default_capability` del config
- Config esperada:
  ```yaml
  gateway:
    auto_route_enabled: true
    default_capability: chat
  ```
- Tests de conformidad: request sin model, request con "router:coding", request con capability inexistente
