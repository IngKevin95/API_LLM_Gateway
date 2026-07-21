---
name: "OPSX: Bulk Archive"
description: Archiva varios changes terminados en una sola pasada, resolviendo conflictos de specs contra lo que realmente está construido
category: Workflow
tags: [workflow, archivo, experimental, masivo]
---

## Objetivo

Archivar N changes de una sola vez en lugar de uno por uno — y cuando dos changes tocan la misma spec de capacidad, determinar cuál es real antes de sincronizar nada.

## Entrada

Ninguna. El comando siempre pregunta la selección.

## Procedimiento

### 1. Reunir candidatos

`openspec list --json` → todos los changes activos. Si no hay ninguno, informa y detente.

### 2. Dejar que la persona elija

**AskUserQuestion**, multi-selección, una opción por change (con su schema) más un atajo "Todos los changes". Cualquier cantidad es válida — incluso una sola selección, aunque el sentido de este comando es agrupar 2 o más. Nunca preseleccionar nada.

### 3. Validar cada change seleccionado

Por change, reunir tres datos:

- **Artefactos** — `openspec status --change "<nombre>" --json`, anota `schemaName` y qué artefactos no están `done`.
- **Tareas** — lee `tasks.md`, cuenta `- [x]` contra `- [ ]`. Sin archivo → registra "Sin tareas."
- **Delta-specs** — lista `openspec/changes/<nombre>/specs/*/spec.md`, y de cada uno extrae los nombres `### Requirement:`.

### 4. Mapear capacidades para detectar conflictos

Arma `capacidad -> [changes que la tocan]`:

```
auth -> [change-a, change-b]   conflicto: 2+ changes
api  -> [change-c]             sin problema: 1 change
```

Toda capacidad con 2 o más changes es un conflicto a resolver antes de archivar.

### 5. Resolver cada conflicto contra el código real

Por cada capacidad en conflicto:

1. Lee el delta-spec de cada change para ver qué afirma.
2. Busca en el código evidencia — archivos, funciones, tests que coincidan.
3. Decide:
   - Solo un lado está realmente implementado → sincroniza solo ese.
   - Ambos están implementados → sincroniza ambos, el change más antiguo primero (su contenido entra primero, las ediciones del change más nuevo ganan donde se superponen).
   - Ninguno está implementado → no sincronices esa capacidad todavía, márcala como advertencia.
4. Anota el razonamiento (lo que encontraste en el código) junto con la decisión — esto va en la tabla de estado del paso siguiente.

**Ejemplo resuelto, un solo ganador:**
```
spec auth tocada por [agregar-oauth, agregar-jwt]
agregar-oauth: el delta agrega "Integración de proveedor OAuth" -> src/auth/oauth.ts existe, implementa el flujo
agregar-jwt:   el delta agrega "Manejo de tokens JWT"           -> nada encontrado en src/
Resolución: sincronizar solo agregar-oauth.
```

**Ejemplo resuelto, ambos reales:**
```
spec api tocada por [agregar-rest-api (2026-01-10), agregar-graphql (2026-01-15)]
Ambos tienen archivos fuente correspondientes (rest.ts, graphql.ts).
Resolución: sincronizar agregar-rest-api primero, después agregar-graphql (orden cronológico, el más nuevo gana las superposiciones).
```

### 6. Mostrar una tabla antes de preguntar nada

```
| Change              | Artefactos | Tareas | Specs      | Conflictos | Estado    |
|---------------------|------------|--------|------------|------------|-----------|
| gestion-schemas     | Completo   | 5/5    | 2 delta    | Ninguno    | Listo     |
| agregar-oauth       | Completo   | 4/4    | 1 delta    | auth (!)   | Listo*    |
| agregar-skill-verify| Falta 1    | 2/5    | Ninguno    | Ninguno    | Alerta    |

* auth: aplicando agregar-oauth y luego agregar-jwt (ambos implementados, orden cronológico)
Advertencias: agregar-skill-verify tiene 1 artefacto incompleto, 3 tareas incompletas
```

### 7. Una sola confirmación para todo el lote

**AskUserQuestion**, pregunta única, opciones aproximadas:
- Archivar los N
- Archivar solo los listos, omitir los incompletos
- Cancelar

Si hay changes incompletos en la mezcla, aclara que llevarán advertencias al archivo.

### 8. Ejecutar, en el orden de resolución

Por cada change confirmado, respetando cualquier orden cronológico del paso 5:

1. Si tiene delta-specs, sincronízalos como lo haría `/opsx:sync` (fusión guiada por criterio, no copia ciega). Registra si la sincronización ocurrió.
2. `mkdir -p openspec/changes/archive` y luego `mv openspec/changes/<nombre> openspec/changes/archive/AAAA-MM-DD-<nombre>`.
3. Registra el resultado: archivado / falló (con error) / omitido (la persona lo descartó).

### 9. Reportar

## Formato de salida

**Todo exitoso**
```
## Archivado masivo completo

Se archivaron N changes:
- <change-1> -> archive/AAAA-MM-DD-<change-1>/
- <change-2> -> archive/AAAA-MM-DD-<change-2>/

Sincronización de specs: N delta-specs sincronizados, M conflictos resueltos (o "sin conflictos")
```

**Resultado mixto**
```
## Archivado masivo completo (parcial)

Se archivaron N changes:
- <change-1> -> archive/AAAA-MM-DD-<change-1>/

Omitidos M (la persona descartó los incompletos):
- <change-2>

Fallaron K:
- <change-3>: el destino de archivo ya existe
```

**Nada para hacer**
```
## No hay changes para archivar

No se encontraron changes activos. Usá `/opsx:new` para arrancar uno.
```

## Límites

- La selección siempre es manual — nunca auto-elegir, ni siquiera con un solo candidato.
- Detectar conflictos antes de mostrar la tabla de estado, no después.
- La evidencia en el código decide el orden de resolución, no solo la fecha del change (la fecha solo desempata cuando ambos lados están implementados).
- Implementación ausente en ambos lados significa omitir-y-advertir, no adivinar-y-sincronizar.
- Una sola confirmación habilita todo el lote — no preguntar change por change.
- Cada change recibe un resultado registrado: archivado, omitido, o fallido.
- `.openspec.yaml` se mueve con su directorio; no manejarlo por separado.
- Los destinos de archivo llevan fecha `AAAA-MM-DD-<nombre>`; una colisión hace fallar solo ese change, sin abortar el resto.
