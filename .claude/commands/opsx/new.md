---
name: "OPSX: New"
description: Arranca un change nuevo usando el flujo experimental basado en artefactos (OPSX)
category: Workflow
tags: [workflow, artefactos, experimental]
---

## Objetivo

Arrancar un change nuevo con el enfoque guiado por artefactos, dejándolo listo para que la persona decida cómo seguir.

## Entrada

`/opsx:new [nombre-cambio | descripción]`. El argumento puede ser un nombre en kebab-case, o una descripción libre de lo que se quiere construir.

## Procedimiento

### 1. Entender qué se va a construir

Sin argumento: usa **AskUserQuestion** (abierta, sin opciones prefijadas) — "¿Qué cambio querés trabajar? Describí qué querés construir o arreglar."

De la respuesta, derivá un nombre kebab-case (p. ej. "agregar autenticación de usuarios" → `agregar-auth-usuarios`).

**Importante**: no avances sin tener claro qué se va a construir.

### 2. Determinar el schema del flujo

Usa el schema por defecto (omití `--schema`) salvo que la persona pida explícitamente otro workflow.

**Usa un schema distinto solo si la persona menciona:**
- Un nombre de schema específico → usá `--schema <nombre>`
- "mostrame los workflows" o "qué workflows hay" → corré `openspec schemas --json` y dejá que elija

**En cualquier otro caso**: omití `--schema` para usar el default.

### 3. Crear el directorio del change

```bash
openspec new change "<nombre-cambio>"
```

Agregá `--schema <nombre>` solo si la persona pidió un workflow específico. Esto crea el change en `openspec/changes/<nombre-cambio>/` con el schema elegido.

### 4. Mostrar el estado de artefactos

```bash
openspec status --change "<nombre-cambio>"
```

Muestra qué artefactos faltan y cuáles ya están listos (con sus dependencias satisfechas).

### 5. Pedir instrucciones del primer artefacto

El primer artefacto depende del schema. Fijate en el estado cuál es el primero marcado "ready".

```bash
openspec instructions <primer-artefacto-id> --change "<nombre-cambio>"
```

Esto devuelve el template y el contexto para crear ese primer artefacto.

### 6. Detente y esperá indicación

No sigas de largo — este comando entrega el punto de partida, no la ejecución.

## Resultado esperado

Al terminar, resumí:
- Nombre y ubicación del change
- Schema/workflow en uso y su secuencia de artefactos
- Estado actual (0/N artefactos completos)
- El template del primer artefacto
- Cierre: "¿Arrancamos con el primer artefacto? Corré `/opsx:continue` o simplemente describime de qué se trata y lo redacto."

## Límites

- No crear ningún artefacto todavía — solo mostrar las instrucciones.
- No avanzar más allá de mostrar el template del primer artefacto.
- Nombre inválido (no kebab-case) → pedir uno válido.
- Change con ese nombre ya existente → sugerir `/opsx:continue` en su lugar.
- Pasar `--schema` únicamente si se usa un workflow no-default.
