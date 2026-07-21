---
name: "OPSX: Apply"
description: Recorre la lista de tareas de un change, implementando y marcando cada ítem (Experimental)
category: Workflow
tags: [workflow, implementación, experimental]
---

## Objetivo

Convertir la lista de tareas de un change en código funcionando, un checkbox a la vez.

## Entrada

`/opsx:apply [nombre-cambio]`. Sin argumento, infiere de la conversación; si sigue ambiguo, pregunta.

## Procedimiento

### 1. Elegir el change

- Nombre explícito → úsalo.
- Sin nombre: infiere del contexto, o auto-selecciona si hay exactamente un change activo.
- Sigue ambiguo → `openspec list --json` y **AskUserQuestion** para que la persona elija.

Anuncia la elección: "Usando change: <nombre>" junto con la sintaxis para cambiarlo (`/opsx:apply <otro>`).

### 2. Diagnosticar el schema antes de tocar nada

```bash
openspec status --change "<nombre-cambio>" --json
```

Extrae `schemaName` y determina qué artefacto contiene la lista de tareas — normalmente `tasks`, pero confírmalo con el estado en vez de asumirlo.

### 3. Pedir instrucciones de apply

```bash
openspec instructions apply --change "<nombre-cambio>" --json
```

La respuesta trae: rutas de `contextFiles`, contadores de progreso, la lista de tareas con su estado, y una instrucción que depende del estado actual:

| `state` | Acción |
|---|---|
| `blocked` | Faltan artefactos requeridos — señala `/opsx:continue` y detente |
| `all_done` | Felicita, sugiere `/opsx:archive` |
| cualquier otro | Continúa a la implementación |

### 4. Cargar contexto

Lee cada archivo listado en `contextFiles`. Para el schema estándar spec-driven eso es proposal, specs, design y tasks; otros schemas pueden listar archivos distintos — confía en lo que devuelve la CLI, no en supuestos.

### 5. Reportar el punto de partida

Antes de escribir código, muestra el schema en uso, "N/M tareas completas" y un vistazo rápido a lo pendiente, más cualquier instrucción dinámica que haya devuelto la CLI.

### 6. Trabajar la lista

Recorre las tareas pendientes en bucle:
1. Anuncia cuál estás empezando.
2. Haz el cambio mínimo de código que requiere.
3. Marca su checkbox: `- [ ]` → `- [x]`.
4. Pasa a la siguiente.

**Detente y muestra el problema en vez de adivinar** cuando:
- la descripción de la tarea es ambigua,
- implementarla revela un hueco de diseño (los artefactos necesitan actualizarse, no solo el código),
- aparece un error o bloqueo,
- la persona interrumpe.

### 7. Cerrar

Reporta las tareas terminadas en esta sesión y el progreso general. Todo completo → sugiere archivar. Si quedó pausado → explica el bloqueo y espera indicación.

## Formato de salida

**Progreso**

```
## Implementando: <nombre-cambio> (schema: <nombre-schema>)

Trabajando en tarea 3/7: <descripción>
[...implementación en curso...]
✓ Tarea completa

Trabajando en tarea 4/7: <descripción>
[...implementación en curso...]
✓ Tarea completa
```

**Completado**

```
## Implementación completa

**Change:** <nombre-cambio>
**Schema:** <nombre-schema>
**Progreso:** 7/7 tareas completas ✓

### Completadas en esta sesión
- [x] Tarea 1
- [x] Tarea 2
...

¡Todas las tareas completas! Listo para archivar este change.
```

**Pausado**

```
## Implementación pausada

**Change:** <nombre-cambio>
**Schema:** <nombre-schema>
**Progreso:** 4/7 tareas completas

### Problema encontrado
<descripción del problema>

**Opciones:**
1. <opción 1>
2. <opción 2>
3. Otro enfoque

¿Cómo querés seguir?
```

## Límites

- No te estanques entre tareas — seguí avanzando hasta terminar o bloquearte de verdad.
- Cargá los archivos de contexto antes de la primera edición, siempre.
- Tarea ambigua → preguntá, no adivines.
- Hueco de diseño que aparece a mitad de la implementación → proponé actualizar el artefacto, no diverjas en silencio.
- Mantené cada cambio acotado a su tarea; resistite a refactors de paso.
- Marcá una tarea en cuanto esté realmente hecha, no en lote al final.
- Errores y bloqueos pausan el bucle — no se resuelven a la fuerza.
- Confiá en `contextFiles`/`outputPath` que devuelve la CLI en vez de nombres de archivo fijos.

## Dónde encaja en el flujo

Apply no está atado a una fase fija — puede correr antes de que existan todos los artefactos (mientras existan las tareas), después de una implementación parcial, o entrelazado con otras acciones de opsx. Si implementar revela que los artefactos estaban mal, decilo y ofrecé actualizarlos en vez de tratar la lista de tareas como verdad absoluta.
