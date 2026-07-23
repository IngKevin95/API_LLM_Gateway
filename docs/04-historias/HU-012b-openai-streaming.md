---
id: HU-012b
titulo: Streaming SSE compatible OpenAI en el endpoint de chat
epica: EP-005
prioridad: Must
complejidad: S
estado: lista
---

# Streaming SSE compatible OpenAI en el endpoint de chat

Como **aplicación integradora**, quiero **recibir la respuesta de `/v1/chat/completions` en streaming (SSE) cuando pido `stream: true`**, para **mostrar tokens incrementalmente como con la API de OpenAI**.

Contexto: rebanada vertical de EP-005 (split de la antigua HU-012). Depende del chat base HU-012a. Actividad 2 del journey.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — streaming | Dado que una petición de chat con `stream: true` | Cuando se envía | Entonces la Gateway transmite tokens incrementalmente en formato SSE compatible OpenAI y cierra con el evento de fin |
| 2 | Error — corte del proveedor a mitad | Dado que un stream en curso cuyo proveedor se corta | Cuando ocurre el corte | Entonces la Gateway emite un evento de error SSE y cierra la conexión emitiendo un evento de error SSE y cerrando el socket TCP, sin colgarla |
| 3 | Edge — cliente aborta | Dado que un cliente que cierra la conexión antes de terminar | Cuando el cliente corta | Entonces la Gateway detiene la generación y libera el recurso del proveedor |
| 4 | Edge — failover en streaming | Dado que el proveedor primario falla antes del primer token | Cuando se inicia el stream | Entonces aplica fallback y comienza el streaming por el siguiente proveedor sin exponer el fallo al cliente |
| 5 | Edge — Stream Idle Timeout | Dado un stream activo | Cuando pasan > 5s sin tokens (TBT) | Entonces se cierra la conexión unilateralmente para prevenir sockets zombies |

## Checklist INVEST

- [x] Independent — el desarrollo puede realizarse con mocks del endpoint base, no requiere HU-012a completada previamente.
- [x] Negotiable — detalle de manejo de corte abierto
- [x] Valuable — experiencia de tokens en vivo
- [x] Estimable — capa de streaming acotada
- [x] Small — 1-3 días
- [x] Testable — stream simulado y cortes forzados

## Notas técnicas

Formato de eventos SSE idéntico al de OpenAI. Liberar recursos al abortar. Failover solo antes del primer token entregado.
