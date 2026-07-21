---
id: HU-004a
titulo: Failover básico de cadena con degradación a local
epica: EP-002
prioridad: Must
complejidad: M
estado: lista
---

# Failover básico de cadena con degradación a local

Como **agente consumidor**, quiero **que si el proveedor primario falla la Gateway pase automáticamente al siguiente de la cadena hasta terminar en un modelo local**, para **que mi petición se complete sin que yo maneje el error**.

Contexto: resiliencia del producto. Cadena de fallback ordenada por capacidad, terminando en modelos locales. El circuit breaker y los timeouts dinámicos se cubren en HU-004b y HU-004c.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy - Failover simple | Dado que el `providerA` devuelve 503 | Cuando el failover engine intercepta el error | Entonces enruta la petición al `providerB` de forma transparente para el cliente |
| 2 | Error - Pool agotado | Dado que todos los providers de la capacidad fallan | Cuando no hay más alternativas | Entonces retorna `HTTP 502 Bad Gateway` al cliente con el detalle del último error |
| 3 | Edge - Retry condicional | Dado que el `providerA` devuelve `429 Too Many Requests` | Cuando el engine evalúa el código | Entonces aplica failover en lugar de retry para proteger la cuota global del proveedor A |
| 4 | Edge - Degradación a modelo local | Dado que todos los providers remotos de la capacidad fallan y hay un modelo local configurado | Cuando se agota el pool remoto | Entonces el router envía la petición al modelo local (ej. Llama 3) sin exponer el fallo al cliente |
| 5 | Error — payload mal formado (400) | Dado que el cliente envía un prompt inválido (ej. JSON malformado) | Cuando llega la petición y el primer proveedor devuelve 400 | Entonces el Gateway NO intenta el failover y retorna 400 al cliente inmediatamente |

## Checklist INVEST

- [x] Independent — depende de HU-002a (routing) entregable
- [x] Negotiable — estrategia de reintento abierta
- [x] Valuable — cumple objetivo de resiliencia ≥99% (consumidor no ve el error)
- [x] Estimable — acotado a la mecánica de cadena de fallback
- [x] Small — un sprint; solo failover básico + degradación local
- [x] Testable — proveedores mockeados con códigos forzados

## Notas técnicas

Límite de intentos por cadena para evitar bucles; respetar timeouts por proveedor. El fallo mid-stream se reporta al cliente (no hay failover transparente mid-stream) y penaliza el score.
