## Why

El Gateway expone `GET /metrics` (con bloque `quota`) y `GET /alerts` (RBAC, EP-EVO-003 SS1) pero
no existe ninguna interfaz visual: un operador solo puede inspeccionar el estado del Gateway con
`curl` o clientes ad-hoc. Sin un dashboard, la detección proactiva de cuota baja (SS1) no llega a
las personas que deben actuar, y no hay forma de monitorear uptime/latencia/proveedores sin leer
JSON crudo.

## What Changes

- Nueva app React (`src/ui/dashboard/`) montada como SPA estática, con `Dashboard.jsx` como shell
  de 4 tabs (Overview, Quotas, Alerts, Providers) siguiendo el prototipo Stitch "Gateway Ops Dark".
  - `Overview.jsx`: uptime, total requests, errors, latencia p50/p95/p99 (mismo payload que
    `GET /metrics`, ya expuesto desde HU-060).
  - `Quotas.jsx`: tabla `[Provider, Model, Limit, Remaining, ResetAt, HealthStatus]` desde el
    bloque `quota` de `/metrics` (HU-EVO-011), refrescada cada 5s.
  - `Alerts.jsx`: lista de `GET /alerts` (HU-EVO-013) con `[Severity, Provider, Model, Message,
    AlertTime]`, fila roja si `severity=critical`.
  - `Providers.jsx`: grid de tarjetas por proveedor con estado healthy/unhealthy, última respuesta
    y circuit breaker status (derivado de `/metrics`).
  - `hooks/useMetrics.js`: polling a `/metrics` cada 5s con cleanup de `setInterval`.
- Notificaciones browser (`hooks/useAlerts.js`): polling a `/alerts` cada 30s, dedup de toasts
  activos por `provider+model` (ventana 10s), `react-toastify` para warning/critical, beep de audio
  en critical, `Notification.requestPermission()` + `new Notification(...)` para alertas fuera del
  foco del navegador, umbral de UI configurable en `localStorage` sin redeploy.
- Ambas HU respetan RBAC de forma pasiva: consumen los endpoints ya filtrados server-side (SS1), no
  duplican lógica de autorización en el cliente.

## Capabilities

### New Capabilities
- `dashboard-ui`: SPA React con 4 tabs consumiendo `/metrics` y `/alerts` con polling.
- `browser-notifications`: toasts + sonido + Notification API sobre alertas de cuota, con dedup y
  umbral configurable.

### Modified Capabilities
- Ninguna en backend; este change es puramente frontend, consumidor de contratos ya estables
  (`metrics-store`, `alerts-endpoint` de `metrics-quota-alertas-rbac`).

## Impact

- Código: `src/ui/dashboard/**` (nuevo), `package.json` raíz extendido con dependencias de
  frontend (`react`, `react-dom`, `react-toastify`, `vite`, `vitest`, `@testing-library/react`) —
  contrastadas contra `.claude/config/stack-allowlist.json`.
- Datos: ninguno (solo lectura de endpoints existentes).
- Servido: build estático de Vite; en Fase 1 se sirve vía `npm run build` + carpeta `dist/`
  montada por un servidor de archivos estáticos ligero (documentado en tasks.md); no se acopla el
  binario Go a servir la SPA en este change (evita ampliar alcance de `cmd/gateway`).
- Sin impacto en `src/internal/**` Go — solo consumo HTTP de contratos ya construidos en SS1.

## Trazabilidad

- Épica: EP-EVO-003
- Sub-slice: EP-EVO-003-SS2
- Historias: HU-EVO-014, HU-EVO-015
