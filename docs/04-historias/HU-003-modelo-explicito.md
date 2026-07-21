---
id: HU-003
titulo: Forzar modelo explícito con política de fallback
epica: EP-001
prioridad: Must
complejidad: S
estado: lista
---

# Forzar modelo explícito con política de fallback

Como **desarrollador integrador**, quiero **enviar un parámetro `model` explícito y que la Gateway use ese modelo**, para **controlar exactamente qué modelo responde cuando lo necesito**.

Contexto: modo explícito del "LLM universal". Complementa el modo automático (HU-002). Encaja en la actividad 2 (enviar petición) aunque lo resuelva el Router.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — modelo disponible | Dado que un `model` explícito que existe y está sano en el Registry | Cuando llega la petición con ese `model` | Entonces la Gateway usa exactamente ese modelo y responde, sin aplicar scoring automático |
| 2 | Error — modelo inexistente | Dado que un `model` que no existe en el Registry | Cuando llega la petición | Entonces responde error 404 "modelo no encontrado" listando modelos válidos, sin caer a otro modelo silenciosamente |
| 3 | Edge — modelo caído con fallback permitido | Dado que un `model` existente pero no-sano y la política configurada permite fallback | Cuando llega la petición | Entonces la Gateway aplica la cadena de fallback de esa capacidad y anota en la respuesta/log que sustituyó el modelo pedido |
| 4 | Edge — modelo caído sin fallback | Dado que un `model` existente pero no-sano y la política prohíbe fallback | Cuando llega la petición | Entonces responde error 503 "modelo solicitado no disponible" sin sustituir |

## Checklist INVEST

- [x] Independent — se apoya en HU-001; entregable por separado
- [x] Negotiable — política de fallback configurable
- [x] Valuable — habilita control explícito del LLM universal
- [x] Estimable — pequeño
- [x] Small — 1-2 días
- [x] Testable — cada rama es un test

## Notas técnicas

La política (permitir/prohibir fallback en modo explícito) es configurable por request o por defecto global.
