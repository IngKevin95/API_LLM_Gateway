---
id: HU-EVO-005
titulo: Quota Manager inicializa contadores por proveedor desde YAML
epica: EP-EVO-001
prioridad: Should
complejidad: S
estado: lista
---

# Quota Manager inicializa contadores por proveedor desde YAML

Como **operador del Gateway**, quiero **que Quota Manager cargue `quota_hint` de cada proveedor en `free-tier.yaml` y lo use como valor inicial de cuota restante**, para **tener un piso realista de cuota antes del primer aprendizaje desde headers**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — init desde YAML | Dado que Groq en YAML tiene `quota_hint: 14400` | Cuando Quota Manager boot | Entonces `Remaining("groq")` devuelve 14400 sin hacer requests |
| 2 | Happy — sobrescribe en runtime | Dado que primer request a Groq devuelve `X-RateLimit-Remaining: 14300` | Cuando Quota Manager aprende el header | Entonces actualiza remaining a 14300 (learned value > quota_hint) |
| 3 | Error — quota_hint 0 o negativo | Dado que un proveedor tiene `quota_hint: 0` | Cuando Quota Manager lo carga | Entonces trata como agotado (remaining = 0) y lo excluye de Router hasta primer learning |
| 4 | Edge — proveedor sin quota_hint en YAML | Dado que un proveedor nuevo no tiene `quota_hint` definido | Cuando Quota Manager lo carga | Entonces asume default 1M tokens (una semana típica de free tier) |
| 5 | Edge — reinicio persiste learned quotas | Dado que antes del reinicio, aprendimos que Mistral tiene 500M remaining | Cuando Gateway reinicia y Quota Manager lee PostgreSQL | Entonces restaura 500M como remaining (no vuelve a YAML quota_hint) |

## Checklist INVEST

- [x] Independent — depende de HU-EVO-002 (YAML cargado)
- [x] Negotiable — default quota_hint configurable
- [x] Valuable — evita crashes cuando proveedor es nuevo
- [x] Estimable — actualización de Quota Manager init
- [x] Small — 1 día
- [x] Testable — tests con YAML + sin YAML

## Notas técnicas

Quota Manager en `src/internal/quota/manager.go` extendido:

```go
func (m *Manager) InitFromRegistry(reg *registry.Registry) {
    for _, provider := range reg.Providers {
        quotaHint := provider.QuotaHint
        if quotaHint <= 0 {
            quotaHint = 1_000_000 // Default 1M
        }
        m.quotas[provider.ID] = Quota{
            Limit: quotaHint,
            Remaining: quotaHint,
            UpdatedAt: time.Now(),
        }
    }
}
```

YAML schema:
```yaml
providers:
  - id: groq
    quota_hint: 14400  # requests, no tokens (Groq usa RPM)
```

---

## Relación con existentes

- Extiende: `src/internal/quota/manager.go` (HU-006, existente)
- Usa: HU-EVO-002 (YAML con quota_hint)
- Requisito para: HU-EVO-007 (learning desde headers sobrescribe quota_hint)

## Nota de alcance — AC5 diferido (2026-07-23)

AC5 ("reinicio persiste learned quotas" vía PostgreSQL) queda cubierto en EP-EVO-001 solo a nivel
unitario (`quota.RestoreRemaining()` probado con valores inyectados directamente, sin fuente
PostgreSQL real). La integración con PostgreSQL en boot (leer/escribir cuota persistida) se difiere
a HU-EVO-008 ("Persistencia async en PostgreSQL de learned quotas", EP-EVO-002, ya en
`docs/03-backlog/backlog.md`), que es el slice correcto para cablear el backing store real — por
acuerdo explícito del equipo: EP-EVO-001 es Fase 1 in-memory-first (Quota Manager sin backing store
externo todavía; `main.go` usa `quota.NewInMemoryManager()`), y cablear PostgreSQL en este slice
excedería su alcance de "adapters + registry + health + quota in-memory" declarado en
`proposal.md`. Este diferimiento no baja la HU de `estado: lista` (los otros 4 AC sí están
cerrados end-to-end); solo el AC5 queda marcado como pendiente de un follow-up.

## Change

Implementado por el openspec change `adapters-proveedores-gratuitos-curados` (EP-EVO-001, branch `feature/ep-evo-001-adapters-gratuitos`). Ver `openspec/changes/adapters-proveedores-gratuitos-curados/`.
