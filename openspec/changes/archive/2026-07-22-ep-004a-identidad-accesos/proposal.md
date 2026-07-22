## Why

El Gateway hoy resuelve y entrega respuestas (EP-001/EP-002) pero sin controlar **quién** entra: cualquier consumidor anónimo accedería a los modelos. EP-004A cierra ese hueco con autenticación (API key, OAuth2/OIDC, mTLS), autorización por scope/RBAC y tenant, y protección de abuso (rate limiting, límites de payload, concurrencia de vision). Es requisito explícito del sponsor (Obj. 4: 100% auth + authz) y la tercera épica `foundational`.

## What Changes

- **AuthN por API key**: exige key válida en header (nunca en URL/query), comparación en tiempo constante, hash almacenado, nunca se loguea la key. 401 ante ausencia/invalidez.
- **AuthZ por scope/RBAC y tenant**: autoriza por `capability:<x>` y aísla por tenant (sin fuga cross-tenant); modelo vetado y `vision` exigen scope explícito. 403 ante violación.
- **Rate limiting + payload**: 429 al exceder cuota por ventana (validación atómica en RAM, <5ms), 413 por payload (10MB texto / 50MB vision-base64 por PRD §2), protección Slowloris (ReadHeaderTimeout).
- **Concurrencia vision**: máximo 2 peticiones vision concurrentes por nodo (429 sin encolar), enrutamiento por menor carga (no Hash L7).
- **OAuth2/OIDC**: valida firma, `aud`/`iss`, expiración de JWT de IdP corporativo (401/403).
- **mTLS**: autenticación por certificado cliente, extrae scope del cert, rechaza revocados/sin-cert/CA no confiable en capa TLS.
- **Guardián de Prompts (opt-in)**: envuelve el último mensaje `user` en un template de optimización, con bypass pasivo ante fallo/timeout, sin alterar tool calling ni streaming.
- Middlewares construidos y probados en aislamiento (httptest); el cableado al request-path HTTP se difiere a EP-005 (API universal).

## Capabilities

### New Capabilities
- `authentication`: verificación de identidad por API key (hash, tiempo constante, solo header), OAuth2/OIDC (firma+aud+iss+exp de JWT) y mTLS (certificado cliente, trust store). Identifica al consumidor para auditoría.
- `authorization`: control de acceso por scope/RBAC y aislamiento multi-tenant; capacidad/modelo vetados y `vision` exigen scope explícito.
- `rate-limiting`: throttling por ventana con validación atómica en RAM (429), límites de payload por tipo (413), protección Slowloris, y límite de concurrencia de vision por nodo con enrutamiento por menor carga.
- `prompt-guardian`: optimización opt-in del prompt (wrapping del último mensaje user) con bypass pasivo, preservando tool calling y streaming.

### Modified Capabilities
<!-- Ninguna spec previa cambia sus requisitos. Los middlewares envuelven el request-path que expondrá EP-005; el router (model-router) no cambia su contrato. -->

## Impact

- Código nuevo (Go): `src/internal/auth` (apikey, oauth2, mtls), `src/internal/authz`, `src/internal/ratelimit`, `src/internal/promptguard`. Nombres tentativos, se fijan en design.md.
- Middlewares `http.Handler` que envuelven handlers; se prueban con `httptest.Server` (sin IdP/red real; JWT firmados en test con clave efímera).
- Dependencias: preferir stdlib (`crypto/*`, `net/http`, `crypto/tls`); validación JWT puede requerir una librería (a evaluar en design.md contra el stack). Rate limit en RAM (sin dependencias).
- Sin breaking changes; sin endpoints nuevos (EP-005). Config: keys/scopes/tenants y límites se leen de configuración (formato a definir; secretos por entorno).

## Trazabilidad

- **Épica**: EP-004A · Identidad y Accesos (`layer: foundational`) — objetivo del PRD: Obj. 4 (seguridad empresarial; 100% auth + authz por scope; secretos protegidos).
- **Historias cubiertas** (por sub-slice):
  - SS1 — `authentication` (API key) + `authorization`: HU-008 (auth API key), HU-009 (authz scope/RBAC/tenant)
  - SS2 — `rate-limiting`: HU-022 (rate limiting + payload + Slowloris), HU-022b (concurrencia vision)
  - SS3 — `authentication` (federada): HU-025a (OAuth2/OIDC), HU-025b (mTLS)
  - SS4 — `prompt-guardian`: HU-027 (Guardián de Prompts opt-in)
