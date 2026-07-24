# Quota Persistence Specification

## Overview
Learned quota values persist asynchronously to PostgreSQL for crash recovery and auditing.

## ADDED Requirements

### Requirement: Async persistence without blocking request
Quota Manager SHALL enqueue learned quotas to async worker and return immediately (<5ms total latency added to response). (HU-EVO-008 AC1)

#### Scenario: Async persistence non-blocking
- **WHEN** LearnFromHeaders() processes response
- **THEN** method enqueues job and returns; DB write happens in background worker; response sent to client in parallel

### Requirement: Restore learned quotas on restart
On Gateway boot, Quota Manager SHALL query PostgreSQL `provider_quotas_learned` and restore latest `remaining` value per provider, preferring learned over hint. (HU-EVO-008 AC2)

#### Scenario: Boot restores learned from DB
- **WHEN** Gateway restarts after learning `remaining: 500000` for "openai"
- **THEN** Manager queries `provider_quotas_learned`, finds latest row, initializes to `remaining: 500000` (not `quota_hint`)

### Requirement: DB unavailability does not block requests
If PostgreSQL is down, async worker logs warnings and continues in-memory; requests proceed without blocking on DB. (HU-EVO-008 AC3)

#### Scenario: DB connection failure graceful fallback
- **WHEN** PostgreSQL is unreachable and worker attempts persist
- **THEN** worker logs error, drops or retries job locally, does NOT block request path

### Requirement: Idempotent persistence with UPSERT
Multiple writes to same (provider_id, model_id, learned_at) SHALL be idempotent using PostgreSQL UPSERT (ON CONFLICT DO UPDATE). (HU-EVO-008 AC4)

#### Scenario: Concurrent writes to same provider-model
- **WHEN** 100 parallel requests write learned quota for same provider+model
- **THEN** DB uses UPSERT to avoid duplicates; final state is consistent

### Requirement: Audit trail of quota changes
Each learned quota write SHALL create a row in `provider_quotas_learned` with timestamp, preserving historical changes for auditing. (HU-EVO-008 AC5)

#### Scenario: Audit trail preserved
- **WHEN** Cerebras quota updates 3 times in 1 minute
- **THEN** table has 3 rows (one per update) with distinct `learned_at` timestamps; auditors can replay history
