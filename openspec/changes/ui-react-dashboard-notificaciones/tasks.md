## 1. Scaffold de frontend (fundacional para este change)

- [ ] 1.1 Extender `.claude/config/stack-allowlist.json` con `react`, `react-dom`, `react-toastify`, `vite`, `vitest`, `@testing-library/react`, `@testing-library/jest-dom`, `jsdom`, `@vitejs/plugin-react`
- [ ] 1.2 Extender `package.json` raíz: dependencies (`react`, `react-dom`, `react-toastify`) + devDependencies (`vite`, `vitest`, `@vitejs/plugin-react`, `@testing-library/react`, `@testing-library/jest-dom`, `jsdom`) + scripts `dev:ui`, `build:ui`, `test:ui`
- [ ] 1.3 Crear `vite.config.js` con proxy dev `/metrics` y `/alerts` → `http://localhost:8080`
- [ ] 1.4 Crear `src/ui/dashboard/index.html` + `src/ui/dashboard/main.jsx` (entrypoint Vite)
- [ ] 1.5 `npm install` limpio, verificar `npm run build:ui` sin error

## 2. Dashboard shell + 4 tabs (HU-EVO-014)

- [ ] 2.1 `src/ui/dashboard/Dashboard.jsx`: shell con navegación de 4 tabs (Overview/Quotas/Alerts/Providers), estilo "Gateway Ops Dark" (dark, tabs, badges semáforo, tipografía monoespaciada para métricas) fiel al prototipo Stitch
- [ ] 2.2 `src/ui/dashboard/hooks/useMetrics.js`: polling a `/metrics` cada 5s con cleanup de `setInterval`, adjunta `Authorization: Bearer <token>` desde `localStorage`
- [ ] 2.3 `src/ui/dashboard/Overview.jsx`: uptime/requests/errors/latencia p50-p95-p99 desde `useMetrics`
- [ ] 2.4 `src/ui/dashboard/Quotas.jsx`: tabla Provider/Model/Limit/Remaining/ResetAt/HealthStatus desde `useMetrics().quota`
- [ ] 2.5 `src/ui/dashboard/Alerts.jsx`: lista Severity/Provider/Model/Message/AlertTime desde `GET /alerts` (fila roja si critical)
- [ ] 2.6 `src/ui/dashboard/Providers.jsx`: grid de tarjetas healthy/unhealthy + última respuesta + circuit breaker, desde `useMetrics().providers`
- [ ] 2.7 Modal de configuración: campo API key/Bearer (persistido en `localStorage`)
- [ ] 2.8 Tests Vitest + Testing Library: AC1-AC7 de HU-EVO-014 (render de 4 tabs, contenido por tab, refresco 5s con fake timers, filtrado RBAC pasivo)

## 3. Notificaciones browser (HU-EVO-015)

- [ ] 3.1 `src/ui/dashboard/hooks/useAlerts.js`: polling a `/alerts` cada 30s
- [ ] 3.2 Integrar `react-toastify`: `ToastContainer` en `Dashboard.jsx`, `toast.warning`/`toast.error` según severidad
- [ ] 3.3 Dedup por clave `provider+model` con ventana 10s (`useRef` Map) — `toast.update` en vez de nuevo `toast`
- [ ] 3.4 `playBeep()`: audio corto embebido (data URI), solo en `critical`, solo primera vez que aparece la alerta
- [ ] 3.5 `notifyBrowser()`: `Notification.requestPermission()` gated por botón explícito en modal de settings; `new Notification(...)` en `critical` si permiso otorgado
- [ ] 3.6 Settings de umbral de notificación en `localStorage` (default 10%), leído por `useAlerts` en cada poll
- [ ] 3.7 Tests Vitest: AC1-AC5 de HU-EVO-015 (toast warning/critical, dedup, beep solo una vez, umbral configurable cambia comportamiento)

## 4. Wiring end-to-end y verificación

- [ ] 4.1 `npm run build:ui` sin error, `npm run test:ui` en verde
- [ ] 4.2 journey_smoke: boot real del Gateway Go + `npm run dev:ui` (o `vite preview` sobre build), navegar los 4 tabs con datos reales de `/metrics`/`/alerts`, confirmar refresco 5s y toasts reales
- [ ] 4.3 ux-fidelity-reviewer: comparar (MCP devtools) las 4 pantallas reales contra el prototipo Stitch (Overview/Quotas/Alerts/Providers)
- [ ] 4.4 wiring-adversarial-verifier: sin stubs, todos los AC de HU-EVO-014/015 con prueba real ejecutada
