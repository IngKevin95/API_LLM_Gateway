# Specification: Automatic Routing by Capability

## ADDED Requirements

### Requirement: Router resolves capability without explicit model name
The Gateway SHALL accept requests with `model: "router:capability"` or without `model` field, and automatically resolve the best available model of that capability from the provider chain, without exposing model selection to the client.

#### Scenario: Automatic routing with router: prefix
- **WHEN** client sends POST /v1/chat/completions with `model: "router:coding"`
- **THEN** Router.Resolve("coding") picks the best coder from available models (ranked by score: quality, speed, availability, quota)
- **AND** request continues to selected adapter without client intervention

#### Scenario: Automatic routing with missing model field
- **WHEN** client sends POST /v1/chat/completions with messages but no `model` field
- **THEN** Router infers capability from context (default: "chat") and picks best model automatically
- **AND** request is processed end-to-end

#### Scenario: Fallback chain on first provider unavailable
- **WHEN** best coder (OpenAI gpt-4) is unavailable (429/500) and fallback chain exists
- **THEN** Router picks next ranked coder (e.g. Anthropic Claude) transparently
- **AND** client sees successful response without knowing fallback occurred

#### Scenario: Capability not found
- **WHEN** client sends `model: "router:unknown-capability"`
- **THEN** Gateway returns 400 Bad Request with message "capability 'unknown-capability' not found"
- **AND** lists available capabilities in error response

### Requirement: Backward compatibility with explicit model names
Requests with explicit `model: "gpt-4"` SHALL continue to work as before without triggering automatic routing.

#### Scenario: Explicit model bypasses routing
- **WHEN** client sends `model: "gpt-4"` (no router: prefix)
- **THEN** Gateway uses that model exactly, skipping routing heuristic
- **AND** failover still applies if that model fails

## Configuration

Capability chains defined in config.yaml:
```yaml
capabilities:
  chat:
    - provider: openai
      models: [gpt-4o, gpt-4-turbo]
  coding:
    - provider: openai
      models: [gpt-4o]
    - provider: anthropic
      models: [claude-opus]
```

Scoring function remains EP-001 Router.Score(quality, speed, availability, quota, cost, latency).
