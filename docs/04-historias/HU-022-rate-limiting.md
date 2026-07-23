---
id: HU-022
titulo: Rate Limiting y protección de Payload
epica: EP-004A
prioridad: Must
complejidad: M
estado: lista
---

# Rate Limiting y protección de Payload

Como **ingeniero de seguridad**, quiero **que la Gateway aplique límites de tasa de peticiones y tamaño de carga útil**, para **prevenir ataques de denegación de servicio (DDoS) o consumos maliciosos masivos**.

Contexto: Capa de defensa determinista in-memory antes de enrutar la petición.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — dentro del límite | Dado que un cliente con límite de 60 req/minuto | Cuando hace su petición número 10 en ese minuto | Entonces la petición se admite y continúa al ruteo |
| 2 | Error — rate limit excedido | Dado que un cliente excede las 60 req/minuto | Cuando hace su petición 61 | Entonces la Gateway bloquea devolviendo un `429 Too Many Requests` en menos de 5ms, sin llamar a LLMs |
| 3 | Error — payload vision excede 50MB | Dado que un cliente envía una imagen en base64 de 51MB (límite de contenido `vision` = 50MB, per PRD §2) | Cuando pasa la autenticación | Entonces la Gateway intercepta el tamaño excediendo 50MB, corta el stream y devuelve `413 Payload Too Large` |
| 4 | Edge — race conditions en límite | Dado que un cliente con 1 token restante de rate limit | Cuando envía 10 peticiones concurrentes | Entonces la validación atómica en RAM permite solo 1 y bloquea 9 |
| 5 | Error — payload de texto excede 10MB | Dado que un cliente envía un JSON de texto plano mayor a 10MB (límite de contenido de texto = 10MB, per PRD §2; distinto del límite de 50MB de `vision`) | Cuando el payload entra al gateway | Entonces el sistema rechaza inmediatamente con `413 Payload Too Large` |
| 6 | Sad path — Ataque Slowloris | Dado que un cliente malicioso abre una conexión TCP | Cuando el timer estricto de lectura (ReadHeaderTimeout de 5s) expira | Entonces el Gateway cierra forzosamente la conexión TCP liberando los recursos |

> El límite de concurrencia y enrutamiento de red específicos de la capacidad `vision` se separan en **HU-022b**.

## Checklist INVEST

- [x] Independent — componente middleware que corre antes del router
- [x] Negotiable — límites hardcodeados inicialmente, dinámicos después
- [x] Valuable — protege la billetera de la empresa y la estabilidad de los nodos al impedir que picos de tráfico maliciosos o accidentales tumben el sistema.
- [x] Estimable — algoritmos estándar (Token Bucket o Sliding Window)
- [x] Small — 1 sprint
- [x] Testable — pruebas de carga controladas en entorno local

## Notas técnicas

Debe funcionar 100% en `Local RAM Cache` (cero red) para mantener overhead mínimo.
**Importante**: Requiere configuración de Sticky Sessions (Hash de API Key) en el Load Balancer para que el límite en RAM sea totalmente exacto sin usar Redis.


> **OpenSpec change**: `ep-004a-identidad-accesos` (EP-004A)
