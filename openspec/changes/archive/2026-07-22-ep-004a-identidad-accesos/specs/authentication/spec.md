## ADDED Requirements

### Requirement: Autenticación por API key
El sistema SHALL exigir una API key válida en el header de cada petición, comparándola en tiempo constante contra un hash almacenado, sin aceptar la key por canales inseguros ni loguearla. (Traza: HU-008)

#### Scenario: Key válida
- **WHEN** llega una petición con una API key de cliente activa en el header
- **THEN** la Gateway la acepta y procesa, y el consumidor queda identificado para auditoría

#### Scenario: Sin key
- **WHEN** llega una petición sin header de autenticación
- **THEN** responde 401 sin procesar ni exponer detalles internos

#### Scenario: Key inválida o revocada
- **WHEN** llega una petición con una API key inexistente o revocada
- **THEN** responde 401 y registra el intento fallido sin loguear la key completa

#### Scenario: Key por canal inseguro
- **WHEN** la petición pone la key en la query string en vez del header
- **THEN** la Gateway no la acepta y responde 401, evitando exponer secretos en URLs/logs

### Requirement: Autenticación OAuth2/OIDC
El sistema SHALL validar firma, `aud`, `iss` y expiración de un JWT de IdP corporativo, rechazando tokens inválidos o con claims incorrectos. (Traza: HU-025a)

#### Scenario: JWT válido
- **WHEN** un cliente envía un JWT válido de un IdP corporativo
- **THEN** la Gateway valida firma, aud/iss y autoriza el acceso

#### Scenario: Token expirado
- **WHEN** el JWT está expirado
- **THEN** rechaza con 401 Unauthorized

#### Scenario: Firma inválida
- **WHEN** el JWT está adulterado (firma inválida)
- **THEN** rechaza con 401 y loguea el intento de intrusión

#### Scenario: Claims incorrectos
- **WHEN** el JWT es criptográficamente válido pero sin el claim `aud` correcto
- **THEN** rechaza con 403 Forbidden

### Requirement: Autenticación mTLS
El sistema SHALL autenticar por certificado cliente contra un trust store, extrayendo el scope del certificado y rechazando revocados, ausentes o de CA no confiable en capa TLS. (Traza: HU-025b)

#### Scenario: Certificado cliente válido
- **WHEN** un servicio interno con certificado cliente válido llama a la Gateway
- **THEN** la conexión se establece y se extrae el scope del certificado

#### Scenario: Certificado revocado o expirado
- **WHEN** el certificado cliente está revocado o expirado
- **THEN** el handshake falla y la petición se aborta

#### Scenario: Sin certificado
- **WHEN** un servicio sin certificado cliente llama al puerto seguro
- **THEN** la Gateway rechaza la conexión en capa TCP/TLS

#### Scenario: CA no confiable
- **WHEN** el certificado cliente es válido pero de una CA fuera del trust store
- **THEN** el handshake mTLS falla con "unknown certificate authority" y la petición se aborta
