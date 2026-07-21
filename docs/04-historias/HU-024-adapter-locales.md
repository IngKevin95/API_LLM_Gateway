---
id: HU-024
titulo: Adapter para modelos locales (Ollama / vLLM / LM Studio)
epica: EP-002
prioridad: Must
complejidad: M
estado: lista
---

# HU-024: Adapter para modelos locales (Ollama / vLLM / LM Studio)

## INVEST
- [x] Independent: asume que existe un stub/mock de enrutamiento (HU-001).
- [x] Negotiable: el detalle de la API local puede variar, pero debe ser OpenAI-compatible.
- [x] Valuable: permite failover a modelos gratuitos sin costo y offline.
- [x] Estimable: complejidad clara al ser compatible con OpenAI.
- [x] Small: un solo adapter.
- [x] Testable: se puede probar con un contenedor de Ollama.

## Criterios de Aceptación (BDD)
| ID | Escenario | Dado (Given) | Cuando (When) | Entonces (Then) |
|---|---|---|---|---|
| 1 | Petición exitosa | Un servidor Ollama está corriendo | La Gateway enruta una petición de chat a Ollama | El Adapter se comunica y devuelve la respuesta en el formato estándar del Gateway |
| 2 | Timeout local | El servidor local está colgado | Se enruta una petición a un modelo local | Se aplica el TTFT o timeout dinámico y falla si se supera |
