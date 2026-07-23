# Reporte de Coherencia Documental — 2026-07-23

## 🔴 CRÍTICO: Inconsistencias Detectadas

---

## 1. Épicas Declaradas vs. Épicas en Backlog

| Épica | En epicas.md | En backlog.md | Historias Anticipadas | Historias en Backlog | Status |
|---|---|---|---|---|---|
| EP-001 | ✅ | ✅ | HU-001, 002a, 002b, 003, 035 | Sí (HU-001-003, HU-035) | ✅ OK |
| EP-002 | ✅ | ✅ | HU-004a-c, 005, 020a-c, 021a-b, 024 | Sí (10 HU) | ✅ OK |
| EP-003 | ✅ | ✅ | HU-006, 007 | Sí (2 HU) | ✅ OK |
| EP-004A | ✅ | ✅ | HU-008, 009, 022, 022b, 025a, 025b, 027 | Sí (7 HU) | ✅ OK |
| EP-004B | ✅ | ✅ | HU-010, 011, 026a-b, 028, 034 | Sí (6 HU) | ✅ OK |
| EP-005 | ✅ | ✅ | HU-012a-c, 013, 016, 033 | Sí (6 HU) | ✅ OK |
| EP-006 | ✅ (Deprecated) | ❌ | — | — | ✅ OK (esperado) |
| EP-007 | ✅ | ✅ | HU-017, 018, 019, 023, 032 | Sí (5 HU) | ✅ OK |
| EP-008 | ✅ | ✅ | HU-029, 030, 031, 036 | Sí (4 HU) | ✅ OK |
| **EP-009** | ✅ | ❌ **FALTA** | HU-038, 039, 040, 041 | ❌ **NO** | 🔴 **BLOQUEADOR** |
| **EP-010** | ✅ (2 definiciones) | ✅ (parcial) | HU-042-048 (viejo) + HU-050-060 (nuevo) | Mezcla confusa | 🔴 **CRÍTICO** |

---

## 2. Historias Huérfanas (Existen en docs/04-historias/ pero NO en backlog.md)

### EP-009 (Completamente ausente de backlog)
```
HU-038 — Implementar Sync Worker (existente en docs/04-historias/)
HU-039 — Write-Ahead Log (WAL) local (existente en docs/04-historias/)
HU-040 — Graceful Shutdown (existente en docs/04-historias/)
HU-041 — Cache Invalidator (existente en docs/04-historias/)
```

**Impacto**: EP-009 está documentada en epicas.md pero sus 4 historias NO aparecen en backlog.md. Esto rompe la trazabilidad bidireccional.

---

### EP-010 Ambigüedad Crítica

**Problema**: EP-010 aparece con dos conjuntos de historias:

#### Set A (Historias previas — buildstate activo):
```
HU-042 — Routing automático por capability
HU-043 — Endpoint responses OpenCode
HU-044 — Parámetros OpenAI completos
HU-045 — Parámetros Anthropic completos
HU-046 — Endpoint /models con metadata
HU-047 — Middleware normalización formatos
HU-048 — Documentación y configuración
```

**Frontmatter**: todas dicen `epica: EP-010`  
**Estado en backlog.md**: ❌ NO APARECEN

#### Set B (Nuevas historias agregadas hoy):
```
HU-050 — Logging OpenAI handler
HU-051 — Debug ProcessChat()
HU-052 — Validar Router.Route()
HU-053 — OmniRoute adapter
HU-054 — Registrar OmniRoute
HU-055 — Test OmniRoute
HU-056 — Alinear IDs proveedores
HU-057 — ProcessEmbedding()
HU-058 — Fix Anthropic /v1/messages
HU-059 — Logging estructurado
HU-060 — Métricas /metrics
```

**Frontmatter**: todas dicen `epica: EP-010`  
**Estado en backlog.md**: ✅ SÍ APARECEN (órdenes 46-56)

---

## 3. Conflicto de Definición de EP-010

### En epicas.md (línea 275):
```markdown
## EP-010 · MVP Fixes & Completeness (Observabilidad y Compatibilidad Universal)

**Histórico anticipadas**:
- HU-050 — Logging OpenAI handler
- HU-051 — Debuguear ProcessChat()
...
```

### En build-state.json (del summary anterior):
```
Slice activo: EP-010 [HU-042, HU-043, HU-044, HU-045, HU-046, HU-047, HU-048]
change=compatibilidad-universal-clientes
fase=tdd
```

**Contradicción**: 
- `epicas.md` define EP-010 como "MVP Fixes & Completeness" (HU-050-060)
- `build-state.json` define EP-010 como "compatibilidad universal clientes" (HU-042-048)

Esto es un **conflicto fundamental** sobre qué es EP-010.

---

## 4. Conteo Total de Historias

| Rango | Épica | En backlog.md | En docs/04-historias/ | Discrepancia |
|---|---|---|---|---|
| HU-001-005 | EP-001 | 3 | 3 | ✅ OK |
| HU-006-035 | EP-002 a EP-008 | 40 | 40 | ✅ OK |
| HU-038-041 | EP-009 | ❌ 0 | ✅ 4 | 🔴 FALTA |
| HU-042-048 | EP-010 (viejo) | ❌ 0 | ✅ 7 | 🔴 FALTA |
| HU-050-060 | EP-010 (nuevo) | ✅ 11 | ✅ 11 | ✅ OK |

**Total esperado**: 45 (pre-EP-009/010) + 4 (EP-009) + 7 (EP-010 viejo) + 11 (EP-010 nuevo) = **67 historias**  
**Total en backlog.md**: 56 historias (faltan HU-038-041, HU-042-048)

---

## 5. Trazabilidad PRD ↔ Épica

### ✅ Cobertura de Objetivos (Correcta según epicas.md)
```
Obj. 1 → EP-001 (Desacople)
Obj. 2 → EP-002, EP-007, EP-008, EP-009 (Resiliencia)
Obj. 3 → EP-001, EP-007, EP-010 (Selección óptima + observabilidad)
Obj. 4 → EP-003, EP-004A, EP-004B, EP-009, EP-010 (Seguridad + gobernanza)
Obj. 5 → EP-005, EP-010 (Compatibilidad universal)
```

### ⚠️ Pero: Si EP-010 solo contiene HU-050-060, entonces:
- EP-009 NO contribuye a su Obj. 4 (auditoría) porque no está en backlog
- EP-010 (viejo, HU-042-048) desaparece de la trazabilidad

---

## 6. Validaciones Fallidas

| Validación | Resultado | Detalles |
|---|---|---|
| **Épica → Historias**: Toda épica en epicas.md tiene historias en backlog | ❌ FALLA | EP-009: 0/4 historias; EP-010 (HU-042-048): 0/7 historias |
| **Historiashuérfanas**: No hay historias sin épica en backlog | ✅ PASA | Todas las HU en backlog tienen épica |
| **Épica huérfana**: No hay épica sin historias | ❌ FALLA | EP-009 existe pero 0 historias en backlog |
| **Frontmatter coherencia**: Épica en HU ↔ Épica en backlog | ⚠️ PARCIAL | HU-038-041 (EP-009), HU-042-048 (EP-010) en frontmatter pero NO en backlog |
| **Bidireccionalidad PRD ↔ Épica ↔ Historias** | ❌ FALLA | Hay épicas sin historias en backlog |

---

## 7. Matriz de Coherencia (Resumen)

```
epicas.md ────→ backlog.md ────→ docs/04-historias/

EP-001 ──→ Sí ──→ HU-001-003, HU-035 ✅ OK
EP-002 ──→ Sí ──→ HU-004-005, HU-020-021, HU-024 ✅ OK
...
EP-008 ──→ Sí ──→ HU-029-031, HU-036 ✅ OK
EP-009 ──→ NO ──→ [Nada] ❌ FALTA EN BACKLOG
EP-010 ──→ Sí ──→ HU-050-060 (pero HU-042-048 también existen) 🔴 CONFLICTO
```

---

## 8. Recomendaciones de Remediación

### **OPCIÓN A: Restaurar el estado original (build-state.json)**
```
1. EP-010 = HU-042-048 (compatibilidad universal clientes)
2. Agregar HU-042-048 a backlog.md (órdenes 45-52)
3. Agregar HU-050-060 como EP-011 (nueva épica de MVP Fixes)
4. Agregar HU-038-041 a backlog.md (EP-009, órdenes 37-41)
5. Reescribir epicas.md para que EP-010 = HU-042-048, crear EP-011 = HU-050-060
```

### **OPCIÓN B: Mantener lo que hiciste (nuestra versión)**
```
1. EP-010 = HU-050-060 (MVP Fixes & Completeness)
2. Cambiar HU-042-048 frontmatter: epica → EP-010-PREVIO (deprecated/archive)
3. Agregar HU-038-041 a backlog.md (EP-009, órdenes 37-41)
4. Agregar HU-042-048 a backlog.md como EP-010-PREVIO (órdenes 43-50) o ignorarlas
5. Actualizar epicas.md para remover HU-042-048 de EP-010, agregar HU-050-060 solo
```

---

## 9. Matriz de Decisión

| Opción | Impacto | Esfuerzo | Coherencia |
|---|---|---|---|
| A (Restaurar) | Retroceder avance; conflicto con build-state | Bajo (revert) | ✅ Perfecta (estado previo) |
| B (Mantener) | Romper continuidad con build-state; rename masivo | Alto (fixes múltiples) | ⚠️ Nueva coherencia si se aplica completo |

---

## 10. Estado Final Recomendado

**Recomendación: OPCIÓN A (Restaurar coherencia)**

Razón: El `build-state.json` es el registry canónico del arnés Factory. Mantenerlo coherente es prerequisito para poder avanzar.

```
Cambios requeridos:

1. Cambiar HU-050-060 frontmatter: epica: EP-010 → epica: EP-011
2. Crear sección EP-011 en epicas.md (copia de la actual EP-010, renombre a EP-011)
3. Restaurar EP-010 en epicas.md: historias = HU-042-048 (descommentar, copiar de backup o validar)
4. Agregar a backlog.md:
   - HU-038-041 (EP-009, órdenes 37-41)
   - HU-042-048 (EP-010, órdenes 43-50)
5. Ejecutar validaciones finales:
   - trazabilidad-auditor
   - story-decomposer-auditor
```

---

## Conclusión

**Estado Actual**: 🔴 **NO COHERENTE**

**Bloqueadores**:
1. EP-009 tiene 4 historias documentadas pero NO en backlog
2. EP-010 tiene 2 definiciones en conflicto (HU-042-048 vs HU-050-060)
3. HU-042-048 existen en docs/04-historias/ pero NO en backlog.md

**Próximo Paso**: Elegir Opción A o B, aplicar remediaciones, re-validar.
