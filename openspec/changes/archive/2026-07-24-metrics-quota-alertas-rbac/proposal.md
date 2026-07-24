## Why

El Gateway hoy aprende cuota por proveedor/modelo en runtime (`quota.Manager`, EP-EVO-002) pero no
la expone en `/metrics`, no genera alertas proactivas cuando la cuota remanente cae bajo un umbral,
y no existe ningún endpoint para que operadores/usuarios consulten esas alertas respetando
multi-tenancy y RBAC. Sin esto, un operador solo se entera de que un proveedor se agotó cuando el
Router ya falló failover hacia él.

## What Changes

- `metrics.Store` invoca `quota.Manager.Snapshot()` en cada request a `/metrics`, exponiendo un
  bloque `quota: [{provider, model, limit, remaining, reset_at, healthy}]` por proveedor y modelo;
  filtra por scopes del token cuando el requester no es admin; responde en <100ms al ser lectura en
  RAM sin I/O.
- Nuevo `alert.Manager` (`src/internal/alert/manager.go`): worker en background que corre cada 1
  minuto, evalúa `quota.Manager.Snapshot()` contra un umbral configurable (default 10%), inserta
  alertas `severity: warning` (remaining < umbral) o `severity: critical` (remaining == 0) en tabla
  PostgreSQL `provider_alerts`, con dedup (no duplica alerta activa, solo actualiza `updated_at`) y
  granularidad por proveedor+modelo.
- Nuevo endpoint `GET /alerts` (`src/internal/handler/alerts.go`) que consulta `provider_alerts`
  filtrando por `tenant_id` del token de auth y por scopes/capabilities autorizados; admin
  (`GATEWAY_ADMIN_TOKEN`) ve todas las alertas sin filtro de tenant; soporta paginación
  (`page`/`limit`).

## Capabilities

### New Capabilities
- `alert-manager`: evaluación periódica de cuota contra umbral configurable y persistencia de
  alertas en PostgreSQL con dedup.
- `alerts-endpoint`: endpoint HTTP `GET /alerts` con filtrado RBAC multi-tenant y paginación.

### Modified Capabilities
- `metrics-store` (extiende `src/internal/metrics/store.go`, HU-060): agrega snapshot de cuota por
  proveedor/modelo al payload de `/metrics`, con filtrado por scopes del requester.

## Impact

- Código: `src/internal/metrics/store.go` (extendido), `src/internal/alert/*` (nuevo),
  `src/internal/handler/alerts.go` (nuevo), routing en `cmd/gateway/main.go` para registrar
  `GET /alerts` y arrancar el worker del Alert Manager.
- Datos: nueva tabla PostgreSQL `provider_alerts` (provider_id, model_id, severity, message,
  alert_time, updated_at, resolved_at, UNIQUE provider_id+model_id+alert_time). Requiere migración
  y conexión real a PostgreSQL (ya prevista/cableada desde EP-EVO-002/HU-EVO-008).
- Dependencias: ninguna nueva (reutiliza `database/sql` + driver Postgres ya presente en el stack
  desde EP-EVO-002).
- Sin impacto en UI directo — este sub-slice es backend puro; alimenta a HU-EVO-014/015 (UI React,
  SS2 de esta misma épica) que consumirán `/metrics` (quota) y `/alerts`.

## Trazabilidad

- Épica: EP-EVO-003
- Historias: HU-EVO-011, HU-EVO-012, HU-EVO-013
