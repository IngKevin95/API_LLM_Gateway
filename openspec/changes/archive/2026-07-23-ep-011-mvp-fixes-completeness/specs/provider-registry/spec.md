# Spec: Provider Registry (Modified)

## MODIFIED Requirements

### Requirement: Provider IDs normalized in config.yaml and buildAdapters()
The system SHALL align provider ID strings between config.yaml and buildAdapters() to prevent mismatch errors.

#### Scenario: google-gemini registered correctly
- **WHEN** config.yaml lists provider "google-gemini" with models ["gemini-2.0-flash"]
- **THEN** buildAdapters() creates GoogleAdapter with ID "google-gemini" (matching config)

#### Scenario: local-ollama registered correctly
- **WHEN** config.yaml lists provider "local-ollama" with endpoint "http://localhost:11434"
- **THEN** buildAdapters() creates OllamaAdapter with ID "local-ollama" (matching config)

#### Scenario: openrouter registered correctly
- **WHEN** config.yaml lists provider "openrouter" with API key
- **THEN** buildAdapters() creates OpenRouterAdapter with ID "openrouter" (matching config)

### Requirement: Capability mapping is declared and validated
The system SHALL validate that each provider in config.yaml declares which capabilities it supports (chat, embeddings, reasoning, vision).

#### Scenario: OmniRoute declares chat capability
- **WHEN** config.yaml lists "omniroute" with capabilities: [chat]
- **THEN** Router can route chat requests to OmniRoute but rejects embeddings requests with "No provider for capability"

#### Scenario: Mismatch detected at startup
- **WHEN** buildAdapters() finds provider in config.yaml but no matching adapter code
- **THEN** startup fails with clear error: "Provider 'unknown-provider' not found in buildAdapters()"
