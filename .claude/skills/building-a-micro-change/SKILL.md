---
name: building-a-micro-change
description: Use for genuine maintenance that is NOT new product capability — a typo, a copy/string tweak, a dependency version bump within the stack allowlist, an infra/config/docs change, or a small bug fix of a few lines that adds no new capability. Lightweight lane — branch fix/*|chore/* → change → regression test only if behavior changes → PR — WITHOUT opening active_slice, an epic (EP-XXX), or an OpenSpec change. HARD LIMITS: escalate to building-a-slice (a full epic) if the change adds a new dependency, creates a new public API/endpoint, or changes domain logic or the data model/invariants. The epic stays the unit for product construction; this lane is out-of-band maintenance only.
---

# Carril ligero para micro-changes

No todo lo que toca el repo es "construir producto". Un typo, un bump de dependencia permitida, un
ajuste de copy — forzarlos por las ocho fases de `building-a-slice` es ceremonia que no compra nada.
Este carril existe para ese mantenimiento genuino, sin diluir los guardarraíles deterministas que ya
protegen al repo.

> Si algo aquí contradice `METODOLOGIA.md`, gana la metodología — siempre.

## Paso 0: el decision gate (obligatorio, sin excepciones)

Antes de tocar una línea, confirma que las tres condiciones se cumplen **a la vez**:

1. El cambio **no añade capacidad de producto nueva** — corrige, ajusta o mantiene algo que ya
   existía.
2. El **alcance es acotado**: pocas líneas, una sola preocupación, no toca varios módulos a la vez.
3. El proyecto ya está en **fase `active`** (scaffold runnable confirmado). Este carril no sirve para
   arrancar un proyecto — eso lo cubre la Fase 0 de `building-a-slice`.

Ejemplos que sí califican: corregir un typo o texto visible, ajustar un valor de configuración,
actualizar una dependencia que ya está en la allowlist, cambios de documentación, un fix de pocas
líneas que no introduce comportamiento nuevo.

## Los límites duros — cruzar cualquiera = STOP, esto ya es una épica

Si mientras trabajas descubres que el cambio hace cualquiera de estas cosas, detente y escala a
`building-a-slice` (abre `EP-XXX` con su DoR):

1. **Agrega una dependencia nueva** fuera de `stack-allowlist.json`. (`stack-guard.sh` ya lo bloquea
   de forma determinista, pero no esperes a que el hook lo atrape — reconócelo antes.)
2. **Crea un endpoint o una API pública nueva.**
3. **Toca lógica de dominio**, o el modelo/invariantes de datos.
4. **El alcance se desborda**: empieza a introducir capacidad, tocar muchos archivos, o mezclar
   preocupaciones distintas.

Regla de desempate: si dudas, es una épica. Este carril nunca es un atajo para esquivar el DoR de un
cambio de producto real.

## El pipeline

1. **Rama tipada** — `fix/<slug>` para corrección, `chore/<slug>` para infra/config/docs/bump.
   `gitflow-guard.sh` exige rama tipada y prohíbe commit/push directo a `main`.
2. **Aplica el cambio, dentro de los límites.** Si al implementar cruzas un límite duro, para ahí
   mismo y escala — no sigas "un poco más" para terminarlo como micro-change.
3. **Test de regresión, solo si cambia comportamiento.** Un fix de bug necesita un test que falle
   antes del cambio y pase después (rojo→verde, delegado a
   `superpowers:test-driven-development`). Un cambio no-conductual (typo, docs, config) no lo exige.
4. **PR a `main`.** `gitflow-guard.sh` impide cualquier otra vía de integración. En la descripción,
   deja explícito que es un micro-change y qué límite duro NO cruza.

## Qué se conserva gratis vs. qué se omite

| Se conserva (hooks deterministas, sin esfuerzo extra) | Se omite por diseño |
|---|---|
| `gitflow-guard` — rama tipada + integración por PR | DoR formal (escenarios G/W/T, INVEST) |
| `stack-guard` — bloquea dependencias nuevas | OpenSpec change + bloque `## Trazabilidad` |
| `lint-typecheck` — estilo y typecheck incremental | `journey_smoke`, `api`, `data`, DoD reducido |
| | Apertura de `active_slice` y decisión de Release Gate |

`scaffold-guard.sh` no entra en juego aquí: no hay slice activo, y el propio decision gate ya exige
que el proyecto esté en fase `active` con el scaffold confirmado.

## Estado y trazabilidad

Un micro-change no escribe `build-state.json` — es mantenimiento fuera de banda, y su rastro vive en
git y en el PR, no en `history[]`. `reflect-nudge.sh` no lo empuja a reflexión posterior: no hay
aprendizaje de épica que capturar en la corrección de un typo.

## Reglas duras

- El decision gate del Paso 0 no es opcional. Ante la duda, trátalo como épica.
- Cruzar un límite duro es un STOP, no una excepción a negociar — siempre escala a
  `building-a-slice`.
- Toda integración pasa por PR a `main`, sin atajos (`gitflow-guard.sh` lo respalda).
- `METODOLOGIA.md` gana cualquier contradicción con este documento.
