# Implementation Tasks: EP-010 Compatibilidad Universal de Clientes

## 1. Core Middleware Infrastructure

- [x] 1.1 Create middleware/format_detector.go with FormatDetector interface
- [x] 1.2 Implement auto-detection logic for OpenAI/Anthropic/Responses formats
- [x] 1.3 Define NormalizedRequest struct in internal/request/normalized.go
- [x] 1.4 Create middleware/normalizer.go to convert detected format to internal
- [x] 1.5 Integrate middleware into request processing pipeline
- [x] 1.6 Add unit tests for format detection (OpenAI, Anthropic, Responses)
- [x] 1.7 Add unit tests for format normalization round-trip

## 2. Automatic Routing by Capability

- [x] 2.1 Extend router.Router to support "router:capability" model prefix
- [x] 2.2 Implement capability inference from request context (default: "chat")
- [x] 2.3 Add fallback chain resolution when primary provider unavailable
- [x] 2.4 Preserve backward compatibility with explicit model names
- [x] 2.5 Add unit tests for automatic routing scenarios
- [x] 2.6 Add integration tests for fallback behavior

## 3. OpenAI Parameter Translation

- [x] 3.1 Create config/openai-parameters.yaml with parameter specs and ranges
- [x] 3.2 Implement parameter validation (temperature 0-2, top_p 0-1, seed range)
- [x] 3.3 Add parameter logging (unknown params logged as warnings)
- [x] 3.4 Implement tool_choice enum validation
- [x] 3.5 Implement response_format parameter support (JSON mode)
- [x] 3.6 Create parameter mapper in internal/adapter/parameter_mapper.go
- [x] 3.7 Add unit tests for OpenAI parameter translation
- [x] 3.8 Add validation tests for out-of-range parameters

## 4. Anthropic Parameter Translation

- [ ] 4.1 Create config/anthropic-parameters.yaml with parameter specs
- [ ] 4.2 Implement temperature validation (0-1 range)
- [ ] 4.3 Implement top_k parameter handling and translation
- [ ] 4.4 Implement thinking (extended thinking) parameter support
- [ ] 4.5 Implement tool_use parameter validation and routing
- [ ] 4.6 Add max_tokens requirement enforcement (always required for Anthropic)
- [ ] 4.7 Add fallback for clients requesting unsupported features
- [ ] 4.8 Add unit tests for Anthropic parameter translation

## 5. New /responses Endpoint (OpenCode Support)

- [ ] 5.1 Add handler for POST /responses in cmd/gateway/handlers.go
- [ ] 5.2 Implement Responses API format parsing (model, input, reasoning_effort)
- [ ] 5.3 Implement reasoning_effort to capability mapping (low/medium/high)
- [ ] 5.4 Add streaming support for /responses endpoint (Server-Sent Events)
- [ ] 5.5 Add request validation (required fields, schema)
- [ ] 5.6 Add response formatting for Responses API output
- [ ] 5.7 Add unit tests for /responses endpoint
- [ ] 5.8 Add integration tests with OpenCode client simulator

## 6. Enhanced /v1/models Endpoint

- [ ] 6.1 Extend /v1/models handler to include capability field per model
- [ ] 6.2 Add supported_parameters field listing model's accepted params
- [ ] 6.3 Add dynamic metadata fields: latency_p95_ms, cost_per_1k_tokens, availability_percent
- [ ] 6.4 Integrate with EP-007 observability metrics for dynamic fields
- [ ] 6.5 Implement ?capability=<name> query parameter filtering
- [ ] 6.6 Add status field (available/degraded/unavailable) from Health Monitor
- [ ] 6.7 Maintain OpenAI /v1/models backward compatibility
- [ ] 6.8 Add unit tests for metadata endpoint

## 7. Client Setup Documentation

- [ ] 7.1 Create docs/14-guides/_TEMPLATE.md for guide structure
- [ ] 7.2 Write openwebui-gateway-setup.md with complete examples
- [ ] 7.3 Write opencode-gateway-setup.md with Responses API examples
- [ ] 7.4 Write free-claude-code-gateway-setup.md with free provider notes
- [ ] 7.5 Write claude-code-gateway-setup.md with IDE integration notes
- [ ] 7.6 Write openhands-gateway-setup.md with LLM_MODEL config
- [ ] 7.7 Write openclaw-gateway-setup.md with voice integration notes
- [ ] 7.8 Write crewai-gateway-setup.md with Python examples
- [ ] 7.9 Write ui-tars-gateway-setup.md with automation context
- [ ] 7.10 Create GATEWAY_CLIENTS.md with comparison matrix and master index
- [ ] 7.11 Test all curl examples from guides against running Gateway
- [ ] 7.12 Test all Python examples from guides (if applicable)

## 8. Integration & Cross-Adapter Wiring

- [ ] 8.1 Wire middleware into main request pipeline (before router)
- [ ] 8.2 Ensure parameter translation works with all adapters (OpenAI, Anthropic, Google, OpenRouter, AIHubMix, local)
- [ ] 8.3 Add integration tests: OpenAI client → Anthropic backend (round-trip)
- [ ] 8.4 Add integration tests: Anthropic client → OpenAI backend (round-trip)
- [ ] 8.5 Add integration tests: Responses API client → any backend
- [ ] 8.6 Verify backward compatibility: existing EP-005 tests still pass
- [ ] 8.7 Add end-to-end test for all 8 clients with minimal params

## 9. Testing & Verification

- [ ] 9.1 Run full test suite (-race flag): all pass
- [ ] 9.2 Run linter (go vet): all pass
- [ ] 9.3 Build binary: no errors
- [ ] 9.4 Manual smoke test: 8 clients connect and exchange messages
- [ ] 9.5 Verify /health endpoint: returns 200
- [ ] 9.6 Verify /metrics endpoint: includes new metrics
- [ ] 9.7 Test parameter edge cases (clamping, validation, logging)
- [ ] 9.8 Test format detection edge cases (ambiguous requests, malformed)
- [ ] 9.9 Verify logs for warnings/errors on unknown parameters
- [ ] 9.10 Performance: measure middleware overhead (target < 1ms p95)

## 10. Finalization & Archival

- [ ] 10.1 Review all code changes for style, safety, and consistency
- [ ] 10.2 Ensure config.yaml has example entries for all 8 client formats
- [ ] 10.3 Add CHANGELOG entry for EP-010
- [ ] 10.4 Merge feature branch to develop via PR
- [ ] 10.5 Tag release (if applicable)
- [ ] 10.6 Archive OpenSpec change (mark all artifacts as final)
