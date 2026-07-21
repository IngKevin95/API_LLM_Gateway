---
name: "OPSX: Verify"
description: Contrasta la implementación contra los artefactos del change antes de archivar
category: Workflow
tags: [workflow, verificación, experimental]
---

## Objetivo

Antes de archivar un change, comprobar que lo implementado realmente cumple lo que dicen sus artefactos (specs, tareas, design) — en tres dimensiones: completitud, corrección y coherencia.

## Entrada

`/opsx:verify [nombre-cambio]`. Si se omite, intenta inferirlo del contexto; si sigue ambiguo, pide selección explícita (ver paso 1). Nunca adivines ni auto-selecciones.

## Procedimiento

### 1. Elegir el change a verificar

Si no llegó nombre: `openspec list --json`, y usa **AskUserQuestion** para ofrecer los changes que ya tengan tareas de implementación. Muestra el schema de cada uno y marca "(En progreso)" los que tengan tareas incompletas.

### 2. Diagnosticar el schema activo

```bash
openspec status --change "<nombre-cambio>" --json
```

Interesa `schemaName` (workflow en uso) y qué artefactos existen para este change.

### 3. Cargar los artefactos

```bash
openspec instructions apply --change "<nombre-cambio>" --json
```

De la respuesta interesan los `contextFiles` — la lista real de artefactos a considerar (proposal, specs, design, tasks u otros según el schema). Léelos todos antes de evaluar nada.

### 4. Armar el esqueleto del reporte

Tres dimensiones a evaluar, cada una con hallazgos de severidad CRÍTICO / ADVERTENCIA / SUGERENCIA:

- **Completitud** — cobertura de tareas y de specs
- **Corrección** — si los requisitos y escenarios están realmente implementados
- **Coherencia** — si el código sigue las decisiones de diseño y los patrones del proyecto

### 5. Evaluar completitud

**Tareas**: si existe `tasks.md`, léelo y cuenta checkboxes `- [ ]` vs `- [x]`. Cada tarea incompleta es un hallazgo CRÍTICO con recomendación puntual ("Completar tarea: <descripción>" o "Marcar como hecha si ya está implementada").

**Cobertura de specs**: si hay delta-specs en `openspec/changes/<nombre-cambio>/specs/`, extrae cada requisito (`### Requirement:`) y busca en el código evidencia de que existe. Requisito sin rastro → hallazgo CRÍTICO ("Requisito no encontrado: <nombre>") con recomendación de implementarlo.

### 6. Evaluar corrección

**Mapeo requisito → implementación**: por cada requisito de los delta-specs, busca en el código evidencia (archivo y rango de líneas) y evalúa si coincide con la intención del requisito. Divergencia → ADVERTENCIA ("La implementación podría divergir del spec: <detalle>") con recomendación de revisar `archivo:líneas` contra el requisito.

**Cobertura de escenarios**: por cada escenario (`#### Scenario:`), revisa si el código lo maneja y si hay test que lo cubra. Escenario sin cobertura → ADVERTENCIA con recomendación de agregar test o implementación.

### 7. Evaluar coherencia

**Fidelidad al diseño**: si existe `design.md`, extrae las decisiones clave (secciones tipo "Decisión:", "Enfoque:", "Arquitectura:") y verifica que la implementación las respete. Contradicción → ADVERTENCIA ("Decisión de diseño no seguida: <decisión>"). Sin `design.md`, salta este chequeo y anótalo.

**Consistencia de patrones**: revisa que el código nuevo respete convenciones de nombres, estructura de carpetas y estilo del proyecto. Desviación notoria → SUGERENCIA con ejemplo del patrón esperado.

### 8. Redactar el reporte final

```
## Reporte de verificación: <nombre-cambio>

### Resumen
| Dimensión    | Estado                |
|--------------|-----------------------|
| Completitud  | X/Y tareas, N reqs    |
| Corrección   | M/N reqs cubiertos    |
| Coherencia   | Seguida/Con hallazgos |
```

Agrupa los hallazgos por severidad (CRÍTICO primero, luego ADVERTENCIA, luego SUGERENCIA), cada uno con su recomendación accionable.

**Veredicto final**:
- Con CRÍTICOS: "X hallazgo(s) crítico(s). Resolver antes de archivar."
- Solo advertencias: "Sin hallazgos críticos. Y advertencia(s) a considerar. Listo para archivar (con mejoras señaladas)."
- Todo en orden: "Todos los chequeos pasaron. Listo para archivar."

## Heurísticas de verificación

- **Completitud**: apóyate en lo objetivamente verificable (checkboxes, lista de requisitos).
- **Corrección**: búsqueda de palabras clave, análisis de rutas de archivo, inferencia razonable — no exijas certeza perfecta.
- **Coherencia**: señala inconsistencias evidentes, no critiques estilo menor.
- **Ante la duda**: preferí SUGERENCIA sobre ADVERTENCIA, y ADVERTENCIA sobre CRÍTICO.
- **Accionabilidad**: todo hallazgo lleva recomendación concreta, con referencia a archivo/línea cuando aplique.

## Degradación elegante

- Solo `tasks.md`: verifica nada más que completitud de tareas.
- Tareas + specs: completitud y corrección, sin coherencia.
- Artefactos completos: las tres dimensiones.
- Siempre deja constancia de qué chequeo se omitió y por qué.

## Límites

- Usa markdown claro: tabla de resumen, listas agrupadas por severidad.
- Referencia código como `archivo.ts:123`.
- Nunca una recomendación vaga tipo "considerar revisar" — siempre específica y accionable.
