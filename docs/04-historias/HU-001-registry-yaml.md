---
id: HU-001
titulo: Cargar providers, models y routing desde YAML (Registry)
epica: EP-001
prioridad: Must
complejidad: M
estado: lista
---

# Cargar providers, models y routing desde YAML (Registry)

Como **operador de la plataforma**, quiero **declarar providers, modelos y reglas de routing en un archivo YAML que la Gateway carga al iniciar**, para **agregar o cambiar proveedores sin tocar código ni recompilar**.

Contexto: el Registry es la fuente declarativa que alimenta al Router. Encaja en la actividad 3 del journey (resolver modelo), pero es prerequisito de todo el Gateway.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — carga válida | Dado que un `config.yaml` con `providers`, `models` (con atributos quality/coding/reasoning/speed/vision/cost/latency) y `routing` por capacidad | Cuando la Gateway inicia | Entonces carga el catálogo en memoria y expone los modelos habilitados<br>And imprime el conteo final en el stdout |
| 2 | Error — YAML inválido | Dado que un `config.yaml` con sintaxis inválida o un campo obligatorio faltante | Cuando la Gateway inicia | Entonces falla el arranque con un error que nombra el archivo, la línea y el campo problemático<br>And no arranca en estado parcial |
| 3 | Edge — secreto embebido | Dado que un `config.yaml` que pone una API key literal en vez de `${VAR}` | Cuando la Gateway carga el archivo | Entonces rechaza la carga (o la clave) y exige referencia a variable de entorno<br>And registra la violación sin imprimir el valor |
| 4 | Edge — capacidad sin modelos | Dado que una capacidad en `routing` que no lista ningún modelo habilitado | Cuando la Gateway inicia | Entonces marca esa capacidad como no disponible y escribe un WARN level en el log<br>And no aborta el resto de la carga |
| 5 | Happy — parámetros de red físicos | Dado un YAML con `max_in_flight` y `stream_idle_timeout` | Cuando el Registry lo carga | Entonces expone estos valores al Failover Engine y Adapters |

## Checklist INVEST

- [x] Independent — se entrega sin esperar otra historia (base del Gateway)
- [x] Negotiable — formato interno del parser abierto
- [x] Valuable — habilita config declarativa, objetivo del PRD
- [x] Estimable — alcance acotado a parseo y validación
- [x] Small — cabe en un sprint
- [x] Testable — cada AC es un test de carga

## Notas técnicas

Validar contra un esquema; nunca loguear valores de secretos; resolver `${VAR}` desde entorno/secret manager.
