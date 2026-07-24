## Why

Cada proveedor de LLM devuelve su cuota real en headers HTTP estándar (`X-RateLimit-*`, `Retry-After`, variantes por proveedor). Hoy el Gateway inicia con `quota_hint` estático del YAML (valores aproximados). Aprendiendo desde headers en tiempo real obtenemos cuota exacta y dinámica, sin esperar auditoría programada. Esto permite que el Router deprioritice automáticamente proveedores con cuota baja (<20%), evitando exhaustiones imprevistas.

## What Changes

- **New adapter method**: Todos los adapters extraen `X-RateLimit-Limit/Remaining/Reset` (+ variantes OpenAI, Anthropic, Groq) y retornan estructura normalizada `QuotaInfo{Limit, Remaining, ResetAt}`.
- **Quota Manager learning**: Post-response, `LearnFromHeaders(providerID, quotaInfo)` actualiza remaining en RAM de forma atómica, detectando resets de ventana y clampeando negativos.
- **Async persistence**: Background worker persiste learned quotas en PostgreSQL (`provider_quotas_learned`) sin bloquear path crítico (<5ms).
- **Router score penalty**: Score decrementa 50% cuando `remaining < limit*0.2`, deprioritizando proveedores con cuota baja.
- **429 handling**: Failover extrae `Retry-After` header y retira temporalmente proveedor; múltiples 429 consecutivos aplican backoff exponencial (30s→60s→120s→tope).
- No breaking changes: todas las extensiones son aditivas; adapters existentes siguen funcionando sin cambios.

## Capabilities

### New Capabilities
- `quota-header-parsing`: Parseo de headers estándar X-RateLimit-* (+ variantes OpenAI/Anthropic/Groq) en todas las respuestas de adapter. Graceful fallback si headers ausentes.
- `quota-learning-ram`: Actualización atómica de remaining en RAM tras cada response. Detecta resets de ventana por ResetAt. Clampea negativos a 0.
- `quota-persistence`: Persistencia asíncrona a PostgreSQL de learned quotas. Background worker, batch UPSERT, crash recovery. Tabla: `provider_quotas_learned`.

### Modified Capabilities
- `quota-manager`: Se extiende con métodos `LearnFromHeaders()` y persistencia. Inicialización desde `QuotaHints` del registry (EP-EVO-001) + restauración desde DB post-reinicio.
- `model-router`: Método `Score()` agrega penalización cuando `remaining < 20% del limite`. No breaking change al algoritmo existente.
- `failover-engine`: Manejo de 429 con extracción de `Retry-After` header. Retiro temporal con duration del header. Backoff exponencial para 429 repetidos. No bloquea mid-stream.
- `health-monitor`: Integración con retiros por 429 (backoff exponencial ya implementado en EP-EVO-001; se sigue utilizando).

## Impact

**Code changes:**
- `src/internal/adapter/adapter.go`: Base struct Response extendido con `QuotaInfo` field.
- `src/internal/adapter/{openai,anthropic,google,groq,generic}/adapter.go`: Método `extractQuota()` nuevo; ya parseado en response.
- `src/internal/quota/manager.go`: Métodos `LearnFromHeaders()` nuevo, `RestoreRemaining()` extensión. Mutex para atomicidad.
- `src/internal/quota/persist.go`: Worker async nuevo, batch UPSERT, retry logic.
- `src/internal/router/router.go`: Línea en `Score()` para penalización <20%.
- `src/internal/failover/failover.go`: Parseo de Retry-After, retiro temporal en `OnRateLimited()`, backoff exponencial.
- Schema PostgreSQL: Tabla `provider_quotas_learned` nueva.

**APIs:**
- GET `/metrics`: Payload extendido con bloque `quota[]{provider, limit, remaining, reset_at, healthy}` (EP-EVO-003 implementa UI; esta épica expone datos).
- No cambios en endpoints públicos.

**Dependencies:**
- PostgreSQL driver (ya existe: `github.com/lib/pq` o similar, verificar import).
- Ningún nuevo package externo.

**Testing:**
- Unit: mock servers con headers variables, atomic updates bajo race, async persistence sin DB real.
- Integration: end-to-end desde adapter.Chat() con headers reales hasta quota.Manager.LearnFromHeaders() hasta health.Blacklist() en 429.

---

## Trazabilidad

**Épica:** EP-EVO-002 (Aprendizaje de Cuota desde Headers HTTP + Persistencia)

**Historias cubiertas:**
- HU-EVO-006: Parsear headers estándar X-RateLimit-* por adapter
- HU-EVO-007: Implementar LearnFromHeaders() en Quota Manager con actualización RAM
- HU-EVO-008: Persistencia asíncrona en PostgreSQL de learned quotas
- HU-EVO-009: Router penaliza score cuando remaining < 20%
- HU-EVO-010: Manejo de 429 con reset timeout y Retry-After
