---
id: HU-034
titulo: Protección TCP contra ataques Slowloris
epica: EP-004B
prioridad: Must
complejidad: S
estado: lista
---

# HU-034: Protección TCP contra ataques Slowloris

## INVEST
- [x] Independent: puede implementarse de forma independiente en la capa de red del framework.
- [x] Negotiable: los timeouts exactos (ReadHeaderTimeout, WriteTimeout) pueden configurarse en YAML.
- [x] Valuable: protege al servidor contra ataques de agotamiento de conexiones (DoS).
- [x] Estimable: usar timeouts estándar del servidor HTTP.
- [x] Small: requiere configuración en el servidor HTTP base.
- [x] Testable: simulable con clientes lentos (ej. slowhttptest).

## Criterios de Aceptación (BDD)
| ID | Escenario | Dado (Given) | Cuando (When) | Entonces (Then) |
|---|---|---|---|---|
| 1 | Ataque de cabeceras lentas | Un cliente abre una conexión y envía cabeceras a 1 byte por segundo | El tiempo excede el ReadHeaderTimeout | El servidor cierra el socket devolviendo un 408 Request Timeout |
