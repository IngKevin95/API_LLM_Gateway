---
id: HU-024
titulo: Adapter para modelos locales (Ollama / vLLM / LM Studio)
epica: EP-002
prioridad: Must
complejidad: M
estado: lista
---

# HU-024: Adapter para modelos locales (Ollama / vLLM / LM Studio)

Como **desarrollador de la plataforma**, quiero **un adapter para servidores locales OpenAI-compatibles (Ollama, vLLM, LM Studio)**, para **degradar a modelos gratuitos y offline como último eslabón de la cadena de failover**.

Contexto: reutiliza la traducción OpenAI-compat (HU-020a). Es el destino de la degradación local en HU-004a.

## INVEST
- [x] Independent: asume que existe un stub/mock de enrutamiento (HU-001).
- [x] Negotiable: el detalle de la API local puede variar, pero debe ser OpenAI-compatible.
- [x] Valuable: permite failover a modelos gratuitos sin costo y offline.
- [x] Estimable: complejidad clara al ser compatible con OpenAI.
- [x] Small: un solo adapter.
- [x] Testable: se mockea el endpoint HTTP local con un httptest.Server (sin red real ni contenedor).

## Criterios de Aceptación (BDD)
| ID | Escenario | Dado (Given) | Cuando (When) | Entonces (Then) |
|---|---|---|---|---|
| 1 | Happy — petición exitosa | Un servidor local OpenAI-compatible responde 200 | La Gateway enruta una petición de chat al modelo local | El Adapter se comunica y devuelve la respuesta en el formato estándar del Gateway |
| 2 | Error — timeout local | El servidor local está colgado (no responde antes del TTFT/timeout dinámico) | Se enruta una petición a un modelo local | Se aplica el timeout y retorna el formato estandarizado de falla para que la Gateway inicie failover |
| 3 | Edge — respuesta no OpenAI-compatible | El servidor local responde 200 pero con un cuerpo que no respeta el esquema OpenAI (JSON mal formado o campos faltantes) | El adapter intenta normalizar la respuesta | El adapter falla con un error estandarizado sin crashear, y la Gateway lo trata como falla del proveedor |

> **OpenSpec change**: `ep-002-resiliencia-conectividad` (EP-002)
