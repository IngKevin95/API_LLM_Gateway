---
name: "OPSX: Sync"
description: Fusiona los delta-specs de un change hacia los specs principales
category: Workflow
tags: [workflow, specs, experimental]
---

## Objetivo

Llevar los delta-specs de un change hacia los specs principales del proyecto. Es una operación **guiada por criterio propio**, no un merge programático: lees los delta-specs y editas directamente los specs principales, aplicando fusión inteligente (por ejemplo, agregar un escenario sin copiar todo el requisito que lo contiene).

## Entrada

`/opsx:sync [nombre-cambio]`. Si se omite, intenta inferirlo del contexto de conversación; si sigue ambiguo, usa **AskUserQuestion** mostrando los changes que tengan delta-specs (carpeta `specs/`). Nunca adivines ni auto-selecciones — la persona elige.

## Procedimiento

### 1. Localizar los delta-specs

Busca en `openspec/changes/<nombre-cambio>/specs/*/spec.md`. Cada archivo delta puede traer estas secciones:

- `## ADDED Requirements` — requisitos nuevos a agregar
- `## MODIFIED Requirements` — cambios sobre requisitos existentes
- `## REMOVED Requirements` — requisitos a eliminar
- `## RENAMED Requirements` — requisitos a renombrar (formato FROM:/TO:)

Si no aparece ningún delta-spec, informa y detente.

### 2. Fusionar cada delta contra el spec principal

Por cada capacidad con delta-spec en `openspec/changes/<nombre-cambio>/specs/<capacidad>/spec.md`:

1. Lee el delta-spec para entender la intención del cambio.
2. Lee el spec principal en `openspec/specs/<capacidad>/spec.md` (puede no existir todavía).
3. Aplica los cambios con criterio:

   **ADDED Requirements** — si el requisito no existe en el spec principal, agrégalo; si ya existe, actualízalo para que coincida (trátalo como un MODIFIED implícito).

   **MODIFIED Requirements** — ubica el requisito en el spec principal y aplica el cambio, que puede ser: agregar escenarios nuevos (sin copiar los existentes), modificar escenarios existentes, o cambiar la descripción del requisito. Todo lo no mencionado en el delta se preserva tal cual.

   **REMOVED Requirements** — elimina el bloque completo del requisito en el spec principal.

   **RENAMED Requirements** — ubica el requisito FROM y renómbralo a TO.

4. Si la capacidad no tiene spec principal todavía, créalo: `openspec/specs/<capacidad>/spec.md` con una sección de Propósito (puede ser breve, marcada como pendiente) y la sección de Requisitos con lo que venga en ADDED.

### 3. Reportar el resultado

Tras aplicar todos los cambios, resume qué capacidades se actualizaron y qué operación se hizo en cada una (agregado / modificado / eliminado / renombrado).

## Formato de referencia del delta-spec

```markdown
## ADDED Requirements

### Requirement: Función nueva
El sistema DEBE hacer algo nuevo.

#### Scenario: Caso básico
- **WHEN** el usuario hace X
- **THEN** el sistema hace Y

## MODIFIED Requirements

### Requirement: Función existente
#### Scenario: Escenario nuevo a agregar
- **WHEN** el usuario hace A
- **THEN** el sistema hace B

## REMOVED Requirements

### Requirement: Función obsoleta

## RENAMED Requirements

- FROM: `### Requirement: Nombre viejo`
- TO: `### Requirement: Nombre nuevo`
```

## Principio clave: fusión inteligente

A diferencia de un merge programático, aquí se permiten **actualizaciones parciales**: para agregar un escenario basta incluirlo bajo MODIFIED, sin copiar los escenarios que ya existían. El delta representa una *intención*, no un reemplazo total — usa tu criterio para fusionar con sensatez.

## Resultado esperado

```
## Specs sincronizados: <nombre-cambio>

Specs principales actualizados:

**<capacidad-1>**:
- Requisito agregado: "Función nueva"
- Requisito modificado: "Función existente" (1 escenario agregado)

**<capacidad-2>**:
- Archivo de spec creado
- Requisito agregado: "Otra función"

Los specs principales quedaron al día. El change sigue activo — archívalo cuando la implementación esté completa.
```

## Límites

- Lee tanto el delta como el spec principal antes de tocar nada.
- Preserva todo contenido del spec principal que el delta no mencione.
- Ante ambigüedad, pregunta antes de aplicar.
- Muestra qué vas cambiando a medida que avanzas.
- La operación debe ser idempotente — correrla dos veces debe dar el mismo resultado.
