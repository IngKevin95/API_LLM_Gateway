---
id: HU-012a
titulo: Endpoint OpenAI-compat de chat (sin streaming) con enrutamiento
epica: EP-005
prioridad: Must
complejidad: M
estado: lista
---

# Endpoint OpenAI-compat de chat (sin streaming) con enrutamiento

Como **aplicación integradora**, quiero **llamar a `/v1/chat/completions` con el contrato de OpenAI (respuesta completa, sin streaming)**, para **usar la Gateway como LLM universal sin cambiar mi cliente**.

Contexto: rebanada vertical de EP-005 (split de la antigua HU-012). Cubre el chat básico. El streaming es HU-012b y embeddings HU-012c. Actividad 2 del journey.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — chat sin modelo | Dado que un cliente OpenAI apuntando a la Gateway, petición a `/v1/chat/completions` sin forzar modelo | Cuando se envía la petición | Entonces la Gateway enruta por capacidad `chat` y responde en el formato de respuesta de OpenAI |
| 2 | Happy — chat con modelo | Dado que una petición que fija `model` explícito válido | Cuando se envía | Entonces usa ese modelo y responde en formato OpenAI |
| 3 | Error — payload malformado | Dado que una petición que no cumple el esquema OpenAI de chat | Cuando se envía | Entonces responde error en formato OpenAI (400) con detalle del campo |
| 4 | Edge — `/v1/models` | Dado que un cliente que consulta el catálogo | Cuando pide `/v1/models` | Entonces devuelve los modelos habilitados en el formato de listado de OpenAI |

## Checklist INVEST

- [x] Independent — depende de HU-002 (routing) entregable
- [x] Negotiable — implementación del contrato abierta
- [x] Valuable — habilita el LLM universal para chat
- [x] Estimable — un endpoint acotado
- [x] Small — un sprint (solo chat no-streaming + models)
- [x] Testable — golden requests OpenAI

## Notas técnicas

Mapear errores internos a códigos OpenAI. Paridad de campos comunes de chat.
