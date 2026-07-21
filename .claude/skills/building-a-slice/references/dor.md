# Definition of Ready — puerta de entrada al slice

La unidad que entra o no entra a construcción es la **épica** (`EP-XXX`), y el checklist se aplica
tanto a ella como a **cada HU** que declare en `hus[]`. Quien decide es agente dor-dod-gatekeeper; sin su
visto bueno no existe `active_slice`.

## Requisitos de la épica y sus historias

1. **La épica existe y traza al PRD**: `EP-XXX` está en `docs/03-backlog/epicas.md` con enlace
   explícito a un objetivo del PRD.
2. **Cobertura declarada**: al menos una HU listada, y la lista coincide exactamente con lo que
   entrará en `hus[]`.
3. **Frontmatter íntegro** en cada `docs/04-historias/HU-XXX.md` (`id, titulo, epica, prioridad,
   complejidad, estado`) con `estado: lista`.
4. **INVEST**: cada HU pasa los seis criterios (Independent, Negotiable, Valuable, Estimable,
   Small, Testable) — no se negocia parcialmente.

## Criterios de aceptación proporcionales, no una cuota fija

Los escenarios Given/When/Then escalan con la `complejidad` declarada, y cubren los modos de fallo
que **de verdad existen** en esa HU:

| Complejidad | Escenarios esperados |
|---|---|
| trivial / baja | 1–2 (happy + el edge/error crítico, si aplica) |
| media | 3 (happy + error + edge) |
| alta | 3–5 (cobertura completa) |

Regla dura: si una rama de error o de borde existe en la HU, **tiene** que tener su escenario — lo
que se evita es fabricar 3–5 escenarios artificiales para una HU trivial que no los necesita.

## Orden de construcción: cimiento antes que negocio

- **Dependencias resueltas**: toda épica/HU de la que depende ya está en `history[]`, o se declara
  explícitamente que no bloquea.
- **Excepción dura para infraestructura fundacional** (auth, acceso a datos, arquitectura base,
  design-system): esa cláusula de "no bloquea" **no aplica**. Tiene que estar construida y
  archivada primero.
- Si esta épica es `layer: business` y arrastra cimiento (`layer: foundational`) sin archivar →
  **STOP**: se extrae ese cimiento a su propia épica y se construye antes. Las fundacionales
  siempre van primero en la cola.

## Tamaño: cuándo trocear en sub_slices

Umbral por defecto (configurable por proyecto): **más de 3 HU** o **3 o más capas tocadas**. Una
épica que lo cruza no entra de una pasada — el DoR la descompone en `sub_slices[]`, cada uno
construido y verificado (`journey_smoke` verde) antes de tocar el siguiente. Es proporcional, no
una regla rígida: una épica de una capa con pocas HU entra directa.

## Otros requisitos de entrada

- **Stack compatible**: nada fuera de `stack-allowlist.json` (PRD §7).
- **Datos de prueba** disponibles o identificables (fixtures sintéticos del dominio).
- **Fuente de diseño confirmada** si el slice toca UI: `design_source.confirmed` en `true`, y este
  slice apunta a pantallas concretas de esa fuente. Nada de UI construida fuera de lo declarado.

## Resultado

**Todo ✓** → agente dor-dod-gatekeeper abre `active_slice` con `epica`, `hus[]`, `phase: dor`,
`gates.dor: true`, el resto de gates en `false` (incluido `wiring_verified: false`), y
`fidelity`/`api` en `null` si no aplican (`fidelity` arranca en `false` si el slice sí tiene UI).

**Algo ✗** → no se abre el slice. Se reporta el faltante puntual y se regresa a discovery.
