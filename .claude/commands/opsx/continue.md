---
name: "OPSX: Continue"
description: Avanza un cambio abierto creando únicamente el siguiente artefacto pendiente
category: Workflow
tags: [workflow, artefactos, openspec]
---

## Objetivo

Detectar en qué punto quedó un cambio en curso y producir el artefacto inmediatamente siguiente, sin saltarse pasos ni adelantarse.

## Entrada

`/opsx:continue [nombre-cambio]`. Si se omite el nombre, intenta inferirlo del contexto de la conversación; si sigue siendo ambiguo, detente y pide selección explícita (paso 1).

## Procedimiento

### 1. Resolver qué cambio continuar

Si no llegó un nombre de cambio explícito:

```bash
openspec list --json
```

Ordena el resultado por `lastModified` (más reciente primero) y usa **AskUserQuestion** para ofrecer los 3-4 cambios más recientes. Por cada opción muestra: nombre, schema (`schema` si existe, si no asumir "spec-driven"), progreso ("2/5 tareas", "completo", "sin tareas") y hace cuánto se tocó. Etiqueta el más reciente como "(Recomendado)".

No auto-selecciones un cambio adivinando por contexto pobre — la elección la hace la persona.

### 2. Diagnosticar el estado

```bash
openspec status --change "<nombre-cambio>" --json
```

De la respuesta interesan tres campos: `schemaName` (workflow en uso), `artifacts` (cada uno con estado `done` / `ready` / `blocked`) e `isComplete` (booleano global).

### 3. Decidir la acción según el estado

**Caso A — hay un artefacto en estado `ready`** (el camino más común):

1. Toma el PRIMER artefacto marcado `ready` en el arreglo `artifacts` — nunca elijas otro aunque parezca más relevante.
2. Pide sus instrucciones:
   ```bash
   openspec instructions <artifact-id> --change "<nombre-cambio>" --json
   ```
3. De ese JSON usarás: `template` (estructura base a rellenar), `instruction` (guía propia del schema), `outputPath` (ruta de escritura), `dependencies` (artefactos ya completados que debes leer antes de escribir). Los campos `context` y `rules` son restricciones para tu criterio — jamás se transcriben al archivo final.
4. Lee las dependencias completadas, redacta el artefacto siguiendo `template` + `instruction`, y escríbelo en `outputPath`.
5. Reporta qué se creó y qué quedó desbloqueado.
6. Detente — un solo artefacto por invocación, sin excepciones.

**Caso B — `isComplete: true`**: felicita, muestra el estado final con el schema usado, sugiere implementar o archivar el cambio, y detente.

**Caso C — ningún artefacto en `ready` y tampoco `isComplete`**: situación anómala para un schema válido; muestra el estado crudo y sugiere revisar la definición del schema en vez de forzar una creación.

### 4. Confirmar progreso tras escribir

```bash
openspec status --change "<nombre-cambio>"
```

## Reglas de construcción por schema (spec-driven)

Cuando el schema activo es el estándar (proposal → specs → design → tasks):

- **proposal.md**: si el motivo del cambio no está claro, pregunta antes de rellenar Por qué / Qué cambia / Capacidades / Impacto. La sección de Capacidades es la bisagra: cada capacidad listada ahí exige luego su propio spec.
- **specs/<capacidad>/spec.md**: un archivo por capacidad declarada en la proposal (nombrado por la capacidad, no por el cambio).
- **design.md**: decisiones técnicas, arquitectura elegida y por qué.
- **tasks.md**: desglose en checkboxes ejecutables.

Para cualquier otro schema, la fuente de verdad es el campo `instruction` devuelto por la CLI — no asumas nombres de artefacto de memoria.

## Resultado esperado

Al cerrar la invocación, deja constancia de: artefacto creado, schema en uso, progreso actual (N/M), qué se desbloqueó, y sugiere `/opsx:continue` para el siguiente paso.

## Límites

- Un artefacto por invocación — nunca más.
- Jamás te saltes artefactos ni cambies su orden.
- Ante ambigüedad de contenido, pregunta antes de escribir.
- Verifica que el archivo quedó escrito en disco antes de reportar avance.
- El orden de artefactos lo define el schema, no una lista fija memorizada.
- `context` y `rules` guían tu redacción pero nunca aparecen literalmente en el artefacto de salida.
