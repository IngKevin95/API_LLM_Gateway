# Qué MCP/LSP apoya cada gate

Los MCP listados aquí son **ejemplos opt-in**, no dependencias duras: el consumidor habilita en su
entorno los que le sirvan según el stack de su PRD. Ningún gate asume que un MCP está disponible;
cuando lo está, se invoca **just-in-time**, solo en la fase que lo necesita — nunca se precarga "por
si acaso".

## LSP para navegación semántica (cualquier fase con código)

El **LSP de TypeScript** permite seguir definiciones y referencias con precisión de compilador —
mejor señal que `grep` sobre código tipado. Lo aprovechan `coherence-three-way` (trazar símbolo →
test) y `simple-design-reviewer` (uso real de una función, detección de duplicación).

## Tabla de MCP por fase/agente

| Fase · agente | MCP (si está habilitado) | Para qué sirve |
|---|---|---|
| `smoke` · `ux-fidelity-reviewer` | **chrome-devtools** *(requerido en slices con UI)* | `new_page`/`navigate_page`, `take_screenshot`, `take_snapshot` para comparar composición/paleta/tipografía del diseño contra la app corriendo. Necesita la app levantada (fase `active`). Es la única excepción a "opt-in": sin este MCP, `fidelity` se queda en `false` y `dod` no cierra — hay que correr el slice donde el MCP esté disponible. |
| revisión · `ux-krug-reviewer` | **chrome-devtools** | `take_snapshot` (árbol accesible), `lighthouse_audit` (Accessibility/Best Practices), `take_screenshot`, `list_console_messages`, todo sobre la UI corriendo. |
| `api` · `api-contract-tester` | — (Newman por CLI, no es MCP) | Contratos de endpoint vía Bash. |
| `data` · `data-consistency-checker` | **postgresql** *(solo si el proyecto usa Postgres)* | `read_query`/`describe_table` para validar consistencia directamente en BD. Con otro almacén (según el PRD) se valida por tests, sin MCP. |
| performance (opcional, fuera del DoD) | **k6** | `execute_k6_test` para carga/latencia, si alguna HU lo exige explícitamente. |

## Reglas de uso

- Opt-in real: si el entorno del consumidor no expone el MCP, el gate se cubre igual con la
  alternativa CLI/test descrita en la tabla.
- `chrome-devtools`, cuando está habilitado, exige la app corriendo — solo tiene sentido en fase
  `active`.
- **figma** (u otro ecosistema de diseño) es otro MCP opcional, útil para consultar diagramas de
  arquitectura del slice; nunca forma parte de un gate.
- Ningún MCP debe recibir datos sensibles/PII real ni secretos — siempre datos sintéticos.
