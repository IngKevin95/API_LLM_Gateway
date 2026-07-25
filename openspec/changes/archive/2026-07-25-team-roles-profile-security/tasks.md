## 1. Backend

- [x] 1.1 `GET /users/me` (`UsersHandler.Me`) resolviendo perfil desde `auth.Identity`
- [x] 1.2 Cablear `GET /users/me` con `identityMiddleware` en `registerUsersRoutes`
- [x] 1.3 Fix: `registerAuthRoutes` (`/sessions`, `/auth/mfa/*`) recibe `identityMiddleware` real
      en vez de un wrap no-op — hueco de wiring detectado en esta sesión (rutas devolvían 401
      siempre, sin importar el token)
- [x] 1.4 JWT de `POST /auth/login` incluye claim `role`
- [x] 1.5 Tests de integración contra PostgreSQL real: `TestUsersHTTP_Me_ReturnsOwnProfile`,
      `TestUsersHTTP_Me_NoIdentity_Returns401`
- [x] 1.6 `go build ./...`, `go vet ./...`, `go test ./...` (unit + integration) verdes

## 2. Frontend

- [x] 2.1 `hooks/useCurrentUser.js` — `GET /users/me`
- [x] 2.2 `TeamRoles.jsx` — tabla + invitar + cambiar rol/suspender (HU-EVO-021)
- [x] 2.3 `ProfileSecurity.jsx` — perfil + API keys + sesiones + 2FA (HU-EVO-022)
- [x] 2.4 `Dashboard.jsx`: tabs "Team" (admin-only) y "Profile & Security" (siempre)
- [x] 2.5 Tests de componente (`TeamRoles.test.jsx`, `ProfileSecurity.test.jsx`) cubriendo los 10
      escenarios G/W/T de HU-EVO-021/022
- [x] 2.6 `npm run test:ui` verde (28/28)

## 3. Verificación end-to-end

- [x] 3.1 Smoke manual contra binario real + PostgreSQL real (docker): login → `GET /users/me`,
      `GET /sessions`, `POST /auth/mfa/enroll` con el JWT emitido — confirma el fix de wiring
- [x] 3.2 `ux-fidelity-reviewer` contra las pantallas Stitch "Team & Roles" y "Profile & Security"
      (MCP chrome-devtools) — FIEL (con desviacion de shell ya aceptada en EP-EVO-003)
- [x] 3.3 `wiring-adversarial-verifier` en contexto en blanco — corrido, 3 huecos reales + 2 menores encontrados y corregidos, reverificado en verde
- [x] 3.4 DoD (`dor-dod-gatekeeper`) — aprobado
- [x] 3.5 PR + archive
