---
id: HU-032
titulo: Cache Semántica (Fase 2) con Vector Search local
epica: EP-007
prioridad: Could
complejidad: M
estado: lista
---

# Cache Semántica (Fase 2) con Vector Search local

Como **administrador de la plataforma**, quiero **habilitar una caché semántica (Vector Search) para respuestas de LLMs**, para **reducir drásticamente los costos y la latencia en peticiones idénticas o semánticamente muy similares**.

Contexto: La historia incluye la integración de una librería ligera de embeddings en memoria (ej. librería zero-deps en el servicio) para calcular la similitud localmente en < 50ms antes de ir a base de datos/Vector DB, evitando así depender de un worker complejo externo.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Hit semántico exacto | Dado que un mock del servicio de embeddings devuelve un valor configurado >0.98 | Cuando el cliente envía la petición | Entonces el Gateway retorna la respuesta desde la caché sin cobrar tokens al proveedor |
| 2 | Edge — Cache Miss (Similitud Baja) | Dado que un mock del servicio de embeddings devuelve un valor configurado <0.98 | Cuando el Gateway busca en la Cache Semántica | Entonces procesa la petición enviándola al LLM externo de forma normal |
| 3 | Timeout de Vector DB | Dado que la base de datos vectorial no responde a tiempo | Cuando ocurre un timeout | Entonces el Gateway interrumpe la búsqueda semántica y ejecuta el LLM de forma normal (Fail-fast asíncrono) |
| 4 | Prevención de Falsos Positivos | Dado que el prompt del usuario es muy corto (ej. 'hola') | Cuando se busca en el caché semántico | Entonces el Gateway omite la validación semántica para evitar coincidencias erróneas por alta densidad dimensional |

## Checklist INVEST

- [x] Independent — Middleware puro sobre el request, no requiere un worker externo.
- [x] Negotiable — La técnica de embedding local y el umbral de similitud son iterables.
- [x] Valuable — Ahorra >50% de costo en entornos de agentes repetitivos.
- [x] Estimable — Lógica vector-search en BD es estándar.
- [x] Small — Limitado a usar una librería de embedding empotrada muy ligera para acotar tamaño.
- [x] Testable — Lanzar 2 requests idénticos; el segundo debe tener 0ms de red hacia el LLM primario.

## Notas técnicas
- Asegurar alineación con NFRs de latencia y uso de caché si aplica. Los asserts de <50ms deben probarse en suites de estrés y NFR para evitar flaky tests.
