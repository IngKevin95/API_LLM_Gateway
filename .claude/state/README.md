# `build-state.json` — estado del arnés

`build-state.json` es la **única fuente de verdad** sobre qué slice (épica) está en construcción.
Es el medio de sincronización entre agentes en modo **secuencial**: la unidad de construcción es la
**épica (`EP-XXX`)**, y las HU que el change cubre quedan listadas en `hus[]` para trazabilidad al
alcance interno. Cada gate del pipeline transiciona el estado, y el agente siguiente lo lee antes de
mover un dedo.

- **Esquema**: `build-state.schema.json` (versionado, JSON Schema draft 2020-12). Para validar:
  ```bash
  # Recomendado (funciona sin plugins de formato):
  python3 -c "import jsonschema,json; jsonschema.Draft202012Validator(json.load(open('.claude/state/build-state.schema.json'))).validate(json.load(open('.claude/state/build-state.json'))); print('OK')"
  # Alternativa con ajv (requiere spec + formatos):
  npx --yes ajv-cli@5 validate --spec=draft2020 -c ajv-formats -s .claude/state/build-state.schema.json -d .claude/state/build-state.json
  ```
- **El estado vivo no se versiona**: `build-state.json` vive en `.gitignore` (es working state del
  proyecto consumidor). El schema y este README sí viajan con el paquete.

## Protocolo: quién lee, quién escribe

1. **Leer antes de tocar nada.** Cualquier agente del pipeline consulta `active_slice` y `gates`
   antes de trabajar. `active_slice` en `null` significa que no hay slice abierto — solo
   `dor-dod-gatekeeper` puede abrir uno nuevo.
2. **Una escritura por transición, nada más.** Quien cierra una fase actualiza `phase`, el gate que
   le corresponde, `updated_at` (ISO 8601 UTC) y `updated_by` (su propio nombre). No toca ningún
   otro campo de paso.
3. **Los gates solo avanzan.** Un gate pasa de `false` a `true` cuando su agente lo aprueba. Si una
   revisión posterior detecta un problema, vuelve a `false` y `phase` retrocede con él.
4. **`fidelity` y `api` toleran `null`** cuando el slice no tiene UI o no expone endpoints (N/A, no
   bloquean el DoD). Lo que `fidelity` **no** puede hacer es quedar en `null` si el slice sí toca UI.
   `wiring_verified`, en cambio, siempre aplica (nunca `null`) y es **prerequisito duro de `dod`**.
5. **Archivado**: al terminar `opsx:archive`, el `active_slice` se mueve a `history[]` con
   `phase: "archived"` y `active_slice` vuelve a `null`.
6. **Dos loops conviven aquí**: el inner loop (por slice) cierra los gates baratos épica por épica;
   los gates pesados se agregan en `releases[]` (outer loop) y los escribe la skill
   `releasing-a-version` una vez por release, no una vez por slice.

## El recorrido de fases

**Inner loop, una vez por slice:**
```
dor → change → red → green → refactor → smoke → api → data → dod → pr → archived
```
Al final de ese recorrido entra el **Release Gate** (outer loop, skill `releasing-a-version`), cuyo
default se computa a partir de las líneas de release del Story Map.

`harness_phase` nace en `authoring`; `load-build-state.sh` lo pasa a `active` en cuanto detecta un
`package.json` en la raíz (señal de que ya existe scaffold de código). Es puramente **informativo**
— no es el gate que decide si se puede avanzar.

### Gate `scaffold` — el paso 1 que nada se salta

Es un gate **de proyecto**, no por-slice: `{ confirmed, confirmed_by, confirmed_at, notes }`. Nace en
`confirmed: false` y solo llega a `true` por **confirmación humana explícita** (a través de la
Fase 0 de `building-a-slice` / `dor-dod-gatekeeper`) — jamás por detección automática. Es la
condición restrictiva para abrir cualquier slice: `scaffold-guard.sh` bloquea escribir código de
slice (fases `red` a `data`) mientras `confirmed` siga en `false`. El arnés no fabrica el scaffold
por sí mismo, solo audita que alguien confirmó que existe.

### `design_source` — el mismo gate, pero para la UI

Su contraparte para proyectos con interfaz: `applies` (¿este proyecto tiene UI?), `confirmed`
(¿una persona confirmó que existe una fuente de diseño declarada y vigente?), `source` (puntero al
prototipo o export). El arnés tampoco genera el prototipo — solo lo respalda `design-source-guard.sh`.

### Gate `fidelity` — estricto, por-slice, inner loop

`true` significa fiel al diseño (o con desviaciones justificadas, y **siempre vía verificación
visual real**); `false` significa desviaciones sin justificar **o** que la verificación visual
simplemente no se corrió; `null` significa que el slice no toca UI. Lo computa
`ux-fidelity-reviewer` durante la fase `smoke`, y lo persiste el `build-orchestrator`. En slices con
UI (`design_source.applies === true`) el único camino para cerrarlo es **MCP de devtools de
navegador** comparando screenshot de la app contra el prototipo — un resultado inconcluso no cuenta
como aprobado, se queda en `false` y el `dod` no cierra hasta correr esa verificación donde haya MCP
disponible.

### Handoff fino del cableado — sobrevive a un reset de contexto

Tres campos por-slice pensados para que una sesión nueva retome el trabajo leyendo disco, no
"confiando" en la conversación anterior:
- **`wiring_checklist[]`** — un item por escenario de aceptación y por punto de integración entre
  capas; nace en `failing` y solo pasa a `passing` tras una prueba real ejecutada, con `evidence`
  adjunta.
- **`progress_log[]`** — bitácora append-only (`{at, by, note}`) que no se pierde en un reset.
- **`sub_slices[]`** — la descomposición de una épica que cruzó el umbral de tamaño (más de 3 HU, o
  3 o más capas tocadas).

Y el gate que corona todo esto: **`wiring_verified`**, cerrado por el `wiring-adversarial-verifier`
— un subagente independiente, en contexto virgen, cuyo trabajo es intentar refutar el slice. `dod`
no puede pasar a `true` sin que este gate lo haga primero; el `dor-dod-gatekeeper` lo deja en
`false` al abrir el DoR.

### Reflexión post-slice — el ciclo que se autocorrige

Cuando un slice se archiva, su entrada en `history[]` puede llevar `reflected` / `reflected_at`. El
hook `reflect-nudge.sh` (evento `Stop`, no bloqueante) sugiere correr `/build:reflect` mientras
quede alguna entrada con `reflected != true`. Quien ejecuta `/build:reflect` es el modelo: detecta
una convención nueva o un error recurrente, **propone** un parche al bloque `factory-build-learnings`
de `CLAUDE.md` (se aplica solo si el usuario lo aprueba) y estampa `reflected: true` para que el
nudge deje de insistir. El razonamiento vive en el modelo — el hook es solo un recordatorio
determinista.

## Quién escribe qué

| Campo / gate | Lo escribe | Cadencia |
|---|---|---|
| `active_slice` (alta) · `gates.dor` · `gates.dod` | `dor-dod-gatekeeper` | slice |
| `gates.coherence_link` | `change-epic-coherence` | slice |
| `gates.tdd` | flujo `superpowers:test-driven-development` (vía `build-orchestrator`) | slice |
| `gates.journey_smoke` · `gates.fidelity` · `phase` (transiciones) · `history[]` (archivado) · `wiring_checklist[]` · `progress_log[]` · `sub_slices[]` | `build-orchestrator` (fidelity desde `ux-fidelity-reviewer`) | slice |
| `gates.api` | `api-contract-tester` | slice |
| `gates.data` | `data-consistency-checker` | slice |
| `gates.wiring_verified` | `wiring-adversarial-verifier` (independiente, contexto virgen) | slice (antes de `dod`) |
| `releases[]` (`security`, `smell`, `ux`, `coherence`, `stack_arch`, `integration`, `status`) | `releasing-a-version` (delega en `security-reviewer`, `simple-design-reviewer`, `ux-krug-reviewer`, `coherence-three-way`, `stack-guardian`) | release |
| `history[].reflected` · `history[].reflected_at` | `/build:reflect` | post-slice (tras archivar) |
| `harness_phase` | `load-build-state.sh` (SessionStart) | — |
