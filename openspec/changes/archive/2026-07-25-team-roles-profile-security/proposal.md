## Why

Los endpoints de administración de usuarios (HU-EVO-017), API keys (HU-EVO-018), sesiones
(HU-EVO-019) y 2FA (HU-EVO-020) ya existen en el backend, pero no tienen interfaz: un admin no
puede gestionar el equipo sin llamar a la API a mano, y ningún usuario puede rotar sus propias
credenciales o activar 2FA desde el dashboard. Esto cierra el círculo de EP-EVO-004 agregando las
dos pantallas de UI que faltaban.

## What Changes

- Nueva tab "Team" (admin-only) en el dashboard: tabla de usuarios con rol/estado/scopes, invitar
  miembro, cambiar rol o suspender.
- Nueva tab "Profile & Security" (cualquier usuario autenticado): perfil propio, gestión de API
  keys (generar/revocar), sesiones activas (cerrar), y activación de 2FA/TOTP.
- Fix de wiring descubierto durante esta implementación: `/sessions` y `/auth/mfa/*` estaban
  registrados en el mux sin pasar por `identityMiddleware`, por lo que `auth.FromContext` nunca
  resolvía identidad y esas rutas devolvían 401 siempre, sin importar el token. Se reordena el
  cableado en `main()` para que `registerAuthRoutes` reciba `identityMiddleware` ya construido.
- Nuevo endpoint `GET /users/me` (no existía): resuelve el perfil del usuario autenticado desde
  `auth.Identity`, necesario para que el frontend muestre "mi" perfil y decida si la tab "Team" es
  visible (HU-EVO-021 AC4) sin depender de que el JWT lleve el rol embebido y sin duplicar lógica
  de RBAC en el cliente.
- JWT de `POST /auth/login` ahora incluye el claim `role`.

## Capabilities

### New Capabilities
<!-- Ninguna: se extiende dashboard-ui con 2 requirements nuevos y user-store con GET /users/me -->

### Modified Capabilities
- `dashboard-ui`: se agregan los requirements "Tab Team con gestión RBAC" y "Tab Profile & Security
  con autoservicio de credenciales".
- `user-store`: se agrega el requirement "GET /users/me expone el perfil propio".

## Impact

- **Frontend:** `src/ui/dashboard/TeamRoles.jsx`, `src/ui/dashboard/ProfileSecurity.jsx`,
  `src/ui/dashboard/hooks/useCurrentUser.js`, `Dashboard.jsx` extendido con 2 tabs condicionales.
- **Backend:** `src/internal/handler/users.go` (`UsersHandler.Me`), `src/cmd/gateway/main.go`
  (reordenamiento de wiring de `registerAuthRoutes`/`registerUsersRoutes` + fix de identidad en
  `/sessions` y `/auth/mfa/*`), `src/internal/handler/auth_handler.go` (claim `role` en JWT).
- **APIs:** nuevo `GET /users/me`; sin cambios de contrato en los endpoints existentes.
- **Dependencias:** ninguna nueva (reutiliza `github.com/golang-jwt/jwt/v5` ya presente).

## Trazabilidad

- Epica: EP-EVO-004
- Historias: HU-EVO-021, HU-EVO-022
- Sub-slice: EP-EVO-004-SS3
