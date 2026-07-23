# Specification: Client Setup Guides

## ADDED Requirements

### Requirement: 8 reproducible client configuration guides (one per tool)
The Gateway SHALL provide comprehensive setup guides for each of 8 supported client tools, including exact environment variables, configuration snippets, curl/Python examples, and troubleshooting.

#### Scenario: OpenWebUI setup guide exists and is complete
- **WHEN** user reads docs/14-guides/openwebui-gateway-setup.md
- **THEN** guide includes: installation steps, OPENAI_BASE_URL config, OPENAI_API_KEY setup, working curl example, troubleshooting section
- **AND** curl example returns successful response from running Gateway

#### Scenario: OpenCode setup guide for Responses API
- **WHEN** user reads docs/14-guides/opencode-gateway-setup.md
- **THEN** guide explains /responses endpoint, OPENCODE_BASE_URL, reasoning_effort usage
- **AND** includes Python example using Responses API format
- **AND** documents fallback behavior

#### Scenario: Free Claude Code setup guide
- **WHEN** user reads docs/14-guides/free-claude-code-gateway-setup.md
- **THEN** guide explains ANTHROPIC_BASE_URL setup, API key from free provider (e.g. AIHubMix)
- **AND** clarifies differences from official Anthropic (rate limits, model availability)
- **AND** includes verification steps

#### Scenario: Master guide with client matrix
- **WHEN** user reads docs/14-guides/GATEWAY_CLIENTS.md
- **THEN** document contains comparison table: Client | Endpoint | Format | Auto-Detect | Parameters
- **AND** links to individual setup guides
- **AND** explains routing capability system

#### Scenario: Each guide is executable
- **WHEN** user runs curl examples from any guide against running Gateway
- **THEN** examples return successful responses
- **AND** demonstrates parameter passing (temperature, tool_choice, etc.) where applicable

### Requirement: Setup guides cover 8 tools: OpenWebUI, OpenCode, Free Claude Code, Claude Code, OpenHands, OpenClaw, CrewAI, UI-TARS
Each tool gets a dedicated markdown file with tool-specific setup.

#### Scenario: All 8 guides exist
- **WHEN** user lists docs/14-guides/
- **THEN** 8 guide files present: openwebui-, opencode-, free-claude-code-, claude-code-, openhands-, openclaw-, crewai-, ui-tars-gateway-setup.md
- **AND** all follow consistent template structure

#### Scenario: Multi-language examples (curl, Python, JS)
- **WHEN** user reads a setup guide
- **THEN** examples provided in at least 2 languages (curl + Python recommended)
- **AND** language-specific setup (env vars for shell, client library usage for Python)

## Configuration

Guides stored in docs/14-guides/ directory.
Consistent template file at docs/14-guides/_TEMPLATE.md.
All examples tested against running Gateway before publication.
