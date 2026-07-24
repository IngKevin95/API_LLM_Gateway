## Context

EP-EVO-001 implementó adapters genéricos data-driven y Health Monitor con detección de 429. Hoy la Quota Manager inicia con `quota_hint` estático del registry YAML (valores aproximados). Cada provider devuelve su cuota real en headers HTTP (X-RateLimit-*, Retry-After, variantes). Para aprovechar esa información en tiempo real, necesitamos:

1. Parsear headers en cada adapter
2. Actualizar remaining en RAM de forma atómica tras cada response
3. Persistir learned quotas en PostgreSQL sin bloquear path crítico
4. Penalizar score del Router cuando cuota está baja
5. Manejar 429 con Retry-After y backoff exponencial

Stakeholders: operadores (auditoría de aprendizaje), ingenieros (menos 429 inesperados), usuarios (failover más inteligente).

## Goals / Non-Goals

**Goals:**
- Quota real y dinámica: remaining actualizado post-response, sin auditoría externa
- Crash recovery: reinicio restaura learned desde PostgreSQL
- No path-critical bloqueado: async persist <5ms en respuesta al usuario
- Router intelligent: deprioritiza automáticamente cuota baja <20%
- 429 respect: Retry-After parsed y aplicado; backoff exponencial evita ban

**Non-Goals:**
- Predicción de reset: cuándo se resetea la ventana (future work: Fase 2)
- Alertas en UI: dashboard con umbral configurable (EP-EVO-003)
- Rate limiting enforcement en Gateway (se delega a provider)
- Cambios en API pública de Gateway (learn es internal)

## Decisions

### Decision 1: Estructura de QuotaInfo en adapter response

**Chosen**: Nuevo field `QuotaInfo` en `Response` struct (existente en EP-EVO-001).
- `QuotaInfo { Limit int64, Remaining int64, ResetAt *time.Time }`
- Parseado por cada adapter en `extractQuota()` (método nuevo, graceful fallback)
- Normalización: todos los adapters retornan mismo schema (X-RateLimit-*, anthropic-ratelimit-*, groq x-ratelimit-*, etc.)

**Alternatives considered**:
- Inyectar QuotaInfo como header en respuesta HTTP (requeriría cambio de API) — rechazado
- Dedicar un paquete `adapter/quota` para lógica compartida — overkill; 2-3 líneas por adapter

**Rationale**: Minimal, data-driven. Cada adapter sabe dónde están sus headers; normalización al level de struct evita duplicación.

---

### Decision 2: Atomicidad de updates en RAM

**Chosen**: Mutex (sync.RWMutex) en quota.Manager. LearnFromHeaders(providerID, quotaInfo) agarra lock, actualiza, suelta.
- Regla de update: si quotaInfo.ResetAt > current.ResetAt, es un reset, copia todo. Si Remaining > current, update conservador (puede ser race de otro request). Ignora si Remaining < current (otro request paralelo lo vio primero).

**Alternatives considered**:
- Atomic operations (sync/atomic) en campos específicos — Go no soporta atomic struct
- Lock-free con channels — overkill para frecuencia de updates (1/request)

**Rationale**: Correctness > performance. Mutex permite transacción lógica de "detectar reset + actualizar" de forma indivisible. RWMutex permite reads concurrentes sin bloqueo (Router.Score() es read-heavy).

---

### Decision 3: Persistencia asíncrona, no sincrónica

**Chosen**: Background worker (goroutine) con channel. LearnFromHeaders() enqueue job, retorna inmediatamente (<1ms). Worker batch writes cada 100ms o 1000 jobs, UPSERT en DB.

**Alternatives considered**:
- Sync writes: bloquearía response en latencia de DB (~10-50ms) — rechazado por SLA
- Redis + background sync: extra dependency — PostgreSQL ya existe

**Rationale**: Path crítico limpio. Crash risk mitigation: si Gateway cae antes de flush, Ram ya tiene learned; post-reinicio RestoreRemaining() lee DB y prioriza lo persistido.

---

### Decision 4: Penalización de Router sin refactor

**Chosen**: Una línea en Router.Score(): `if remaining < limit*0.2 { score *= 0.5 }` (penalización 50%).
- Integración mínima: Router ya conoce `quota.Manager`; solo llama `quota.Remaining(providerID, modelID)`.

**Alternatives considered**:
- Crear scorer separado — indirection innecesaria
- Penalización fija vs % — % es más explícita (20% del limit es el trigger, 50% es el castigo)

**Rationale**: YAGNI. Scoring es internal; failover chain respeta penalización porque elige proveedor con mejor score.

---

### Decision 5: Manejo de 429 con Retry-After

**Chosen**: adapter.ProviderError gana field `RetryAfter *time.Duration`. generic.Adapter.checkStatus() parsea header. failover.Engine.Complete() llama `e.OnRateLimited(providerID, pe.RetryAfter)` si Status==429. health.Monitor.RetireOn429() acepta duration (ya existe desde EP-EVO-001).

**Alternatives considered**:
- Hardcode 30s default — pierde respeto a header del proveedor
- Distintos defaults por proveedor — complexity; 30s es seguro fallback

**Rationale**: HTTP standard says Retry-After (segundos o fecha HTTP). Respetarlo evita throttling repetido. Backoff exponencial (ya en health monitor) maneja abuso.

---

### Decision 6: Tabla PostgreSQL de histórico

**Chosen**: `provider_quotas_learned` con columnas `(id, provider_id, model_id, limit, remaining, reset_at, learned_at)` + unique constraint `(provider_id, model_id, learned_at)` para auditoría sin llenar DB.

**Alternatives considered**:
- ON CONFLICT + UPDATE (UPSERT): pierde histórico si se sobrescribe
- Separar current (simple) + histórico (audit table) — complexity

**Rationale**: Auditoría + crash recovery en una tabla. `learned_at` timestamp natural index para consultas `WHERE learned_at > ?`.

---

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| **Parser vulnerability (malformed headers)** | parseRetryAfter() usa time.ParseDuration() + RFC1123; malformed → defaulting, no panic. Tests con headers inválidos. |
| **Race in learning (2 requests simultaneous, conflicting remaining)** | Mutex en LearnFromHeaders(). Last writer wins (conservador: toma valor más alto). Logs si conflicto. |
| **DB down, async worker fails** | Background worker logs error, continúa usando RAM. POST-REINICIO: si DB unavailable, Gateway boot waits con timeout (fallback: skip restore, continúa con hint). |
| **Retry-After header malicious (huge number)** | Capped a 240s (4 min). Exponential backoff stops at 120s; 2nd 429 = 240s total. |
| **Learning stale (header lag behind real quota)** | Inherent to headers; no mitigation sin out-of-band audit. Future: Fase 2 agrega predicción de reset. |
| **Penalización score demasiado agresiva (pierde proveedor viable)** | 50% es castigo; failover tries non-penalized first, but falls back. Configurable en future. |

---

## Migration Plan

**Phase 1: Deploy (zero-downtime)**
1. Adapter changes (extractQuota) + Quota.LearnFromHeaders() backward compatible (DB down = noop).
2. Background worker starts silently, enqueues jobs. If DB down, worker accumulates in memory (bounded queue, drops old if full).
3. Router.Score() change live: no breaking API change, just internal scoring tweak.
4. Health.RetireOn429() already live from EP-EVO-001; failover now passes real duration.
5. Create PostgreSQL table `provider_quotas_learned` (can be done pre-deploy or in-band first insert).

**Phase 2: Observe (1-2 weeks)**
- Monitor logs for parser errors, DB write latency, conflict rates.
- Alerts if async worker queue grows unbounded (sign of DB trouble).

**Phase 3: Rollback (if needed)**
- Set LEARNING_ENABLED=false env var → LearnFromHeaders() returns immediately, no learning.
- DB keeps historical data; re-enabling reads it.

---

## Open Questions

1. **DB connection pooling**: How many workers write simultaneously? Should we batch more aggressively or tune pool size?
2. **Penalización threshold (20%) y castigo (50%)**: Should these be configurable per deployment? (Today: hardcoded.)
3. **Retry-After cap (240s)**: Is 4 min reasonable, or should it be per-provider? (Today: global.)
4. **Model-level vs provider-level quota**: HU-EVO-006/007/008 treat provider+model as key (some providers report per-model). Is composite key ("openai.gpt-4" vs "openai.gpt-3.5") necessary, or group by provider only? (Today: model-aware, but grouped per provider in learn.)
5. **Learning after cache hit**: If response comes from cache (future caching layer), skip learning? (Today: N/A, all live.)
