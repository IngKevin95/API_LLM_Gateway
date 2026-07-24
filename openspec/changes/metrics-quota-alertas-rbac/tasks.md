## 1. Snapshot de cuota en /metrics (HU-EVO-011)

- [x] 1.1 Agregar método público `quota.Manager.Snapshot() []QuotaSnapshot` si no existe aún (lectura pura desde el mapa en RAM)
- [x] 1.2 Definir tipo `metrics.QuotaSnapshot{Provider, Model, Limit, Remaining, ResetAt, Healthy}` en `src/internal/metrics/store.go`
- [x] 1.3 Extender `Store.GetMetrics()` para poblar campo `Quota []QuotaSnapshot` desde `quota.Manager.Snapshot()`
- [x] 1.4 Filtrar `Quota` por scopes/capabilities del requester cuando no es admin (reutilizar `AuthContext`)
- [x] 1.5 Verificar que proveedor sin quota learned expone `remaining: <quota_hint>` y `learned_at: null`
- [x] 1.6 Test de latencia: `GET /metrics` con 125 cuotas (25 proveedores x 5 modelos) < 100ms
- [x] 1.7 Tests unitarios: AC1-AC5 de HU-EVO-011

## 2. Alert Manager (HU-EVO-012)

- [x] 2.1 Migración SQL: tabla `provider_alerts` (provider_id, model_id, severity, message, alert_time, updated_at, resolved_at) -- DESVIACIÓN: sin `UNIQUE(provider_id, model_id, alert_time)` (design.md ya señala que ese unique es insuficiente por sí solo); el dedup real lo hace `generateAlert` con un `UPDATE ... WHERE resolved_at IS NULL` antes del `INSERT` (ver 2.6), verificado con Postgres real en `internal/alert/manager_integration_test.go`.
- [x] 2.2 Implementar `src/internal/alert/manager.go`: `Manager.Run(ctx, interval)` con ticker propio en goroutine
- [x] 2.3 Implementar `Manager.Check(ctx)`: evalúa `quota.Manager.Snapshot()` contra umbral configurable (`GATEWAY_ALERT_THRESHOLD`, default 0.10)
- [x] 2.4 Generar alerta `severity: "warning"` cuando `remaining < limit * umbral`
- [x] 2.5 Generar alerta `severity: "critical"` cuando `remaining == 0`
- [x] 2.6 Implementar dedup: consultar alerta activa (sin `resolved_at`) por (provider_id, model_id) antes de insertar; si existe, solo actualizar `updated_at`
- [x] 2.7 Arrancar worker en `cmd/gateway/main.go` solo si hay conexión PostgreSQL configurada (fail-soft: WARN y no arranca si no hay DB)
- [x] 2.8 Tests unitarios: AC1-AC5 de HU-EVO-012

## 3. Endpoint GET /alerts con RBAC (HU-EVO-013)

- [x] 3.1 Implementar `src/internal/handler/alerts.go`: `GetAlerts(w, r)` leyendo `AuthContext` del contexto de request
- [x] 3.2 Query SQL con filtro `tenant_id = auth.TenantID` salvo `auth.IsAdmin` -- DESVIACIÓN: `provider_alerts` no tiene columna `tenant_id` (una alerta de cuota pertenece al proveedor/modelo, no a un tenant; no hay tal concepto en el schema de HU-EVO-012). El aislamiento se logra vía scope de capacidad (3.3), consistente con `internal/authz/authz.go` (no hay "tenant" propio de alertas en el resto del Gateway). Admin = mismo bearer estático `GATEWAY_ADMIN_TOKEN` que ya protege `/metrics` (`handler.WithAdmin`).
- [x] 3.3 Filtro por capacidad -- DESVIACIÓN: no existe tabla `models` en PostgreSQL (los modelos viven en el Registry YAML, no en DB), así que el filtro no es un `IN (SELECT ...)` SQL sino un post-filtro en memoria: `AlertsHandler.allowed()` resuelve capacidades de (provider, model) vía `CapabilityLookup` (inyectado desde `registry.Registry.Providers()` en `cmd/gateway/main.go:buildCapabilityLookup`) y compara contra `auth.Identity.Scopes`. Verificado con Postgres real filtrando por RBAC en `internal/handler/alerts_integration_test.go`.
- [x] 3.4 Implementar paginación `page`/`limit` (default `limit=50`) aplicada después del filtrado RBAC
- [x] 3.5 Registrar ruta `GET /alerts` en el router HTTP junto a los demás endpoints protegidos
- [x] 3.6 Tests unitarios: AC1-AC5 de HU-EVO-013

## 4. Wiring end-to-end y verificación

- [x] 4.1 Test de integración: boot del Gateway con PostgreSQL real (o testcontainer), Groq bajo umbral, Alert Manager corre, `GET /alerts` devuelve la alerta filtrada por tenant
- [x] 4.2 Test de integración: `GET /metrics` expone bloque `quota` poblado end-to-end (no mock aislado)
- [x] 4.3 `go build ./... && go vet ./... && go test ./... -race` en verde, sin regresión sobre EP-EVO-001/002
- [x] 4.4 journey_smoke: boot real, `GET /metrics` con quota no vacío, `GET /alerts` responde 200 filtrado por tenant del token de prueba
