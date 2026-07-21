---
id: HU-012c
titulo: Endpoint OpenAI-compat de embeddings con enrutamiento
epica: EP-005
prioridad: Must
complejidad: S
estado: lista
---

# Endpoint OpenAI-compat de embeddings con enrutamiento

Como **aplicación integradora**, quiero **llamar a `/v1/embeddings` con el contrato de OpenAI**, para **obtener vectores desde la Gateway sin acoplarme a un proveedor de embeddings concreto**.

Contexto: rebanada vertical de EP-005 (split de la antigua HU-012). Endpoint independiente del chat. Actividad 2 del journey.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — embeddings | Dado que una petición a `/v1/embeddings` con uno o varios textos | Cuando se envía | Entonces responde vectores en el formato OpenAI, enrutando a un modelo de capacidad `embedding` |
| 2 | Error — sin modelo de embedding | Dado que ningún modelo de capacidad `embedding` habilitado/sano | Cuando se envía la petición | Entonces responde 503 "sin proveedor de embeddings disponible" en formato OpenAI |
| 3 | Error — payload malformado | Dado que una petición sin campo de entrada válido | Cuando se envía | Entonces responde error 400 en formato OpenAI con detalle |
| 4 | Edge — lote grande | Dado que una petición con muchos textos que excede el límite del proveedor | Cuando se envía | Entonces la Gateway lotea o rechaza con un error claro, sin truncar silenciosamente |

## Checklist INVEST

- [x] Independent — depende de HU-002 (routing); paralelo a HU-012a
- [x] Negotiable — estrategia de loteo abierta
- [x] Valuable — habilita embeddings vía LLM universal
- [x] Estimable — un endpoint acotado
- [x] Small — 1-3 días
- [x] Testable — golden requests de embeddings

## Notas técnicas

Formato de respuesta de embeddings de OpenAI. Definir política de loteo/límite por proveedor.
