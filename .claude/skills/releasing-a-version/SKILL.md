---
name: releasing-a-version
description: Use when closing a release of the product — runs the heavy review gates ONCE over the accumulated diff of a release line (security, design/smell, UX/Krug, three-way coherence, architecture, and full end-to-end integration with real deps), instead of per epic. This is the outer loop; the per-epic inner loop lives in building-a-slice. Trigger after archiving an epic when the user accepts the Release Gate, or when a Story Map release line is complete. Records results in build-state.json releases[].
---

# Release Gate — el outer loop del arnés

Mientras `building-a-slice` construye épica por épica, rápido y barato (inner loop), esta skill hace
lo contrario: corre las revisiones **caras** una sola vez por release, sobre el **diff acumulado** de
todas las épicas que la componen. El resultado práctico: el costo de los agentes pesados pasa de
escalar con el número de épicas a escalar con el número de releases.

## Qué cuenta como "una release"

Una línea de release completa del Story Map (`docs/02-user-story-map/`). Ejemplo: **R1-mvp** agrupa
EP-001 + EP-002 + EP-003 + EP-004 bajo un objetivo compartido ("resolver un caso de punta a punta con
resultado explicable"). Todas esas épicas se auditan juntas, como bloque.

## Cuándo se dispara

- `building-a-slice` pregunta con un default calculado al archivar una épica, y el usuario acepta.
- O de forma explícita: "corre el Release Gate de R1".

## Cómo opera

- Única fuente de verdad: `.claude/state/build-state.json`, específicamente `releases[]` — se lee
  antes de empezar y se escribe una entrada por release procesada.
- El alcance es el diff acumulado: desde el merge previo a la primera épica de la release hasta
  `main`.
- Cada gate se delega a un subagente que devuelve síntesis condensada, no ejecución paralela de
  escritura — así se protege el presupuesto de contexto.

## Los seis gates del pipeline

| Gate | Qué revisa | Subagente | Referencia |
|---|---|---|---|
| `security` | Vectores generales + foco de dominio del consumidor (PII/secretos/authz, claves server-side de servicios externos) sobre todo el diff | `security-reviewer` | — |
| `smell` | Las 4 reglas de Beck + code smells sobre el diff acumulado | `simple-design-reviewer` | `building-a-slice/references/simple-design.md` |
| `ux` | Krug + Lighthouse sobre la UI ensamblada de la release (`null` si no hay UI) | `ux-krug-reviewer` | `building-a-slice/references/krug-ux.md` |
| `coherence` | Trazabilidad triple AC↔change↔código de **todas** las HU de la release | `coherence-three-way` | — |
| `stack_arch` | Arquitectura contra el PRD del consumidor: capa de servicios externos en la frontera declarada, capa de decisión determinista sin IA | `stack-guardian` | — |
| `integration` | Journey completo de la release, de punta a punta, con **dependencias reales** — no stubs | skill `verify`/`run` (+ MCP chrome-devtools) | `references/release-dod.md` |

Checklist de cierre completo: `references/release-dod.md`.

## Procedimiento

1. Lee `build-state.json`; ubica la release y sus épicas cruzando con
   `docs/02-user-story-map/`.
2. Crea o actualiza la entrada correspondiente en `releases[]` con `status: pending`.
3. Dispara los subagentes de revisión en paralelo sobre el diff acumulado.
4. Corre el gate de integración con la skill `verify`/`run` — journey completo, deps reales, sin
   atajos.
5. Si todos los gates cierran (o quedan `null` cuando no aplican) → `status: passed`, se escriben los
   `gates` y `updated_by: releasing-a-version`. Si algo falla → `status: failed`, se listan los
   hallazgos bloqueantes, se corrigen como un slice normal (dentro de `building-a-slice`) y se
   re-corre el Release Gate completo.

## Reglas duras

- Este loop no duplica al inner loop: aquí no hay TDD ni gates por-slice.
- `integration` con dependencias reales es obligatorio para llegar a `status: passed` — es
  precisamente el gate que faltaba cuando "todo estaba en verde pero el producto no funcionaba al
  final". Con todo stubbeado no hay release aprobada.
- Cualquier contradicción con `METODOLOGIA.md` la resuelve la metodología, no esta skill.
