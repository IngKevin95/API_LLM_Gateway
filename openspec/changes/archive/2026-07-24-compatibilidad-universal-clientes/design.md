## Context

EP-005 (API universal compatible) estableció endpoints para 2 clientes: OpenWebUI (OpenAI-compat) y Free Claude Code (Anthropic). EP-010 amplía soporte a 8 herramientas diferentes (OpenCode, Claude Code, OpenHands, OpenClaw, CrewAI, UI-TARS) sin cambios en el cliente.

Desafío: cada herramienta espera un formato distinto (OpenAI, Anthropic, Responses API) y parámetros variados. Hoy requiere un adapter manual por nueva herramienta; EP-010 lo automatiza con middleware de normalización y routing por capability.

## Goals / Non-Goals

**Goals:**
- 8 clientes (OpenWebUI, OpenCode, Claude Code, Free Claude Code, OpenHands, OpenClaw, CrewAI, UI-TARS) funcionan sin cambios de código cliente
- Routing automático por capability sin `model` explícito
- Parámetros completos (temperature, top_p, seed, thinking, etc.) traducidos sin error
- Formato auto-detect: entrada OpenAI/Anthropic/Responses enruta correctamente sin endpoint distinto
- Backward compatible: requests existentes siguen funcionando

**Non-Goals:**
- Cambios en adapters backend (OpenAI, Anthropic, Google, OpenRouter, etc.) — se amplifican pero no se reescriben
- Cambios en Registry/Router core (EP-001)
- UI de configuración — solo docs y env vars
- Validación de parámetros a nivel negocio (eso es role del cliente o adapter)

## Decisions

### Decision 1: Pipeline middleware para normalización

**Choice**: Agregar middleware pre-router que normaliza entrada a un formato interno antes de que Router resuelva capability y Adapter traduzca al backend.

```
Cliente (OpenAI/Anthropic/Responses)
    ↓
[Middleware: Format Auto-Detect]
    ↓
Formato interno normalizado
    ↓
Router (resuelve capability → modelo)
    ↓
Adapter (traduce a backend nativo)
    ↓
Proveedor (OpenAI/Anthropic/Google/etc)
```

**Why**: Desacopla detección de formato de lógica de routing. Múltiples formatos de entrada → una ruta de procesamiento. Sin esto, habría N endpoints o N ramas de lógica condicional.

**Alternatives**: 
- (Rejected) Endpoint distinto por formato (/v1/chat vs /v1/messages vs /responses). Expone complejidad al cliente.
- (Rejected) Lógica de detección en cada adapter. Duplica código y dificulta mantenimiento.

### Decision 2: Routing automático via "router:capability" y ausencia de model

**Choice**: Acepta `model: "router:coding"` o `model` ausente. En ambos casos, Router.Resolve elige automáticamente el mejor modelo de la cadena de la capability.

**Why**: Habilita el principio central de EP-001 (agentes no nombran modelos, consumen capacidades). Un cliente puede pedir `capability:coding` sin saber qué modelos hay en la registry.

**Backward compat**: Requests con `model: "gpt-4"` explícito siguen funcionando igual.

### Decision 3: Formato interno normalizado

**Choice**: Un struct interno `NormalizedRequest` que vive solo dentro del Gateway, nunca se expone externamente.

```go
type NormalizedRequest struct {
  Capability string           // "chat", "coding", "reasoning", etc.
  Messages   []Message         // formato interno
  Parameters map[string]interface{} // temp, top_p, seed, thinking, etc.
  Streaming  bool
}
```

**Why**: Abstraer diferencias OpenAI/Anthropic/Responses. Adapters traducen NormalizedRequest → ProviderRequest nativa.

### Decision 4: Traducción de parámetros en Adapter

**Choice**: Cada adapter (OpenAI, Anthropic, Google, etc.) recibe NormalizedRequest y traduce parámetros a su schema nativo. P. ej. Anthropic adapter mapea `temperature` (0-2) a `temperature` (0-1) con clipping.

**Why**: Centraliza lógica de traducción. Si Anthropic cambia su schema, solo se toca ese adapter.

### Decision 5: Archivo de configuración expandido

**Choice**: config.yaml registra no solo `model`, sino también qué parámetros cada modelo soporta (opcional, defaults sensatos).

```yaml
models:
  - name: gpt-4o
    provider: openai
    capability: coding
    supports:
      - temperature
      - top_p
      - seed
```

**Why**: Futuro-proof. Si un cliente consulta `/v1/models`, puede saber qué parámetros son válidos.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| **Traducción de parámetros falla silenciosamente** | Cada adapter loguea qué parámetros mapeó y cuáles ignoró. Si un parámetro es esencial (p. ej. `thinking` en Claude), el adapter rechaza la request si no está disponible. |
| **Rendimiento del middleware** | Format auto-detect es O(1) (inspecciona 3 campos). Normalización es copia superficial. P95 < 1ms overhead. Medible vía /metrics. |
| **Fragmentación de formatos de entrada** | Hoy soportamos 3 (OpenAI, Anthropic, Responses). Si llega un cuarto, solo hay que agregar un detector más al middleware, sin tocar adapters. |
| **Compatibilidad hacia atrás se rompe** | No — todo es aditivo. Requests viejas sin cambios siguen siendo válidas. |

## Migration Plan

**Fase 1 (esta épica)**: Middleware + normalización + 7 nuevas capacidades. Requests existentes sin `model` o con `router:*` se enrutan automáticamente.

**Fase 2 (futura)**: Dashboard para visualizar rutas y parámetros por herramienta.

**Rollback**: Feature flag `FORMAT_AUTO_DETECT=false` desactiva el middleware y vuelve al comportamiento EP-005. No requiere redeploy, solo env var.

## Open Questions

1. ¿Todos los 8 clientes del scope soportan streaming, o hay alguno que solo espera unary? (Afecta a design de bufferizado en middleware.)
2. ¿Está permitido ignorar parámetros desconocidos (lenient) o rechazar con error (strict)? (Afecta a tolerancia.)
3. ¿Se registra en audit log qué formato de entrada se detectó? (Seguridad / debugging.)
