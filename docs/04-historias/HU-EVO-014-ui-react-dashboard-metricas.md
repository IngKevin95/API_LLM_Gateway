---
id: HU-EVO-014
titulo: UI React - Dashboard de métricas con tabs (Overview, Quotas, Alerts, Providers)
epica: EP-EVO-003
prioridad: Must
complejidad: L
estado: draft
---

# UI React - Dashboard de métricas con tabs

Como **operador del Gateway**, quiero **una interfaz React que muestre métricas, cuota por proveedor/modelo, alertas activas, y estado de proveedores en tiempo real**, para **monitorear el Gateway sin APIs de texto**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — dashboard carga | Dado que usuario entra en `/dashboard` | Cuando página abre | Entonces renderiza 4 tabs (Overview, Quotas, Alerts, Providers) y es interactivo |
| 2 | Happy — Overview tab | Dado que Gateway lleva 10h online | Cuando Overview se abre | Entonces muestra uptime, total requests, errors, latency p50/p95/p99 (igual que HU-060 JSON) |
| 3 | Happy — Quotas tab | Dado que 5 proveedores se conocen | Cuando Quotas tab se abre | Entonces tabla mostrando [Provider, Model, Limit, Remaining, ResetAt, HealthStatus] actualizada cada 5s |
| 4 | Happy — Alerts tab | Dado que hay 3 alertas activas | Cuando Alerts tab se abre | Entonces lista alertas filtradas por tenant/scope, mostrando [Severity, Provider, Model, Message, AlertTime], rojo si critical |
| 5 | Happy — Providers tab | Dado que 5 proveedores existen | Cuando Providers tab se abre | Entonces muestra estado (healthy/unhealthy), última respuesta, circuit breaker status |
| 6 | Edge — refresh automático | Dado que usuario en Quotas tab | Cuando pasan 5s | Entonces datos se actualizan automáticamente sin reload manual |
| 7 | Edge — filtrado por tenant | Dado que usuario T1 accede | Cuando datos se renderizan | Entonces solo ve alertas/modelos de T1 (respeta RBAC) |

## Checklist INVEST

- [x] Independent — depende de HU-EVO-011/012/013 (endpoints)
- [x] Negotiable — diseño UI extensible
- [x] Valuable — visibilidad operacional
- [x] Estimable — React + recharts + polling
- [x] Small/Medium — 3-4 días
- [x] Testable — cypress E2E, verifica tabs + data

## Notas técnicas

React app en `src/ui/dashboard/`:
```
Dashboard.jsx
├── Overview.jsx (uptime, requests, latency)
├── Quotas.jsx (tabla + charts)
├── Alerts.jsx (tabla roja/amarilla)
├── Providers.jsx (estado + history)
└── hooks/useMetrics.js (fetch cada 5s)
```

`useMetrics.js`:
```jsx
const [metrics, setMetrics] = useState(null);
useEffect(() => {
  const interval = setInterval(async () => {
    const res = await fetch('/metrics');
    setMetrics(await res.json());
  }, 5000);
  return () => clearInterval(interval);
}, []);
```

---

## Relación con existentes

- Usa: HU-EVO-011 (`/metrics`), HU-EVO-013 (`/alerts`)
- Integra: HU-009 (auth via API key/token)
- Requisito para: visibilidad operacional
