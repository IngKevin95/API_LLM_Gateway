## Context

EP-EVO-004-SS1 agrega la primera capa de identidad persistente del Gateway (hasta ahora, solo
tokens estáticos y env vars). Mismo patrón de infraestructura que `alert.Manager` (HU-EVO-012) y
`quota.PostgresPersister` (HU-EVO-008): migración idempotente en boot, opt-in vía DSN, fail-soft
sin DB.

## Goals / Non-Goals

**Goals**: CRUD admin de usuarios; API keys por usuario con hash+prefijo enmascarado; reemplazo
gradual de `GATEWAY_API_KEYS` sin breaking change.

**Non-Goals** (diferido a SS2/SS3 de EP-EVO-004): login por password, sesiones, 2FA/TOTP, UI de
administración (Team & Roles, Profile & Security) — estos viven en HU-EVO-019/020/021/022.

## Decisiones

1. **Subject de `auth.Identity` = ID de usuario, no email.** El resto del Gateway usa `Subject`
   solo para auditoría/logs; acá además sirve para resolver ownership en
   `/users/{id}/api-keys` sin una consulta extra. Alternativa descartada: mantener email como
   Subject y resolver ownership por lookup adicional — más simple pero un round-trip extra por
   request.
2. **`identityMiddleware` prueba PostgreSQL primero, legacy después.** Evita invalidar de golpe
   las keys ya emitidas vía `GATEWAY_API_KEYS` en despliegues existentes.
3. **Hash sha256 + comparación en tiempo constante**, igual criterio que `apikey.Store` ya
   existente — no se introduce un algoritmo nuevo sin justificación de seguridad adicional en este
   slice (bcrypt/argon2 quedan reservados para login por password, HU-EVO-019/020).
4. **Fail-soft sin DSN**: `/users` y `/users/{id}/api-keys` responden `503` en vez de nil-pointer
   panic o 500 genérico, consistente con `/alerts` (HU-EVO-012/013) cuando no hay PostgreSQL
   configurada.

## Riesgos / Mitigación

- **Migración de instalaciones con `GATEWAY_API_KEYS` en producción**: mitigado por el fallback en
  `identityMiddleware` (ver Decisión 2); no exige migración inmediata.
- **IDs numéricos secuenciales como identificador público de usuario/key**: aceptado en este slice
  (mismo patrón que `provider_alerts.id`); no es un secreto, no hay enumeración de valor (las keys
  siguen siendo opacas, solo el ID de fila es secuencial).

## Migration Plan

Sin migración de datos: tablas nuevas, sin tocar `apikey.Store` ni `GATEWAY_API_KEYS`. Reversible
apagando `GATEWAY_USERS_POSTGRES_DSN` (el Gateway vuelve a depender solo del seed legacy).
