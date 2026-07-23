# Tasks: MVP Fixes & Completeness

## Phase 1: Handler Fixes (HU-050, HU-051, HU-052, HU-057, HU-058)

### HU-050: Logging to OpenAI handler
- [ ] Add slog JSON logger initialization in cmd/gateway/main.go
- [ ] Pass logger to ProcessChat handler
- [ ] Emit log on receipt: `{"level": "info", "action": "request_received", "component": "openai-handler", "model": "...", "tokens_prompt": N}`
- [ ] Emit log on success: `{"level": "info", "action": "request_completed", "latency_ms": N, "tokens_completion": M}`
- [ ] Emit log on error: `{"level": "error", "error": "...", "stack": "..."}`
- [ ] Test: `go test ./internal/processor -v -run TestProcessChat_Logging`

### HU-051: Fix ProcessChat()
- [ ] Read gateway.Processor.ProcessChat() current code
- [ ] Identify why it returns 500 (missing router call? nil adapter? response mapping broken?)
- [ ] Add router.Route(capability: "chat") call with score logic
- [ ] Call adapter.Do() and handle response
- [ ] Map provider response back to OpenAI schema (choices + usage)
- [ ] Return HTTP 200 with valid JSON
- [ ] Test: `go test ./internal/processor -v -run TestProcessChat` (happy path + error cases)

### HU-052: Validate Router.Route()
- [ ] Read router.Route() score computation
- [ ] Add unit test: Route returns highest-score provider
- [ ] Add unit test: Route respects max_latency constraint
- [ ] Add unit test: Route falls back to secondary provider if primary unavailable
- [ ] Add integration test: E2E request through router + adapter
- [ ] Test: `go test ./internal/router -v`

### HU-057: Implement ProcessEmbedding()
- [ ] Similar to ProcessChat but route to embedding-capable providers only
- [ ] Map request (input text + model) to provider schema
- [ ] Call adapter.Do()
- [ ] Map response (embedding vectors + usage) back to OpenAI schema
- [ ] Return HTTP 200 with valid embeddings array
- [ ] Test: `go test ./internal/processor -v -run TestProcessEmbedding`

### HU-058: Fix ProcessMessages()
- [ ] Read gateway.Processor.ProcessMessages() current code
- [ ] Identify mapping errors (why 400?)
- [ ] Map request (messages array + system prompt) to Anthropic schema
- [ ] Call adapter.Do()
- [ ] Map response (content blocks + usage) back to Anthropic schema
- [ ] Return HTTP 200 with valid content
- [ ] Test: `go test ./internal/processor -v -run TestProcessMessages`

## Phase 2: OmniRoute Adapter (HU-053, HU-054, HU-055)

### HU-053: Create OmniRoute Adapter
- [ ] Create `internal/adapter/omniroute/omniroute.go`
- [ ] Implement `adapter.Adapter` interface (NormalizeRequest, Do, NormalizeResponse, Capabilities)
- [ ] Handle connection to OmniRoute endpoint (default http://localhost:11434 via config)
- [ ] Translate system prompt + tool definitions to OpenAI-compat format
- [ ] Handle timeout (TTFT 2s for chat)
- [ ] Return proper ProviderError on failure
- [ ] Test: `go test ./internal/adapter/omniroute -v`

### HU-054: Register in buildAdapters()
- [ ] Open cmd/gateway/main.go buildAdapters() function
- [ ] Add OmniRoute instantiation: `adapters["omniroute"] = omniroute.New(cfg)`
- [ ] Verify config.yaml has "omniroute" provider with endpoint + capabilities
- [ ] Test: `go run ./cmd/gateway -config config.yaml` (check startup logs, no panic)

### HU-055: Test OmniRoute connectivity
- [ ] Start OmniRoute: `docker run -p 11434:11434 omniroute:latest`
- [ ] Run Gateway: `go run ./cmd/gateway`
- [ ] Test E2E: `curl -X POST http://localhost:8080/v1/chat/completions -H "Content-Type: application/json" -d '{"model":"...", "messages": [...]}'`
- [ ] Verify response is HTTP 200 with valid choices
- [ ] Test: `./test-endpoints.sh` or similar (verify OmniRoute in provider list)

## Phase 3: Config & Normalization (HU-056)

### HU-056: Align provider IDs
- [ ] Audit config.yaml: list all provider IDs
- [ ] Audit buildAdapters(): list all adapter keys
- [ ] Create mapping table showing mismatches (e.g., "google-gemini" in config vs "google" in buildAdapters)
- [ ] Decide: update config.yaml to match buildAdapters() keys (preferred) OR update buildAdapters() to match config.yaml
- [ ] Apply fix
- [ ] Add startup validation: for each provider in config.yaml, verify adapter exists in buildAdapters()
- [ ] Test: `go run ./cmd/gateway -config config.yaml` (startup must pass validation)

## Phase 4: Observability (HU-059, HU-060)

### HU-059: Structured logging in all handlers
- [ ] Apply logging pattern from HU-050 to ProcessEmbedding (HU-057) and ProcessMessages (HU-058)
- [ ] Ensure request_id is propagated from middleware → handler → router → adapter
- [ ] Log at every boundary: request received, router selected provider, adapter called, response returned
- [ ] Error logs include stack trace
- [ ] Test: `go test ./internal/processor -v -run Logging` (verify JSON output)

### HU-060: Implement /metrics endpoint
- [ ] Create metrics.Collector (tracks latency, success_rate, tokens per provider in rolling 5-min window)
- [ ] Update adapter.Do() to record timing + status after each call
- [ ] Expose GET /metrics handler in cmd/gateway/main.go
- [ ] Format: Prometheus plain text (e.g., `gateway_latency_p95_ms{provider="openai"} 1234`)
- [ ] Test: `curl http://localhost:8080/metrics` (verify p50/p95/p99 for each provider, success_rate, tokens)

## Validation Checklist

- [ ] All 5 handlers return correct HTTP status (200 for success, 4xx/5xx for errors)
- [ ] All handlers emit structured JSON logs
- [ ] Router.Route() selects provider with highest score
- [ ] OmniRoute adapter is registered and connectable
- [ ] Config provider IDs match buildAdapters() keys
- [ ] /metrics endpoint serves valid Prometheus format with real data
- [ ] go build ./... succeeds
- [ ] go vet ./... succeeds
- [ ] go test -race ./... succeeds
- [ ] E2E journey: request → router → adapter → response → logs → metrics all working
