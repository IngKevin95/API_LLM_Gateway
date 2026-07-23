---
name: "OPSX: Fast Forward"
description: Crea un change y genera de una sola pasada todos los artefactos que hacen falta para implementar
category: Workflow
tags: [workflow, artefactos, experimental]
---

## Objetivo

Saltar el ida-y-vuelta artefacto-por-artefacto de `/opsx:continue`: crear el change y encadenar la generación de todo lo necesario hasta quedar listo para implementar.

## Entrada

`/opsx:ff [nombre-cambio | descripción]`. El argumento puede ser un nombre en kebab-case, o una descripción libre de lo que se quiere construir.

## Procedimiento

### 1. Entender qué se va a construir

Sin argumento: usa **AskUserQuestion** (abierta, sin opciones prefijadas) — "¿Qué querés construir o arreglar?". De la respuesta, derivá un nombre kebab-case ("agregar login con Google" → `agregar-login-google`).

No avances sin tener claro qué se va a construir — es la base de todo lo que sigue.

### 2. Crear el directorio del change

```bash
openspec new change "<nombre-cambio>"
```

Si ya existe un change con ese nombre, pregunta si continuarlo (→ `/opsx:continue`) o crear uno nuevo con otro nombre.

### 3. Averiguar el orden de construcción

```bash
openspec status --change "<nombre-cambio>" --json
```

De ahí interesan `applyRequires` (los artefactos que deben existir antes de poder implementar, p. ej. `["tasks"]`) y `artifacts` (cada uno con su estado y dependencias).

### 4. Encadenar artefactos hasta quedar listo

Usa una lista de tareas (TodoWrite/TaskCreate) para trackear el avance. Repite mientras haya artefactos con dependencias satisfechas:

1. Toma el artefacto `ready`, pide sus instrucciones:
   ```bash
   openspec instructions <artifact-id> --change "<nombre-cambio>" --json
   ```
2. De la respuesta usa `template` (estructura), `instruction` (guía del schema), `outputPath` (dónde escribir) y `dependencies` (artefactos previos a leer). `context` y `rules` son restricciones para tu criterio — nunca se copian al archivo.
3. Lee las dependencias completadas, redacta el artefacto y escríbelo en `outputPath`.
4. Muestra progreso breve: "✓ <artifact-id> creado".
5. Vuelve a consultar `openspec status --change "<nombre-cambio>" --json` y revisa si cada artefacto en `applyRequires` ya está `done`.

Si algún artefacto necesita contexto que no tenés, usa **AskUserQuestion** puntualmente y seguí — preferí decisiones razonables antes que frenar el impulso.

Detente cuando todos los artefactos de `applyRequires` estén completos.

### 5. Mostrar el estado final

```bash
openspec status --change "<nombre-cambio>"
```

## Resultado esperado

Resume: nombre y ubicación del change, artefactos creados con descripción breve, y el mensaje "Todos los artefactos creados — listo para implementar." Cierra sugiriendo `/opsx:apply` para arrancar la implementación.

## Guía para la creación de artefactos

- La fuente de verdad es siempre el campo `instruction` que devuelve la CLI para cada tipo de artefacto — el schema define el contenido.
- Lee las dependencias antes de escribir el artefacto que las necesita.
- Usa `template` como punto de partida, no como plantilla rígida — rellena con el contexto real del change.

## Límites

- Crear TODOS los artefactos que exija `applyRequires` del schema, sin saltar ninguno.
- Leer siempre las dependencias antes de crear el siguiente artefacto.
- Contexto críticamente ambiguo → preguntar; si es ambigüedad menor, decidir y avanzar.
- Change con nombre repetido → preguntar si continuar el existente o crear otro.
- Verificar que cada archivo quedó escrito en disco antes de pasar al siguiente artefacto.
