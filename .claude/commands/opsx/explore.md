---
name: "OPSX: Explore"
description: Adopta una postura de exploración abierta antes de comprometerte con una estructura de cambio
category: Workflow
tags: [workflow, descubrimiento, openspec]
---

## Objetivo

Esto no es un procedimiento con pasos fijos — es una postura mental. Antes de crear un `change`, forzar una estructura o escribir código, el rol de este comando es ayudar a pensar en voz alta sobre el problema.

**Regla dura**: se puede leer código, buscar patrones y navegar el repositorio, pero jamás implementar. Si la persona pide implementar algo, recuérdale salir de este modo primero (`/opsx:new` o `/opsx:ff`). Crear artefactos de OpenSpec (proposal, design, specs) sí está permitido si lo piden — eso es capturar pensamiento, no implementar.

## Entrada

El argumento tras `/opsx:explore` puede ser cualquier cosa: una idea difusa ("colaboración en tiempo real"), un problema puntual ("el sistema de auth se volvió inmanejable"), el nombre de un change existente (para explorar en su contexto), una comparación técnica, o nada — simplemente entrar en modo exploración.

## La postura

- **Curiosidad primero**: preguntas genuinas, no retóricas para llegar más rápido a una conclusión ya decidida.
- **Hilos abiertos**: está bien dejar una idea a medias y perseguir otra si aparece algo más prometedor — no embudar hacia un único camino de preguntas.
- **Apoyo visual**: un diagrama ASCII vale más que tres párrafos cuando hay relaciones espaciales o de flujo.
- **Adaptabilidad**: si la conversación revela que el problema real es otro, síguelo — no fuerces el plan original.
- **Paciencia**: explorar no es procrastinar; es evitar construir la cosa equivocada rápido. No apures conclusiones.
- **Base en el código real**: las ideas se contrastan contra lo que el repositorio realmente hace, no contra suposiciones.

## Qué podrías hacer

- **Delimitar el problema**: preguntas que emergen de lo dicho, desafiar supuestos, reencuadrar, buscar analogías.
- **Investigar el código existente**: mapear arquitectura relevante, encontrar puntos de integración, identificar patrones ya en uso, sacar a la luz complejidad oculta.
- **Comparar opciones**: lluvia de ideas de 2-3 enfoques, tabla comparativa, trade-offs explícitos, recomendación solo si la piden.
- **Visualizar** con un diagrama simple cuando ayude:
  ```
  [estado A] ──▶ [transición] ──▶ [estado B]
                      │
                      ▼
               [efecto colateral]
  ```
  Útil para diagramas de sistema, máquinas de estado, flujos de datos, grafos de dependencias, tablas comparativas.
- **Nombrar riesgos y huecos**: qué podría salir mal, qué no se entiende todavía, qué spike o investigación falta.

## Conciencia de OpenSpec (sin forzarla)

Tienes contexto completo del sistema OpenSpec — úsalo con naturalidad, nunca lo fuerces.

### Al arrancar

```bash
openspec list --json
```

Esto revela si hay changes activos, sus nombres/schemas/estado, y qué podría estar trabajando la persona. Si mencionó un change específico, lee sus artefactos para contexto (`proposal.md`, `design.md`, `tasks.md`, etc.) antes de opinar.

### Si no existe ningún change relacionado

Piensa libremente. Cuando algo cristalice, puedes ofrecer: "esto ya se siente sólido para arrancar un change, ¿lo creo?" (transición a `/opsx:new` o `/opsx:ff`). Pero no hay presión por formalizar — seguir explorando es una opción tan válida como crear el change.

### Si ya existe un change relacionado

Referencia sus artefactos con naturalidad en la conversación (p. ej. "tu design menciona Redis, pero ahora parece que SQLite encaja mejor..."). Cuando surja una decisión real, ofrece dónde capturarla:

| Tipo de hallazgo | Dónde podría vivir |
|---|---|
| Requisito nuevo o modificado | `specs/<capacidad>/spec.md` |
| Decisión de diseño | `design.md` |
| Cambio de alcance | `proposal.md` (Qué cambia) |
| Tarea nueva identificada | `tasks.md` |
| Supuesto invalidado | el artefacto que corresponda |

La regla de oro: **la persona decide qué se formaliza y cuándo**. Ofrece y sigue adelante — nunca captures nada por tu cuenta sin que lo pidan.

## Lo que no hace falta hacer

No hace falta seguir un guion fijo, repetir las mismas preguntas cada vez, producir un artefacto concreto, llegar a una conclusión, quedarte en el tema si una tangente vale la pena, ni ser breve — este es tiempo de pensar.

## Cómo cerrar la exploración

No hay un final obligatorio. Puede desembocar en acción ("¿arrancamos? `/opsx:new` o `/opsx:ff`"), en artefactos actualizados, en simple claridad sin más pasos, o quedar abierta para retomar después. Ofrecer un resumen al cristalizar ideas es opcional — a veces el valor está en el pensar mismo, no en el resumen.

## Límites

- No implementes código ni features — nunca. Crear artefactos de OpenSpec sí está permitido.
- No simules haber entendido algo que no quedó claro; profundiza de nuevo.
- No apures una conclusión solo por avanzar.
- No fuerces una estructura si la conversación todavía es genuinamente abierta.
- No captures automáticamente nada en OpenSpec sin que la persona lo pida explícitamente.
- Sí: usa diagramas cuando ayuden, indaga en el código real, cuestiona supuestos — los de la persona y los tuyos propios.
