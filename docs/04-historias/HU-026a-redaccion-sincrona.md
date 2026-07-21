---
id: HU-026a
titulo: Redacción Síncrona de Secretos
epica: EP-004B
prioridad: Must
complejidad: S
estado: lista
---

# Redacción Síncrona de Secretos

Como **oficial de seguridad**, quiero **que ningún secreto llegue al modelo mediante redacción en memoria (< 10ms)**, para **prevenir fugas de API keys en prompts**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy - Payload sin secretos | Dado que el texto no contiene PII | Cuando pasa por el escaneo síncrono | Entonces el payload es aprobado |
| 2 | Happy - Redacción | Dado que hay un email | Cuando se escanea | Entonces se enmascara |
| 3 | Error - Timeout Regex | Dado que un regex toma > 50ms | Cuando se evalúa | Entonces devuelve un error HTTP 500 |
| 4 | Edge - Base64 excluido | Dado un bloque masivo Base64 | Cuando se escanea síncronamente | Entonces se omite para evitar timeouts delegando al escaner asíncrono |

## Checklist INVEST

- [x] Independent — Motor DLP que se inserta como pipe antes del router.
- [x] Negotiable — Patrones regex iniciales (tarjetas, SSN) vs ML NLP (futuro).
- [x] Valuable — Evita fugas de datos confidenciales al proveedor externo (OpenAI/Anthropic).
- [x] Estimable — Lógica de regex y reemplazo de strings tiene complejidad predecible O(N).
- [x] Small — Limitarse a regex básicos mantiene el scope en < 1 sprint.
- [x] Testable — Inyectar PII sintética y asertar que sale censurada con ***.

## Notas técnicas
- Asegurar alineación con NFRs de latencia y uso de caché si aplica. La restricción de `< 10ms` se evaluará en pruebas de carga (NFR tests) para evitar flakiness en CI funcional.
