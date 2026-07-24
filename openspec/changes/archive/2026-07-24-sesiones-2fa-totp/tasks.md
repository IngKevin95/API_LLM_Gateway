## 1. Base de Datos
- [x] 1.1 Crear script de migración para tabla `sessions`.
- [x] 1.2 Actualizar tabla `users` con columnas `mfa_secret` (text/bytea) y `mfa_enabled` (boolean).
- [x] 1.3 Crear repositorio/interfaz para gestionar la tabla de sesiones (crear, obtener todas, eliminar una, eliminar excepto actual).
- [x] 1.4 Actualizar el modelo de `users` en el backend para reflejar los nuevos campos (mfa).

## 2. API Sesiones
- [x] 2.1 Implementar endpoints `GET /sessions` y `DELETE /sessions` (con soporte param `except_current`).
- [x] 2.2 Implementar endpoint `DELETE /sessions/:id`.
- [x] 2.3 Escribir tests de integración para las operaciones de sesión.

## 3. Autenticación y TOTP
- [x] 3.1 Integrar la librería OTP (e.g. `github.com/pquerna/otp/totp`).
- [x] 3.2 Implementar endpoint `POST /auth/mfa/enroll` (generar QR/URI).
- [x] 3.3 Implementar endpoint `POST /auth/mfa/verify` (validar código y activar mfa).
- [x] 3.4 Implementar endpoint `POST /auth/mfa/disable` (desactivar mfa).
- [x] 3.5 Escribir tests de integración para TOTP.

## 4. Middleware y JWT
- [x] 4.1 Modificar el `identityMiddleware` o lógica de JWT para añadir `sid` al token de sesión.
- [x] 4.2 Configurar el middleware para verificar que la sesión siga activa en DB si existe un `sid` en el token.
