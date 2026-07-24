## ADDED Requirements

### Requirement: Gestión de API keys por usuario
El Gateway SHALL permitir a cada usuario generar, listar y revocar sus propias API keys
(admin puede operar sobre cualquier usuario), persistiendo solo el hash de la key. (Traza:
HU-EVO-018)

#### Scenario: Generar key
- **GIVEN** un usuario autenticado
- **WHEN** hace `POST /users/{id}/api-keys` con un nombre descriptivo
- **THEN** recibe la key en texto plano una sola vez en la respuesta; solo su hash se persiste

#### Scenario: Listar keys enmascaradas
- **GIVEN** un usuario con 3 keys generadas
- **WHEN** hace `GET /users/{id}/api-keys`
- **THEN** ve nombre, prefijo enmascarado, scopes, fecha de creación y último uso — nunca la key completa

#### Scenario: Revocar key
- **GIVEN** una key activa
- **WHEN** hace `DELETE /users/{id}/api-keys/{keyId}`
- **THEN** esa key deja de autenticar inmediatamente (siguiente request con ella: `401`)

#### Scenario: Revocar key de otro usuario
- **GIVEN** un usuario no-admin
- **WHEN** intenta revocar una key de otro usuario
- **THEN** recibe `403 Forbidden`

#### Scenario: Último uso se actualiza
- **GIVEN** una key nunca usada (`last_used_at` null)
- **WHEN** se autentica una request con ella
- **THEN** `last_used_at` se actualiza a la fecha/hora de esa request
