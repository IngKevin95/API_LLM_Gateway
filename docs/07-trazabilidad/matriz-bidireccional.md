# Matriz de Trazabilidad Bidireccional

Mapeo completo: PRD Objectives → Épicas → Historias → Componentes arquitectónicos.

## ✅ PRD Objectives → Épicas

| Objetivo PRD | Épica | Descripción |
|---|---|---|
| Desacoplamiento agente-proveedor | EP-001 | Enrutamiento dinámico por capacidad |
| Selección automática de modelo | EP-002 | Scoring inteligente (calidad, velocidad, costo) |
| Disponibilidad 99.9% | EP-004B | Failover y circuito breaker |
| Auditoría y compliance | EP-004A | Logging inmutable, PII redacción, encriptación |
| Baja latencia (< 100ms overhead) | EP-001, EP-002 | Algoritmos O(1/n), sin I/O en path crítico |
| Throughput 500 RPS (~43M/día) | EP-002, EP-007 | Scaling horizontal, rate limiting, quota |
| Soporte multiproveedor | EP-003 | Adapters genéricos (OpenAI, Anthropic, Google, local) |
| Configuración declarativa | EP-006 | YAML schema, providers + models + routing |
| Cero pérdida de eventos | EP-009 | WAL local, graceful shutdown, sync async |
| Seguridad de datos | EP-004A | Envelope encryption (DEK/KEK), auth RBAC |

## ✅ Épicas → Historias (Muestra)

### EP-001: Enrutamiento Dinámico

| Historia | Fase | AC Count | Componentes |
|----------|------|----------|------------|
| HU-006 | 1 | 5 | Registry |
| HU-007 | 1 | 5 | Model Router |
| HU-008 | 1 | 5 | Scoring Engine |
| HU-025 | 2 | 5 | Request Validation |

**Cobertura**: 4/4 casos críticos del flujo EP-001 (Registry → Router → Scoring → Validation)

### EP-002: Scoring Dinámico

| Historia | Fase | AC Count | Componentes |
|----------|------|----------|------------|
| HU-008 | 1 | 5 | Scoring Engine |
| HU-009 | 2 | 5 | Health Monitor |
| HU-011 | 3 | 5 | Learning Engine |

**Cobertura**: 3/3 pilares (base formula, health data, histórico)

### EP-004A: Auditoría & Compliance

| Historia | Fase | AC Count | Componentes |
|----------|------|----------|------------|
| HU-010 | 1 | 5 | Audit Logger |
| HU-020 | 1 | 5 | Envelope Encryption |
| HU-021 | 1 | 5 | PII Redactor |
| HU-030 | 2 | 5 | Right-to-Forget |
| HU-031 | 1 | 5 | Auth & AuthZ |
| HU-032 | 2 | 5 | Semantic Cache |

**Cobertura**: 6 historias cubren trazabilidad + seguridad + compliance

### EP-009: Sincronización & Persistencia

| Historia | Fase | AC Count | Componentes |
|----------|------|----------|------------|
| HU-038 | 1 | 5 | Sync Worker |
| HU-039 | 1 | 5 | Write-Ahead Log |
| HU-040 | 1 | 5 | Graceful Shutdown |
| HU-041 | 2 | 5 | Cache Invalidator |

**Cobertura**: 4 historias cubren persistencia multi-layer (WAL → async sync → shutdown limpio)

## ✅ Historias → Componentes (Matriz Cruzada)

Cada historia implementa 1-2 componentes principales:

| Historia | Componente Primario | Componente Secundario | Fase |
|----------|---|---|---|
| HU-006 | Registry | - | 1 |
| HU-007 | Model Router | - | 1 |
| HU-008 | Scoring Engine | Health Monitor | 1 |
| HU-009 | Health Monitor | Scoring Engine | 2 |
| HU-010 | Audit Logger | WAL | 1 |
| HU-011 | Learning Engine | - | 3 |
| HU-012 | Rate Limiter | - | 1 |
| HU-013 | Quota Manager | - | 1 |
| HU-014 | Failover | Circuit Breaker | 2 |
| HU-018 | Adapter OpenAI | - | 1 |
| HU-020 | Envelope Encryption | KMS Client | 1 |
| HU-021 | PII Redactor | - | 1 |
| HU-024 | Adapter Local | - | 1 |
| HU-025 | Request Validator | - | 2 |
| HU-030 | Right-to-Forget | Audit Logger | 2 |
| HU-031 | Auth & AuthZ | - | 1 |
| HU-032 | Semantic Cache | - | 2 |
| HU-038 | Sync Worker | WAL | 1 |
| HU-039 | Write-Ahead Log | - | 1 |
| HU-040 | Graceful Shutdown | Sync Worker | 1 |
| HU-041 | Cache Invalidator | Quota Manager | 2 |

## ✅ Validación de Cobertura

### Fase 1 MVP (Constructible)

**Épicas críticas**: EP-001, EP-002, EP-004A, EP-009, EP-003, EP-006

**Historias Fase 1**: 28 historias (Must + Should con complejidad S/M)

**Componentes necesarios Fase 1**: 14 componentes
- Registry, Model Router, Scoring Engine, Health Monitor
- Rate Limiter, Quota Manager, Circuit Breaker
- Adapters (OpenAI, Google, Anthropic, Local)
- Audit Logger, Envelope Encryption, PII Redactor
- Write-Ahead Log, Sync Worker, Auth & AuthZ

**Status**: ✅ COBERTURA COMPLETA FASE 1
- Cada objetivo PRD cubierto por ≥1 épica
- Cada épica cubierta por ≥3 historias
- Cada historia mapea a 1-2 componentes

### Fase 2 (Post-MVP)

**Historias Could/Nice-to-have**: HU-009 (Health Monitor improvement), HU-011 (Learning), HU-025 (Validation), HU-030 (GDPR right-to-forget), HU-032 (Cache), HU-041 (Cache invalidator)

**Componentes Fase 2**: Health Monitor enhancements, Learning Engine, Cache layer, GDPR tooling

## ✅ Gap Analysis

| Tipo | Encontrado | Resolución |
|------|-----------|-----------|
| PRD objetivo sin épica | 0 | ✅ Todos cubren ≥1 épica |
| Épica sin historias | 0 | ✅ Todas tienen 3-6 historias |
| Historia sin AC | 0 | ✅ 48/48 historias tienen 5 AC en G/W/T |
| Historia sin componente | 0 | ✅ Todas mapean a 1-2 componentes |
| Componente sin historia | 0 | ✅ Los 18 componentes tienen soporte de historias |
| AC sin criterio verificable | 0 | ✅ Todos en formato Given/When/Then testeable |

## ✅ Decisiones de Trazabilidad

1. **Uno-a-muchos permitido**: 1 historia puede cubrir 2 componentes (ej: HU-008 → Scoring + Health Monitor)
2. **Muchos-a-uno permitido**: múltiples historias pueden refinar 1 componente (ej: HU-008, HU-009, HU-011 → Scoring/Health/Learning)
3. **Fase 1 aislada**: Todas las historias Must de Fase 1 son implementables sin Fase 2
4. **AC → Tests**: Cada AC G/W/T es convertible a test case (unitario o integración)

## Referencias

- PRD: `docs/01-prd/api-llm-gateway.md`
- Épicas: `docs/03-backlog/epicas.md`
- Historias: `docs/04-historias/HU-*.md` (48 archivos)
- Arquitectura: `docs/11-architecture/api-llm-gateway.md` (18 componentes)
- Flows: `docs/06-flows/EP-*.md` (4 épicas críticas)
