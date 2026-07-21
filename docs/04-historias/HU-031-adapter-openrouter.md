---
id: HU-031
titulo: Adapter para OpenRouter
epica: EP-008
prioridad: Should
complejidad: M
estado: lista
---

# Adapter para OpenRouter

Como **desarrollador de la plataforma**, quiero **un adapter para la API de OpenRouter**, para **poder consumir cualquier modelo de código abierto u open-weights ofrecido en su catálogo sin acoplar agentes**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — proxy transparente | Dado que un request de chat | Cuando el router escoge un modelo como `anthropic/claude-3-haiku` vía OpenRouter | Entonces el adapter inyecta los headers `HTTP-Referer` y `X-Title` requeridos por OpenRouter y encamina el tráfico |
| 2 | Error — modelo no disponible | Dado que openRouter reporta que el modelo upstream está saturado | Cuando el adapter intenta invocarlo | Entonces el adapter traduce el error upstream a un 503 estándar para que el Gateway inicie failover |
| 3 | Edge — Timeout OpenRouter | Dado que OpenRouter tarda demasiado en responder | Cuando el request supera el límite TTFT | Entonces el adapter devuelve error 504 Gateway Timeout |

## Checklist INVEST

- [x] Independent — Adaptador aislado.
- [x] Negotiable — Soporte para fallback interno de OpenRouter vs forzar fallback local del Gateway.
- [x] Valuable — Da acceso a cientos de modelos open source con una sola integración.
- [x] Estimable — API casi idéntica a OpenAI, muy predecible.
- [x] Small — Implementación rápida por compatibilidad de contrato.
- [x] Testable — Suite de contrato estándar OpenAI-compat.

## Notas técnicas
- Asegurar alineación con NFRs de latencia y uso de caché si aplica.
