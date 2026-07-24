## 1. Persistencia y Modelos (DB)

- [ ] 1.1 Crear script de migración para tabla `sessions`.
- [ ] 1.2 Actualizar tabla `users` con columnas `mfa_secret` (text/bytea) y `mfa_enabled` (boolean).
- [ ] 1.3 Crear repositorio/interfaz para gestionar la tabla de sesiones (crear, obtener todas, eliminar una, eliminar excepto actual).
- [ ] 1.4 Actualizar el modelo de `users` en el backend para reflejar los nuevos campos (mfa).

## 2. API Sesiones

- [ ] 2.1 Implementar endpoints `GET /sessions` y `DELETE /sessions` (con soporte param `except_current`).
- [ ] 2.2 Implementar endpoint `DELETE /sessions/:id`.
- [ ] 2.3 Escribir tests de integración para las operaciones de sesión.

## 3. Autenticación y TOTP

- [ ] 3.1 Integrar la librería OTP (e.g. `github.com/pquerna/otp/totp`).
- [ ] 3.2 Implementar endpoint `POST /auth/mfa/enroll` (generar QR/URI).
- [ ] 3.3 Implementar endpoint `POST /auth/mfa/verify` (validar código y activar mfa).
- [ ] 3.4 Implementar endpoint `POST /auth/mfa/disable` (desactivar mfa).
- [ ] 3.5 Escribir tests de integración para TOTP.

## 4. Integración

- [ ] 4.1 Modificar el `identityMiddleware` o lógica de JWT para añadir `sid` al token de sesión.
- [ ] 4.2 Configurar el middleware para verificar que la sesión siga activa en DB si existe un `sid` en el token.
