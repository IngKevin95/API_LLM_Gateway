# Tasks — EP-EVO-002 Implementation Checklist

## 1. Adapter Header Parsing (HU-EVO-006)

- [x] 1.1 Define `QuotaInfo` struct in `src/internal/adapter/types.go`: `{Limit int64, Remaining int64, ResetAt *time.Time}`
- [x] 1.2 Extend `Response` struct to include `QuotaInfo` field
- [x] 1.3 Add `extractQuota()` method to base adapter (common parsing logic)
- [x] 1.4 Implement OpenAI headers parsing in `src/internal/adapter/openai/adapter.go`
- [x] 1.5 Implement Anthropic headers parsing (anthropic-ratelimit-* variants)
- [x] 1.6 Implement Groq headers parsing (case-insensitive x-ratelimit-*)
- [x] 1.7 Implement Google headers parsing (google-specific variants)
- [x] 1.8 Add graceful fallback for missing headers (return empty QuotaInfo)
- [x] 1.9 Parse Retry-After header (seconds and RFC1123 date formats)
- [x] 1.10 Unit tests: mock HTTP responses with/without headers (5 scenarios per HU-EVO-006 AC)
- [x] 1.11 Test race conditions: concurrent parsing does not corrupt state

## 2. Quota Manager Learning in RAM (HU-EVO-007)

- [x] 2.1 Add `LearnFromHeaders(providerID, modelID, quotaInfo)` method to `quota.Manager`
- [x] 2.2 Implement mutex protection for atomic updates (sync.RWMutex)
- [x] 2.3 Implement reset detection logic (compare ResetAt timestamps)
- [x] 2.4 Implement conservative update strategy (learned > current, or reset detected)
- [x] 2.5 Clamp negative remaining to 0; mark as exhausted
- [x] 2.6 Unit tests: atomic updates under race conditions (`go test -race`)
- [x] 2.7 Unit tests: reset detection and reactivation
- [x] 2.8 Unit tests: 10 concurrent requests, final state consistency
- [x] 2.9 Integrate LearnFromHeaders() call into `quota.Commit()` when response.QuotaInfo present

## 3. Quota Persistence to PostgreSQL (HU-EVO-008)

- [x] 3.1 Create PostgreSQL schema: `provider_quotas_learned` table (id, provider_id, model_id, limit, remaining, reset_at, learned_at)
- [x] 3.2 Add unique constraint on (provider_id, model_id, learned_at)
- [x] 3.3 Implement background worker in `src/internal/quota/persist.go` (goroutine + channel)
- [x] 3.4 Implement batch writes (100ms or 1000 jobs threshold)
- [x] 3.5 Implement UPSERT logic (ON CONFLICT DO UPDATE) for idempotence
- [x] 3.6 Add retry logic for transient DB failures
- [x] 3.7 Implement `RestoreRemaining()` method to read from DB on boot
- [x] 3.8 Wire RestoreRemaining() into Gateway boot sequence
- [x] 3.9 Unit tests: async enqueue <5ms latency
- [x] 3.10 Unit tests: boot restores learned from DB
- [x] 3.11 Unit tests: DB down does not block requests (graceful fallback)
- [x] 3.12 Unit tests: UPSERT idempotence under concurrent writes
- [x] 3.13 Integration tests: end-to-end learn → persist → restore flow

## 4. Router Score Penalization (HU-EVO-009)

- [x] 4.1 Extend `Router.Score()` method to check `remaining < limit * 0.2`
- [x] 4.2 Apply 50% score multiplier (score *= 0.5) if below 20%
- [x] 4.3 Update failover chain generation to respect penalización
- [x] 4.4 Ensure Router.Score() reads from quota.Manager atomically
- [x] 4.5 Unit tests: penalización applied (15% remaining)
- [x] 4.6 Unit tests: no penalización if > 20%
- [x] 4.7 Unit tests: remaining = 0 highly penalized
- [x] 4.8 Unit tests: failover chain respects penalization (non-penalized first)
- [x] 4.9 Integration tests: end-to-end scoring with learned quotas

## 5. Failover 429 Handling (HU-EVO-010)

- [x] 5.1 Extend `adapter.ProviderError` with `RetryAfter *time.Duration` field
- [x] 5.2 Add Retry-After parsing in adapter response handling (seconds + RFC1123)
- [x] 5.3 Wire failover to extract RetryAfter from ProviderError
- [x] 5.4 Wire failover to call `health.Monitor.RetireOn429(providerID, retryAfter)`
- [x] 5.5 Default RetryAfter to 30s if header absent
- [x] 5.6 Implement mid-stream 429 detection (abort, no transparent failover)
- [x] 5.7 Health Monitor backoff exponential logic (30s → 60s → 120s → cap 120s)
- [x] 5.8 Unit tests: Retry-After extraction (seconds, RFC1123 date)
- [x] 5.9 Unit tests: 429 without Retry-After defaults to 30s
- [x] 5.10 Unit tests: mid-stream 429 aborts (no failover)
- [x] 5.11 Unit tests: backoff exponential (5 × 429 → tope 120s)
- [x] 5.12 Integration tests: end-to-end 429 handling → retire → failover

## 6. Database & Boot Integration

- [x] 6.1 Migrate PostgreSQL: `provider_quotas_learned` table creation
- [x] 6.2 Wire boot sequence in `cmd/gateway/main.go` to call RestoreRemaining()
- [x] 6.3 Graceful DB connection handling (timeouts, retries)
- [x] 6.4 Add env vars: `GATEWAY_QUOTA_DB_*` for connection details (optional; default to main DB)
- [x] 6.5 Add metrics tracking for learning (learned_updates_total, persist_errors_total)

## 7. End-to-End Integration Tests

- [x] 7.1 Test: boot → load free-tier.yaml → initialize quotas from hint → learn from headers → persist → restart → restore
- [x] 7.2 Test: 429 response → extract Retry-After → retire provider → failover → reactivate after timeout
- [x] 7.3 Test: penalización applies correctly when remaining < 20% during routing
- [x] 7.4 Test: concurrent requests learn + persist without race conditions

## 8. Smoke Test & Deployment Readiness

- [x] 8.1 Build: `go build ./...` clean (no errors, warnings OK)
- [x] 8.2 Vet: `go vet ./...` clean
- [x] 8.3 Tests: `go test ./... -race` 100% pass (all packages)
- [x] 8.4 Boot: binario arranca con free-tier.yaml + DB, /health responds 200, /metrics includes quota block
- [x] 8.5 Verify learning: simulate provider response with headers, confirm remaining updated in /metrics
- [x] 8.6 Verify persistence: stop gateway, restart, confirm learned quotas restored
- [x] 8.7 Document: update README with quota-learning feature overview
