---
name: "OPSX: Archive"
description: Mueve un change terminado al archivo, sincronizando antes sus delta-specs
category: Workflow
tags: [workflow, archivo, experimental]
---

## Objetivo

Cerrar un change: confirmar que realmente está terminado, sincronizar sus specs si corresponde, y moverlo al archivo.

## Entrada

`/opsx:archive [nombre-cambio]`. Sin argumento, infiere del contexto; si sigue ambiguo, pregunta — nunca adivines.

## Procedimiento

### 1. Elegir el change

Sin nombre: `openspec list --json`, restringido a changes activos (no archivados), y **AskUserQuestion** para elegir. Muestra el schema de cada uno. Auto-selección no está permitida acá — la persona siempre elige.

### 2. Verificar completitud de artefactos

```bash
openspec status --change "<nombre-cambio>" --json
```

Lee `schemaName` y el estado de cada artefacto. Cualquier artefacto que no esté `done`:
- avisa, listando qué falta,
- pide confirmación,
- avanza solo si se confirma.

### 3. Verificar completitud de tareas

Lee el archivo de tareas (típicamente `tasks.md`) y cuenta `- [ ]` vs `- [x]`. Tareas incompletas → avisa con el conteo, confirma, y avanza. Sin archivo de tareas → salta este chequeo en silencio.

### 4. Verificar el estado de sincronización de delta-specs

Revisa `openspec/changes/<nombre-cambio>/specs/`. Si no hay nada, salta al paso 5.

Si hay delta-specs, compáralos contra su spec principal en `openspec/specs/<capacidad>/spec.md` para determinar qué cambiaría (agregados/modificaciones/eliminaciones/renombres), y muestra ese resumen antes de preguntar nada.

Luego ofrece:
- Sin sincronizar todavía → "Sincronizar ahora (recomendado)" / "Archivar sin sincronizar"
- Ya sincronizado → "Archivar ahora" / "Sincronizar de todos modos" / "Cancelar"

Elegir sincronizar corre la misma lógica que `/opsx:sync`. En cualquier caso, el archivado sigue después.

### 5. Moverlo

```bash
mkdir -p openspec/changes/archive
```

Nombre destino: `AAAA-MM-DD-<nombre-cambio>` (fecha de hoy).

- Si el destino ya existe → falla con error; sugiere renombrar el archivo existente o esperar un día.
- Si no:
  ```bash
  mv openspec/changes/<nombre-cambio> openspec/changes/archive/AAAA-MM-DD-<nombre-cambio>
  ```

### 6. Reportar

Resume: nombre del change, schema, ruta de archivo, resultado de la sincronización de specs (sincronizado / omitido / sin delta-specs), y una nota de cualquier advertencia arrastrada de los pasos 2-3.

## Formato de salida

**Éxito**

```
## Archivo completo

**Change:** <nombre-cambio>
**Schema:** <nombre-schema>
**Archivado en:** openspec/changes/archive/AAAA-MM-DD-<nombre-cambio>/
**Specs:** ✓ Sincronizados con los specs principales

Todos los artefactos completos. Todas las tareas completas.
```

**Éxito, sin delta-specs**

```
## Archivo completo

**Change:** <nombre-cambio>
**Schema:** <nombre-schema>
**Archivado en:** openspec/changes/archive/AAAA-MM-DD-<nombre-cambio>/
**Specs:** Sin delta-specs

Todos los artefactos completos. Todas las tareas completas.
```

**Éxito con advertencias**

```
## Archivo completo (con advertencias)

**Change:** <nombre-cambio>
**Schema:** <nombre-schema>
**Archivado en:** openspec/changes/archive/AAAA-MM-DD-<nombre-cambio>/
**Specs:** Sincronización omitida (la persona eligió omitirla)

**Advertencias:**
- Archivado con 2 artefactos incompletos
- Archivado con 3 tareas incompletas
- Se omitió la sincronización de delta-specs (la persona eligió omitirla)

Revisá el archivo si esto no fue intencional.
```

**Falla (el destino ya existe)**

```
## Archivo fallido

**Change:** <nombre-cambio>
**Destino:** openspec/changes/archive/AAAA-MM-DD-<nombre-cambio>/

El directorio de archivo destino ya existe.

**Opciones:**
1. Renombrar el archivo existente
2. Borrar el archivo existente si es un duplicado
3. Esperar a otra fecha para archivar
```

## Límites

- Sin auto-selección — siempre preguntar qué change cuando no se da uno explícito.
- Basar los chequeos de completitud en `openspec status --json` y el grafo de artefactos, no en inspección visual.
- Las advertencias informan, no bloquean — la persona decide si avanza.
- `.openspec.yaml` viaja con el directorio durante el movimiento; no tocarlo por separado.
- Dejar siempre un resumen claro y específico de lo que pasó.
- La sincronización, cuando se pide, sigue el enfoque de `/opsx:sync` (fusión guiada por criterio), no una copia ciega.
- Si hay delta-specs, correr siempre la comparación y mostrarla antes de preguntar — no preguntar a ciegas.
