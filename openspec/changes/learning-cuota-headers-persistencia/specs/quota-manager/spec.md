# Quota Manager Specification — Delta

## ADDED Requirements

### Requirement: Learn quota from response headers
Quota Manager SHALL expose `LearnFromHeaders(providerID, modelID, quotaInfo)` method that updates in-memory remaining quota atomically, detecting window resets and clamping negatives to 0. (HU-EVO-007)

#### Scenario: LearnFromHeaders updates remaining atomically
- **WHEN** post-response, adapter calls `qm.LearnFromHeaders("openai", "", QuotaInfo{Remaining: 9950})`
- **THEN** immediate `qm.Remaining("openai", "")` returns 9950; no race conditions under `go test -race`

#### Scenario: Reset detection and reactivation
- **WHEN** previous `remaining=0, resetAt=yesterday`, new response has `resetAt=today` with `remaining=<new>`
- **THEN** Manager detects reset, updates remaining, reactivates provider in Router

### Requirement: Async persistence to PostgreSQL
Quota Manager SHALL enqueue learned quotas to background worker for async DB persist without blocking response (<5ms overhead). Worker batch-writes via UPSERT. (HU-EVO-008)

#### Scenario: Async enqueue non-blocking
- **WHEN** LearnFromHeaders() called
- **THEN** method enqueues job and returns immediately; DB write happens in parallel; response unblocked

#### Scenario: Boot restore from PostgreSQL
- **WHEN** Gateway starts, Quota Manager initializes
- **THEN** queries `provider_quotas_learned`, restores latest learned values per provider, preferring learned over `quota_hint`

## MODIFIED Requirements

### Requirement: Commit confirms actual consumption and adjusts balances
El Quota Manager SHALL accept real consumption post-execution, adjust saldos, y triggear async persistence de learned quotas si response headers indican cambio. Si response incluye QuotaInfo, llama LearnFromHeaders() automáticamente dentro de Commit().

FROM: `Commit confirma el consumo real post-ejecución y ajusta saldos.`

TO: `Commit() acepta consumption real, ajusta balances, y si la respuesta incluye headers de cuota, automáticamente invoca LearnFromHeaders() para aprender del servidor.`

#### Scenario: Commit triggers learning
- **WHEN** response.QuotaInfo != nil in Commit()
- **THEN** Commit() calls LearnFromHeaders() internally before returning; learned quota updated and enqueued for persistence
