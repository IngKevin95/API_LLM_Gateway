# Specification: Format Auto-Detection Middleware

## ADDED Requirements

### Requirement: Automatic detection of request format (OpenAI, Anthropic, Responses)
The middleware SHALL inspect incoming request to /v1/chat/completions, /v1/messages, /responses and auto-detect format (OpenAI-compatible, Anthropic Messages, Responses API) without requiring explicit Content-Type or Accept headers, routing to correct handler.

#### Scenario: OpenAI-compatible request auto-detection
- **WHEN** client sends POST /v1/chat/completions with `{messages: [...], model: "gpt-4"}`
- **THEN** middleware detects OpenAI format (presence of messages array, model field)
- **AND** normalizes to internal format
- **AND** routes through OpenAI-compatible handler

#### Scenario: Anthropic request auto-detection
- **WHEN** client sends POST /v1/messages with `{model: "claude", messages: [...]}`
- **THEN** middleware detects Anthropic format (messages array + model)
- **AND** normalizes to internal format
- **AND** routes through Anthropic-compatible handler

#### Scenario: Responses API auto-detection
- **WHEN** client sends POST /responses with `{input: [...], model: "gpt-5", reasoning_effort: "medium"}`
- **THEN** middleware detects Responses API format (presence of input + reasoning_effort)
- **AND** normalizes to internal format with inferred capability

#### Scenario: Ambiguous format (messages + model) defaults to OpenAI
- **WHEN** client sends request with both messages (generic) and model
- **THEN** middleware defaults to OpenAI-compatible interpretation
- **AND** processes as OpenAI format

#### Scenario: Malformed request (missing required fields)
- **WHEN** client sends request missing messages or input
- **THEN** middleware returns 400 Bad Request with schema error
- **AND** does not attempt format detection on invalid payload

### Requirement: Format normalization is lossless
Translation from OpenAI → internal → Anthropic SHALL preserve semantic meaning (e.g. tool definitions, message roles).

#### Scenario: Tool calling round-trip
- **WHEN** client sends OpenAI tool_use format request
- **THEN** middleware normalizes to internal representation
- **AND** Anthropic adapter reformats as Anthropic tool_use_block
- **AND** response tool use results round-trip back to OpenAI format

## Configuration

Format detection heuristics defined in middleware/format_detector.go.
Detection priority: Responses API (most specific) > Anthropic (specific structure) > OpenAI (default).
