# Checklist de cierre del Release Gate (outer loop)

Una release — una línea de release del Story Map — no se declara cerrada hasta que TODO lo de abajo
está en verde. Se corre **una única vez** sobre el diff acumulado de todas sus épicas, y el resultado
se escribe en `build-state.json` → `releases[]`.

- [ ] **`security`** — `security-reviewer` no reporta hallazgos CRÍTICO/ALTO en el diff completo de
  la release; claves de servicios externos solo server-side; PII/datos sensibles regulados (según el
  PRD del consumidor) nunca persistidos en crudo; toda salida de servicio externo/IA tratada como
  entrada no confiable y validada contra esquema antes de llegar a la capa de decisión.
- [ ] **`smell`** — `simple-design-reviewer` sin bloqueantes sobre el diff acumulado; las 4 reglas de
  Beck respetadas.
- [ ] **`ux`** — `ux-krug-reviewer` aprueba la UI ya ensamblada (o queda `null` si la release no tiene
  UI); Lighthouse/accesibilidad corrido si la app está en pie.
- [ ] **`coherence`** — `coherence-three-way` confirma trazabilidad AC↔change↔código para **todas**
  las HU de **todas** las épicas de la release, sin ninguna huérfana.
- [ ] **`stack_arch`** — `stack-guardian` confirma que la arquitectura respeta el PRD del consumidor:
  capa de servicios externos/IA en la frontera server-side declarada (sin decidir), capa de decisión
  determinista sin IA, ninguna clave de servicio externo expuesta en cliente.
- [ ] **`integration`** — el journey completo de la release se recorre de punta a punta con
  **dependencias reales** del proyecto (servicios externos/IA y capa de decisión reales, según el
  PRD del consumidor) — nunca stubs. Verificado ejecutando (skill `verify`/`run`, más MCP
  `chrome-devtools`).

## Resultado

Todo ✓ (o `null` donde no aplica) → `releases[].status: "passed"`, se escriben los `gates` y
`updated_by: releasing-a-version`. Cualquier ✗ → `status: "failed"`, con la lista de hallazgos
bloqueantes; se corrigen como un slice normal dentro de `building-a-slice` y se re-corre el Release
Gate completo.

> `integration` es el gate que la era anterior no tenía: todos los gates por-slice en verde y aun así
> el producto no funcionaba de punta a punta al terminar. Sin `integration` verificado con
> dependencias reales, no hay release cerrada.
