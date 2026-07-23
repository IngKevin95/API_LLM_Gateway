# Spec: Anthropic Handler (Modified)

## MODIFIED Requirements

### Requirement: ProcessMessages returns HTTP 200 with valid response
The system SHALL fix ProcessMessages() to return HTTP 200 status (instead of 400) with content block populated from Anthropic response.

#### Scenario: Valid messages request returns 200
- **WHEN** ProcessMessages receives valid request (messages array + model)
- **THEN** response is HTTP 200 with {"content": [{"type": "text", "text": "..."}], "usage": {...}}

#### Scenario: Invalid request returns 400
- **WHEN** ProcessMessages receives malformed request (missing messages, invalid schema)
- **THEN** response is HTTP 400 with error message

#### Scenario: System prompt is preserved
- **WHEN** ProcessMessages receives system prompt in request
- **THEN** system prompt is passed to Anthropic unchanged

### Requirement: Anthropic handler emits structured logging
The system SHALL emit JSON logs from Anthropic handler with request_id, model, provider, latency.

#### Scenario: Successful request logs metrics
- **WHEN** ProcessMessages completes successfully
- **THEN** log includes {"level": "info", "request_id": "...", "provider": "anthropic", "latency_ms": N, "tokens_input": N, "tokens_output": N}
