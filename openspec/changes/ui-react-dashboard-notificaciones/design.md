## Context

SS1 de EP-EVO-003 dejó `GET /metrics` (con bloque `quota`) y `GET /alerts` (RBAC) mergeados en
`develop`. Este change (SS2) es la primera pieza de frontend del proyecto: no hay convención previa
de UI en el repo, así que se fija aquí el patrón para futuras SPAs. Fuente de diseño confirmada:
proyecto Stitch 12981760791975432480 "API LLM Gateway - Dashboard de Métricas", design system
"Gateway Ops Dark" (dark, denso en datos, tabs, badges semáforo, tipografía monoespaciada para
métricas), 4 pantallas (Overview/Quotas/Alerts/Providers).

## Goals / Non-Goals

**Goals:**
- Dashboard React con 4 tabs, fiel al prototipo Stitch, consumiendo `/metrics` y `/alerts` reales.
- Notificaciones browser (toast + sonido + Notification API) accionables sin mantener el dashboard
  en foco.
- Polling simple (setInterval), sin WebSocket — consistente con el patrón ya usado en el propio
  backend (`/metrics` es un endpoint de lectura barata, <100ms).

**Non-Goals:**
- No se implementa autenticación de UI (login) — el dashboard asume que el operador ya tiene un API
  key/token que pega en un campo de configuración local (`localStorage`), reutilizando el mismo
  mecanismo Bearer que ya usan `/metrics` y `/alerts`.
- No se acopla el binario `cmd/gateway` a servir la SPA en este change (se sirve por separado vía
  `vite preview` / servidor estático; una épica futura puede decidir embeberla).
- No se agregan gráficas de series de tiempo (sparklines de Providers.jsx se resuelven con datos ya
  disponibles en `/metrics`, sin nueva fuente de datos histórica).

## Decisions

### 1. Vite + React (sin framework SSR) para la SPA
Se elige Vite (dev server rápido, build estático simple) en vez de Next.js: el dashboard es un
cliente puro de APIs ya existentes, no necesita SSR/routing de servidor. Alinea con
`stack-allowlist.json` (a extender con `vite`, `react`, `react-dom`).
- Alternativa descartada: Next.js — rechazado por sobre-ingeniería para un cliente HTTP puro sin
  necesidad de SSR ni rutas de servidor propias.

### 2. Autenticación de UI vía campo de configuración + localStorage, no login propio
El dashboard pide un API key/Bearer token una sola vez (modal de configuración), lo persiste en
`localStorage`, y lo adjunta como header `Authorization: Bearer <token>` en cada fetch a
`/metrics`/`/alerts`. RBAC real sigue viviendo 100% server-side (SS1); el cliente nunca decide qué
mostrar más allá de lo que el servidor ya filtró.
- Alternativa descartada: implementar un flujo de login con sesión — rechazado porque el Gateway no
  tiene modelo de usuarios/sesiones, solo API keys estáticas (HU-009); un login propio duplicaría
  ese concepto sin necesidad.

### 3. `useMetrics`/`useAlerts` como hooks independientes, cada uno con su propio intervalo
`useMetrics` hace polling cada 5s (AC3/AC6 HU-EVO-014); `useAlerts` hace polling cada 30s (nota
técnica HU-EVO-015) y es responsable exclusivo de disparar notificaciones — separa "refrescar datos
visibles" de "vigilar y notificar", evitando que cada refresco de Quotas dispare un re-chequeo de
notificaciones cada 5s (ruido de toasts).
- Alternativa descartada: un solo hook combinado — rechazado porque acopla dos cadencias distintas
  (UI en tiempo casi real vs. vigilancia de alertas) y complica el dedup de toasts.

### 4. Dedup de toasts por clave `provider+model` con ventana de 10s, en memoria del hook
`useAlerts` mantiene un `Map<string, {toastId, lastSeenAt}>` en un `useRef`; antes de emitir un
toast nuevo, revisa si ya existe uno activo para la misma clave dentro de los últimos 10s — si
existe, llama `toast.update(toastId, ...)` en vez de `toast(...)`. El beep y la `Notification`
del sistema solo se disparan la primera vez que aparece la alerta (no en cada poll mientras sigue
activa), usando el mismo `Map` para no repetir el sonido cada 30s.
- Alternativa descartada: dedup por `alert.id` del backend — insuficiente porque el backend
  actualiza `updated_at` en la misma fila (SS1 dedup), así que distintos polls devuelven el mismo
  `id`; el criterio real de "ya la mostré" debe vivir en el cliente, no asumir unicidad de payload.

### 5. Umbral de notificación configurable en `localStorage`, desacoplado del umbral del backend
El Alert Manager (SS1) ya filtra/clasifica server-side con `GATEWAY_ALERT_THRESHOLD` (default 10%).
La UI agrega un umbral *adicional* propio (AC5 HU-EVO-015, default 10%, editable en la modal de
configuración) que solo decide qué severidad de alerta ya recibida dispara notificación visual —
no reclasifica ni sobrescribe la severidad que asignó el servidor.
- Alternativa descartada: exponer `GATEWAY_ALERT_THRESHOLD` como configurable vía UI y hacer un
  `PATCH` al backend — fuera de alcance de este change (requeriría endpoint de administración
  nuevo, no pedido por ningún AC de HU-EVO-014/015).

## Risks / Mitigation

- **Riesgo**: `Notification.requestPermission()` requiere gesto de usuario en navegadores modernos;
  si se llama automáticamente al montar, el navegador la ignora silenciosamente.
  Mitigación: el permiso se pide desde un botón explícito en la modal de configuración de
  notificaciones, no en el `useEffect` de montaje.
- **Riesgo**: polling paralelo de `/metrics` (5s) y `/alerts` (30s) desde múltiples tabs abiertas
  del operador puede generar carga innecesaria.
  Mitigación: fuera de alcance de este change (no hay AC que lo pida); documentado como deuda
  conocida para una épica futura de "single tab leader election" si se vuelve un problema real.
- **Riesgo**: sin build previo de frontend en el repo, el hook `stack-guard.sh` podría no reconocer
  las dependencias nuevas.
  Mitigación: se agregan explícitamente a `.claude/config/stack-allowlist.json` antes de instalar,
  como parte de tasks.md grupo 1.

## Migration Plan

1. Extender `package.json` raíz con dependencias de frontend + scripts `dev:ui`/`build:ui`/
   `test:ui` (no reemplaza el `package.json` existente, que ya tiene `openai` como dependencia de
   otro componente).
2. Crear `src/ui/dashboard/` desde cero: `Dashboard.jsx`, 4 componentes de tab, 2 hooks, tests con
   Vitest + Testing Library.
3. `vite.config.js` con proxy de dev (`/metrics`, `/alerts` → `http://localhost:8080`) para que el
   dev server no choque con CORS contra el binario Go real durante `journey_smoke`.
4. No se toca `cmd/gateway/main.go` ni ningún paquete `src/internal/**`.
