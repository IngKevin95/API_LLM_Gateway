---
id: HU-EVO-003
titulo: Extender conformance_test.go para validar cada ProviderSpec
epica: EP-EVO-001
prioridad: Should
complejidad: M
estado: lista
---

# Extender conformance_test.go para validar cada ProviderSpec

Como **desarrollador del Gateway**, quiero **que cada ProviderSpec pase conformance tests automáticamente antes de producción**, para **garantizar que nuevos proveedores implementan Chat/Stream/Embed correctamente**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — spec pasa conformance | Dado que Groq está en `free-tier.yaml` con spec válido | Cuando corre `go test ./...` | Entonces conformance_test.go itera todos los providers, llama Chat/Stream/Embed, y verifica respuestas ✓ |
| 2 | Happy — respuesta normalizada | Dado que un provider devuelve respuesta en formato nativo (ej: Groq OpenAI-compatible) | Cuando adapter normaliza a schema interno | Entonces respuesta tiene `Content, Model, Usage` poblados, sin error |
| 3 | Error — spec sin modelo default | Dado que un provider en `free-tier.yaml` tiene `models: []` | Cuando conformance test intenta seleccionar un modelo | Entonces falla gracefully con `ErrNoModelAvailable` y el test reporta qué fue |
| 4 | Edge — timeout en test | Dado que un provider es muy lento (>10s) | Cuando conformance test llama Chat() con timeout 5s | Entonces aborta y reporta timeout sin bloquear suite completa |
| 5 | Edge — múltiples specs en paralelo | Dado que hay 5 providers a validar | Cuando conformance_test corre en paralelo (t.Parallel()) | Entonces todos completan sin race conditions ni port conflicts |

## Checklist INVEST

- [x] Independent — depende de HU-EVO-001/002 (adapter y YAML)
- [x] Negotiable — suites de test extensible
- [x] Valuable — detecta regressions antes de producción
- [x] Estimable — tabla-driven tests con mocks
- [x] Small — 1-2 días
- [x] Testable — tests de los tests (meta)

## Notas técnicas

Archivo `src/internal/adapter/conformance_test.go` extendido:

```go
func TestConformanceAllProviders(t *testing.T) {
    // Carga free-tier.yaml
    specs := registry.Load("config/providers/free-tier.yaml")
    
    for _, spec := range specs {
        t.Run(spec.ID, func(t *testing.T) {
            t.Parallel()
            // Test Chat, Stream, Embed con spec mock
            // Verifica respuesta normalizada
        })
    }
}
```

Mocks: usar `httptest.Server` devolviendo respuestas de cada proveedor en formato nativo.

---

## Relación con existentes

- Extiende: `src/internal/adapter/conformance_test.go` (existente)
- Usa: HU-EVO-002 (free-tier.yaml), HU-EVO-001 (adapter genérico)
- Bloqueada por: HU-EVO-001 (implementación del adapter)

## Change

Implementado por el openspec change `adapters-proveedores-gratuitos-curados` (EP-EVO-001, branch `feature/ep-evo-001-adapters-gratuitos`). Ver `openspec/changes/adapters-proveedores-gratuitos-curados/`.
