## Context

`quota.Manager` (EP-EVO-002) ya aprende y persiste cuota por proveedor/modelo en PostgreSQL, y
`metrics.Store` (HU-060, extendido en EP-EVO-001) ya expone `/metrics` con providers/latency. Este
change conecta ambos (snapshot de cuota en `/metrics`) y agrega dos piezas nuevas: un worker que
vigila la cuota y persiste alertas, y un endpoint HTTP RBAC-aware para consultarlas. RBAC/auth
multi-tenant ya existe (HU-009) y se reutiliza sin modificar su contrato.

## Goals / Non-Goals

**Goals:**
- Exponer cuota remanente por proveedor/modelo en `/metrics`, filtrada por scope del requester.
- Detectar proactivamente cuota baja (warning <10%, critical =0%) sin esperar a que el Router
  falle failover.
- Persistir alertas en PostgreSQL con dedup, evitando spam de alertas repetidas por el mismo
  evento.
- Exponer `GET /alerts` respetando tenant y scopes, con paginación.

**Non-Goals:**
- No se agrega UI en este sub-slice (SS2 de EP-EVO-003 la consume después).
- No se cambia el contrato de autenticación existente (HU-009); solo se lee `AuthContext` ya
  resuelto por middleware.
- No se implementa canal de notificación push (browser/email) — eso es HU-EVO-015, sub-slice SS2.

## Decisions

### 1. Snapshot de cuota vía método `quota.Manager.Snapshot()`, sin nuevo estado
`metrics.Store.GetMetrics()` llama `s.quota.Snapshot()` (lectura pura desde el mapa en RAM que ya
mantiene `quota.Manager`) en cada request; no se cachea entre requests porque el volumen esperado
(cientos de proveedores/modelos) mantiene la respuesta bajo 100ms sin necesidad de cache.
- Alternativa descartada: cachear snapshot con TTL — rechazada por complejidad innecesaria dado el
  volumen actual; se revisita si el conteo de proveedores crece 10x.

### 2. `alert.Manager` como worker independiente con su propio ticker
Se crea `src/internal/alert/manager.go` con `Manager.Run(ctx, interval)` arrancado desde
`cmd/gateway/main.go` en una goroutine propia (no acoplado al ciclo de request), leyendo
`quota.Manager.Snapshot()` cada 1 minuto (`interval` configurable vía flag/env, default `time.Minute`).
Umbral configurable vía `GATEWAY_ALERT_THRESHOLD` (default 0.10).
- Alternativa descartada: evaluar alertas dentro del propio Quota Manager al aprender cada header —
  rechazada porque acopla la ruta caliente de requests con I/O a PostgreSQL (latencia); un worker
  periódico desacopla escritura de alertas del request path.

### 3. Dedup por índice único `(provider_id, model_id, alert_time)` + upsert de `updated_at`
Antes de insertar, `Manager.generateAlert` consulta si existe una alerta sin `resolved_at` para
`(provider_id, model_id)`; si existe, solo actualiza `updated_at`; si no, inserta nueva fila con
`alert_time = now()`. Esto evita que una revisión cada minuto genere 60 alertas/hora para el mismo
evento sostenido.
- Alternativa descartada: usar `UNIQUE(provider_id, model_id, alert_time)` como única defensa —
  insuficiente porque `alert_time` cambia en cada corrida; se necesita el chequeo lógico de "alerta
  activa sin resolver" antes del insert.

### 4. `GET /alerts` reutiliza `AuthContext` de HU-009, filtra en la query SQL
El handler nuevo obtiene `auth := r.Context().Value("auth").(AuthContext)` (mismo mecanismo que
otros endpoints protegidos) y construye la query con `WHERE tenant_id = $1 AND model_id IN (SELECT
model_id FROM models WHERE capability = ANY($2))`, salvo que `auth.IsAdmin` (token ==
`GATEWAY_ADMIN_TOKEN`), en cuyo caso omite el filtro de tenant. Paginación vía `LIMIT`/`OFFSET`
calculados de `page`/`limit` query params (default `limit=50`).
- Alternativa descartada: filtrar en memoria tras traer todas las alertas — rechazada por no
  escalar con el volumen de alertas y por exponer momentáneamente datos de otros tenants en el
  proceso Go antes de descartarlos.

## Risks / Mitigation

- **Riesgo**: el worker de alertas satura PostgreSQL si el catálogo de proveedores crece mucho.
  Mitigación: batch de upserts en una sola transacción por corrida, no un INSERT por
  proveedor/modelo.
- **Riesgo**: RBAC mal filtrado expondría alertas de otro tenant (dato sensible operativo).
  Mitigación: filtro en la query SQL (no en memoria) + test de integración específico
  (HU-EVO-013 AC2) que verifica que T1 nunca ve alertas de T2/T3.
- **Riesgo**: `quota.Manager.Snapshot()` no expone aún un método público — si no existe, hay que
  agregarlo sin romper el contrato interno actual de `quota.Manager` (verificar en TDD).

## Migration Plan

1. Migración SQL nueva: `provider_alerts` (additiva, no toca tablas existentes de EP-EVO-002).
2. `metrics.Store` extendido de forma retrocompatible (campo `quota` nuevo y opcional en el JSON
   de respuesta; clientes existentes de `/metrics` no se rompen).
3. `alert.Manager` arranca solo si hay conexión PostgreSQL configurada (reutiliza el mismo check
   fail-soft que ya usa la persistencia de quota de EP-EVO-002); si no hay DB, loguea WARN y no
   arranca el worker (no bloquea boot del Gateway).
4. `GET /alerts` se registra en el router HTTP junto a los demás endpoints protegidos existentes.
