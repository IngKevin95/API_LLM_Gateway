# Spec: Router (Modified)

## MODIFIED Requirements

### Requirement: Router.Route() chooses provider by highest score
The system SHALL validate that Router.Route() computes score = f(quality, latency, cost, quota, availability) and selects provider with highest score.

#### Scenario: High-quality provider selected
- **WHEN** Router.Route(capability: "chat") is called with OpenAI (quality=0.95) and Ollama (quality=0.60) available
- **THEN** OpenAI is selected (higher score)

#### Scenario: Fallback to available provider
- **WHEN** primary provider (OpenAI) is unavailable (health check failing)
- **THEN** Router selects next available provider with acceptable score

#### Scenario: Score logged for debugging
- **WHEN** Router.Route() selects provider
- **THEN** structured log includes {"action": "route_selected", "capability": "chat", "selected_provider": "openai", "score": 0.87, "runner_up": "anthropic", "runner_up_score": 0.81}

### Requirement: Router validates cost and latency constraints
The system SHALL ensure selected provider meets max_cost and max_latency constraints (if specified by client).

#### Scenario: Client specifies max latency
- **WHEN** request includes header X-Max-Latency-Ms: 500
- **THEN** Router only considers providers with p95 latency < 500ms
