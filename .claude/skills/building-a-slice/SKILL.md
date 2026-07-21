---
name: building-a-slice
description: Use when building, continuing, or shipping a product epic (EP-XXX) end to end — drives the fast inner-loop pipeline (DoR → OpenSpec change linked to its epic → TDD → journey-smoke → api/data → reduced DoD → PR+archive → ask for Release Gate) using the build-state.json file to synchronize agents. Heavy reviews (security/design/UX/three-way coherence/architecture/integration) run once per release in the releasing-a-version skill, not per epic. The epic is the build unit; the HUs it covers are its internal scope. Delegates to opsx:* for changes and superpowers:test-driven-development for TDD; never reimplements them.
---

# Construcción de un slice (épica `EP-XXX`)

Este skill mueve una épica desde discovery ya cerrada hasta código integrado en `main`. No
implementa el motor de changes (`opsx:*`) ni el motor de TDD
(`superpowers:test-driven-development`): los invoca en secuencia y anota cada transición en
`.claude/state/build-state.json`.

## 0 · Unidad de trabajo

Un **slice** equivale exactamente a: una épica (`EP-XXX`) + un OpenSpec change + una rama +
un PR. Las historias de usuario que caen bajo esa épica (viven en `docs/04-historias/`) son
alcance interno del change, listadas en `active_slice.hus[]` — nunca se abre un slice por HU
suelta, eso sería fragmentar el trabajo por debajo de su unidad natural.

**Escape hatch de mantenimiento**: si lo que tienes enfrente es un typo, un ajuste de copy/config,
un bump de dependencia ya permitida o un fix de pocas líneas que no agrega capacidad, no abras
épica — usa `building-a-micro-change` (carril `fix/*`/`chore/*` directo a PR). Si ese
micro-cambio termina tocando una dependencia nueva, un endpoint nuevo, o lógica/datos de dominio,
detente y ábrelo aquí como épica: esos son los límites duros que separan un carril del otro.

## 1 · Precondiciones antes de abrir cualquier slice

Dos guardas bloquean la apertura del primer slice de un proyecto. Ambas se preguntan de forma
explícita al usuario (AskUserQuestion) y ambas están respaldadas por un hook determinista además
del gate del agente gatekeeper.

**1a. Scaffold runnable.** Lee `scaffold.confirmed` en `build-state.json`.
- `true` → continúa.
- `false` → pregunta *"¿Existe un scaffold runnable (arranca vacío, el script build/dev corre sin
  error)?"*. Si NO: **STOP**, indica crearlo según `stack-allowlist.json`/PRD técnico (el arnés no
  lo genera, es agnóstico al stack). Si SÍ: escribe `scaffold.confirmed=true` con
  `confirmed_by`/`confirmed_at`/`notes` y continúa. Hook: `scaffold-guard.sh` bloquea escribir
  código de slice sin esto.

**1b. Fuente de diseño (solo si el proyecto tiene UI).** Lee `design_source` en `build-state.json`.
- `applies` ausente → pregunta si el proyecto tiene UI y fija el valor.
- `applies === false` → N/A, sigue de largo.
- `applies === true && confirmed === false` → pregunta *"¿Hay una fuente de diseño declarada
  (prototipo/export) para la UI?"*. Si NO: **STOP**, indica declararla; no se abre slice con UI
  sin esto. Si SÍ: registra `design_source.source`, `confirmed=true`, `confirmed_by`,
  `confirmed_at`, `notes`. Hook: `design-source-guard.sh`.

Ambas quedan también como criterio duro del DoR (agente gatekeeper).

## 2 · Cómo piensa este loop

- **Fuente de verdad única**: `build-state.json` (esquema y protocolo completo en
  `.claude/state/README.md`). Se lee antes de cada acción, se escribe una vez por transición.
- **Un slice activo a la vez**, secuencial, sin saltar gates.
- **Carga perezosa de contexto**: cada `references/<tema>.md` se abre solo cuando la fase que lo
  necesita empieza.
- **Subagentes para explorar y para revisar pesado**, no para escribir en paralelo — devuelven
  síntesis condensada y protegen el presupuesto de atención de quien cablea.
- **Cada iteración es headless**: no hay memoria de conversación entre pasadas. El estado se
  reconstruye desde disco (git + `build-state.json` + logs), nunca desde lo que "se recuerda" haber
  hecho. Por eso existe `wiring_checklist[]` — un ítem por escenario AC y por punto de integración
  entre capas, que nace `failing` y solo pasa a `passing` cuando una prueba real lo demuestra. Deja
  una entrada en `progress_log[]` en cada hito. Mientras quede algo `failing`, el slice sigue
  abierto — ver `references/state-protocol.md`.

**Dos velocidades de revisión.** Este skill es el *inner loop*: rápido, por épica, sin subagentes
de revisión pesada, con meta de ~20 minutos por épica y el producto siempre caminando de punta a
punta. Seguridad, diseño, UX, coherencia triple, arquitectura e integración con dependencias reales
son el *outer loop* — corren una vez por release en `releasing-a-version`, no aquí.

**Tres reglas de secuenciación que gobiernan qué se puede construir cuándo:**

1. *Esqueleto antes que músculo.* El primer slice de cada release arma el journey completo más
   flaco posible (primera a última actividad del backbone del Story Map), aunque cada paso sea un
   stub. Cada slice siguiente engorda un paso sin romper `journey_smoke`. Nunca se levantan capas
   horizontales que "se ensamblan al final".
2. *Fundación antes que negocio.* Épicas `layer: foundational` (auth, acceso a datos, arquitectura
   base, design-system) se construyen antes que cualquier `layer: business`. El DoR rechaza una
   épica de negocio que depende de fundación no construida y exige extraerla primero. Así el slice
   de negocio solo toca lógica de negocio, sin quemar contexto en infraestructura.
3. *Trocear antes que sobrecargar.* Una épica que pasa el umbral de tamaño (por defecto >3 HU o
   ≥3 capas tocadas, configurable) no entra de una pasada: el DoR la parte en `sub_slices[]`
   construidos uno a la vez, cada uno con su propio `journey_smoke` verde antes de avanzar al
   siguiente. La exploración se reparte "ancho antes que profundo" con subagentes solo-lectura por
   área (front/back/datos); el cableado real lo hace siempre la sesión principal, nunca un
   subagente en paralelo.

## 3 · El pipeline

| # | Fase | Qué hace | A quién delega | Gate que cierra | Referencia |
|---|---|---|---|---|---|
| 1 | `dor` | Confirma Definition of Ready | agente gatekeeper | `dor` | `dor.md` |
| 2 | `change` | `opsx:new` + itera `opsx:continue` hasta completar todos los artefactos + sección `## Trazabilidad` + valida el enlace | `opsx:new`, `openspec-complete-artifacts`, `change-epic-coherence` | `coherence_link` | `link-change-epic.md` |
| 3 | `tdd` | Ciclo red→green→refactor | `superpowers:test-driven-development` | `tdd` | — |
| 4 | `smoke` | Recorre el journey acumulado end-to-end con el runner fuera-de-chat, en sesión virgen; en UI exige verificación visual real | runner `integration-check`, skill `verify`/`run` + MCP chrome-devtools, `ux-fidelity-reviewer` | `journey_smoke`, `fidelity` | `integration-check.md`, `mcp-map.md` |
| 5 | `api/data` | Contratos de API y consistencia de datos, si el slice los toca | `api-contract-tester`, `data-consistency-checker` | `api`, `data` | `newman-tests.md`, `data-consistency.md` |
| 6 | `dod` | Verificación adversarial independiente del cableado primero; DoD reducido después | agente wiring-adversarial-verifier, agente gatekeeper | `wiring_verified`, `dod` | `dod.md`, `integration-check.md` |
| 7 | `pr` | Abre PR y archiva el change en el mismo PR | `opsx:archive`, `opsx:sync` | — | `gitflow.md` |
| 8 | `release?` | Pregunta si corresponde correr el Release Gate | usuario, default computado | — | ver §5 |

Los gates `stack`, `security`, `smell`, `ux` y la coherencia triple **no viven aquí**: son del
Release Gate. Las dependencias siguen vigiladas en vivo por `stack-guard.sh`; lint/tsc/gitflow por
sus propios hooks.

`fidelity` es la excepción: sí vive en el inner loop porque se calcula en `smoke`, con la app ya
corriendo, y es un gate por-slice (distinto del `ux-krug-reviewer` de usabilidad, que sí es del
Release Gate). Para slices con UI (`design_source.applies===true`) no basta con que el prompt
"describa" la UI — el agente tiene que cargar la página, mirar la salida real y leer la consola vía
MCP chrome-devtools (screenshot app vs. prototipo). Sin ese MCP disponible, `fidelity` se queda en
`false` (ya no existe un estado "inconcluso que pasa") y `dod` no cierra hasta correr el slice donde
sí haya MCP. Se exige además cobertura total: ninguna pantalla del prototipo sin construir dentro
del alcance, ninguna pantalla de la app sin HU/épica que la respalde, y un recorrido de clics reales
como tenant no-admin.

Mapa MCP/LSP por gate: `references/mcp-map.md`. Protocolo de estado: `references/state-protocol.md`.

## 3.1 · Detalles de fase "change": generación completa de artefactos

La fase "change" genera **todos** los artefactos del workflow OpenSpec antes de pasar a TDD. Secuencia:

1. **`opsx:new`**: scaffoldea el change y muestra template del primer artefacto (proposal)
2. **Redacta proposal.md** con Por qué / Qué cambia / Capacidades / Impacto
3. **`openspec-complete-artifacts`**: delega a agente que itera automáticamente:
   - Genera `specs/<capacidad>/spec.md` (uno por capacidad en proposal)
   - Genera `design.md` (decisiones técnicas)
   - Genera `tasks.md` (checklist ejecutable)
   - Se detiene cuando no hay más artefactos ready
4. **Añade bloque `## Trazabilidad`** al final de proposal.md: épica + historias (audita `change-epic-coherence`)
5. Pasa a gate `coherence_link` ✓

Resultado: cuando entra a TDD, el change tiene propuesta, especificaciones, diseño y tareas — no es un documento incompleto.

## 4 · Al archivar: ¿toca Release Gate?

Tras el `opsx:archive` de la épica, pregunta al usuario si corresponde correr el Release Gate ahora.
El default se calcula, no se adivina, cruzando contra las líneas de release del Story Map
(`docs/02-user-story-map/`):

- Si la épica cierra una línea de release completa (todas sus épicas ya en `history[]`) → default
  "sí, correr `releasing-a-version` ahora".
- Si no la cierra → default "seguir con la siguiente épica", salvo que ya haya ≥2 épicas
  archivadas desde la última entrada de `releases[]` — ahí el nudge recomienda correrlo de todos
  modos aunque no sea el default.

El usuario puede pisar el default en cualquier caso.

## 5 · Cómo arrancar una sesión

1. Resuelve §1 (scaffold + fuente de diseño) si aún no está confirmado — sin esto no hay slice.
2. Identifica la épica objetivo en `docs/03-backlog/epicas.md` y reúne sus HU
   (`epica: EP-XXX` en `docs/04-historias/`) para poblar `hus[]`.
3. Lee `build-state.json`: si `active_slice` existe, retoma en su primer gate `false`; si es
   `null`, arranca en `dor`.
4. El agente orchestrator conduce el pipeline automáticamente, o ejecuta fase por fase respetando
   el orden de gates (si prefieres control manual).

## 6 · Invariantes que no se negocian

- Sin `scaffold.confirmed=true` no se abre slice ni se escribe código de slice — lo hace cumplir
  `scaffold-guard.sh`. El arnés exige el scaffold, jamás lo genera.
- En `harness_phase: authoring` (proyecto sin `package.json` aún) puedes avanzar precondiciones +
  `dor` + `change`, pero los gates de código (`tdd`, `journey_smoke`, `api`, `data`) esperan a que
  el scaffold esté confirmado.
- El enlace change↔épica vive en `## Trazabilidad` dentro del cuerpo markdown de `proposal.md` —
  **nunca** en frontmatter YAML, porque rompe `openspec validate`. Ver `link-change-epic.md`.
- Toda integración a `main` pasa por PR; `gitflow-guard.sh` bloquea commit/push directo.
- `dod` no cierra sin `wiring_verified: true` primero. La verificación adversarial es independiente
  a propósito: el checklist declarativo del gatekeeper es un piso mínimo, no el arreglo completo —
  el mismo agente no puede ser generador y verificador de su propio trabajo sin auto-confirmarse.
- El alcance acordado se construye completo, no un MVP recortado. Diferir o recortar alcance es una
  decisión de equipo explícita, nunca algo que el modelo decide solo. Toda verificación es
  ejecutada — nunca por inspección visual del código (ver `METODOLOGIA.md` y el bloque del arnés en
  `CLAUDE.md`).
- Ante cualquier contradicción entre esta skill y `METODOLOGIA.md`, gana la metodología.
