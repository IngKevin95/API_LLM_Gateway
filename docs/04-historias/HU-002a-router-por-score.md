---
id: HU-002a
titulo: Resolver capacidad a modelo por score (Router automático)
epica: EP-001
prioridad: Must
complejidad: S
estado: lista
---

# Resolver capacidad a modelo por score (Router automático)

Como **agente consumidor**, quiero **pedir una capacidad (coding, reasoning, vision, image, embedding, chat) sin nombrar el modelo y que la Gateway elija el óptimo por score**, para **obtener la mejor respuesta disponible sin acoplarme a un proveedor**.

Contexto: núcleo del desacople. El score combina calidad, velocidad, disponibilidad, cuota restante, costo y latencia.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy path — Elección óptima | Dado que hay 3 modelos disponibles para "chat" con distintos pesos (costo, latencia) y cuota > 0 | Cuando el cliente pide la capacidad "chat" | Entonces el Router retorna la lista ordenada de fallbacks colocando al modelo de mayor score total en primer lugar |
| 2 | Edge — Filtro por estado | Dado que el modelo de mayor score teórico está deshabilitado en configuración YAML o marcado unhealthy/sin cuota | Cuando el Router evalúa los scores | Entonces el Router excluye a ese modelo de la lista y el de segundo mejor score toma el primer lugar |
| 3 | Edge — Context Window Validation | Dado que la suma de tokens estimados del request supera el max_context_window del modelo (considerando un buffer de seguridad del 20%) | Cuando el Router evalúa los candidatos | Entonces descarta ese modelo antes de computar el score para evitar errores upstream |

## Checklist INVEST

- [x] Independent — depende solo de HU-001 (Registry) ya entregable
- [x] Negotiable — fórmula de score ajustable
- [x] Valuable — realiza el objetivo de desacople y selección óptima
- [x] Estimable — algoritmo de scoring acotado
- [x] Small — un sprint
- [x] Testable — scores fijos → resultado determinista

## Notas técnicas

Los pesos iniciales del score son fijos (config YAML); el ajuste dinámico con histórico es HU-019. La estimación de tokens usa una interfaz genérica `ITokenizer` o la validación recae sobre la capa de Adapter (ej. `tiktoken-go` para OpenAI, tokenizador nativo para Anthropic/Llama).

- Riesgo de tamaño: complejidad S (los casos de error y desempate se movieron a HU-002b).
