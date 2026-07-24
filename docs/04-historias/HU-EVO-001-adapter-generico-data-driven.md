---
id: HU-EVO-001
titulo: Crear adapter genérico data-driven
epica: EP-EVO-001
prioridad: Must
complejidad: M
estado: lista
---

# Crear adapter genérico data-driven

Como **arquitecto del Gateway**, quiero **que cada proveedor sea una configuración declarativa (baseURL, authHeader, formato) sin código Go nuevo**, para **agregar 20+ proveedores sin duplicar lógica de adapters**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — adapter OpenAI-compatible | Dado que un `ProviderSpec{baseURL: "https://api.groq.com/openai/v1", authHeader: "Authorization", format: "openai"}` | Cuando se instancia el adapter genérico con ese spec | Entonces implementa `Chat/Stream/Embed` reenviando a Groq con headers correctos, devuelve respuesta OpenAI-compatible |
| 2 | Happy — adapter Claude-compatible | Dado que un `ProviderSpec{baseURL: "https://api.anthropic.com", authHeader: "x-api-key", format: "claude"}` | Cuando se instancia el adapter genérico con ese spec | Entonces traduce request OpenAI → Claude format, devuelve respuesta normalizada |
| 3 | Error — spec inválido | Dado que un `ProviderSpec` con baseURL vacía o formato desconocido | Cuando se intenta instanciar | Entonces retorna error `ErrInvalidProviderSpec` sin hacer request |
| 4 | Edge — headers extra por proveedor | Dado que un spec con `headers: {"X-Custom": "value"}` | Cuando envía request | Entonces inyecta headers extra en cada request sin sobrescribir autenticación |
| 5 | Edge — timeout por adapter | Dado que un spec con `timeoutMs: 5000` diferente del global | Cuando corre timeout en ejecución | Entonces respeta el timeout del spec, no el global |

## Checklist INVEST

- [x] Independent — depende de Registry (HU-EVO-002) cargando el spec
- [x] Negotiable — formato de spec abierto, extensible
- [x] Valuable — permite agregar proveedores sin código Go
- [x] Estimable — wrapping de `openai.Adapter` + traducción de formato
- [x] Small — 2-3 días
- [x] Testable — conformance_test.go con specs mock

## Notas técnicas

Struct `ProviderSpec` en `src/internal/adapter/types.go`:
```go
type ProviderSpec struct {
    ID           string
    BaseURL      string
    AuthType     string // "apikey" | "bearer"
    AuthHeader   string // "Authorization" | "x-api-key"
    Format       string // "openai" | "claude" | "gemini"
    Headers      map[string]string
    TimeoutMs    int
}
```

Adapter genérico en `src/internal/adapter/generic/adapter.go` implementa `Adapter` interface.

---

## Relación con existentes

- Usa: `src/internal/adapter/openai.Adapter` (reutiliza wrapping)
- Extiende: `src/internal/registry/registry.go` (carga spec desde YAML)
- Requiere: HU-EVO-002 (Registry ampliado con `free-tier.yaml`)

## Change

Implementado por el openspec change `adapters-proveedores-gratuitos-curados` (EP-EVO-001, branch `feature/ep-evo-001-adapters-gratuitos`). Ver `openspec/changes/adapters-proveedores-gratuitos-curados/`.
