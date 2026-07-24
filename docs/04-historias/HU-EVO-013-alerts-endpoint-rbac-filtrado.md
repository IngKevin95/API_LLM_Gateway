---
id: HU-EVO-013
titulo: Filtrado RBAC en GET /alerts (respeta tenant + scopes)
epica: EP-EVO-003
prioridad: Should
complejidad: M
estado: lista
---

# Filtrado RBAC en GET /alerts (respeta tenant + scopes)

Como **usuario del Gateway**, quiero **que GET `/alerts?tenant=T1` devuelva solo alertas relevantes a mi tenant y capacidades autorizadas**, para **no ver alertas de otros tenants ni modelos que no puedo usar**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — cliente ve sus alertas | Dado que cliente T1 con scope `capability:coding` tiene alerta en Groq mixtral | Cuando GET `/alerts` se ejecuta con su API key | Entonces devuelve esa alerta (pertenece a su tenant + scope) |
| 2 | Happy — cliente NO ve otros tenants | Dado que cliente T1 genera request | Cuando GET `/alerts` se ejecuta | Entonces no devuelve alertas de T2, T3 (filtra por tenant del token) |
| 3 | Happy — admin ve todas | Dado que admin token (GATEWAY_ADMIN_TOKEN) se usa | Cuando GET `/alerts` sin filtro tenant | Entonces devuelve todas las alertas del Gateway |
| 4 | Error — scope insuficiente | Dado que usuario tiene scope `capability:chat` pero alerta es de `capability:vision` | Cuando GET `/alerts` se ejecuta | Entonces filtra la alerta (usuario no ve alertas de capacidades no autorizadas) |
| 5 | Edge — paginación | Dado que hay 1000 alertas | Cuando GET `/alerts?page=2&limit=50` se ejecuta | Entonces devuelve 50 alertas, page 2, respetando RBAC |

## Checklist INVEST

- [x] Independent — depende de HU-EVO-012 (alerts table)
- [x] Negotiable — schema de query extensible
- [x] Valuable — multi-tenant, privacidad de datos
- [x] Estimable — query con JOINs + auth checks
- [x] Small — 1-2 días
- [x] Testable — mock auth context, verifica filtering

## Notas técnicas

Endpoint en `src/internal/handler/alerts.go`:

```go
func (h *Handler) GetAlerts(w http.ResponseWriter, r *http.Request) {
    auth := r.Context().Value("auth").(AuthContext)
    
    // Query: SELECT * FROM provider_alerts WHERE tenant_id = auth.TenantID
    // AND model_id IN (SELECT model_id FROM models WHERE capability IN auth.Scopes)
    
    alerts := h.db.GetAlerts(
        ctx,
        auth.TenantID,
        auth.Scopes,
        pagination,
    )
    
    json.NewEncoder(w).Encode(alerts)
}
```

---

## Relación con existentes

- Crea: endpoint `GET /alerts`
- Usa: HU-EVO-012 (table provider_alerts)
- Integra: HU-009 (RBAC existente)
- Alimenta: HU-EVO-015 (notificaciones)

## Change

Implementado en `openspec/changes/metrics-quota-alertas-rbac`. Ver `tasks.md` para el detalle de cobertura y desviaciones documentadas.
