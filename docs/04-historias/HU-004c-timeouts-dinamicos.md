---
id: HU-004c
titulo: Timeouts dinámicos por capacidad y Stream Idle Timeout
epica: EP-002
prioridad: Must
complejidad: M
estado: lista
---

# Timeouts dinámicos por capacidad y Stream Idle Timeout

Como **operador**, quiero **umbrales de TTFT adaptativos por capacidad y un corte por inactividad mid-stream**, para **liberar conexiones colgadas sin penalizar a los modelos de razonamiento que piensan antes de emitir tokens**.

Contexto: complementa el failover básico (HU-004a). Separa la política de timeouts del mecanismo de reintento.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Edge — TTFT excedido (estándar) | Dado que el proveedor primario tarda más del umbral estricto pre-stream (ej. 2.0s dinámico) para una capacidad estándar (chat/código) | Cuando ocurre el timeout pre-stream | Entonces el Gateway aborta la conexión primaria y ejecuta el failover silenciosamente |
| 2 | Happy — Timeout dinámico reasoning | Dado que se envía una petición con la capacidad `reasoning` | Cuando el Gateway monitorea el TTFT | Entonces el applies configured timeout_reasoning (ej. < 30s, configurable) para permitir el pensamiento prolongado antes del primer token, sin disparar failover |
| 3 | Edge — Stream Idle Timeout | Dado que el proveedor deja de emitir tokens por más del Stream Idle Timeout (configurable por modelo en YAML, ej. 5s) | Cuando el Gateway monitorea el flujo TBT | Entonces corta el socket unilateralmente y penaliza el score del proveedor |

## Checklist INVEST

- [x] Independent — depende de HU-004a (failover básico) entregable
- [x] Negotiable — umbrales concretos configurables por YAML
- [x] Valuable — evita cortar modelos de reasoning y libera streams colgados
- [x] Estimable — acotado a la lógica de timeouts por capacidad
- [x] Small — un sprint
- [x] Testable — se simulan latencias de TTFT y pausas mid-stream

## Notas técnicas

El TTFT externo (del proveedor) no incluye el overhead del Guardián de Prompts. Los valores viven en el YAML por modelo/capacidad.
