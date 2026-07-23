# Specification: Parameter Translation for OpenAI

## ADDED Requirements

### Requirement: Translate OpenAI parameters to internal normalized format
When a request arrives with OpenAI-compatible parameters (temperature, top_p, seed, tool_choice, response_format), the Gateway SHALL validate ranges and translate to normalized internal format without error, passing through to backend provider adapters.

#### Scenario: Temperature parameter validation and translation
- **WHEN** client sends `temperature: 1.5` (valid for OpenAI 0-2 range)
- **THEN** middleware validates 0 ≤ temperature ≤ 2
- **AND** translates to internal NormalizedRequest.Parameters["temperature"] = 1.5
- **AND** adapter receives value and maps to backend provider format

#### Scenario: Top_p parameter clamping
- **WHEN** client sends `top_p: 1.2` (exceeds valid range 0-1)
- **THEN** middleware logs warning "top_p 1.2 exceeds [0,1], clamping to 1.0"
- **AND** continues with clamped value

#### Scenario: Seed parameter passthrough
- **WHEN** client sends `seed: 42`
- **THEN** middleware translates to internal format (seed is backend-agnostic integer)
- **AND** adapter forwards to provider that supports reproducibility

#### Scenario: Tool_choice enum validation
- **WHEN** client sends `tool_choice: "auto"` (valid enum: auto|required|<tool_id>)
- **THEN** middleware validates against OpenAI enum spec
- **AND** translates to adapter-specific tool_choice format
- **AND** invalid enum "invalid" returns 400 Bad Request

#### Scenario: Unknown parameters are logged but not rejected
- **WHEN** client sends `temperature: 1.0, custom_param: "value"`
- **THEN** middleware logs "unknown parameter 'custom_param' ignored"
- **AND** continues processing with known parameters only

### Requirement: Response format parameter (JSON mode, etc.)
The system SHALL support `response_format` parameter to request structured output.

#### Scenario: JSON mode request
- **WHEN** client sends `response_format: {"type": "json_object"}`
- **THEN** middleware stores format requirement
- **AND** adapter enforces JSON schema validation on response

## Configuration

Parameter specs and ranges defined in config/openai-parameters.yaml.
Mapping tables for providers defined in internal/adapter/parameter_mapper.go.
