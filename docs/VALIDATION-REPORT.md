# Reporte de Validación de Artefactos — 2026-07-23

## Estado de Alineación

### 1. PRD vs Épicas

**PRD Objetivos** (docs/01-prd/api-llm-gateway.md):
1. Obj.1 — Desacople total (0 refs a proveedor/modelo en agentes)
2. Obj.2 — Resiliencia (failover transparente; ≥99% éxito con ≥1 proveedor sano)
3. Obj.3 — Selección óptima (overhead routing < 50ms)
4. Obj.4 — Seguridad (100% auth + authz + auditoría)
5. Obj.5 — Compatibilidad universal (OpenAI/Anthropic-compat)

**Épicas Existentes** (docs/03-backlog/epicas.md):
- ✓ EP-001 · Enrutamiento por capacidad → Obj.1, Obj.3
- ✓ EP-002 · Resiliencia y Conectividad Base → Obj.2
- ✓ EP-003 · Gobernanza de consumo → Obj.4
- ✓ EP-004A · Identidad y Accesos → Obj.4
- ✓ EP-004B · Seguridad, Protección y Auditoría → Obj.4
- ✓ EP-005 · API universal compatible → Obj.5
- ✓ EP-007 · Observabilidad y aprendizaje → Obj.3, Obj.4
- ✓ EP-008 · Ecosistema de Adaptadores Secundarios → Obj.2
- ✓ EP-009 · Sincronización Asincronista → Obj.2

**Estado**: Épicas 001-009 OK ✓

---

### 2. EP-010 Status

**Archivo en .claude/state/**: EP-010-ARCHIVAL.md, EP-010-COMMITS-PLAN.md, EP-010-REVISION-DOCUMENTAL.md

**Hallazgo**: ⚠️ EP-010 NO está en `epicas.md`, pero tiene archivos en harness.
- Parece que EP-010 fue completado/archivado
- NO está formalmente documentado en `docs/03-backlog/epicas.md`

**Acción Requerida**: Revisar qué era EP-010 y si debe agregarse formalmente.

---

### 3. EP-011 a EP-017 (Mis épicas nuevas)

**Status**: ❌ **CRÍTICO**

Las épicas que acabo de crear (EP-011 a EP-017) tienen:
- ✓ Historias documentadas (HU-050 a HU-060) en `docs/04-historias/`
- ✓ AC en formato Given/When/Then
- ✓ Trazabilidad HU → Épica
- ❌ **NO ESTÁN** en `docs/03-backlog/epicas.md` (falta trazabilidad Épica → PRD)
- ❌ **NO ESTÁN** en `docs/03-backlog/backlog.md`
- ❌ **NO ESTÁN** en `docs/05-priorizacion/` (no priorizadas formalmente)

| Épica | Descripción | Está en epicas.md? | Está en backlog.md? | Priorizada? |
|---|---|---|---|---|
| EP-011 | Debug `/v1/chat/completions` | ❌ | ❌ | ❌ |
| EP-012 | Adaptador OmniRoute | ❌ | ❌ | ❌ |
| EP-013 | Normalizar IDs proveedores | ❌ | ❌ | ❌ |
| EP-014 | Implementar `/v1/embeddings` | ❌ | ❌ | ❌ |
| EP-015 | Fix `/v1/messages` (Anthropic) | ❌ | ❌ | ❌ |
| EP-016 | Logging estructurado | ❌ | ❌ | ❌ |
| EP-017 | Métricas reales | ❌ | ❌ | ❌ |

---

### 4. Trazabilidad Épicas → PRD Objetivos

Mis épicas (EP-011 a EP-017) deberían mapear a:

| Épica | Objetivos PRD | Mapeo |
|---|---|---|
| EP-011 | Obj.3, Obj.5 | Observabilidad de routing + API compatibility (chat endpoint) |
| EP-012 | Obj.2 | Resiliencia (OmniRoute como fallback local) |
| EP-013 | Obj.1, Obj.3 | Desacople (normalización de IDs) |
| EP-014 | Obj.5 | Compatibilidad universal (OpenAI embeddings endpoint) |
| EP-015 | Obj.5 | Compatibilidad universal (Anthropic endpoint) |
| EP-016 | Obj.4 | Gobernanza (auditoría via logs) |
| EP-017 | Obj.4, Obj.3 | Gobernanza (métricas) + selección óptima (p50/p95/p99 latencia) |

**Recomendación**: Estas épicas NO son nuevas líneas de trabajo, son **fixes a MVP incompleto**. Deberían ser **sub-slices de EP-005** (API universal compatible) + **HU faltantes en EP-002**.

---

### 5. ¿Dónde deberían estar?

**Opción A**: Agregar EP-011 a EP-017 como épicas formales en `epicas.md`
- Requiere: trazabilidad explícita a PRD
- Requiere: agregación en `backlog.md`
- Requiere: priorización formal

**Opción B** (Recomendado): Estas son **trabajos de completitud de MVP**
- Debería llamarlos EP-010 (continuar lo que se archivó)
- O distribuir entre EP-005 (Adaptadores) y crear una épica nueva "EP-010 · MVP Fixes & Completeness"

---

### 6. Historias Huérfanas (Nuevas pero sin épica formal)

| HU | Épica | En docs/04-historias/? | En backlog.md? |
|---|---|---|---|
| HU-050 | EP-011 | ✓ | ❌ |
| HU-051 | EP-011 | ✓ | ❌ |
| HU-052 | EP-011 | ✓ | ❌ |
| HU-053 | EP-012 | ✓ | ❌ |
| HU-054 | EP-012 | ✓ | ❌ |
| HU-055 | EP-012 | ✓ | ❌ |
| HU-056 | EP-013 | ✓ | ❌ |
| HU-057 | EP-014 | ✓ | ❌ |
| HU-058 | EP-015 | ✓ | ❌ |
| HU-059 | EP-016 | ✓ | ❌ |
| HU-060 | EP-017 | ✓ | ❌ |

---

## Acciones Requeridas Antes de Implementar

### CRÍTICO (BLOQUEA)
1. **Decidir**: ¿EP-011 a EP-017 son épicas formales o sub-épicas de EP-010/EP-005?
2. **Agregar EP-010** (o EP-011-017) a `docs/03-backlog/epicas.md` con:
   - Descripción clara
   - Trazabilidad a PRD objetivos
   - Historias anticipadas (HU-050 a HU-060)
3. **Actualizar** `docs/03-backlog/backlog.md` con todas las historias nuevas
4. **Priorizar** en `docs/05-priorizacion/` (si existe ese archivo)

### IMPORTANTE (para cleanroom)
5. **Validar AC** de todas las HU (dado/cuando/entonces)
6. **Validar INVEST** de todas las HU
7. **Revisar trazabilidad** bidireccional (PRD → Épica → HU ← AC)

---

## Plan de Remediación

### Opción Recomendada: EP-010 · MVP Fixes & Completeness

```markdown
## EP-010 · MVP Fixes & Completeness

| Campo | Valor |
|---|---|
| Objetivo(s) del PRD cubiertos | Obj.3, Obj.4, Obj.5 |
| Capa (build) | completeness |
| Métrica de éxito | POST /v1/chat/completions devuelve 200 con content válido; OmniRoute disponible como proveedor; /v1/embeddings y /v1/messages funcionan; logging estructurado en todos los handlers |

**De qué se trata**: Completar handlers MVP (chat, embeddings, messages) que devuelven errores; crear adaptador OmniRoute; normalizar configuración; añadir observabilidad.

**Historias**:
- HU-050, HU-051, HU-052 (Debug chat/completions)
- HU-053, HU-054, HU-055 (OmniRoute adapter)
- HU-056 (Normalizar IDs)
- HU-057 (Embeddings)
- HU-058 (Messages)
- HU-059 (Logging)
- HU-060 (Métricas)
```

---

## Checklist de Validación

- [ ] EP-010/EP-011 formalmente declarado en `epicas.md`
- [ ] Trazabilidad clara EP → PRD objetivos
- [ ] Todas las HU-050-060 en `backlog.md`
- [ ] Todas las HU tienen AC en G/W/T
- [ ] Todas las HU pasan INVEST check
- [ ] Backlog ordenado por priorización
- [ ] build-state.json refleja estado actual
- [ ] No hay HU "huérfanas" (sin épica)

---

## Recomendación Final

✅ **Proceder con**: Opción "EP-010 · MVP Fixes & Completeness"

Esto porque:
1. Epistemológicamente, es **completitud de MVP**, no nueva funcionalidad
2. Está alineado con PRD Obj.3, Obj.4, Obj.5 (observabilidad, compatibilidad, gobernanza)
3. Las historias (HU-050-060) ya están bien documentadas
4. El arnés Factory lo soporta (ep_id es configurable)

**Proximos Pasos**:
1. Crear/formalizar EP-010 en `epicas.md`
2. Agregar HU-050 a HU-060 en `backlog.md`
3. Ejecutar Factory auditors (invest-validator, bdd-validator, etc.)
4. Luego: comenzar build slice por slice

