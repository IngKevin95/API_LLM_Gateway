---
id: HU-EVO-002
titulo: Cargar catálogo free-tier.yaml en Registry
epica: EP-EVO-001
prioridad: Must
complejidad: S
estado: lista
---

# Cargar catálogo free-tier.yaml en Registry

Como **operador del Gateway**, quiero **que Registry cargue `config/providers/free-tier.yaml` con 5 proveedores gratuitos prioritarios** (Groq, Cerebras, Mistral, Gemini, Cloudflare AI), para **agregar 1.19B tokens/mes documentados sin tocar código**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — carga YAML válido | Dado que existe `config/providers/free-tier.yaml` con 5 proveedores | Cuando Registry.Load() se ejecuta en boot | Entonces deserializa cada proveedor, lo registra con spec completo, y está disponible para routing |
| 2 | Happy — provider en YAML sobrescribe config.yaml | Dado que Groq está en ambos archivos | Cuando Registry carga free-tier.yaml después de config.yaml | Entonces la versión free-tier (con quota_hint más realista) es la que queda activa |
| 3 | Error — YAML malformado | Dado que `free-tier.yaml` tiene sintaxis JSON en lugar de YAML | Cuando Registry.Load() parsea | Entonces retorna `ErrInvalidConfig` y no inicia el Gateway (fail-fast) |
| 4 | Edge — proveedor sin modelo default | Dado que un proveedor en YAML tiene `models: []` vacío | Cuando Router intenta resolver una capacidad | Entonces lo excluye automáticamente del scoring sin crash |
| 5 | Edge — quota_hint negativo | Dado que un proveedor tiene `quota_hint: -100` | Cuando Quota Manager lo consulta | Entonces trata como cuota agotada (remaining = 0) y lo retira de la selección |

## Checklist INVEST

- [x] Independent — depende de HU-EVO-001 (adapter genérico)
- [x] Negotiable — YAML structure abierta
- [x] Valuable — habilita 5 proveedores sin código Go
- [x] Estimable — extensión de Registry.Load() existente
- [x] Small — 1 día
- [x] Testable — tests con YAML válido/inválido/edge

## Notas técnicas

Archivo `config/providers/free-tier.yaml` (ejemplo):
```yaml
providers:
  - id: groq
    type: groq
    baseUrl: https://api.groq.com/openai/v1
    authHeader: Authorization
    format: openai
    models:
      - id: mixtral-8x7b-32768
        name: Mixtral 8x7b
        capabilities: [chat, coding]
        quality_score: 85
        cost_per_1m_tokens: 0
    quota_hint: 14400  # 30 RPM * 24 * 20 days
    circuit_breaker:
      failure_threshold_percent: 50
      reset_timeout_sec: 30
  # ... 4 más
```

Registry en `src/internal/registry/registry.go` extendido:
```go
func (r *Registry) LoadFromFile(path string) error {
    // Load config.yaml
    // Load free-tier.yaml y sobrescribe matching providers
    // Merge y valida
}
```

---

## Relación con existentes

- Extiende: `src/internal/registry/registry.go` (LoadFromFile)
- Usa: HU-EVO-001 (adapter genérico para estos proveedores)
- Requiere: config/providers/free-tier.yaml creado en Fase 1

## Change

Implementado por el openspec change `adapters-proveedores-gratuitos-curados` (EP-EVO-001, branch `feature/ep-evo-001-adapters-gratuitos`). Ver `openspec/changes/adapters-proveedores-gratuitos-curados/`.
