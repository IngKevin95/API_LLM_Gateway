---
id: HU-EVO-020
titulo: 2FA/TOTP opcional por usuario
epica: EP-EVO-004
prioridad: Should
complejidad: M
estado: lista
---

# 2FA/TOTP opcional por usuario

Como **usuario del Gateway**, quiero **activar verificación en dos pasos (TOTP) en mi cuenta**, para **reducir el riesgo de que alguien acceda con solo mi contraseña o API key robada**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — enrolar 2FA | Dado que no tengo 2FA activo | Cuando hago `POST /auth/mfa/enroll` | Entonces recibo un secreto TOTP + QR compatible con Google Authenticator/Authy, sin activarlo todavía (queda pendiente de confirmación) |
| 2 | Happy — confirmar enrolamiento | Dado que generé un secreto TOTP pendiente | Cuando hago `POST /auth/mfa/verify` con el código de 6 dígitos correcto de mi app authenticator | Entonces 2FA queda activo en mi cuenta |
| 3 | Happy — desactivar 2FA | Dado que tengo 2FA activo | Cuando lo desactivo confirmando mi password actual | Entonces 2FA queda deshabilitado |
| 4 | Error — código incorrecto en verify | Dado que tengo un secreto TOTP pendiente | Cuando envío un código de 6 dígitos incorrecto | Entonces recibo 400, el 2FA sigue sin activarse |
| 5 | Edge — código con ventana de tiempo expirada | Dado que tengo 2FA activo | Cuando envío un código TOTP válido pero fuera de la ventana de tolerancia (±30s configurado) | Entonces se rechaza como inválido, sin revelar si el código era "casi correcto" |

## Checklist INVEST

- [x] Independent — módulo propio sobre `users` (HU-EVO-017), no bloquea el resto de la épica si se difiere
- [x] Negotiable — librería TOTP específica es detalle de implementación (estándar RFC 6238)
- [x] Valuable — expectativa básica de seguridad en un producto empresarial con acceso a API keys de LLM
- [x] Estimable — columna `mfa_secret`/`mfa_enabled` en `users` + 2 endpoints + validación TOTP
- [x] Small/Medium — 2 días
- [x] Testable — test genera secreto, calcula código TOTP válido con la misma librería, verifica aceptación; prueba código inválido/expirado

## Notas técnicas

Usar una librería TOTP estándar (RFC 6238) ya validada, no implementar el algoritmo desde cero. El secreto se persiste cifrado (no en texto plano) en la misma base PostgreSQL usada por `users`.

---

## Relación con existentes

- Depende de: HU-EVO-017 (tabla `users`)
- Requisito para: HU-EVO-022 (UI Profile & Security, sección Security/2FA)
