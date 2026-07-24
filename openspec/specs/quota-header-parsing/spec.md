# Quota Header Parsing Specification

## Overview
Adapters extract rate-limit information from HTTP response headers and normalize into `QuotaInfo` struct.

## ADDED Requirements

### Requirement: Normalize X-RateLimit headers (OpenAI standard)
Each adapter SHALL extract `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset` from HTTP response headers and normalize into `QuotaInfo{Limit, Remaining, ResetAt}`. (HU-EVO-006 AC1)

#### Scenario: OpenAI headers extracted correctly
- **WHEN** adapter.Chat() receives response with headers `X-RateLimit-Limit-Requests: 10000, X-RateLimit-Remaining-Requests: 9950, X-RateLimit-Reset-Requests: 1721760000`
- **THEN** adapter returns `QuotaInfo{Limit: 10000, Remaining: 9950, ResetAt: time.Unix(1721760000)}`

### Requirement: Normalize Anthropic provider headers
Anthropic adapters SHALL extract `anthropic-ratelimit-requests-limit`, `anthropic-ratelimit-requests-remaining`, `anthropic-ratelimit-requests-reset` and normalize to standard schema. (HU-EVO-006 AC2)

#### Scenario: Anthropic headers case-insensitive
- **WHEN** adapter receives response with Anthropic-specific headers (case-insensitive variants)
- **THEN** adapter normalizes and returns `QuotaInfo` with exact same struct as OpenAI format

### Requirement: Case-insensitive header parsing (Groq)
All adapters SHALL parse headers case-insensitively per HTTP spec (`x-ratelimit-limit-requests` same as `X-RateLimit-Limit-Requests`). (HU-EVO-006 AC3)

#### Scenario: Groq lowercase headers
- **WHEN** adapter receives lowercase `x-ratelimit-limit-requests: 30, x-ratelimit-remaining-requests: 29`
- **THEN** adapter normalizes and returns `QuotaInfo{Limit: 30, Remaining: 29}`

### Requirement: Graceful fallback for missing headers
Adapters SHALL return `QuotaInfo{Limit: 0, Remaining: 0, ResetAt: nil}` without crashing if no rate-limit headers present. (HU-EVO-006 AC4)

#### Scenario: Provider without quota headers
- **WHEN** adapter receives response with no X-RateLimit-* headers
- **THEN** adapter returns empty `QuotaInfo` without error; calling code detects `Limit == 0` and skips learning

### Requirement: Parse multiple reset header formats
Adapters SHALL detect and parse both `X-RateLimit-Reset: <unix-timestamp>` (seconds) and `Retry-After: <seconds>` or `Retry-After: <RFC1123-date>` formats, converting to normalized `time.Time`. (HU-EVO-006 AC5)

#### Scenario: Unix timestamp reset
- **WHEN** response has `X-RateLimit-Reset: 1721760000`
- **THEN** adapter converts to `time.Time` and sets `ResetAt`

#### Scenario: Retry-After in seconds
- **WHEN** response has `Retry-After: 60`
- **THEN** adapter calculates `ResetAt = now.Add(60 * time.Second)`

#### Scenario: Retry-After in RFC1123 date
- **WHEN** response has `Retry-After: Wed, 23 Jul 2026 19:00:00 GMT`
- **THEN** adapter parses RFC1123 and converts to `time.Time`
