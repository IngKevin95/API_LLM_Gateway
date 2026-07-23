# Spec: OpenAI Handler (Modified)

## MODIFIED Requirements

### Requirement: ProcessChat returns HTTP 200 with valid response
The system SHALL fix ProcessChat() to return HTTP 200 status (instead of 500) with choice.content populated from provider response.

#### Scenario: Valid chat request returns 200
- **WHEN** ProcessChat receives valid request and provider succeeds
- **THEN** response is HTTP 200 with {"choices": [{"content": "..."}], "usage": {...}}

#### Scenario: Invalid request returns 400
- **WHEN** ProcessChat receives malformed request (missing messages, invalid schema)
- **THEN** response is HTTP 400 with error message

#### Scenario: Provider timeout returns 504
- **WHEN** ProcessChat exceeds TTFT (2s for chat)
- **THEN** response is HTTP 504 Gateway Timeout

### Requirement: OpenAI handler emits structured logging
The system SHALL emit JSON logs from OpenAI handler with request_id, model, provider, latency, token usage.

#### Scenario: Successful request logs metrics
- **WHEN** ProcessChat completes successfully
- **THEN** log includes {"level": "info", "request_id": "...", "provider": "openai", "latency_ms": N, "tokens_prompt": N, "tokens_completion": N}
