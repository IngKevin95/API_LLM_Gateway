---
name: factory-flujo-cc
description: "Meta-skill iterativa que documenta un cambio o funcionalidad evolutiva: lee la documentación existente en `docs/`, hace preguntas de discovery, genera épicas nuevas (marcadas como evolutivas), escribe historias + AC, mapea relaciones con lo existente, y produce un plan de construcción listo para pasar a `/building-a-slice`. Orquesta el flujo de discovery evolutivo (contexto → descubrimiento → descomposición → AC → plan). Se activa cuando hay documentación base ya lista pero una nueva funcionalidad/cambio necesita discovery, especificación y planeación."
category: Discovery
tags: [meta, evolutivo, change-control, end-to-end, recovery, factory]
---

# factory-flujo-cc

Control de cambios evolutivo. Lee lo existente, descubre la nueva funcionalidad, genera épicas marcadas como `tipo: evolutivo`, y produce un plan de construcción con trazabilidad explícita hacia lo que ya existe.

No genera ningún artefacto por sí misma: orquesta las skills de escritura (historias, AC, descomposición) en orden, decide cuándo el usuario debe revisar, y persiste estado para poder retomar.

## El pipeline que recorre

```
1. Ingesta: descripción de nueva funcionalidad
   └─ Leer docs/ completo para contexto existente

2. Discovery iterativo (agente: change-discovery-agent)
   └─ Itera preguntas: ¿quiénes, qué duele, impacto, dependencias, riesgos?
      → Output: brief descubierto

3. Descomposición en épicas evolutivas (skill: factory-descomponer-cambio-a-epicas)
   └─ Reutiliza lógica de factory-descomponer-prd-a-epicas
      → docs/03-backlog/epicas-evolutivas.md (IDs: EP-EVO-XXX)
      → Validación: change-decomposer-auditor

4. (loop) Historias + AC por épica
   └─ factory-escribir-historia-usuario (HU-EVO-XXX) — REUTILIZA SKILL EXISTENTE
      + factory-escribir-criterios-aceptacion-bdd — REUTILIZA SKILL EXISTENTE
      → docs/04-historias/HU-EVO-*.md
      → Validación: invest-validator (existente), bdd-validator (existente)

5. Mapeo de relaciones (agentes: change-impact-analyzer + change-risk-detector)
   └─ change-impact-analyzer: analiza qué épicas/historias existentes se tocan
      → docs/07-cambios/CC-001-<descripcion>.md (reporte de impacto)
      → Valida trazabilidad: reutiliza trazabilidad-auditor
   └─ change-risk-detector: detecta riesgos de cambios en producción
      → Añade sección "## Riesgos de cambios en producción" al reporte de impacto
      → Frena si riesgo ALTO (requiere confirmación explícita del usuario)

6. Plan de construcción (agente: construction-sequencer)
   └─ Ordena slices por dependencias, estima tiempo, identifica paralelización
      → docs/07-cambios/PLAN-CC-001.md (qué slices construir, en qué orden)
```

Pasos 1-4 obligatorios y secuenciales. Paso 5-6 refinados iterativamente.

## Qué necesita antes de arrancar

- Una **descripción breve** de la nueva funcionalidad (lenguaje natural).
- Acceso lectura a `docs/` existente para entender contexto.
- Opcionalmente: lista de EP/HU existentes que pueden verse afectadas (si el usuario la tiene).

## Estado persistente y reanudación

Tras cada paso importante, escribe `docs/.flow-cc-state.json`:

```json
{
  "last_completed_step": "epicas-evolutivas",
  "status": "in-progress",              // NEW: in-progress | blocked | completed
  "failed_attempts": 0,                 // NEW: contador para recovery
  "change_id": "CC-001",
  "description": "Añadir notificaciones por email",
  "params": {"scope": "users que no han activado 2FA", "impact": "high"},
  "timestamp": "2026-07-23T15:30:00Z",
  "epicas_generadas": ["EP-EVO-001", "EP-EVO-002"],
  "historias_pending": ["HU-EVO-001", "HU-EVO-002", "HU-EVO-003"]
}
```

Si el archivo existe al invocar:
1. Si `status: "completed"` → ofertar limpiar y empezar nuevo CC.
2. Si `status: "in-progress"` + `timestamp > 24h` → sugerir abortar (sesión vieja).
3. Si `status: "blocked"` + `failed_attempts > 2` → ofertar reiniciar desde cero como default.
4. Si `status: "in-progress"` + `timestamp < 24h` → preguntar reanudar vs reiniciar.

Nunca asumir.

## Cómo funciona cada checkpoint

Al cerrar un paso:

1. Se muestra el artefacto generado — resumen, path, cambios clave.
2. Se reporta veredicto del revisor (si corrió).
3. Se pregunta: **seguir / corregir / refinar / abortar**.
4. `seguir` avanza al próximo paso. `corregir` reinvoca la skill actual con feedback. `refinar` permite iterar sin avanzar (cambiar detalles). `abortar` persiste estado y cierra.

## Procedimiento paso a paso

1. **Detectar estado previo**
   - ¿Existe `.flow-cc-state.json`? → preguntar reanudar/reiniciar.

2. **Ingesta + contexto** (2-5 min)
   - Pregunta: "¿Describe brevemente la nueva funcionalidad?"
   - Lee `docs/` (PRD, épicas, historias, AC existentes).
   - Resumen al usuario: "Entiendo que el proyecto ya tiene [X feature], y ahora necesitas [Y]".

3. **Discovery iterativo — delegado a `change-discovery-agent`** (15-25 min)
   - Agent itera preguntas:
     - **Quiénes se benefician**? (nuevos usuarios, usuarios existentes)
     - **Qué problema resuelve** que hoy no se resuelve?
     - **Qué puede romperse** o cambiar en lo existente?
     - **Dependencias técnicas** — BD, API, UI, integraciones?
     - **Timeline o restricciones**?
     - **Success metrics** — cómo medimos si funciona?
   - Agent valida coherencia (¿choca con non-goals del PRD?)
   - Output: brief descubierto en markdown.
   - Checkpoint: usuario confirma o pide iteración ("refina X").

4. **Descomposición en épicas evolutivas — delegado a `factory-descomponer-cambio-a-epicas`** (15-30 min)
   - Skill invoca agente `change-decomposer-auditor` automáticamente.
   - Output: `docs/03-backlog/epicas-evolutivas.md` con EP-EVO-XXX:
     ```yaml
     id: EP-EVO-001
     titulo: Backend notificaciones email 2FA
     tipo: evolutivo
     extiende: [EP-003]
     impacta: [HU-045, HU-046]
     riesgo: medio
     fecha_estimada: 2026-08-15
     ```
   - Auditor valida: estructura, referencias, granularidad.
   - Si PASS: "Épicas validadas, continuamos".
   - Si FAIL o WARN: Detener, mostrar feedback, permitir iteración.
   - Checkpoint: usuario aprueba o pide refinar.

5. **Historias + AC — REUTILIZA skills existentes** (30-60 min, en loop)
   - Por cada EP-EVO, invocar en secuencia:
     - `factory-escribir-historia-usuario` (HU-EVO-XXX) — skill existente SIN CAMBIOS
     - `factory-escribir-criterios-aceptacion-bdd` — skill existente SIN CAMBIOS
   - Validadores automáticos (existentes):
     - `invest-validator` — valida INVEST sobre HU-EVO.
     - `bdd-validator` — valida G/W/T sobre AC.
   - Output: `docs/04-historias/HU-EVO-*.md` con frontmatter + AC.
   - Checkpoint por lote: "¿Estas historias están OK, o refino alguna?"

6. **Mapeo de relaciones — delegado a `change-impact-analyzer` + `change-risk-detector`** (15-25 min)
   - `change-impact-analyzer` analiza impacto:
     - ¿Qué épicas/historias existentes se tocan directamente?
     - ¿Qué impacto indirecto hay?
     - ¿Dependencias técnicas?
     - ¿Riesgos de regresión?
     - ¿Choca con non-goals?
   - `change-risk-detector` valida riesgos en producción:
     - ¿La épica extendida está en producción?
     - ¿Las HU impactadas tienen AC maduras?
     - ¿Hay cobertura de tests?
   - Valida trazabilidad con `trazabilidad-auditor` (existente).
   - Output: `docs/07-cambios/CC-001-<descripcion>.md` con sección "## Riesgos de cambios en producción".
   - **Checkpoint crítico**: Si hay riesgo ALTO, frena el flujo → requiere confirmación explícita del usuario antes de continuar.

7. **Plan de construcción — delegado a `construction-sequencer`** (15-20 min)
   - Agent ordena slices por dependencias, estima tiempo, identifica paralelización.
   - Output: `docs/07-cambios/PLAN-CC-001.md`:
     - Secuencia de slices (Slice 1, Slice 2, Slice 2b paralelo, Slice 3).
     - Dependencias explícitas entre slices.
     - Criterios de verificación por slice.
     - Estimación (días) + buffer.
     - Critical path + riesgos.
   - Checkpoint: "¿El plan es realista?"

8. **Cierre**
   - Resumen ejecutivo:
     - Épicas evolutivas generadas (cantidad y nombres).
     - Historias totales (nuevas + existentes impactadas).
     - Plan de construcción (duración, riesgos).
     - Siguiente paso: "Pasar a `/building-a-slice EP-EVO-001` para construir".
   - Sugerir auditoría final: `/factory:revisar` (skill `revisar-calidad-documental`, existente).
   - Marcar `.flow-cc-state.json` con `last_completed_step: done` o eliminar.

## Reglas no negociables

1. **Siempre leer lo existente primero**: sin eso no hay trazabilidad ni relaciones claras.
2. **Cada épica evolutiva DEBE tener `extiende` y `impacta` explícitos**: no flotar huérfanos en el aire.
   - Si EP-EVO no extiende nada, no es evolutivo → reconceptualizar como nuevo.
   - Si EP-EVO no impacta nada, es aislado → validar que es intencional.
3. **Las historias EVO heredan el formato INVEST del flujo normal**: rol específico, beneficio externo, AC verificables en G/W/T.
4. **Reutilizar skills existentes SIN MODIFICAR**: historias se escriben con `factory-escribir-historia-usuario`, AC con `factory-escribir-criterios-aceptacion-bdd`. Solo el ID cambia (HU-EVO en lugar de HU).
5. **Reutilizar validadores existentes**: `invest-validator`, `bdd-validator`, `trazabilidad-auditor` se invocan automáticamente, nada nuevo.
6. **Un issue bloqueante detiene el avance**: impacto crítico sin mitigación → no se genera plan sobre cambios peligrosos.
7. **El estado se persiste siempre**, incluso si se aborta a mitad de camino.
8. **Un riesgo ALTO detectado por `change-risk-detector` requiere confirmación explícita del usuario**: nunca se asume "seguir". Si el cambio toca código en producción con riesgo no mitigable, el usuario debe revisar, reconocer el riesgo, y confirmar "proceder igual" antes de continuar al plan de construcción.

## A quién le entrega el resultado

Al completar: sugerir verificación visual con `/factory:revisar` enfocado en coherencia de cambios existentes.

El usuario pasa después directo a `/building-a-slice` con las EP-EVO generadas.

## Qué evitar

- ❌ **No leer contexto existente**: genera épicas huérfanas o que pisan historias actuales.
- ❌ **Saltarse el mapeo de relaciones**: no se ve el riesgo de regresión hasta que es tarde.
- ❌ **Generar historias EVO sin trazabilidad a lo viejo**: imposible auditar impacto.
- ❌ **Suprimir checkpoints para "ir rápido"**: justamente acá es donde el usuario puede frenar un cambio peligroso.
- ❌ **Mezclar PR definitivo con discovery**: esta skill produce especificación, no implementación.
