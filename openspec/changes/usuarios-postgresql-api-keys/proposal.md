## Why

El Gateway hoy resuelve identidad no-admin exclusivamente vía `GATEWAY_API_KEYS` (env var estática,
`apikey.Store` in-memory): dar de alta o revocar un consumidor exige editar la variable y
redeployar. No hay tabla de usuarios ni CRUD administrativo, y las API keys no tienen dueño,
nombre, ni fecha de último uso — no se puede auditar ni rotar credenciales sin fricción operativa.

## What Changes

- Store de usuarios persistente en PostgreSQL (`internal/user`): tabla `users` (email, role
  admin/operator, status invited/active/suspended, tenant, scopes), CRUD admin-only vía
  `POST/GET /users` y `PATCH /users/{id}`, aislamiento multi-tenant en `GET /users` para admins
  no-globales, `409` ante email duplicado, `403` ante escritura no-admin.
- Gestión de API keys por usuario (`internal/user/apikeys.go`): tabla `api_keys` (hash sha256,
  prefijo enmascarado, scopes, timestamps), `POST/GET /users/{id}/api-keys` y
  `DELETE /users/{id}/api-keys/{keyId}`, la key en claro se devuelve una única vez al generarla,
  revocación inmediata, `last_used_at` actualizado en cada autenticación, ownership (dueño o admin)
  exigido para revocar.
- `cmd/gateway/main.go`: nuevo `identityMiddleware` que resuelve identidad primero contra el store
  de PostgreSQL (`userKeys.Authenticate`) y cae a `apikey.Store`/`GATEWAY_API_KEYS` legacy si no
  matchea, cableado en `/metrics` y `/alerts` — reemplazo gradual sin romper despliegues que no
  migraron. Usuario `suspended` pierde acceso inmediato aunque su key siga sin revocar.

## Capabilities

### New Capabilities
- `user-store`: identidad persistente de usuarios (rol, estado, tenant) con CRUD admin-only sobre
  PostgreSQL.
- `api-keys`: ciclo de vida de credenciales por usuario (generar/listar/revocar) sobre PostgreSQL,
  reemplazando gradualmente el seed estático `GATEWAY_API_KEYS`.

### Modified Capabilities
(ninguna capability existente se modifica en su contrato; `authentication`/`authorization`
ganan una fuente adicional de identidad sin cambiar su interfaz pública)

## Impacto

- Nuevas tablas PostgreSQL (`users`, `api_keys`), migraciones idempotentes `CREATE TABLE IF NOT
  EXISTS` aplicadas en boot, mismo patrón que `provider_alerts` (HU-EVO-012) y `learned_quota`
  (HU-EVO-008).
- Opt-in vía `GATEWAY_USERS_POSTGRES_DSN` (fallback: `GATEWAY_QUOTA_POSTGRES_DSN`); sin DSN
  configurada, `/users` y `/users/{id}/api-keys` responden `503` fail-soft, sin tumbar el boot.
- Sin breaking changes en `/metrics` ni `/alerts`: siguen aceptando el token admin estático y las
  keys legacy de `GATEWAY_API_KEYS`.

## Trazabilidad

Epica: EP-EVO-004
Historias: HU-EVO-017, HU-EVO-018
Sub-slice: EP-EVO-004-SS1
