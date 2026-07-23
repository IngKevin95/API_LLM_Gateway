## Context

EP-001/EP-002 entregan el stack de resolución+entrega sin control de acceso. EP-004A añade la capa de identidad y accesos como **middlewares `http.Handler`** que envolverán el request-path (endpoints en EP-005). ADR-001: Go idiomático, preferir stdlib. Todo se prueba en aislamiento con `httptest` (sin IdP ni red real; JWT y certs generados en test). Datos sensibles: las API keys/JWT nunca se loguean; secretos por entorno.

## Goals / Non-Goals

**Goals:**
- AuthN por API key (hash, tiempo constante, solo header), OAuth2/OIDC (JWT) y mTLS.
- AuthZ por scope/RBAC + aislamiento multi-tenant.
- Rate limiting atómico en RAM (<5ms), límites de payload por tipo, protección Slowloris, concurrencia vision por nodo.
- Guardián de prompts opt-in con bypass pasivo.

**Non-Goals:**
- Endpoints HTTP (EP-005) — los middlewares se cablean ahí.
- Gestión de cuota acumulada por ventana larga y costo (EP-003; aquí solo rate limiting de abuso por segundo/minuto).
- Auditoría/redacción de PII y secretos server-side rotación (EP-004B).

## Decisions

- **Patrón middleware**: cada preocupación es un `func(http.Handler) http.Handler`. Se componen en una cadena; el orden (authN → authZ → rate limit → payload → guardian) se fija en EP-005 al montar el handler. Alternativa descartada: lógica inline en cada endpoint — no compone ni se testea aislado.
- **API key** (`auth/apikey`): almacena `sha256(key)`; compara con `subtle.ConstantTimeCompare`. Acepta solo header (`Authorization: Bearer` / `X-API-Key`), rechaza key en query/body con 401. Nunca loguea la key (a lo sumo un prefijo enmascarado). El consumidor identificado se inyecta en el `context` para auditoría.
- **AuthZ** (`authz`): la credencial porta un set de scopes (`capability:coding`, `capability:vision:trusted`, …), un `tenant`, y opcionalmente modelos vetados. Verifica capacidad solicitada, modelo forzado y tenant; `vision` exige `capability:vision:trusted` (porque deshabilita el DLP). 403 + log del intento ante violación; los recursos de otro tenant no son enumerables.
- **Rate limiting** (`ratelimit`): ventana fija por credencial en RAM con validación atómica (`sync.Mutex` por clave o `atomic`); 429 en <5ms sin tocar LLMs; correcto bajo concurrencia (test con -race). Payload: `http.MaxBytesReader` con límite por content-type (10MB texto / 50MB vision-base64, per PRD §2) → 413. Slowloris: `ReadHeaderTimeout` (ya en el scaffold `main`).
- **Vision concurrency** (`ratelimit`): semáforo por nodo (canal buffered de tamaño 2 o contador atómico); 3ra petición vision concurrente → 429 sin encolar. El enrutamiento por menor carga (no Hash L7) es una nota de balanceo L7 que se materializa en despliegue; aquí se modela el límite de concurrencia local.
- **OAuth2/OIDC** (`auth/oauth2`): valida firma, `aud`, `iss` y `exp` del JWT. **Dependencia**: `github.com/golang-jwt/jwt/v5` (validación robusta de JWT; implementar verificación RS256/JWKS a mano es propenso a errores de seguridad). Se registra como dependencia justificada. Claves públicas del IdP inyectadas (JWKS mockeable en test). Firma inválida/exp → 401; `aud` incorrecto → 403.
- **mTLS** (`auth/mtls`): `tls.Config{ClientAuth: RequireAndVerifyClientCert, ClientCAs: trustStore}`; extrae el scope del certificado cliente (`Subject`/OU o SAN). Revocado/expirado/sin-cert/CA no confiable → falla el handshake TLS. Test con `httptest.NewTLSServer` + certs generados con `crypto/x509`.
- **Prompt guardian** (`promptguard`): opt-in por header/param; muta el último mensaje `user` envolviéndolo en un template, preservando el texto original y la sintaxis de tool calling; bypass pasivo ante prompt inválido o si la optimización excede 100ms; no altera el flujo de streaming (opera sobre el request, no la respuesta).

## Risks / Trade-offs

- **JWT/mTLS son superficie de seguridad** → apoyarse en `crypto/tls` y `golang-jwt` (probados) en vez de criptografía a mano; tests negativos exhaustivos (firma adulterada, aud incorrecto, CA no confiable, cert revocado).
- **Rate limit concurrente** → race conditions; validación atómica + tests `-race` (HU-022 AC4: 10 concurrentes, 1 pasa).
- **Secretos en logs** → invariante: keys/JWT nunca se loguean; solo prefijos enmascarados. Verificado por tests.
- **Guardian muta el request** → riesgo de romper tool calling/streaming; bypass pasivo + tests que preservan tools y no tocan la respuesta.

## Migration Plan

Aditivo sobre EP-001/EP-002 (en develop). Los middlewares no se activan hasta que EP-005 los monte sobre los endpoints. Sin migración de datos. Rollback = revertir el PR.

## Open Questions

- Formato de configuración de credenciales/scopes/tenants y límites: se define en el sub-slice que lo necesite, alineado al `config.yaml`/entorno. No bloqueante para SS1.
