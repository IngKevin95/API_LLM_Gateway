## 1. Sub-slice 1 — Contrato Adapter + OpenAI (HU-020a/b/c)

- [x] 1.1 Definir contrato `Adapter` (Chat/Stream/Embed) + tipos Request/Response/TokenStream/Embedding y `*ProviderError{Status,Retryable,Provider}` en `src/internal/adapter`
- [x] 1.2 (test-first) OpenAI chat: happy (traducción a /v1/chat/completions), tool calling preservado, 500/timeout→ProviderError (HU-020a AC1/2/3) con httptest.Server
- [x] 1.3 Implementar `adapter/openai` chat que hace verde 1.2
- [x] 1.4 (test-first) OpenAI streaming: SSE feliz, failover pre-primer-token, corte mid-stream por Stream Idle Timeout (HU-020b AC1/2/3)
- [x] 1.5 Implementar streaming OpenAI (io.Pipe + timer idle) que hace verde 1.4
- [x] 1.6 (test-first) OpenAI embeddings: happy /v1/embeddings, lote grande sin truncar, modelo no soportado→ProviderError (HU-020c AC1/2/3)
- [x] 1.7 Implementar embeddings OpenAI que hace verde 1.6
- [x] 1.8 journey_smoke SS1: httptest mock OpenAI, adapter chat/stream/embed devuelven respuesta normalizada o ProviderError; suite verde con -race

## 2. Sub-slice 2 — Anthropic + local (HU-021a/b, HU-024)

- [x] 2.1 (test-first) Anthropic chat: traducción de roles (system), tools→tool_use, seed ignorado con WARN, max_tokens default 4096, 5xx/429→ProviderError (HU-021a AC1-5)
- [x] 2.2 Implementar `adapter/anthropic` chat que hace verde 2.1
- [x] 2.3 (test-first) Anthropic streaming: eventos nativos→chunks SSE OpenAI, failover pre-primer-token, corte mid-stream (HU-021b AC1/2/3)
- [x] 2.4 Implementar streaming Anthropic que hace verde 2.3
- [x] 2.5 (test-first) Adapter local OpenAI-compat: happy, timeout local, respuesta no compatible→ProviderError sin crashear (HU-024 AC1/2/3)
- [x] 2.6 Implementar `adapter/local` (reutiliza openai-compat con base_url) que hace verde 2.5
- [x] 2.7 journey_smoke SS2: httptest mocks Anthropic + local; los 3 adapters conviven bajo el contrato; suite verde con -race

## 3. Sub-slice 3 — Failover Engine + timeouts (HU-004a, HU-004c)

- [x] 3.1 (test-first) Failover: 503→siguiente, pool agotado→502, 429→failover (no retry), degradación a local, 400→sin failover (HU-004a AC1-5) con adapters mock
- [x] 3.2 Implementar `src/internal/failover` que consume router.Resolve + adapters, con límite de intentos = largo de cadena, que hace verde 3.1
- [x] 3.3 (test-first) Timeouts dinámicos: TTFT estricto chat/código→abort+failover, reasoning relajado sin failover, Stream Idle Timeout→corta+penaliza (HU-004c AC1/2/3)
- [x] 3.4 Implementar timeouts por capacidad (leídos del routing/registry) que hace verde 3.3
- [x] 3.5 journey_smoke SS3: cadena real (registry+router) + adapters mock; una petición se completa vía failover degradando a local; suite verde con -race

## 4. Sub-slice 4 — Circuit Breaker + Health Monitor (HU-004b, HU-005)

- [x] 4.1 (test-first) Circuit Breaker pasivo: marca inalcanzable tras fallo/Max In-Flight, fast-fail 0-I/O al exceder límite, reactivación tras backoff+health (HU-004b AC1/2/3) con -race
- [x] 4.2 Implementar `src/internal/breaker` (atomic in-flight + cooldown) que hace verde 4.1
- [x] 4.3 (test-first) Health Monitor: sano se mantiene, caída→retiro, reactivación, todos no-sanos→503, intermitente con histéresis N/M (HU-005 AC1-5) con ticker mockeable
- [x] 4.4 Implementar `src/internal/health` (ticker + histéresis + RWMutex) que implementa router.HealthSource, que hace verde 4.3
- [x] 4.5 Integración de componentes (nivel librería, probada en journey-smoke SS4): `health.Monitor` implementa `router.HealthSource` (reemplaza `StaticHealth`) y el breaker se integra en `failover.Engine` (Allow gate + Trip en error retryable). El wiring del **entrypoint HTTP en `cmd/gateway`** (handler de completions consumiendo router+failover+health) se **difiere explícitamente a EP-005** (API universal), donde existe el endpoint; sin ese consumidor, cablearlo en `main` sería código sin uso. No es recorte silencioso: los componentes ya wire juntos, verificado por `failover/breaker_integration_test.go` (TestIntegration_FullStack_HealthAndBreaker)
- [x] 4.6 journey_smoke SS4: health monitor retira/reactiva un mock; breaker previene Failover Suicide bajo concurrencia; router usa salud viva; suite verde con -race

## 5. Cierre de épica

- [x] 5.1 Coherencia triple AC↔specs↔tests (coherence-three-way) sin huecos
- [x] 5.2 Verificación adversarial de cableado (wiring-adversarial-verifier) → wiring_verified
- [x] 5.3 DoD reducido (dor-dod-gatekeeper) → dod
- [ ] 5.4 PR + opsx:archive del change en el mismo PR
