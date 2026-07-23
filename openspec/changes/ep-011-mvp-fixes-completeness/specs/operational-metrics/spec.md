# Spec: Operational Metrics

## ADDED Requirements

### Requirement: Metrics endpoint exposes latency percentiles
The system SHALL expose GET /metrics with latency percentiles (p50, p95, p99) per provider and endpoint, in Prometheus format.

#### Scenario: /metrics returns p95 latency for OpenAI
- **WHEN** client calls GET /metrics
- **THEN** response includes "gateway_latency_p95_ms{provider=\"openai\",endpoint=\"/v1/chat/completions\"} 1234"

### Requirement: Success rate tracking per provider
The system SHALL track success_rate (successful responses / total requests) per provider and expose in /metrics.

#### Scenario: Success rate for OmniRoute at 100%
- **WHEN** OmniRoute has 50 successful, 0 failed requests in current window
- **THEN** /metrics includes "gateway_success_rate{provider=\"omniroute\"} 1.0"

### Requirement: Token usage accounting
The system SHALL track total tokens (input + output) consumed per provider in current operational window and expose in /metrics.

#### Scenario: Token usage for Anthropic
- **WHEN** Anthropic has processed 100K input + 50K output tokens
- **THEN** /metrics includes "gateway_tokens_total{provider=\"anthropic\",type=\"input\"} 100000" and "...type=\"output\"} 50000"

### Requirement: Metrics window is rolling (last N minutes)
The system SHALL calculate metrics over a rolling window (default 5 minutes) to show real-time health.

#### Scenario: New request updates p95
- **WHEN** new request completes in 500ms (faster than previous p95 of 1200ms)
- **THEN** next /metrics call may show updated p95 (metric moves toward 500ms over time as old slow requests fall out of window)
