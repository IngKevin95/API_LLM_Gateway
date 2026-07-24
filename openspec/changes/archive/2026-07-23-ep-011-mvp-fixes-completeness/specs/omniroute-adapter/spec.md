# Spec: OmniRoute Adapter

## ADDED Requirements

### Requirement: OmniRoute adapter implements Adapter interface
The system SHALL implement an adapter for OmniRoute (local, free, OpenAI-compatible) that conforms to the `adapter.Adapter` interface.

#### Scenario: Adapter registered at startup
- **WHEN** Gateway starts with OmniRoute in config.yaml
- **THEN** adapter is instantiated and available in router

### Requirement: OmniRoute schema translation
The system SHALL translate gateway-normalized request/response to OmniRoute OpenAI-compatible format.

#### Scenario: Chat request translates correctly
- **WHEN** ProcessChat sends request to OmniRoute
- **THEN** System Prompt and Tool Definitions are mapped to OpenAI format

#### Scenario: Chat response translates back
- **WHEN** OmniRoute returns OpenAI-compat response
- **THEN** Gateway response includes choice.content and usage

### Requirement: OmniRoute error handling
The system SHALL handle OmniRoute connection errors (timeout, 500, network unavailable) and report as ProviderError with proper retry signal.

#### Scenario: OmniRoute timeout triggers failover
- **WHEN** OmniRoute does not respond within TTFT (2s for chat)
- **THEN** Gateway reports ErrTTFT and router moves to next provider
