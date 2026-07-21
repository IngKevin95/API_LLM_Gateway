---
id: HU-002b
titulo: Manejo de errores y desempates en el enrutamiento
epica: EP-001
prioridad: Must
complejidad: S
estado: lista
---

# Manejo de errores y desempates en el enrutamiento

Como **desarrollador de la plataforma**, quiero **un manejo de errores robusto y desempates deterministas en el Router**, para **evitar caídas cuando no hay modelos aptos o existen puntajes idénticos**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Sad path — Capacidad desconocida | Dado que un cliente solicita una capacidad no definida (ej. "quantum_computing") | Cuando el Router intenta resolverla | Entonces retorna HTTP 400 Bad Request indicando capacidad no soportada |
| 2 | Sad path — Ningún modelo cumple | Dado que un cliente pide "vision" pero todos los modelos de "vision" tienen cuota agotada o están caídos (health=false) | Cuando el Router filtra la lista | Entonces retorna HTTP 503 Service Unavailable y no inicia failover inútil |
| 3 | Edge — Empate de score | Dado que dos modelos aptos obtienen exactamente el mismo score tras la evaluación | Cuando el Router ordena la lista de fallbacks | Entonces desempata consistentemente priorizando el de menor costo, o en su defecto por orden alfabético del ID |

## Checklist INVEST
- [x] Independent — no colisiona con el desarrollo de otros endpoints y módulos
- [x] Negotiable — los detalles finos de la implementación son adaptables por el equipo
- [x] Valuable — garantiza que los clientes nunca se queden colgados esperando una respuesta infinita y reciban un feedback claro del fallo, mejorando enormemente la DX (Developer Experience).
- [x] Estimable — el esfuerzo está bien delimitado por los criterios BDD
- [x] Small (separada de HU-002a)
- [x] Testable — sus flujos edge y happy path son verificables por pruebas unitarias

## Notas técnicas
- El fallback debe respetar estrictamente el timeout dinámico.
