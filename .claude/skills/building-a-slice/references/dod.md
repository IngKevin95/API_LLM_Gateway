# DoD por slice (inner loop) — checklist de cierre

agente dor-dod-gatekeeper decide si un slice cierra leyendo los gates de `active_slice` en
`build-state.json`. Ninguna épica se archiva ni se mergea con un ítem de esta lista sin cumplir.

Este es el DoD **reducido**: rápido, por épica. Las revisiones caras — seguridad, diseño, UX,
trazabilidad triple completa, arquitectura — viven una sola vez por release en el Release Gate
(`skills/releasing-a-version/references/release-dod.md`), no aquí.

## Gates de agente

- **`tdd`** — cada escenario Given/When/Then de cada HU de la épica tiene su test; ciclo
  red→green→refactor cerrado; suite en verde.
- **`journey_smoke`** — el journey acumulado hasta esta épica (backbone del Story Map cubierto hasta
  aquí) se recorre de punta a punta sin romperse, verificado ejecutando (skill `verify`/`run`, más
  MCP `chrome-devtools` si la app corre) — nunca por lectura de código.
- **`coherence_link`** — `change-epic-coherence` confirma el bloque `## Trazabilidad` y que
  `openspec validate` pasa. Es el chequeo barato de esta trazabilidad; la versión completa
  (AC↔change↔código de toda la release) es del Release Gate.
- **`data`** — si el slice toca datos, `data-consistency-checker` confirma que la salida de
  servicios externos se validó contra esquema antes de llegar a la capa de decisión determinista.
- **`api`** — `api-contract-tester` (Newman) en 100% verde, o `null` si el slice no expone endpoints.
- **`fidelity`** (solo slices con UI) — `ux-fidelity-reviewer` devuelve FIEL, o DESVIACIONES todas
  justificadas → `true`; `null` si no hay UI. Para UI real (`design_source.applies===true`) no se
  acepta un veredicto "inconcluso": sin verificación visual real vía MCP `chrome-devtools`
  (screenshot app vs. prototipo) el gate queda `false` y bloquea `dod` hasta correr el slice con ese
  MCP disponible. Exige cobertura completa: nada del prototipo sin construir, nada en la app sin
  HU/épica que lo respalde.
- **`wiring_verified`** — un agente wiring-adversarial-verifier en contexto virgen, **distinto** de quien
  construyó, intenta activamente romper el slice (stubs olvidados, rutas sin cablear, AC sin test,
  ítems de `wiring_checklist[]` todavía `failing`). Solo pasa a `true` si no encuentra huecos. Es
  prerequisito duro de `dod`: el checklist declarativo de abajo es un piso mínimo, no sustituto de
  esta verificación independiente.

## Condiciones adicionales de cierre

- **Sub-slices**: si la épica se partió (`sub_slices[]` no vacío), todos en `status: done` con su
  propio `journey_smoke` verde. Ninguno pendiente.
- **OpenSpec**: todas las tasks marcadas `[x]`; el `archive` del change va en el **mismo PR**, nunca
  en uno separado.
- **Trazabilidad documental**: back-reference del change escrita en la épica y en cada HU de
  `hus[]`.
- **Hooks (automáticos, no gates de agente)**: `lint-typecheck.sh` en verde, `stack-guard.sh` sin
  dependencias fuera de `stack-allowlist.json`, `gitflow-guard.sh` confirmando rama `feature/*` y
  cero commits directos a `main`.

## Resultado

Todo lo anterior en verde → `gates.dod: true`, procede `opsx:archive` y el PR se abre/mergea. Algo
pendiente → se listan los gates abiertos y se regresa el control al `build-orchestrator`.

> Fuera de este checklist quedan `security`, `smell`, `ux`, `coherence` (trazabilidad triple
> completa) y `stack_arch` — todos del Release Gate en `releasing-a-version`. Las dependencias sí se
> vigilan por slice, pero vía hook (`stack-guard.sh`), no vía subagente de revisión.
