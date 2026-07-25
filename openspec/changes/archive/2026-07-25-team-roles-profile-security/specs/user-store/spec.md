## ADDED Requirements

### Requirement: GET /users/me expone el perfil propio
El Gateway SHALL exponer `GET /users/me`, resolviendo el usuario autenticado desde `auth.Identity`
(no desde `AdminContext`), de forma que cualquier usuario autenticado -- admin u operator -- vea su
propio perfil sin necesitar permisos de administrador. (Traza: HU-EVO-022)

#### Scenario: Usuario autenticado ve su propio perfil
- **GIVEN** un usuario autenticado con un JWT o API key válida
- **WHEN** hace `GET /users/me`
- **THEN** recibe `200` con su propio registro (`id`, `email`, `role`, `status`, `tenant`, `scopes`)

#### Scenario: Sin identidad resuelta
- **GIVEN** una petición sin token válido
- **WHEN** hace `GET /users/me`
- **THEN** recibe `401 Unauthorized`, sin exponer ningún dato de otro usuario
