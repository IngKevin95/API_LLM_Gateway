# rate-limiting Specification

## Purpose
TBD - created by archiving change ep-004a-identidad-accesos. Update Purpose after archive.
## Requirements
### Requirement: Rate limiting por ventana con validación atómica
El sistema SHALL limitar las peticiones por credencial en una ventana temporal con validación atómica en RAM, respondiendo 429 en <5ms sin invocar LLMs, y correcto bajo concurrencia. (Traza: HU-022 AC1/AC2/AC4)

#### Scenario: Dentro del límite
- **WHEN** un cliente con límite de 60 req/minuto hace su petición número 10 en ese minuto
- **THEN** la petición se admite y continúa al ruteo

#### Scenario: Rate limit excedido
- **WHEN** un cliente excede las 60 req/minuto y hace la petición 61
- **THEN** la Gateway responde 429 Too Many Requests en <5ms, sin llamar a LLMs

#### Scenario: Concurrencia en el límite
- **WHEN** un cliente con 1 token restante envía 10 peticiones concurrentes
- **THEN** la validación atómica en RAM admite solo 1 y bloquea 9

### Requirement: Límites de payload por tipo de contenido
El sistema SHALL rechazar con 413 los payloads que exceden el límite de su tipo (10MB texto, 50MB vision/base64 por PRD §2), cortando el stream sin cargar el cuerpo completo. (Traza: HU-022 AC3/AC5)

#### Scenario: Payload de texto excede 10MB
- **WHEN** un cliente envía un JSON de texto plano mayor a 10MB
- **THEN** el sistema rechaza inmediatamente con 413 Payload Too Large

#### Scenario: Payload vision excede 50MB
- **WHEN** un cliente envía una imagen base64 de 51MB (sobre el límite de 50MB)
- **THEN** la Gateway corta el stream y devuelve 413 Payload Too Large

### Requirement: Protección Slowloris
El sistema SHALL cerrar conexiones cuyo envío de headers exceda un timeout estricto de lectura, liberando recursos. (Traza: HU-022 AC6)

#### Scenario: Conexión lenta
- **WHEN** un cliente malicioso abre una conexión TCP y el ReadHeaderTimeout (5s) expira
- **THEN** el Gateway cierra forzosamente la conexión TCP liberando los recursos

### Requirement: Límite de concurrencia de vision por nodo
El sistema SHALL limitar a 2 las peticiones `vision` concurrentes por nodo, respondiendo 429 sin encolar la tercera, y enrutar vision por menor carga en vez de Hash L7. (Traza: HU-022b)

#### Scenario: Dentro del límite vision
- **WHEN** un nodo con 1 petición vision activa (límite 2) recibe una segunda vision
- **THEN** devuelve 200 sin timeout

#### Scenario: Concurrencia vision excedida
- **WHEN** un nodo ya atiende 2 peticiones vision y llega una tercera
- **THEN** responde 429 Too Many Requests sin encolarla, evitando bufferbloat

#### Scenario: Enrutamiento dedicado de vision
- **WHEN** llega una petición `vision` al balanceador
- **THEN** se enruta por política de menor carga (no por Hash L7 de API key), aceptando un leve drift de cuota en RAM local a favor de la estabilidad del cluster

