## 1. Adapter genérico data-driven (HU-EVO-001)

- [x] 1.1 Definir `ProviderSpec` (`BaseURL`, `AuthHeader`, `Format`, `Headers`, `TimeoutMs`) en `src/internal/adapter/generic`
- [x] 1.2 Implementar validación de spec y `ErrInvalidProviderSpec`
- [ ] 1.3 Extraer traducción OpenAI compartida a paquete interno reutilizable por `generic` y `adapter/openai` — DIFERIDO: se reimplementó traducción propia y mínima dentro de `generic` (duplicación aceptada por presupuesto de tiempo, cero riesgo de regresión sobre `adapter/openai` existente; ver progress_log 2026-07-23T22:15:00Z)
- [ ] 1.4 Extraer traducción Claude compartida a paquete interno reutilizable por `generic` y `adapter/anthropic` — DIFERIDO: mismo motivo que 1.3
- [x] 1.5 Implementar `Chat/Stream/Embed` del adapter genérico seleccionando estrategia por `Format`
- [x] 1.6 Implementar inyección de headers extra sin sobrescribir el header de autenticación
- [x] 1.7 Implementar timeout por spec independiente del timeout global
- [x] 1.8 Tests unitarios: AC1-AC5 de HU-EVO-001 (Groq openai-compat, Claude-compat, spec inválido, headers extra, timeout)

## 2. Registry carga free-tier.yaml (HU-EVO-002)

- [x] 2.1 Definir schema de `config/providers/free-tier.yaml` (5 proveedores: Groq, Cerebras, Mistral, Gemini, Cloudflare AI)
- [x] 2.2 Extender `Registry.Load()` para leer `free-tier.yaml` tras `config.yaml`
- [x] 2.3 Implementar merge con precedencia de `free-tier.yaml` sobre `config.yaml` por `providerID`
- [x] 2.4 Log INFO listando providers sobrescritos por `free-tier.yaml`
- [x] 2.5 Validación fail-fast (`ErrInvalidConfig`) ante YAML malformado
- [x] 2.6 Excluir del scoring providers con `models: []` vacío
- [x] 2.7 Tests unitarios: AC1-AC5 de HU-EVO-002

## 3. Conformance test extendido (HU-EVO-003)

- [x] 3.1 Agregar caso table-driven en `conformance_test.go` que itera `Registry.AllProviderSpecs()`
- [x] 3.2 Levantar `httptest.Server` simulando cada `format` (openai/claude) para validar Chat/Stream/Embed
- [x] 3.3 Verificar normalización `Content/Model/Usage` en cada respuesta
- [x] 3.4 Manejar spec sin modelo default con `ErrNoModelAvailable` sin abortar suite
- [x] 3.5 Timeout individual por proveedor vía `context.WithTimeout`
- [x] 3.6 Ejecutar casos con `t.Parallel()` y verificar ausencia de race conditions (`go test -race`)

## 4. Health Monitor: detección de 429 (HU-EVO-004)

- [x] 4.1 Agregar mapa `providerID -> retiredUntil` en `src/internal/health/health.go`
- [x] 4.2 Detectar 429 y calcular retiro desde `Retry-After` (o default 30s si ausente) — incluye propagación real end-to-end `ProviderError.RetryAfter` -> `failover.Engine.OnRateLimited` -> `health.Monitor.RetireOn429` (fix de sesión 2026-07-23T23:10:00Z)
- [x] 4.3 Reactivación automática al vencer `retiredUntil`
- [x] 4.4 Abortar streams mid-stream ante 429 sin failover transparente
- [x] 4.5 Backoff exponencial ante 429 repetidos (30s → 60s → 120s, tope configurable)
- [x] 4.6 Exponer estado de retiro vía `HealthSource` existente (sin nueva interfaz)
- [x] 4.7 Tests unitarios: AC1-AC5 de HU-EVO-004

## 5. Quota Manager: init desde quota_hint (HU-EVO-005)

- [x] 5.1 Inicializar `remaining` por proveedor desde `quota_hint` en boot (`src/internal/quota/manager.go`)
- [x] 5.2 Tratar `quota_hint <= 0` como agotado (`remaining = 0`)
- [x] 5.3 Default de 1M tokens cuando `quota_hint` está ausente
- [x] 5.4 Precedencia: valor aprendido en runtime (headers) sobrescribe `quota_hint` inicial
- [ ] 5.5 Precedencia: valor restaurado desde PostgreSQL en boot sobrescribe `quota_hint` inicial — DIFERIDO por acuerdo explícito de equipo a HU-EVO-008 (EP-EVO-002, Persistencia async PostgreSQL); `RestoreRemaining()` existe y está testeado unitariamente con valor inyectado a mano, pero `cmd/gateway/main.go` no conecta a PostgreSQL en este slice (in-memory-first, ver `docs/04-historias/HU-EVO-005-quota-manager-inicializar-contadores.md#nota-de-alcance-ac5-diferido-2026-07-23`)
- [x] 5.6 Tests unitarios: AC1-AC5 de HU-EVO-005 (AC5 con evidencia parcial, ver 5.5)

## 6. Integración y verificación de cableado

- [x] 6.1 Verificar integración adapter genérico <- Registry.Load (INT-adapter-registry)
- [x] 6.2 Verificar integración Registry providers -> Quota Manager init (INT-registry-quota)
- [x] 6.3 Verificar integración Registry providers -> Health Monitor 429 (INT-registry-health)
- [x] 6.4 Verificar integración adapter genérico <- conformance_test.go (INT-adapter-conformance)
- [x] 6.5 `go build ./...`, `go vet ./...`, `go test ./... -race` en verde
- [x] 6.6 Smoke end-to-end: boot con `free-tier.yaml`, `/health` y `/metrics` responden 200 con contenido real (providers/quota poblados, no solo HTTP 200)
