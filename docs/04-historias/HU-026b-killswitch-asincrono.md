---
id: HU-026b
titulo: Kill-Switch Asíncrono de PII
epica: EP-004B
prioridad: Should
complejidad: M
estado: lista
---

# Kill-Switch Asíncrono de PII

Como **oficial de seguridad**, quiero **un análisis de PII profundo en paralelo**, para **abortar flujos que estén filtrando información personal sin penalizar latencia**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — kill-switch TCP | Dado que un payload largo que inicia streaming al proveedor | Cuando el worker detecta PII a los 200ms | Entonces aborta abruptamente el stream TCP y devuelve error |
| 2 | Error — Timeout | Dado que el payload es limpio pero el escáner se demora | Cuando el stream finaliza exitosamente antes | Entonces el escáner aborta su trabajo y se desecha |
| 3 | Edge — Falso positivo (termina rápido) | Dado que un payload contiene un patrón ambiguo susceptible de clasificarse como PII | Cuando el motor NLP del escáner asíncrono lo evalúa y lo descarta como falso positivo | Entonces permite que el stream principal continúe sin interrupción |
| 4 | Sad path — Stream finaliza antes | Dado que el LLM genera la respuesta más rápido de lo que el escáner asíncrono puede procesar y se detecta PII al final | Cuando la conexión ya se cerró | Entonces el Gateway registra un incidente grave post-mortem para auditoría manual (fuga real) |

## Checklist INVEST

- [x] Independent — Worker en background que lee respuestas sin bloquear el stream.
- [x] Negotiable — Puede usar un modelo NLP ligero local vs regex complejos.
- [x] Valuable — Protege la BD de auditoría de almacenar PII que el LLM pudiera generar.
- [x] Estimable — Desacoplado del path crítico, arquitectura worker estándar.
- [x] Small — Reutiliza el motor DLP de 026a, aplicándolo a la salida.
- [x] Testable — Tests de concurrencia y validación de inserciones en base de datos.

## Notas técnicas
- Asegurar alineación con NFRs de latencia y uso de caché si aplica.
