---
id: HU-030
titulo: Adapter para Google (Gemini)
epica: EP-008
prioridad: Should
complejidad: M
estado: lista
---

# Adapter para Google (Gemini)

Como **desarrollador de la plataforma**, quiero **un adapter para la API de Google (Gemini)**, para **soportar capacidades de visión y reasoning avanzadas nativas de Google**.

Contexto: La API de Google no es compatible nativamente con OpenAI, requiere mapeo profundo de payload.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — chat con Visión | Dado que un payload con imagen en Base64 | Cuando el router elige Gemini | Entonces el adapter formatea la imagen a `inlineData` de Google y retorna texto |
| 2 | Error — payload muy grande | Dado que la imagen excede límites de Google | Cuando el adapter falla pre-vuelo | Entonces falla con 400 limpio sin penalizar cuota |
| 3 | Edge — Límite de concurrencia nativo (429) | Dado que se satura la cuota concurrente de Google | Cuando Gemini retorna 429 | Entonces se inicia el failover hacia el siguiente proveedor |
| 4 | Edge — Mapeo de System Prompt | Dado que el request (OpenAI) tiene un mensaje 'system' en el array | Cuando el adapter traduce a Gemini | Entonces extrae el 'system' y lo inyecta en el campo 'systemInstruction' de Gemini |

## Checklist INVEST

- [x] Independent — Adaptador aislado; implementa una interfaz Go ya existente.
- [x] Negotiable — Streaming inicial puede omitirse si la API de Gemini es compleja.
- [x] Valuable — Abre acceso a Gemini 1.5 Pro y su contexto masivo de 1M tokens.
- [x] Estimable — Mapear JSON A a JSON B.
- [x] Small — Un solo proveedor.
- [x] Testable — Suite de contrato con respuestas mockeadas de Google API.

## Notas técnicas
- Asegurar alineación con NFRs de latencia y uso de caché si aplica.
