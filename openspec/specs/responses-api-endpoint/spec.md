# Specification: Responses API Endpoint

## ADDED Requirements

### Requirement: New /responses endpoint supporting Responses API format
The Gateway SHALL expose POST `/responses` endpoint that accepts Responses API format requests (with `reasoning_effort`, `input`, etc.) and translates them internally to normalized format for routing and execution.

#### Scenario: Successful Responses API request
- **WHEN** OpenCode sends POST /responses with `{"model": "gpt-5", "input": [...], "reasoning_effort": "medium"}`
- **THEN** Gateway translates to internal NormalizedRequest with capability inferred from reasoning_effort
- **AND** routes to appropriate provider (e.g. OpenAI o1-preview for reasoning)
- **AND** returns response in Responses API format

#### Scenario: Streaming Responses API
- **WHEN** client sends /responses with `stream: true`
- **THEN** Gateway returns Server-Sent Events stream in Responses API format
- **AND** each event contains reasoning token count and delta content

#### Scenario: Responses API without reasoning_effort (defaults to "medium")
- **WHEN** client sends /responses without reasoning_effort field
- **THEN** Gateway defaults to "medium" and routes accordingly
- **AND** continues processing

### Requirement: Content-Type negotiation for Responses API
The /responses endpoint SHALL auto-detect Responses API format without requiring Content-Type header, but MUST validate schema matches.

#### Scenario: Missing required fields in Responses API
- **WHEN** client sends /responses missing required `input` field
- **THEN** Gateway returns 400 Bad Request with schema validation error
- **AND** error message specifies missing field

## Configuration

Responses API format mapper defined in middleware/format_detector.go.
Reasoning effort values: "low", "medium", "high" map to internal reasoning capability chain.
