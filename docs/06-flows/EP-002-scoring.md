# EP-002: Flujo de Scoring Dinámico

Algoritmo de puntuación para selección de modelo en tiempo de ejecución.

```mermaid
graph TD
    A["Health Monitor<br/>(HU-009)<br/>recolecta métricas"] -->|availability<br/>latency<br/>throughput| B["Learning Engine<br/>(HU-011)<br/>histórico de decisiones"]
    C["Model Registry<br/>atributos fijos<br/>quality, coding, speed, cost"] -->|score base| D["Scoring Compute<br/>(HU-008)<br/>formula dinámico"]
    A -->|datos en vivo| D
    B -->|feedback histórico| D
    D -->|score = f(quality, speed,<br/>availability, cost,<br/>latency, tokens_remaining)| E["Ranking<br/>modelos por score"]
    E -->|top 1| F["Circuit Breaker<br/>(HU-012)<br/>in-flight ≤ max"]
    F -->|OK| G["✅ Modelo seleccionado<br/>con score + razón"]
    F -->|OPEN| H["❌ Fallback<br/>siguiente modelo"]
    H -->|retry scoring| E
```

## Historias Críticas

| Historia | Fase | Componente |
|----------|------|-----------|
| HU-008 | 1 | Scoring Engine (fórmula dinámica) |
| HU-009 | 2 | Health Monitor (métricas periódicas) |
| HU-011 | 3 | Learning Engine (histórico + ajustes) |
| HU-012 | 1 | Circuit Breaker (in-flight limiter) |
| HU-013 | 1 | Quota Manager (tokens/cuota restante) |

## Fórmula de Score (MVP Fase 1)

```
score = (quality × 0.3) + (speed × 0.2) + (availability × 0.3) + (cost_efficiency × 0.2)

Donde:
- quality: [0, 1] (modelo benchmark vs. baseline)
- speed: [0, 1] (TTFT inverso normalizado)
- availability: [0, 1] (health check last 24h)
- cost_efficiency: [1 - (cost / max_cost), 0] (token price normalized)
```

## Fase 2+ Mejoras
- Learning Engine (HU-011): ajustar pesos por tipo de prompt (code > reasoning > vision)
- Análisis de errores pasados: reducir score de modelos que fallaron
- Predicción de cuota agotada

## Edge Cases
- **Modelo 1 caído**: Score = 0 (out of rotation)
- **Múltiples modelos con score idéntico**: Tie-breaker = latencia última conocida
- **Todos los modelos caídos**: Error 503 (no fallback disponible)

## SLA Asociado
- **Scoring latency**: < 10ms p99 (O(n) ranking)
- **Learning loop**: < 1s feedback (async, no blocking)
