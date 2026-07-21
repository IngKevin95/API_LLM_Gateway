---
id: HU-019
titulo: Ajustar pesos del score con datos históricos (Learning Engine)
epica: EP-007
prioridad: Could
complejidad: M
estado: lista
---

# Ajustar pesos del score con datos históricos (Learning Engine)

Como **operador de la plataforma**, quiero **que la Gateway ajuste los pesos del score de selección usando el histórico de resultados**, para **que el enrutamiento mejore solo con el tiempo sin ajuste manual**.

Contexto: autoaprendizaje del enrutamiento, cierre de la visión. Actividad 5, fase final de construcción.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — ajuste con evidencia | Dado que el histórico tiene >100 peticiones en `coding` y el modelo X tiene una latencia p95 10% menor que el resto | Cuando corre el ciclo de aprendizaje | Entonces los pesos se actualizan (delta > 0.1) y el Router empieza a preferir el modelo X para `coding` |
| 2 | Error — datos insuficientes | Dado que un histórico con menos de N registros (umbral configurable) | Cuando corre el ciclo de aprendizaje | Entonces no ajusta los pesos (delta = 0) y conserva los pesos por defecto, evitando sobreajuste |
| 3 | Edge — límites de seguridad | Dado que el aprendizaje intenta empujar todo el tráfico a un único modelo | Cuando se aplica el ajuste | Entonces se respetan topes/guardrails de diversificación para no crear un punto único de fallo |
| 4 | Edge — reversibilidad | Dado que un ajuste que degrada el success_rate observado | Cuando se detecta la degradación | Entonces el sistema puede revertir al conjunto de pesos anterior (rollback) de forma auditable |

## Checklist INVEST

- [x] Independent — depende de HU-018 (histórico) entregable
- [x] Negotiable — algoritmo de aprendizaje abierto
- [x] Valuable — optimización automática continua
- [x] Estimable — acotable a un ajuste simple inicial
- [x] Small — un sprint para la versión inicial (heurística), no el óptimo
- [x] Testable — histórico simulado → ajuste esperado

## Notas técnicas

Empezar con heurística simple y explicable; guardar versiones de pesos para rollback; guardrails de diversificación.
