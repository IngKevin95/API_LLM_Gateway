## 1. Store de usuarios (HU-EVO-017)
- [x] 1.1 Tabla `users` + migración idempotente (`internal/user/user.go`)
- [x] 1.2 `Store.Create` con `ErrEmailExists` (AC5) y default `status=invited` (AC1)
- [x] 1.3 `Store.List` con filtro por tenant / admin global (AC2)
- [x] 1.4 `Store.Patch` (rol/estado) (AC3)
- [x] 1.5 Handler HTTP `POST/GET /users`, `PATCH /users/{id}` con 403 no-admin (AC4)

## 2. API keys por usuario (HU-EVO-018)
- [x] 2.1 Tabla `api_keys` + migración idempotente (`internal/user/apikeys.go`)
- [x] 2.2 `KeyStore.Generate` (hash sha256, key en claro solo en la respuesta) (AC1)
- [x] 2.3 `KeyStore.List` enmascarado (AC2)
- [x] 2.4 `KeyStore.Revoke` con distinción 404/403 por ownership (AC3/AC4)
- [x] 2.5 `KeyStore.Authenticate` (constant-time, corta acceso si usuario suspendido, actualiza
      `last_used_at`) (AC5, y HU-EVO-017 AC3 end-to-end)
- [x] 2.6 Handler HTTP `POST/GET /users/{id}/api-keys`, `DELETE .../{keyId}`

## 3. Wiring en cmd/gateway
- [x] 3.1 `registerUsersRoutes` (503 fail-soft sin DSN)
- [x] 3.2 `identityMiddleware`: PostgreSQL primero, `apikey.Store` legacy como fallback, cableado en
      `/metrics` y `/alerts`

## 4. Evidencia
- [x] 4.1 Tests de integración contra PostgreSQL real (docker), mismo patrón que `alert.Manager`:
      `internal/user/user_integration_test.go`, `internal/handler/users_integration_test.go`
- [x] 4.2 `go build/vet/test -race` 100% verde, sin regresión
- [x] 4.3 Journey smoke real: binario compilado, boot con `GATEWAY_USERS_POSTGRES_DSN` real, curl
      end-to-end de los 8 endpoints (create/list/patch/generate/list-masked/revoke/metrics-auth/401)
