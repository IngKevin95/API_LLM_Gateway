## Context

Actualmente el sistema permite la creación de usuarios y la autenticación mediante API keys. Sin embargo, no existe un concepto de "sesión" para usuarios que inician sesión a través de la UI u otro medio interactivo; los tokens de sesión no están atados a un registro persistente que pueda ser revocado de forma independiente. Además, no se soporta Autenticación de Dos Factores (2FA/TOTP). Estas características son críticas para el producto final según las épicas planificadas.

## Goals / Non-Goals

**Goals:**
- Crear la tabla `sessions` en PostgreSQL para rastrear todas las sesiones activas por usuario, con metadatos útiles (IP, user agent, última actividad).
- Proveer endpoints CRUD-like para listar y eliminar sesiones (individual o masivo con `except_current`).
- Añadir la configuración 2FA a la tabla `users` (columnas `mfa_secret` y `mfa_enabled`).
- Proveer endpoints para enrolamiento y verificación de TOTP (`/auth/mfa/enroll`, `/auth/mfa/verify`).
- Usar una librería estándar (RFC 6238) en Go para la generación y validación de TOTP, como `github.com/pquerna/otp/totp`.

**Non-Goals:**
- No se implementará geolocalización avanzada estricta o pagada; una resolución IP a ciudad/país en texto es suficiente, o se omitirá si requiere licencias costosas en esta etapa.
- En esta etapa no se modifica la UI, es un backend-only slice.
- No se forzará 2FA a nivel global, será enteramente opcional por usuario.

## Decisions

- **Persistencia del secreto MFA:** Se almacenará en la tabla `users` de forma segura. Si bien el secreto se suele almacenar en texto plano en la BD en muchos sistemas (ya que su protección asume la protección de la BD), podemos añadir encriptación a nivel de aplicación (AES) si se requiere, pero por ahora se inserta como `bytea` o `varchar` en un estado seguro.
- **Identificador de sesión:** El token JWT que se emita deberá incluir el `session_id` en sus claims (`sid`) para que, al interceptar la request en el middleware, se pueda verificar si la sesión sigue existiendo en DB o fue revocada.
- **Cache vs BD:** Por simplicidad, se consultará la BD directamente en el middleware. Para mitigar impacto en rendimiento en el futuro, se puede usar Redis, pero la iteración actual usará la tabla `sessions` directamente.

## Risks / Trade-offs

- **Risk:** Carga en BD por cada request si el middleware valida el `sid` en la tabla `sessions`.
  **Mitigation:** Añadir un índice en `id` de `sessions` y planear una capa de cache (ej. Redis) para el futuro, o aceptar el trade-off en un producto de bajo QPS en sus inicios.
- **Risk:** Librería TOTP obsoleta o vulnerable.
  **Mitigation:** Usar `github.com/pquerna/otp` (ampliamente utilizada y auditada en la comunidad Go).
