---
id: HU-020c
titulo: Adapter OpenAI — embeddings
epica: EP-002
prioridad: Must
complejidad: S
estado: lista
---

# Adapter OpenAI — embeddings

Como **desarrollador de la plataforma**, quiero **que el adapter de OpenAI soporte la capacidad `embedding`**, para **vectorizar texto a través de la Gateway sin acoplarme al proveedor**.

Contexto: depende de HU-020a (chat base). Aísla la ruta de embeddings.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — embeddings | Dado que un payload para vectorizar texto | Cuando el router selecciona un modelo text-embedding de OpenAI | Entonces el adapter redirige a `/v1/embeddings` y retorna los vectores normalizados |
| 2 | Edge — lote grande | Dado que un payload con un lote grande de textos a vectorizar | Cuando el adapter procesa la petición | Entonces respeta el límite de batch del proveedor (particionando o rechazando con error claro) sin truncar silenciosamente |
| 3 | Error — modelo no soportado | Dado que se solicita un modelo de embedding inexistente en OpenAI | Cuando el adapter intenta la llamada | Entonces retorna falla estandarizada para que la Gateway aplique fallback |

## Checklist INVEST

- [x] Independent — depende de HU-020a entregable
- [x] Negotiable — estrategia de loteo abierta
- [x] Valuable — habilita embeddings sin acoplarse a proveedor
- [x] Estimable — endpoint acotado
- [x] Small — solo embeddings
- [x] Testable — se mockea `/v1/embeddings`

## Notas técnicas

Comparte cliente HTTP y manejo de errores con HU-020a.

> **OpenSpec change**: `ep-002-resiliencia-conectividad` (EP-002)
