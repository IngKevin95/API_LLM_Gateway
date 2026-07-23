# Design: MVP Fixes & Completeness

## Architecture

### Handler Flow
```
Request (OpenAI/Anthropic-compat format)
  ↓ [Validation]
  ↓ [Router.Route() with score + logging]
  ↓ [Select adapter (OpenAI/Anthropic/OmniRoute/Google/etc)]
  ↓ [Adapter.NormalizeRequest() → provider schema]
  ↓ [Adapter.Do() → HTTP call + metrics]
  ↓ [Adapter.NormalizeResponse() ← provider schema]
  ↓ [Logging: latency, tokens, success/error]
  ↓ Response (OpenAI/Anthropic-compat format)
```

### New Components

#### OmniRoute Adapter (internal/adapter/omniroute/)
- Implements `adapter.Adapter` interface
- Translates system prompt + tool definitions to OpenAI format
- Handles timeout (TTFT < 2s), error parsing
- Returns `Response` with `Choices` and `Usage` populated

#### Structured Logging
- `slog` JSON logger from stdlib (Go 1.21+)
- Fields: timestamp, level, request_id, component, action, details
- Request ID propagated via `context.Context` through handler → router → adapter
- Error logs include stack trace at ERROR level

#### Metrics Collection (in-memory)
- Rolling window: last 5 minutes of requests
- Per provider: latency (p50/p95/p99), success_rate, token counts
- Endpoint `/metrics` returns Prometheus format
- No external dependencies (math/stats built-in)

### Modified Components

#### ProcessChat (gateway.go)
- **BEFORE**: 500 error, no handling of router output
- **AFTER**: Call router.Route() → get adapter → call adapter.Do() → normalize response → return HTTP 200/error

#### ProcessEmbedding (gateway.go)
- **BEFORE**: 501 "not implemented"
- **AFTER**: Similar to ProcessChat but routes to embedding-capable providers

#### ProcessMessages (gateway.go)
- **BEFORE**: 400 error, invalid mapping
- **AFTER**: Similar to ProcessChat but uses Anthropic schema

#### Router.Route()
- **BEFORE**: Fixed weight heuristic
- **AFTER**: Call health.Monitor for real-time metrics → compute score → log decision

#### Provider Registry (config.yaml ↔ buildAdapters())
- **BEFORE**: Mismatch between config IDs and adapter instantiation
- **AFTER**: Config IDs must match buildAdapters() keys exactly; validation at startup

## Trade-offs

| Decision | Rationale | Cost |
|---|---|---|
| In-memory metrics (no DB) | MVP speed, no external deps | Metrics lost on restart; no historical trends |
| Rolling 5-min window | Real-time health snapshot | Limited historical depth; operators need dashboard for trends |
| Structured logging to stdout | Simplicity, operational readiness | No log aggregation; ops must parse JSON or use log collector |
| OmniRoute as free local adapter | Enables offline testing, reduces API quota usage | Requires local binary running; not for production scale |

## Deployment Notes

1. **OmniRoute**: Docker or binary must be running at configured endpoint
2. **Logging**: JSON output to stdout; use `docker logs --follow` or log collector
3. **Metrics**: `/metrics` endpoint serves plain text every 5 minutes; scrape with Prometheus-compatible tool
4. **Config**: `config.yaml` provider IDs must match keys in buildAdapters() code

## Testing Strategy

- **Unit tests**: Each handler (ProcessChat, ProcessEmbedding, ProcessMessages), Router.Route() score logic, Adapter interface implementations
- **Integration tests**: E2E journey from request through all handlers to response (e.g., `go test ./internal/processor -v`)
- **Smoke tests**: Curl against real handlers (`curl -X POST http://localhost:8080/v1/chat/completions ...`)
