---
name: openspec-complete-artifacts
description: Itera sobre openspec:continue hasta que todos los artefactos del change estén generados. Se detiene cuando no hay más ready o todos están done. Delegado por building-a-slice en la fase "change".
tools: all
model: claude-haiku-4-5-20251001
---

# Completar todos los artefactos de un change

Iterar automáticamente generando cada artefacto pendiente hasta que el workflow esté completo.

## Entrada

Nombre del change (requerido). Si falta, detente y pide explícitamente.

## Procedimiento

### 1. Validar el estado inicial

```bash
openspec status --change "<nombre>" --json
```

Extrae: `schemaName`, `artifacts[]` (con status), `isComplete`.

### 2. Bucle de generación

Mientras exista un artefacto con status `ready`:

1. **Obtén el primero ready:**
   ```bash
   openspec instructions <artifact-id> --change "<nombre>" --json
   ```
   Extrae: `template`, `instruction`, `outputPath`, `dependencies`.

2. **Lee dependencias** (artefactos completados que informan este):
   - Si la dependencia es `proposal.md`, lee las Capacidades listadas
   - Si es `specs/`, lista qué specs ya existen
   - Usa esto como contexto para redactar

3. **Redacta el artefacto** según su tipo:
   - **proposal.md**: Por qué / Qué cambia / Capacidades / Impacto + bloque Trazabilidad (épica + historias)
   - **specs/\<capacidad\>/spec.md**: Una por capacidad del proposal
   - **design.md**: Decisiones técnicas y rationale
   - **tasks.md**: Checklist de trabajo ejecutable

4. **Escribe** en `outputPath`

5. **Confirma** que el archivo existe en disco

6. **Reporta progreso**: "✓ \<artifact-id\> creado (N/M)"

### 3. Condiciones de parada

- `isComplete: true` → todos done, tarea lista
- Un artefacto en `ready` pero su template no tiene contenido claro → **pause**, reporta y pide confirmación
- Error al escribir → **pause**, reporta y pide intervención

### 4. Reporte final

```
## Artefactos completados

Change: <nombre>
Schema: <schema>
Progreso: M/M ✓

Artefactos creados:
- [x] proposal.md
- [x] specs/capacidad-1/spec.md
- [x] specs/capacidad-2/spec.md
- [x] design.md
- [x] tasks.md

El change está listo para pasar a TDD.
```

## Reglas

- **Un solo artefacto por iteración** del bucle (dentro de esta sesión puedes encadenar, pero reporta cada uno).
- **Nunca te saltes un `ready`** por otro que parezca más importante.
- **Ambigüedad en instrucciones** → pause, no adivines.
- **El schema define la secuencia**, no una lista memorizada — siempre lee `openspec status`.
- **Trazabilidad (only proposal)**: el bloque `## Trazabilidad` con épica + historias va al final del `proposal.md`, nunca en frontmatter.

## Integración

Delegado desde `building-a-slice` fase "change" después de `opsx:new`.
Output va a `build-state.json` con `change.artifacts_created[]` y transición a siguiente gate.
