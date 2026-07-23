# Spec: Structured Logging

## ADDED Requirements

### Requirement: Structured logging in JSON format
The system SHALL emit structured logs in JSON format from all handlers and critical paths, with fixed fields: timestamp, level, request_id, component, action, details.

#### Scenario: OpenAI handler logs request
- **WHEN** POST /v1/chat/completions receives request
- **THEN** handler logs {"timestamp": "ISO8601", "level": "info", "request_id": "UUID", "component": "openai-handler", "action": "received_request", "provider": "openai", "model": "...", "tokens_prompt": N}

### Requirement: Request ID propagation
The system SHALL extract or generate request_id from X-Request-ID header and include in all logs for a single request.

#### Scenario: Request ID from header
- **WHEN** client sends X-Request-ID: abc123
- **THEN** all logs include "request_id": "abc123"

#### Scenario: Request ID generated if missing
- **WHEN** client does not send X-Request-ID
- **THEN** handler generates UUID and propagates through logs

### Requirement: Error logging includes stack trace for debugging
The system SHALL log error stack traces at ERROR level with full context (function, line, error message).

#### Scenario: Handler error logs full stack
- **WHEN** ProcessChat encounters error (e.g., invalid response from provider)
- **THEN** log includes {"level": "error", "error": "...", "stack": "...", "request_id": "..."}
