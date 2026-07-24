# Specification: Parameter Translation for Anthropic

## ADDED Requirements

### Requirement: Translate Anthropic parameters to internal normalized format
When a request arrives with Anthropic-compatible parameters (temperature, top_k, thinking, tool_use), the Gateway SHALL validate ranges and translate to normalized internal format, passing through to backend adapters.

#### Scenario: Temperature parameter validation
- **WHEN** client sends `temperature: 0.5` (Anthropic valid range 0-1)
- **THEN** middleware validates 0 ≤ temperature ≤ 1
- **AND** translates to internal NormalizedRequest.Parameters["temperature"] = 0.5

#### Scenario: Top_k parameter handling
- **WHEN** client sends `top_k: 100` (Anthropic parameter for nucleus sampling)
- **THEN** middleware translates to internal format
- **AND** adapter converts to provider-native format (e.g. top_k for some, top_p for others)

#### Scenario: Extended thinking (thinking block) parameter
- **WHEN** client sends `thinking: {"type": "enabled", "budget": 5000}`
- **THEN** middleware extracts thinking budget and routes to reasoning-capable model (Claude Opus Extended)
- **AND** response includes thinking blocks if model supports

#### Scenario: Tool use parameter validation
- **WHEN** client sends `tool_use: {"enabled": true}` with tools array
- **THEN** middleware validates tool schema matches Anthropic spec
- **AND** forwards to adapter for tool calling setup

#### Scenario: Thinking disabled fallback
- **WHEN** client requests thinking but backend doesn't support (e.g. free tier)
- **THEN** middleware logs warning "thinking not supported by gpt-4o-mini, proceeding without"
- **AND** processes request without thinking blocks

### Requirement: Max tokens and token counting
The system SHALL correctly handle Anthropic's `max_tokens` requirement (always required, unlike OpenAI).

#### Scenario: Max tokens required
- **WHEN** client sends request to Anthropic-compatible endpoint without max_tokens
- **THEN** Gateway returns 400 Bad Request "max_tokens required for Anthropic models"

## Configuration

Anthropic parameter specs defined in config/anthropic-parameters.yaml.
Thinking support per model defined in config.yaml model registry.
