# Protocolo de `build-state.json`

El estado es el canal de comunicación entre agentes en un modelo secuencial: cada uno lee antes de
actuar, escribe su resultado, y pasa el testigo al siguiente. El esquema completo y las reglas base
viven en `.claude/state/README.md` y `.claude/state/build-state.schema.json`; este documento cubre
el uso operativo del día a día.

## Leer el estado

```bash
python3 -c 'import json;print(json.dumps(json.load(open(".claude/state/build-state.json")),indent=2,ensure_ascii=False))'
```

- `active_slice == null` → no hay slice abierto; solo la fase `dor` puede abrir uno nuevo.
- `active_slice` no-null → localiza el **primer gate en `false`**: por ahí se retoma.

## Por qué el estado vive en disco y no en la conversación

Cada iteración arranca headless, en contexto virgen, y reconstruye lo que pasó **desde disco**
(historial de git + este `build-state.json` + logs de build) — nunca desde "lo que la sesión
anterior recuerda haber hecho". Alargar una sesión para evitar esto degrada el razonamiento (el
ruido acumulado termina en deriva). Tres estructuras sostienen esa reconstrucción:

- **`wiring_checklist[]`** — un ítem por cada escenario AC de cada HU, y otro por cada punto de
  integración entre capas. Nace `failing`; solo pasa a `passing` cuando una prueba real lo demuestra
  (con `evidence` adjunta) — nunca por inspección visual del diff. Mientras algo siga `failing`, el
  slice sigue sin cablear: no se cierra `wiring_verified` ni `dod`.
- **`progress_log[]`** — bitácora append-only (`{at, by, note}`): qué se hizo, qué queda. Una entrada
  por hito basta para que la próxima sesión retome sin repetir trabajo.
- **`sub_slices[]`** — cuando la épica cruzó el umbral de tamaño (más de 3 HU, o 3+ capas tocadas),
  aquí vive el troceo. Cada sub-slice se construye uno a la vez, con su propio `journey_smoke` verde
  antes de tocar el siguiente.

`load-build-state.sh` inyecta al arrancar solo lo pendiente: los ítems `failing` y la última entrada
de bitácora — carga justo-a-tiempo, no el historial completo.

## Escribir el estado

Una transición = una escritura, patrón lee-modifica-escribe con timestamp UTC:

```bash
python3 - <<'PY'
import json,datetime
p=".claude/state/build-state.json"; d=json.load(open(p))
s=d["active_slice"]
s["gates"]["journey_smoke"]=True  # el gate que se cierra
s["phase"]="smoke"               # la fase a la que se avanza
s["updated_at"]=datetime.datetime.now(datetime.timezone.utc).isoformat()
s["updated_by"]="build-orchestrator"  # quién hizo la escritura
json.dump(d,open(p,"w"),indent=2,ensure_ascii=False)
PY
```

## Dos capas de gates

- **Inner loop (por slice)**: `dor`, `tdd`, `journey_smoke`, `coherence_link`, `data`, `fidelity`,
  `api`, `wiring_verified`, `dod`, todos dentro de `active_slice.gates`. Los escribe el flujo de
  `building-a-slice`/`build-orchestrator`. `wiring_verified` es prerequisito duro de `dod` — lo
  cierra un verificador adversarial independiente, jamás el mismo agente que construyó.
- **Outer loop (por release)**: los gates pesados — `security`, `smell`, `ux`, `coherence`,
  `stack_arch`, `integration` — no viven en ningún slice: viven en una entrada de `releases[]`, que
  escribe `releasing-a-version`. Una release equivale a una línea de release del Story Map.

## Reglas operativas

1. No saltar gates: `dod` no cierra con gates abiertos, y en particular exige `wiring_verified: true`
   primero.
2. Retroceso permitido: si una revisión posterior invalida algo, su gate vuelve a `false` y `phase`
   retrocede con él.
3. `fidelity`/`api` pueden ser `null` cuando no aplican (no cuentan como pendientes para el DoD).
   Excepción: `fidelity` no puede quedar `null` si el slice tiene UI — ahí debe llegar a `true` por
   verificación visual real. `wiring_verified`, en cambio, nunca admite `null`.
4. Archivar mueve `active_slice` a `history[]` con `phase: "archived"` y deja `active_slice: null`.
5. Validar después de escribir:
   ```bash
   python3 -c "import jsonschema,json; jsonschema.Draft202012Validator(json.load(open('.claude/state/build-state.schema.json'))).validate(json.load(open('.claude/state/build-state.json'))); print('OK')"
   ```

## Quién escribe qué

Ver `.claude/state/README.md` § "Quién escribe qué". Solo `build-orchestrator` transiciona `phase` y
archiva; cada reviewer toca únicamente su propio gate.
