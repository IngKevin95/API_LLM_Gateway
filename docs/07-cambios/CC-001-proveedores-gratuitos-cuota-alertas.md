---
id: CC-001
titulo: Cambio evolutivo — Proveedores gratuitos + Aprendizaje de cuota + Dashboard de alertas
estado: draft
fecha_creacion: 2026-07-23
responsable: Kevin Beltrán
---

# CC-001: Proveedores Gratuitos + Aprendizaje de Cuota + Dashboard de Alertas

## Resumen ejecutivo

**Objetivo:** Ampliar el API LLM Gateway con:
1. Catálogo curado de ~5 proveedores gratuitos nuevos en Fase 1 (Groq, Cerebras, Mistral, Gemini, Cloudflare AI) usando adapter genérico data-driven
2. Sistema de aprendizaje automático de cuota desde headers HTTP (X-RateLimit-*, etc.) con persistencia en PostgreSQL
3. Dashboard React con alertas por tenant respetando RBAC, UI para visualizar métricas/cuota en tiempo real

**Épicas evolutivas:** EP-EVO-001, EP-EVO-002, EP-EVO-003 (15 HU-EVO total)

**Estimación:** 4-5 sprints (fines julio + agosto 2026)

**Riesgo global:** MEDIO — cambios acotados a adapters + quota manager + dashboard; no toca auth/failover core

---

## Impacto en épicas existentes

### Extensión directa (sin breaking changes)

| Épica existente | Cómo se extiende | Impacto | HU-EVO relacionada |
|---|---|---|---|
| **EP-001** (Enrutamiento) | Router ahora penaliza score cuando remaining < 20% | **Bajo** — solo suma lógica al scoring existente | HU-EVO-009 |
| **EP-002** (Resiliencia) | Health Monitor + Failover respeta 429 con Retry-After | **Bajo** — extiende blacklist existente | HU-EVO-004, HU-EVO-010 |
| **EP-003** (Gobernanza) | Quota Manager aprende desde headers, persiste en DB | **Medio** — cambia fuente de verdad de cuota (YAML → learned + DB) | HU-EVO-005 a HU-EVO-010 |
| **EP-004A** (Auth/AuthZ) | `/alerts` endpoint respeta tenant + scopes | **Bajo** — filtra alertas, no cambia autenticación | HU-EVO-013 |

### Impacto en historias existentes

**Historias que permanecen sin cambios (verde):**
- HU-001 a HU-003 (Registry, Router automático) — continúan igual, adaptan a nuevos proveedores
- HU-004a/b/c (Failover/Circuit Breaker) — se integran nuevos 429 handlers
- HU-005 (Health Monitor) — se extiende con detección de 429
- HU-006 (Quota Manager) — **EXTIENDE SIGNIFICATIVAMENTE** → aprendizaje automático + persistencia
- HU-008/009 (Auth/AuthZ) — se reutilizan para filtrado de alertas
- HU-060 (Métricas) — **EXTIENDE** con bloque `quota[]`

**Historias que se amplían (naranja):**
- **HU-006** (Contabilizar cuota): De "contadores iniciales en RAM desde YAML" → "aprendizaje dinámico desde headers + persistencia en DB"
- **HU-060** (Métricas): De "uptime/requests/latency JSON" → "+ bloque quota[] con desglose por proveedor/modelo"

---

## Riesgos de cambios en producción

### Riesgo crítico: NINGUNO
No hay breaking changes en APIs públicas ni en auth/failover core.

### Riesgo alto: Cuota Manager - aprendizaje mal calibrado

**Escenario:** Un proveedor devuelve headers de cuota inconsistentes → Router penaliza incorrectamente → disponibilidad baja

**Probabilidad:** MEDIA (algunos proveedores tienen headers no-estándar o malformados)

**Impacto:** ALTO (usuarios ven degradación de disponibilidad)

**Mitigación:**
- AC en HU-EVO-006: Test de parseo con mock servers (Groq, Cerebras, OpenAI, Anthropic)
- AC en HU-EVO-007: Race condition tests + atomic updates
- AC en HU-EVO-009: Penalización suave (50%) no dramática
- Monitoreo: Alert Manager detecta si cuota aprendida cae rápidamente (< 10 min) = señal de bug

**Upgrade path:** Si falla mitigación → rollback penalización (30 min) sin rollback de código

---

### Riesgo alto: PostgreSQL persistencia async falla silenciosamente

**Escenario:** Sync Worker cae → learned quotas no se persisten → restart pierde estado

**Probabilidad:** BAJA (async worker + batch writes probado en HU-038/039)

**Impacto:** MEDIO (reinicio resetea cuota a `quota_hint` del YAML, puede ser anticuado)

**Mitigación:**
- AC en HU-EVO-008: DB down no bloquea requests (fail-open)
- AC en HU-EVO-008: Async job failed = warn log + retry exponencial
- Health Monitor: detecta si learned_at timestamp no se actualiza > 5 min = alerta

**Upgrade path:** Si persistencia falla → fall back a RAM-only (graceful) + log critical

---

### Riesgo medio: Dashboard UI React - RBAC no sincroniza

**Escenario:** Usuario T1 ve alertas de T2 por bug en filtrado RBAC

**Probabilidad:** MEDIA (frontend auth es propenso a olvidos)

**Impacto:** MEDIO (información sensible expuesta)

**Mitigación:**
- AC en HU-EVO-013: Query-level filtrado (filtro en WHERE, no frontend)
- AC en HU-EVO-014: Valores filtrados vienen ya del backend, no se filtran en UI
- Test de seguridad: Mock auth context T1/T2, verifica respuesta no contiene datos del otro

**Upgrade path:** Si falla → revert a JSON-only `/metrics` (no UI) + re-test HU-EVO-013

---

### Riesgo bajo: Compatibilidad de nuevos adapters

**Escenario:** Un proveedor nuevo tiene bug específico → conformance tests pasan pero real requests fallan

**Probabilidad:** BAJA (adapter genérico probado con ≥3 mocks)

**Impacto:** BAJO (proveedor simplemente no funciona, failover al siguiente)

**Mitigación:**
- AC en HU-EVO-003: Conformance test corre contra mock servers + headless browser
- AC en HU-EVO-001/002: Cada nuevo proveedor tiene quota_hint + circuit breaker timeout

**Upgrade path:** Deshabilitar proveedor en `free-tier.yaml` sin redeploy

---

## Trazabilidad bidireccional

### Épicas evolutivas → Épicas existentes

```
EP-EVO-001 (Adapters)
  └─ extiende EP-002 (Resiliencia)
     └─ impacta HU-024 (Adapters locales), HU-020* (OpenAI), HU-021* (Anthropic)

EP-EVO-002 (Aprendizaje cuota)
  └─ extiende EP-003 (Gobernanza)
     └─ impacta HU-006 (Cuota), HU-017 (Métricas por modelo)

EP-EVO-003 (Dashboard + Alertas)
  └─ extiende EP-001 (Enrutamiento) + EP-004A (Auth/AuthZ)
     └─ impacta HU-017 (Métricas), HU-023 (Dashboard), HU-009 (RBAC)
```

### Historias evolutivas → Historias existentes

| HU-EVO | Extiende | Bloqueada por | Integra con |
|---|---|---|---|
| HU-EVO-001 | HU-024 (adapters locales) | — | HU-002a (router) |
| HU-EVO-006 | HU-006 (cuota) | HU-EVO-001 | HU-004c (timeouts) |
| HU-EVO-011 | HU-060 (métricas) | HU-EVO-007 | HU-009 (RBAC) |

---

## Cambios de arquitectura

### Cambios en Quota Manager (`src/internal/quota/manager.go`)

**Antes (HU-006):**
```
config.yaml → quota_hint → Reserve/Commit → RAM counter
                             ↓
                          PostgreSQL (async, optional)
```

**Después (HU-EVO-005 a HU-EVO-010):**
```
config.yaml → quota_hint (fallback)
     ↓
Adapter response headers (X-RateLimit-*) → LearnFromHeaders()
     ↓
RAM counter (O(1) lookup) + PostgreSQL async persist
     ↓
Router.Score() penaliza si remaining < 20%
```

**Impacto:** Cuota es ahora "learned" (dinámico) en lugar de "static" (estático del YAML)

---

### Cambios en Router (`src/internal/router/router.go`)

**Nuevo factor en scoring:**
```go
if remaining < limit*0.2 {
    score *= 0.5  // Penalización 50%
}
```

**Impacto:** Bajo — solo suma lógica, no cambia algoritmo

---

### Cambios en `/metrics` endpoint

**Antes (HU-060):**
```json
{
  "uptime_seconds": 3600,
  "requests": {...},
  "latency": {...},
  "providers": [...]
}
```

**Después (HU-EVO-011):**
```json
{
  "uptime_seconds": 3600,
  "requests": {...},
  "latency": {...},
  "providers": [...],
  "quota": [
    {"provider": "groq", "model": "mixtral", "limit": 14400, "remaining": 14200, "reset_at": "...", "healthy": true},
    ...
  ]
}
```

**Impacto:** Extensión aditiva, no breaking

---

### Nuevos endpoints

- **GET `/alerts`** (HU-EVO-013): Filtra alertas por tenant + scope
- **POST `/settings/quota-alert-threshold`** (HU-EVO-012 optional): Configura umbral de alertas

---

## Plan de rollback

**Fase 1 (Adapters):** Si bug en adapter genérico → deshabilitar proveedor en YAML, no code rollback

**Fase 2 (Aprendizaje):** Si bug en LearnFromHeaders() → desactivar learning (usa solo quota_hint), rollback 1 commit

**Fase 3 (Dashboard):** Si bug en UI → revert a JSON `/metrics` only (HU-060 sigue funcionando)

---

## Validación de trazabilidad

✅ Cada HU-EVO tiene bloque "Relación con existentes" (ver archivos individuales)

✅ Cada EP-EVO extiende épica existente con `extiende` y `impacta` explícitos

✅ No hay épicas huérfanas (todas linkedadas a objetivos PRD)

✅ Riesgos identificados y mitigación documentada

**Estado:** Listo para Paso 7 (Plan de construcción)

---

## Próximos pasos

1. **Paso 7** (Plan de construcción): Ordenar slices, estimar, identificar paralelización
2. **Paso 8** (Cierre): Resumen ejecutivo, sugerir `/factory:revisar`, archivo `.flow-cc-state.json` done
