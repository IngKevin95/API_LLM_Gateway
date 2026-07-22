## 1. Sub-slice 1 — AuthN API key + AuthZ (HU-008, HU-009)

- [x] 1.1 (test-first) Middleware API key: key válida (header)→200+identidad en ctx; sin key→401; inválida/revocada→401 sin loguear key; key en query→401 (HU-008 AC1-4) con httptest
- [x] 1.2 Implementar `src/internal/auth/apikey` (sha256 + subtle.ConstantTimeCompare, solo header, identidad en context) que hace verde 1.1
- [x] 1.3 (test-first) Middleware AuthZ: scope permitido→autoriza; capacidad fuera de scope→403; cruce de tenant→403 no enumerable; modelo vetado→403; vision sin scope trusted→403 (HU-009 AC1-5)
- [x] 1.4 Implementar `src/internal/authz` (scopes + tenant + modelos vetados, lee identidad del ctx) que hace verde 1.3
- [x] 1.5 journey_smoke SS1: cadena authN→authZ sobre un handler dummy vía httptest; 200 para credencial válida+scope, 401/403 en los caminos negativos; suite verde

## 2. Sub-slice 2 — Rate limiting + payload + vision (HU-022, HU-022b)

> Precondición: corregir HU-022 (aclarar límites 10MB texto / 50MB vision; traducir AC al español) antes de abrir SS2.

- [x] 2.1 (test-first) Rate limit por ventana: dentro→pasa; excedido→429 <5ms sin LLM; 10 concurrentes con 1 token→1 pasa 9 bloquean (HU-022 AC1/2/4) con -race
- [x] 2.2 Implementar `src/internal/ratelimit` (ventana fija atómica en RAM) que hace verde 2.1
- [x] 2.3 (test-first) Límite de payload por tipo: texto >10MB→413; vision/base64 >50MB→413 sin cargar cuerpo; Slowloris ReadHeaderTimeout (HU-022 AC3/5/6)
- [x] 2.4 Implementar límites de payload (http.MaxBytesReader por content-type) + confirmar ReadHeaderTimeout en el server
- [x] 2.5 (test-first) Concurrencia vision: 2 activas OK, 3ra→429 sin encolar; enrutamiento por menor carga (HU-022b AC1/2/3) con -race
- [x] 2.6 Implementar semáforo de concurrencia vision por nodo que hace verde 2.5
- [x] 2.7 journey_smoke SS2: middlewares de rate/payload/vision sobre handler dummy; suite verde con -race

## 3. Sub-slice 3 — OAuth2/OIDC + mTLS (HU-025a, HU-025b)

- [x] 3.1 Añadir `github.com/golang-jwt/jwt/v5` a go.mod (justificado en design.md)
- [x] 3.2 (test-first) OAuth2/OIDC: JWT válido→autoriza; expirado→401; firma inválida→401+log; aud incorrecto→403 (HU-025a AC1-4) con JWT firmados en test
- [x] 3.3 Implementar `src/internal/auth/oauth2` (valida firma+aud+iss+exp, JWKS inyectable) que hace verde 3.2
- [x] 3.4 (test-first) mTLS: cert válido→scope extraído; revocado/expirado→handshake falla; sin cert→rechazo TLS; CA no confiable→unknown CA (HU-025b AC1-4) con httptest.NewTLSServer + certs generados
- [x] 3.5 Implementar `src/internal/auth/mtls` (tls.Config RequireAndVerifyClientCert + trust store, extrae scope del cert) que hace verde 3.4
- [x] 3.6 journey_smoke SS3: los 3 métodos AuthN (apikey/oauth2/mtls) conviven bajo una interfaz Authenticator común; suite verde

## 4. Sub-slice 4 — Guardián de Prompts (HU-027)

> Precondición: corregir HU-027 (alinear título↔AC: los AC son de optimización, no de jailbreak; traducir frases en inglés) antes de abrir SS4.

- [x] 4.1 (test-first) Guardián opt-in: envuelve último msg user preservando original; prompt inválido→bypass; tool calling intacto; overhead >100ms→bypass; stream no altera respuesta (HU-027 AC1-5)
- [x] 4.2 Implementar `src/internal/promptguard` (wrapping opt-in con timeout de 100ms y bypass pasivo) que hace verde 4.1
- [x] 4.3 journey_smoke SS4: guardián sobre un request con/sin opt-in; preserva tools y streaming; suite verde

## 5. Cierre de épica

- [x] 5.X Coherencia triple AC↔specs↔tests (coherence-three-way) sin huecos
- [x] 5.X Verificación adversarial de cableado (wiring-adversarial-verifier) → wiring_verified
- [x] 5.X DoD reducido (dor-dod-gatekeeper) → dod
- [ ] 5.4 PR + opsx:archive del change en el mismo PR
