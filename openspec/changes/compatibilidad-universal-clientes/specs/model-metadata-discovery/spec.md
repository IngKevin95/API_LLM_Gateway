# Specification: Model Metadata Discovery via /v1/models

## ADDED Requirements

### Requirement: Expanded /v1/models endpoint with capability and performance metadata
The Gateway SHALL return enhanced /v1/models response listing all available models with capability, latency, cost, and support status metadata.

#### Scenario: List models with capability
- **WHEN** client sends GET /v1/models
- **THEN** response includes all registered models with `capability` field (chat, coding, reasoning, vision, embedding)
- **AND** each model lists supported parameters (temperature, top_p, tool_use, thinking, etc.)

#### Scenario: Filter models by capability query parameter
- **WHEN** client sends GET /v1/models?capability=coding
- **THEN** response includes only models supporting coding capability
- **AND** preserves full metadata for each matching model

#### Scenario: Model metadata includes performance indicators
- **WHEN** client sends GET /v1/models
- **THEN** each model includes: `latency_p95_ms`, `cost_per_1k_tokens`, `availability_percent`
- **AND** all values are sourced from recent observability metrics (EP-007)

#### Scenario: Model metadata includes parameter support
- **WHEN** client inspects model metadata
- **THEN** field `supported_parameters` lists which parameters this model accepts (e.g. [temperature, top_p, seed])
- **AND** client can validate parameter compatibility before sending request

#### Scenario: Model status indicates availability
- **WHEN** client queries /v1/models
- **THEN** each model includes `status` field (available, degraded, unavailable)
- **AND** status reflects current health check results (EP-002 Health Monitor)

### Requirement: Backward compatibility with OpenAI /v1/models format
The response SHALL include OpenAI-compatible fields (id, object, created, owned_by) plus new extensions without breaking existing clients.

#### Scenario: Legacy client still works
- **WHEN** client sends GET /v1/models and ignores new fields
- **THEN** response still contains id, object, created for backward compat
- **AND** new fields (capability, latency_p95_ms) are additional

## Configuration

Model metadata sourced from:
- Registry (config.yaml) for static fields (name, capability)
- Observability system (EP-007) for dynamic fields (latency, cost, availability)
- Health Monitor (EP-002) for status
