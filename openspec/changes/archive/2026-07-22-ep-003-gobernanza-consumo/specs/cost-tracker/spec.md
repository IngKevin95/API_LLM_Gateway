# Cost Tracker Specification

## 1. Responsibilities
Atribuir y registrar el costo transaccional (en dólares o unidades) a cada petición despachada exitosamente. Provee los datos crudos para visualización en dashboards y para cálculos de KPI de rentabilidad.

## 2. Interface
```go
package cost

import "context"

type CostRecord struct {
    AgentID      string
    ProviderID   string
    Model        string
    PromptTokens int
    CompletionTokens int
    Cost         float64 // Costo calculado
}

type Tracker interface {
    // Track registra el costo atribuido a un modelo y proveedor específico.
    // Usado tras resolverse el provider exitoso (post-failover).
    Track(ctx context.Context, record CostRecord) error
}
```

## 3. Comportamientos y Reglas
1. **Resolución de Costos**: Requiere conocer el `cost_per_token` del modelo y proveedor (proveniente del `Registry`).
2. **Atribución Justa**: Si la petición hace fallback/failover, el costo se atribuye únicamente al proveedor que emite una respuesta HTTP 200 (HU-007 AC4).
3. **Costo Desconocido o Gratuito**: Si un modelo carece de tarifa configurada, se reporta costo desconocido (`0.0` y marcado como warning, HU-007 AC2). Si el modelo es gratuito (costo explícito `0.0`), se registra sin alterar (HU-007 AC3).
4. **Streaming (Aborto)**: Si una conexión TCP por streaming se aborta antes de terminar, la telemetría debe capturar solo los tokens emitidos (CompletionTokens) para no cobrar un response truncado a tarifa completa (HU-007 AC5).
5. **Persistencia asíncrona**: Emite los eventos que luego son almacenados por el Event Store/WAL, al igual que los eventos de auditoría.
