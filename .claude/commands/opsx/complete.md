---
name: "OPSX: Complete"
description: Genera todos los artefactos faltantes de un change (specs, design, tasks) iterando automáticamente
category: Workflow
tags: [workflow, artefactos, openspec, batch]
---

## Objetivo

Generar automáticamente todos los artefactos pendientes de un change OpenSpec, desde proposal hasta tasks, sin saltar pasos ni artefactos intermedios.

## Entrada

`/opsx:complete [nombre-cambio]`. Si falta el nombre, intenta inferirlo del contexto o pide selección explícita.

## Procedimiento

### 1. Resolver qué change completar

Si no hay nombre explícito:

```bash
openspec list --json
```

Ordena por `lastModified` y usa **AskUserQuestion** para ofrecer los 3-4 más recientes. Etiqueta el más reciente como "(Recomendado)".

### 2. Delegar al agente

Invoca el agente `openspec-complete-artifacts` con:
- Nombre del change
- Instrucción: "Genera todos los artefactos pendientes (specs, design, tasks) iterando con `opsx:continue` hasta que no haya más ready"

### 3. El agente itera

El agente:
- Lee `openspec status` para identificar qué está ready
- Itera mientras haya artefactos ready:
  - Obtiene instrucciones con `openspec instructions`
  - Lee dependencias (proposal, specs anteriores, etc.)
  - Redacta y escribe el artefacto
  - Confirma en disco
- Se detiene cuando `isComplete: true` o encuentra ambigüedad

### 4. Reporta resultado

Resultado esperado: todos los artefactos creados.

```
✓ proposal.md
✓ specs/capacidad-1/spec.md
✓ specs/capacidad-2/spec.md
✓ design.md
✓ tasks.md

Change listo para continuar a TDD.
```

## Cuándo usar

- El proposal ya existe pero faltan specs/design/tasks
- Retomás un change medio y querés generar todo lo faltante de una vez
- La fábrica `building-a-slice` no lo completó automáticamente (fix de la novedad reportada)

## Límites

- No edita artefactos ya creados — solo genera los faltantes
- Si hay ambigüedad en el contenido, pausa y pide intervención
- Un change a la vez (no batch)

## Salida

Reporte de qué se creó, progreso total (N/M), qué quedó desbloqueado, e invitación a implementar o archivar el change.
