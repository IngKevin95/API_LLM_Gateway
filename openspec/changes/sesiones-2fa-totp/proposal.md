## Why

El Gateway necesita gestionar múltiples sesiones por usuario y ofrecer 2FA/TOTP como opción de seguridad, para revocar accesos comprometidos (sesiones) y reducir el riesgo ante el robo de contraseñas o API keys (2FA). Estas características elevan el nivel de seguridad del producto a estándares empresariales básicos.

## What Changes

- Tabla de sesiones para gestionar persistencia de login y revocación.
- Soporte para listar sesiones activas y cerrar sesiones (individuales o todas excepto la actual).
- Soporte para enrolamiento TOTP (generación de secreto y validación de código).
- Soporte para confirmación y desactivación de 2FA.
- Middleware/Autenticación debe soportar la validación de JWT basados en la nueva estructura de sesión y/o manejar el flujo de TOTP si aplica a futuro (en esta etapa solo habilitamos los endpoints de gestión de sesiones y TOTP de acuerdo a HU-EVO-019 y HU-EVO-020).

## Capabilities

### New Capabilities
- `user-sessions`: Gestión de sesiones activas, expiración, listado y revocación.
- `user-2fa-totp`: Enrolamiento, validación y gestión de TOTP por usuario.

### Modified Capabilities
<!-- None for this slice -->

## Impact

- **Base de datos:** Nuevas tablas `sessions` y nuevas columnas en `users` para `mfa_secret` y `mfa_enabled`.
- **APIs:** Nuevos endpoints en `/sessions` y `/auth/mfa/*`.
- **Dependencias:** Añadir una librería para TOTP (compatible con RFC 6238).

## Trazabilidad

- HU-EVO-019
- HU-EVO-020
