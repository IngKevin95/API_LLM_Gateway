---
id: HU-004b
titulo: Circuit Breaker pasivo y Max In-Flight
epica: EP-002
prioridad: Must
complejidad: M
estado: lista
---

# Circuit Breaker pasivo y Max In-Flight

Como **operador**, quiero **un circuit breaker que marque proveedores inalcanzables y limite peticiones en vuelo**, para **evitar el Failover Suicide que agota el pool de conexiones cuando un proveedor se degrada**.

Contexto: complementa el failover básico (HU-004a). El Max In-Flight es configurable por proveedor en YAML.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Circuit Breaker Pasivo | Dado que un proveedor principal devuelve error de red (429/500/timeout) o supera su Max In-Flight (configurable en YAML, ej. 300-500 para top tiers) | Cuando ocurre un failover hacia el secundario | Entonces el proveedor principal se marca como inalcanzable temporalmente (durante `cooldown_ms`) para evitar Failover Suicide de peticiones concurrentes |
| 2 | Edge — Max In-Flight excedido | Dado que las peticiones en curso superan el Max In-Flight configurado del proveedor | Cuando el cliente envía una nueva | Entonces el Circuit Breaker pasivo hace fast-fail (0 I/O) sin esperar el timeout, para prevenir Failover Suicide |
| 3 | Edge — reactivación tras backoff | Dado que un proveedor fue marcado inalcanzable | Cuando transcurre el periodo de gracia (backoff fijo, ej. 30s) y el health check lo reporta sano | Entonces el proveedor vuelve a la cadena de fallback |

## Checklist INVEST

- [x] Independent — depende de HU-004a (failover básico) entregable
- [x] Negotiable — algoritmo del breaker y valores por defecto abiertos
- [x] Valuable — evita el colapso del pool de conexiones bajo carga (estabilidad a 500 RPS)
- [x] Estimable — acotado al contador atómico + estado por proveedor
- [x] Small — un sprint
- [x] Testable — se fuerza el conteo de in-flight y los códigos de error

## Notas técnicas

Max In-Flight declarativo por proveedor en YAML (no hardcodeado). El fast-fail preserva el SLA de 0 I/O de la capa determinista.

> **OpenSpec change**: `ep-002-resiliencia-conectividad` (EP-002)
