---
name: building-a-slice
description: Build or continue a product epic (EP-XXX) end-to-end through the full pipeline. Invokes build-orchestrator to drive the inner loop (DoR → OpenSpec change → TDD → smoke → api/data → DoD → PR). The epic is the build unit.
---

# Construcción de épica EP-XXX

**INSTRUCCIÓN**: Cuando este skill sea invocado, **inmediatamente invoca el agente `build-orchestrator`** con el texto que el usuario proporcionó. No es documentación pasiva — es el punto de entrada activo del pipeline.

**Entry point único** para construcción de épicas. Invoca automáticamente `build-orchestrator` que valida precondiciones y orquesta todo.

Invocar este skill con la épica objetivo (EP-NNN). Si hay `active_slice` en construcción, invocar con `continue` para retomar desde donde paró.

## Entrada

- **Nueva épica**: "Construye EP-003" → orquestador junta HU con `epica: EP-003` de `docs/04-historias/`
- **Retomar**: "Continúa" → orquestador lee `active_slice`, ubica primer gate abierto y retoma

## Precondiciones (verificadas automáticamente)

- `scaffold.confirmed === true` en `build-state.json` (si no: STOP, scaffold es obligatorio, no hay épica sin código runnable)
- `design_source.confirmed === true` si la épica toca UI
- Épica existe en `docs/03-backlog/epicas.md` con ≥1 HU asociada

## Pipeline (8 fases, orquestadas automáticamente)

| Fase | Delegación | Gate |
|---|---|---|
| 1. DoR | `dor-dod-gatekeeper` agent | `dor` |
| 2. Change | `opsx:*` commands + `change-epic-coherence` agent | `coherence_link` |
| 3. TDD | `superpowers:test-driven-development` skill | `tdd` |
| 4. Smoke | `integration-check` + `ux-fidelity-reviewer` agent (si UI) | `journey_smoke`, `fidelity` |
| 5. API/Data | `api-contract-tester`, `data-consistency-checker` agents | `api`, `data` |
| 6. DoD | `wiring-adversarial-verifier` (primero) + `dor-dod-gatekeeper` (después) | `wiring_verified`, `dod` |
| 7. PR | `opsx:archive`, `opsx:sync` commands | — |
| 8. Release? | Pregunta al usuario (default calculado) | — |

## Quién hace qué

- **Este skill** (`/building-a-slice`): Punto de entrada único, invoca automáticamente `build-orchestrator`
- **build-orchestrator agent**: Orquesta todo el pipeline, valida precondiciones, lee/escribe `build-state.json`, delega a agentes y skills
- **Agentes especializados**: Cada uno resuelve su gate (DoR, coherencia, fidelidad, cableado, etc.)
- **Skills de operación**: `opsx:*` para cambios OpenSpec, `superpowers:test-driven-development` para TDD

## Estado persistido

Fuente única: `.claude/state/build-state.json`. Se lee antes de cada fase, se escribe después cada transición. Sin estado en conversación — headless. Cada sesión deja `progress_log[]` para auditoría.

**No hay ambigüedad**: cuando invocas `/building-a-slice` con la épica, automáticamente se dispara `build-orchestrator` para conducir el resto.

---

## Invocación

Formato:
```
/building-a-slice Construye EP-NNN
/building-a-slice Continúa
```

**Esto automáticamente invoca el agente `build-orchestrator`** que:
1. Valida precondiciones (scaffold.confirmed, design_source si UI)
2. Lee épica y HU desde `docs/03-backlog/` y `docs/04-historias/`
3. Orquesta las 8 fases del pipeline
4. Persiste estado en `build-state.json`

No es necesario invocar agentes manualmente — todo fluye a través de este skill.
