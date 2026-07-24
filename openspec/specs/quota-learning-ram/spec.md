# Quota Learning (RAM) Specification

## Overview
Quota Manager learns actual provider limits from response headers and maintains in-memory state atomically.

## ADDED Requirements

### Requirement: Atomic RAM update of remaining quota
Quota Manager SHALL expose `LearnFromHeaders(providerID, modelID, quotaInfo)` method that atomically updates `remaining` value in RAM using mutex protection. (HU-EVO-007 AC1)

#### Scenario: Successful remaining update
- **WHEN** LearnFromHeaders() receives `QuotaInfo{Remaining: 9950}`
- **THEN** immediate call to `Remaining("openai", "")` returns 9950 (atomic, no race)

### Requirement: Learned value takes precedence over hint
When learned `Limit` from headers exceeds `quota_hint` from YAML, Quota Manager SHALL use the learned value (server is authoritative). (HU-EVO-007 AC2)

#### Scenario: Learned limit higher than hint
- **WHEN** quota_hint says `limit: 500000` but server header says `X-RateLimit-Limit: 1000000`
- **THEN** Quota Manager sets `limit: 1000000` (learned is trusted)

### Requirement: Clamp negative remaining to zero
If headers report negative `remaining` (consumptio exceeded limit mid-stream), clamp to 0 and mark as exhausted. (HU-EVO-007 AC3)

#### Scenario: Overshoot clamped
- **WHEN** server devuelve `Remaining: -100`
- **THEN** Quota Manager sets `remaining: 0` and flags provider as exhausted

### Requirement: Thread-safe concurrent updates
Quota Manager SHALL handle multiple parallel requests updating remaining concurrently without data loss or corruption. Last-write-wins strategy. (HU-EVO-007 AC4)

#### Scenario: 10 concurrent requests update remaining
- **WHEN** 10 requests from same provider arrive simultaneously and all call LearnFromHeaders()
- **THEN** final `remaining` value is consistent; no data races detected by `go test -race`

### Requirement: Detect and react to quota window reset
When `ResetAt` timestamp crosses current time (window reset), Quota Manager SHALL detect reset and update `remaining` accordingly, reactivating provider if it was exhausted. (HU-EVO-007 AC5)

#### Scenario: Quota window reset detection
- **WHEN** previous state had `remaining: 0, reset_at: yesterday + 24h`, and new response arrives with `reset_at: today + 24h` and `remaining: <new>`
- **THEN** Quota Manager detects reset (ResetAt changed), updates remaining, and reactivates in Router
