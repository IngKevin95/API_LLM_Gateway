---
id: HU-027
titulo: Guardián de Prompts (Seguridad y Jailbreak)
epica: EP-004A
prioridad: Should
complejidad: M
estado: lista
---

# Guardián de Prompts (Seguridad y Jailbreak)

Como **desarrollador integrador**, quiero **un guardián de prompts opt-in (activable por header o parámetro)**, para **optimizar mis prompts inyectando reglas semánticas y de formato automáticamente antes de que lleguen al modelo final**.

## Notas Técnicas
- El guardián muta el último mensaje del array con rol "user", envolviendo el texto original en el template de optimización.

Contexto: La optimización debe ocurrir dinámicamente inyectando instrucciones de sistema comprobadas (ej. "piensa paso a paso", "actúa como experto") o reescribiendo la solicitud si el usuario activa el modo `optimizado: true`.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy - Prompts optimizados | Dado que el guardián de prompts está habilitado | Cuando se recibe un prompt | Entonces el último mensaje con rol `user` queda envuelto en el template de optimización definido (instrucciones de sistema añadidas) y el texto original se preserva íntegro dentro del wrapper |
| 2 | Error - Prompt inválido | Dado un prompt malformado | Cuando el guardián intenta reestructurarlo | Entonces bypasses optimization without exception y permite pasar el prompt original |
| 3 | Edge - Sin alteración de tool calling | Dado un prompt con llamadas a funciones definidas | Cuando se aplica la optimización semántica | Entonces la sintaxis de tool calling se mantiene intacta |
| 4 | Edge - Overhead excesivo | Dado que la optimización excede 100ms | Cuando el middleware evalúa el tiempo | Entonces aborta la optimización y envía el prompt original |
| 5 | Edge - Streaming Petición | Dado que el payload original especifica `stream: true` | Cuando el Guardián de Prompts procesa la optimización | Entonces no altera la respuesta token a token |

## Checklist INVEST

- [x] Independent — Servicio opt-in; si falla, se hace bypass pasivo.
- [x] Negotiable — Optimización básica (borrar whitespaces) vs Jailbreak detection.
- [x] Valuable — Reduce tokens consumidos y bloquea inyecciones maliciosas.
- [x] Estimable — Basado en heurísticas conocidas y estáticas.
- [x] Small — Limitar a validaciones sintácticas en la V1.
- [x] Testable — Enviar prompts maliciosos y medir si los bloquea con 400 Bad Request.

## Notas técnicas
- Asegurar alineación con NFRs de latencia y uso de caché si aplica.
