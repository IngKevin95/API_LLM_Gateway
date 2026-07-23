---
name: "OPSX: Onboard"
description: Recorrido guiado — completa un ciclo entero de OpenSpec narrando cada etapa
category: Workflow
tags: [workflow, onboarding, tutorial, aprendizaje]
---

## Objetivo

Llevar a la persona por su primer ciclo completo de OpenSpec: de la idea a la implementación, usando una tarea real de su propio repositorio. Es una clase práctica — se hace trabajo real mientras se explica cada paso.

## Requisito previo

```bash
openspec status --json 2>&1 || echo "NOT_INITIALIZED"
```

Si no está inicializado, detente aquí:
> OpenSpec todavía no está configurado en este proyecto. Corre `openspec init` primero y vuelve a `/opsx:onboard`.

## Patrón narrativo

Cada etapa clave sigue: **Contexto** (explicar el concepto) → **Acción** (ejecutar de verdad en el repo) → **Resultado** (mostrar lo producido) → **Pausa** (esperar confirmación antes de seguir). No todas las etapas necesitan las cuatro partes — algunas son solo Contexto + Acción.

## Etapa 1 — Bienvenida

```
## ¡Bienvenido/a a OpenSpec!

Vamos a recorrer un ciclo completo de cambio —de la idea a la implementación— usando una tarea real de tu código. En el camino, aprenderás el flujo haciéndolo.

**Qué haremos:**
1. Elegir una tarea pequeña y real en tu código
2. Explorar el problema brevemente
3. Crear un change (el contenedor de nuestro trabajo)
4. Construir los artefactos: proposal → specs → design → tasks
5. Implementar las tareas
6. Archivar el change completado

**Tiempo estimado:** 15-20 minutos

Empecemos por encontrar algo en qué trabajar.
```

## Etapa 2 — Selección de tarea

**Análisis del repositorio.** Busca oportunidades pequeñas de mejora:

1. Comentarios `TODO`, `FIXME`, `HACK`, `XXX` en archivos de código.
2. Manejo de errores ausente — bloques `catch` que tragan errores, operaciones riesgosas sin try-catch.
3. Funciones sin tests — cruza `src/` contra los directorios de test.
4. Tipos débiles — `any` en TypeScript (`: any`, `as any`).
5. Artefactos de debug — `console.log`, `console.debug`, `debugger` fuera de código de depuración.
6. Validación de entrada ausente en manejadores de input de usuario.

Revisa también actividad reciente de git:
```bash
git log --oneline -10 2>/dev/null || echo "No git history"
```

Presenta 3-4 sugerencias concretas:

```
## Sugerencias de tarea

Escaneando tu repositorio, encontré estos buenos puntos de partida:

**1. [Tarea más prometedora]**
   Ubicación: `src/ruta/archivo.ts:42`
   Alcance: ~1-2 archivos, ~20-30 líneas
   Por qué es buena: [razón breve]

**2. [Segunda tarea]**
   Ubicación: `src/otro/archivo.ts`
   Alcance: ~1 archivo, ~15 líneas
   Por qué es buena: [razón breve]

**3. [Tercera tarea]**
   Ubicación: [ubicación]
   Alcance: [estimado]
   Por qué es buena: [razón breve]

**4. ¿Otra cosa?**
   Cuéntame qué te gustaría trabajar.

¿Cuál te interesa? (Elige un número o describe la tuya)
```

Si no se encuentra nada obvio: pregunta directamente qué le gustaría construir o arreglar.

**Contención de alcance.** Si la persona elige algo demasiado grande (feature mayor, trabajo de varios días):

```
Esa tarea vale la pena, pero probablemente es más grande de lo ideal para tu primer recorrido de OpenSpec.

Para aprender el flujo, más pequeño es mejor — te deja ver el ciclo completo sin perderte en detalles de implementación.

**Opciones:**
1. **Recortarla** — ¿cuál sería el pedazo más pequeño y útil de [tu tarea]? ¿Quizás solo [recorte específico]?
2. **Elegir otra** — alguna de las otras sugerencias, o una tarea distinta más chica.
3. **Hacerla igual** — si de verdad quieres encararla, podemos. Solo ten en cuenta que tomará más tiempo.

¿Qué prefieres?
```

Es una barrera blanda: si insiste, respeta su elección.

## Etapa 3 — Demostración de exploración

Una vez elegida la tarea, muestra brevemente el modo exploración:

```
Antes de crear un change, déjame mostrarte rápido el **modo exploración** — así es como piensas los problemas antes de comprometerte con una dirección.
```

Dedica 1-2 minutos a investigar el código relevante: léelo, dibuja un diagrama ASCII si ayuda, anota consideraciones.

```
## Exploración rápida

[Tu análisis breve — qué encontraste, qué consideraciones surgen]

┌─────────────────────────────────────────┐
│   [Opcional: diagrama ASCII si ayuda]   │
└─────────────────────────────────────────┘

El modo exploración (`/opsx:explore`) es justamente para este tipo de pensamiento — investigar antes de implementar. Puedes usarlo cuando necesites pensar un problema.

Ahora sí, creemos un change para contener nuestro trabajo.
```

**Pausa** — espera reconocimiento antes de continuar.

## Etapa 4 — Crear el change

**Contexto:**
```
## Creando un Change

Un "change" en OpenSpec es un contenedor para todo el pensamiento y la planificación alrededor de una pieza de trabajo. Vive en `openspec/changes/<nombre>/` y contiene tus artefactos — proposal, specs, design, tasks.

Voy a crear uno para nuestra tarea.
```

**Acción:**
```bash
openspec new change "<nombre-derivado>"
```

**Resultado:**
```
Creado: `openspec/changes/<nombre>/`

Estructura de la carpeta:
```
openspec/changes/<nombre>/
├── proposal.md    ← Por qué hacemos esto (vacío, lo llenamos ahora)
├── design.md      ← Cómo lo construiremos (vacío)
├── specs/         ← Requisitos detallados (vacío)
└── tasks.md       ← Checklist de implementación (vacío)
```

Ahora llenemos el primer artefacto — la proposal.
```

## Etapa 5 — Proposal

**Contexto:**
```
## La Proposal

La proposal captura **por qué** hacemos este cambio y **qué** implica a alto nivel. Es el "elevator pitch" del trabajo.

Voy a redactar una basada en nuestra tarea.
```

**Acción** — redacta el borrador (sin guardar todavía):

```
Aquí va un borrador de proposal:

---

## Por qué

[1-2 frases explicando el problema/oportunidad]

## Qué cambia

[Puntos de lo que será distinto]

## Capacidades

### Capacidades nuevas
- `<nombre-capacidad>`: [descripción breve]

### Capacidades modificadas
<!-- Si se modifica comportamiento existente -->

## Impacto

- `src/ruta/archivo.ts`: [qué cambia]
- [otros archivos si aplica]

---

¿Esto captura la intención? Puedo ajustar antes de guardarlo.
```

**Pausa** — espera aprobación o feedback.

Tras la aprobación, obtén las instrucciones y guarda:
```bash
openspec instructions proposal --change "<nombre-cambio>" --json
```
Escribe el contenido en `openspec/changes/<nombre-cambio>/proposal.md`.

```
Proposal guardada. Este es tu documento del "por qué" — siempre puedes volver y refinarlo a medida que el entendimiento evoluciona.

Sigue: specs.
```

## Etapa 6 — Specs

**Contexto:**
```
## Specs

Los specs definen **qué** construimos en términos precisos y verificables. Usan un formato de requisito/escenario que deja el comportamiento esperado cristalino.

Para una tarea pequeña como esta, quizás solo necesitemos un archivo de spec.
```

**Acción:**
```bash
mkdir -p openspec/changes/<nombre-cambio>/specs/<nombre-capacidad>
```

Redacta el contenido:

```
Aquí está el spec:

---

## Requisitos AGREGADOS

### Requisito: <Nombre>

<Descripción de lo que el sistema debe hacer>

#### Escenario: <nombre del escenario>

- **CUANDO** <condición disparadora>
- **ENTONCES** <resultado esperado>
- **Y** <resultado adicional si hace falta>

---

Este formato — CUANDO/ENTONCES/Y — hace los requisitos verificables. Casi se leen como casos de prueba.
```

Guarda en `openspec/changes/<nombre-cambio>/specs/<capacidad>/spec.md`.

## Etapa 7 — Design

**Contexto:**
```
## Design

El design captura **cómo** lo construiremos — decisiones técnicas, tradeoffs, enfoque.

Para cambios pequeños, puede ser breve. Está bien — no todo cambio necesita una discusión de diseño profunda.
```

**Acción** — redacta design.md:

```
Aquí está el design:

---

## Contexto

[Contexto breve sobre el estado actual]

## Metas / No-metas

**Metas:**
- [Qué buscamos lograr]

**No-metas:**
- [Qué queda explícitamente fuera de alcance]

## Decisiones

### Decisión 1: [decisión clave]

[Explicación del enfoque y su razón]

---

Para una tarea pequeña, esto captura las decisiones clave sin sobre-ingeniería.
```

Guarda en `openspec/changes/<nombre-cambio>/design.md`.

## Etapa 8 — Tasks

**Contexto:**
```
## Tasks

Por último, desglosamos el trabajo en tareas de implementación — checkboxes que guían la fase de apply.

Deben ser pequeñas, claras y en orden lógico.
```

**Acción** — genera tareas a partir de specs y design:

```
Aquí están las tareas de implementación:

---

## 1. [Categoría o archivo]

- [ ] 1.1 [Tarea específica]
- [ ] 1.2 [Tarea específica]

## 2. Verificar

- [ ] 2.1 [Paso de verificación]

---

Cada checkbox se vuelve una unidad de trabajo en la fase de apply. ¿Listos para implementar?
```

**Pausa** — espera confirmación de que está listo/a para implementar.

Guarda en `openspec/changes/<nombre-cambio>/tasks.md`.

## Etapa 9 — Apply (implementación)

**Contexto:**
```
## Implementación

Ahora implementamos cada tarea, marcándola al terminar. Voy anunciando cada una y ocasionalmente señalo cómo specs/design informaron el enfoque.
```

**Acción** — por cada tarea:

1. Anuncia: "Trabajando en tarea N: [descripción]"
2. Implementa el cambio en el código
3. Referencia specs/design con naturalidad: "El spec dice X, por eso hago Y"
4. Marca completada en tasks.md: `- [ ]` → `- [x]`
5. Estado breve: "✓ Tarea N completa"

Mantén la narración liviana — no sobre-expliques cada línea de código.

Al terminar todas:

```
## Implementación completa

Todas las tareas listas:
- [x] Tarea 1
- [x] Tarea 2
- [x] ...

¡El change está implementado! Un paso más — archivémoslo.
```

## Etapa 10 — Archive

**Contexto:**
```
## Archivando

Cuando un change está completo, lo archivamos. Esto lo mueve de `openspec/changes/` a `openspec/changes/archive/AAAA-MM-DD-<nombre>/`.

Los changes archivados se vuelven el historial de decisiones de tu proyecto — siempre puedes encontrarlos después para entender por qué algo se construyó de cierta manera.
```

**Acción:**
```bash
openspec archive "<nombre-cambio>"
```

**Resultado:**
```
Archivado en: `openspec/changes/archive/AAAA-MM-DD-<nombre>/`

El change ahora es parte del historial de tu proyecto. El código está en tu repositorio, el registro de la decisión está preservado.
```

## Etapa 11 — Cierre y próximos pasos

```
## ¡Felicitaciones!

Acabas de completar un ciclo completo de OpenSpec:

1. **Explore** — pensaste el problema
2. **New** — creaste un contenedor de change
3. **Proposal** — capturaste el POR QUÉ
4. **Specs** — definiste el QUÉ en detalle
5. **Design** — decidiste el CÓMO
6. **Tasks** — lo desglosaste en pasos
7. **Apply** — implementaste el trabajo
8. **Archive** — preservaste el registro

Este mismo ritmo funciona para cambios de cualquier tamaño — un fix chico o una feature grande.

Prueba `/opsx:new` o `/opsx:ff` en algo que realmente quieras construir. ¡Ya tienes el ritmo!
```

## Manejo de salidas anticipadas

**Si la persona quiere parar a mitad de camino** (lo dice explícitamente, pide pausar, o parece desconectada):

```
¡Sin problema! Tu change quedó guardado en `openspec/changes/<nombre>/`.

Para retomarlo después:
- `/opsx:continue <nombre>` — retomar creación de artefactos
- `/opsx:apply <nombre>` — saltar directo a implementación (si ya hay tasks)

El trabajo no se pierde. Vuelve cuando quieras.
```

Sal con gracia, sin presión.

**Si solo quiere la referencia de comandos** (pide ver los comandos o saltarse el tutorial):

```
## Referencia rápida de OpenSpec

| Comando | Qué hace |
|---------|----------|
| `/opsx:explore` | Pensar problemas (sin cambios de código) |
| `/opsx:new <nombre>` | Nuevo change, paso a paso |
| `/opsx:ff <nombre>` | Fast-forward: todos los artefactos de una |
| `/opsx:continue <nombre>` | Continuar un change existente |
| `/opsx:apply <nombre>` | Implementar tareas |
| `/opsx:verify <nombre>` | Verificar implementación |
| `/opsx:archive <nombre>` | Archivar al terminar |

Prueba `/opsx:new` para tu primer change, o `/opsx:ff` si quieres ir rápido.
```

## Límites

- Sigue el patrón Contexto → Acción → Resultado → Pausa en las transiciones clave (tras explorar, tras el borrador de proposal, tras las tasks, tras archivar).
- Mantén la narración liviana durante la implementación — enseña sin dar cátedra.
- No te saltes etapas aunque el change sea chico — el objetivo es enseñar el flujo completo.
- Pausa para reconocimiento en los puntos marcados, pero sin pausar de más.
- Maneja las salidas con gracia — nunca presiones para continuar.
- Usa tareas reales del repositorio — no simules ni uses ejemplos ficticios.
- Ajusta el alcance con suavidad — guía hacia tareas más chicas pero respeta la elección de la persona.
