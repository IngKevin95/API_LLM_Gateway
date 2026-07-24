---
id: PLAN-CC-001
titulo: Plan de construcción — CC-001 Proveedores gratuitos + Cuota + Dashboard
estado: draft
fecha_creacion: 2026-07-23
---

# PLAN-CC-001: Plan de Construcción

Ordenamiento de slices, dependencias, paralelización, estimación y criterios de verificación.

---

## Secuencia de slices

### **Slice 1: Adapter Genérico Data-Driven (HU-EVO-001 a HU-EVO-005)**

**Épica:** EP-EVO-001

**Historias:** 
- HU-EVO-001: Crear adapter genérico data-driven
- HU-EVO-002: Cargar `free-tier.yaml` en Registry
- HU-EVO-003: Extender `conformance_test.go`
- HU-EVO-004: Health Monitor detecta 429
- HU-EVO-005: Quota Manager inicializa contadores

**Dependencias:**
- Prerequisito: HU-001 (Registry), HU-002a (Router), HU-005 (Health Monitor), HU-006 (Quota Manager) — **YA LISTOS**

**Estimación:** 5-6 días

**Verificación:**
- ✅ 5 nuevos proveedores (Groq, Cerebras, Mistral, Gemini, Cloudflare) cargan sin error
- ✅ `go test ./src/internal/adapter/conformance_test.go` pasa con 5 nuevos specs
- ✅ Health Monitor retira proveedor en 429 durante 30s, lo reactiva después
- ✅ `Quota.Remaining("groq")` devuelve `quota_hint` del YAML antes de primer request

**Bloques de código:**
1. `src/internal/adapter/generic/adapter.go` (200-300 LOC) — implementar `Adapter` interface genérica
2. `config/providers/free-tier.yaml` (50-80 LOC) — definir 5 proveedores
3. `src/internal/registry/registry.go` (50-100 LOC) — extender `LoadFromFile`
4. `src/internal/health/health.go` (100-150 LOC) — extender con 429 detection
5. Tests: `conformance_test.go` (+200 LOC con mocks)

**Slice 1 Entregable:** `src/gateway-bin` arranca, lista 5 nuevos proveedores, failover funciona con 429

---

### **Slice 2: Aprendizaje de Cuota desde Headers (HU-EVO-006 a HU-EVO-010)**

**Épica:** EP-EVO-002

**Historias:**
- HU-EVO-006: Parsear headers X-RateLimit-*
- HU-EVO-007: `LearnFromHeaders()` en Quota Manager con RAM
- HU-EVO-008: Persistencia async en PostgreSQL
- HU-EVO-009: Router penaliza score cuando remaining < 20%
- HU-EVO-010: Manejo de 429 con reset timeout

**Dependencias:**
- **Bloquea en:** Slice 1 (necesita adapter genérico parseando headers)
- **Integra con:** HU-004a/c (Failover), HU-006 (Quota Manager)

**Estimación:** 6-7 días

**Verificación:**
- ✅ Adapter.Chat() devuelve `QuotaInfo{Limit, Remaining, ResetAt}` para OpenAI/Anthropic/Groq
- ✅ `LearnFromHeaders()` actualiza RAM en < 1ms (race condition test)
- ✅ PostgreSQL persiste learned quota cada 100ms sin bloquear requests
- ✅ Router penaliza score en 50% cuando remaining < 20%
- ✅ 429 con `Retry-After: 60` retira proveedor por 60s exacto

**Bloques de código:**
1. `src/internal/adapter/types.go` (+20 LOC) — struct `QuotaInfo`
2. `src/internal/adapter/{openai,anthropic,google}/adapter.go` (+50 LOC/adapter) — parseHeaders()
3. `src/internal/quota/manager.go` (+150-200 LOC) — `LearnFromHeaders()`, atomicidad
4. `src/internal/quota/persist.go` (100-150 LOC nuevo) — async batch writer
5. `src/internal/router/router.go` (+10 LOC) — penalización
6. `src/internal/failover/failover.go` (+50 LOC) — 429 detection
7. Schema PostgreSQL: `CREATE TABLE provider_quotas_learned` (10-15 LOC SQL)

**Slice 2 Entregable:** `quota.Manager.Snapshot()` devuelve learned quotas, persistent en DB, Router respeta penalización

---

### **Slice 3: Dashboard + Alertas (HU-EVO-011 a HU-EVO-015)**

**Épica:** EP-EVO-003

**Historias:**
- HU-EVO-011: Metrics.Store quota snapshot
- HU-EVO-012: Alert Manager cuota < umbral
- HU-EVO-013: GET `/alerts` con filtrado RBAC
- HU-EVO-014: UI React dashboard
- HU-EVO-015: Notificaciones browser

**Dependencias:**
- **Bloquea en:** Slice 2 (necesita `quota.Manager.Snapshot()` y learned quotas)
- **Integra con:** HU-060 (Métricas), HU-009 (RBAC)

**Estimación:** 7-8 días

**Verificación:**
- ✅ GET `/metrics` devuelve bloque `quota[]` con remaining actual
- ✅ Alert Manager genera alertas cuando remaining < 10% (configurable)
- ✅ GET `/alerts?tenant=T1` devuelve solo alertas de T1 + scopes autorizados
- ✅ Dashboard React carga, tabs funcionales, auto-refresh cada 5s
- ✅ Toast notificación aparece cuando remaining cae bajo umbral

**Bloques de código:**
1. `src/internal/metrics/store.go` (+50 LOC) — extender `Snapshot()` con quota
2. `src/internal/alert/manager.go` (150-200 LOC nuevo) — Alert Manager + dedup
3. `src/internal/handler/alerts.go` (80-120 LOC nuevo) — GET `/alerts` endpoint + filtrado
4. Schema PostgreSQL: `CREATE TABLE provider_alerts` (10-15 LOC SQL)
5. `src/ui/dashboard/` (1500+ LOC React) — componentes Dashboard/Overview/Quotas/Alerts/Providers + hooks

**Slice 3 Entregable:** Dashboard accesible en `/dashboard`, UI mostrando cuota real-time con alertas

---

## Paralelización posible

❌ **NO paralelizable.** Dependencia estricta 1 → 2 → 3.

- Slice 1 produce adapter genérico que Slice 2 necesita
- Slice 2 produce learned quotas que Slice 3 necesita

**Dentro de cada slice,** cierta paralelización local:
- Slice 1: adapter genérico (`adapter.go`) + Registry load + tests pueden correr paralelo tras código estar
- Slice 2: parseo headers en cada adapter puede hacerse paralelo con Router penalización + Failover 429
- Slice 3: Alert Manager puede compilarse/testearse mientras UI se desarrolla (en realidad compilarán paralelo)

---

## Estimación total

| Slice | Días | Slack | Total |
|---|---|---|---|
| 1 | 5-6 | 1 | **6-7 días** |
| 2 | 6-7 | 1 | **7-8 días** |
| 3 | 7-8 | 1 | **8-9 días** |
| **Buffer** | — | 2-3 | **2-3 días** |

**Total critical path:** ~21-24 días (3-4 semanas)

**Recomendación:** Planificar para fines de agosto 2026

---

## Criterios de verificación por slice

### Slice 1 ✅

```bash
# Compile
go build ./cmd/gateway/main.go

# Registry carga nuevos proveedores
curl -H "Authorization: Bearer $ADMIN_TOKEN" http://localhost:8080/v1/models | jq '.[] | select(.id | startswith("groq"))'

# Conformance tests pasan
go test ./src/internal/adapter/conformance_test.go -v

# Health Monitor detecta 429
# [test manual con mock server devolviendo 429]
```

### Slice 2 ✅

```bash
# LearnFromHeaders atomicity
go test ./src/internal/quota/manager_test.go -run TestLearnFromHeaders -race

# Router penalización
go test ./src/internal/router/router_test.go -run TestScorePenalty

# PostgreSQL persistent
# [query provider_quotas_learned table, verifica inserts]
```

### Slice 3 ✅

```bash
# GET /metrics devuelve quota
curl http://localhost:8080/metrics | jq '.quota'

# GET /alerts filtra por tenant
curl "http://localhost:8080/alerts?tenant=T1" -H "Authorization: Bearer T1_TOKEN" | jq '.[] | .tenant_id'

# Dashboard opens, tabs interactive
open http://localhost:8080/dashboard
# Verifica: Overview tab con uptime/requests, Quotas tab con tabla, Alerts tab con notificaciones
```

---

## Risk & Mitigation

| Risk | Probability | Mitigation |
|---|---|---|
| Headers parsing bug | MEDIA | Conformance test con 3+ mocks | 
| Learned quota inconsistency | MEDIA | Atomic updates + race tests |
| DB persistencia falla | BAJA | Fail-open (RAM-only graceful) |
| RBAC filtrado falla | MEDIA | Query-level filtrado, no UI-side |

---

## Success Criteria (Global)

✅ Gateway boots con 5 nuevos proveedores sin error
✅ Request a Groq → headers parseados, remaining aprendido, persisted en DB
✅ Dashboard abierto → muestra cuota real-time de cada proveedor/modelo
✅ Alert Manager genera alertas cuando cuota < 10%, respeta tenant + RBAC
✅ 429 del proveedor → retira, reactiva tras Retry-After timeout
✅ Todos los tests de conformance + unit + integration pasan

---

## Próximo paso

**Paso 8:** Cierre (resumen, `/factory:revisar`, persiste `.flow-cc-state.json`)
