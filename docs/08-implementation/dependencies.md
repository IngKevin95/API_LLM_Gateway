# Matriz de Dependencias Fase 1

Versiones pinned y justificaciones para Go 1.21+ backend.

## Go Modules Core

| Módulo | Versión | Razón | Usado Por |
|---|---|---|---|
| `github.com/gin-gonic/gin` | v1.9.1+ | HTTP router (REST/streaming), performance | Registry, Router (HTTP handlers) |
| `github.com/golang-jwt/jwt/v5` | v5.0.0+ | JWT validation (O(1), zero crypto overhead) | Auth & AuthZ (HU-031) |
| `github.com/google/uuid` | v1.3.0+ | Request ID generation (distributed tracing) | Audit Logger (HU-010) |
| `go.uber.org/zap` | v1.26.0+ | Structured logging (async, minimal latency) | Audit Logger (HU-010) |
| `github.com/lib/pq` | v1.10.9+ | PostgreSQL driver (native, minimal overhead) | Sync Worker (HU-038), DB persistence |
| `golang.org/x/crypto` | v0.14.0+ | AES-256 encryption (DEK operations) | Envelope Encryption (HU-020) |
| `github.com/go-jose/go-jose/v4` | v4.0.1+ | JWE handling (KMS wrapper) | Envelope Encryption (HU-020) |

## LLM SDK Adapters

| SDK | Versión | Adaptado Por | Endpoint | Auth |
|---|---|---|---|---|
| `github.com/openai/go-openai` | v1.18.0+ | OpenAI Adapter (HU-018) | api.openai.com/v1 | OPENAI_API_KEY |
| `github.com/anthropics/sdk-go` | v0.1.0+ | Anthropic Adapter | api.anthropic.com | ANTHROPIC_API_KEY |
| `cloud.google.com/go/aiplatform` | v1.53.0+ | Google Adapter (Vertex AI) | aiplatform.googleapis.com | GOOGLE_APPLICATION_CREDENTIALS |
| (Custom) `ollama/go-client` | v0.1.0+ | Local Adapter (HU-024) | localhost:11434 | None (local) |
| (Custom) `vllm-go` | v0.1.0+ | vLLM Adapter (HU-024) | localhost:8000 | None (local) |

## Async & Concurrency

| Módulo | Versión | Razón | Usado Por |
|---|---|---|---|
| `golang.org/x/sync` | v0.4.0+ | Semaphores, error groups (goroutine coordination) | Sync Worker (HU-038), Graceful Shutdown |
| `context` (stdlib) | Go 1.21 | Context cancellation, timeouts | All components |

## Storage & Persistence

| Componente | Versión | Razón | Usado Por |
|---|---|---|---|
| PostgreSQL | 13.0+ | Primary audit store (ACID, WAL recovery) | Sync Worker (HU-038) |

## Testing & Quality

| Módulo | Versión | Razón | Usado Por |
|---|---|---|---|
| `testing` (stdlib) | Go 1.21 | Unit tests (no framework overhead) | All tests |
| `github.com/stretchr/testify` | v1.8.4+ | Assertions + mock helpers | Integration tests |
| `github.com/golang/mock` | v1.6.0+ | Interface mocking | Component mocking |
| `go.uber.org/goleak` | v1.2.1+ | Goroutine leak detection | Graceful Shutdown tests |

## Monitoring & Tracing

| Módulo | Versión | Razón | Usado Por |
|---|---|---|---|
| `github.com/prometheus/client_golang` | v1.17.0+ | Prometheus metrics | Health Monitor (HU-009) |
| `go.opentelemetry.io/otel` | v1.21.0+ | Distributed tracing | Request Validator (HU-025) |

## Encryption & Security

| Módulo | Versión | Razón | Usado Por |
|---|---|---|---|
| `golang.org/x/oauth2` | v0.13.0+ | OAuth2 token exchange | Auth & AuthZ (HU-031) |

## Build & Deployment

| Tool | Versión | Razón | Ubicación |
|---|---|---|---|
| Go compiler | 1.21+ | Static binary, goroutines | Dockerfile, .github/workflows |
| Alpine Linux | 3.18+ | Minimal base (musl libc) | Dockerfile |
| Docker | 20.10+ | Container runtime | docker-compose.yml |
| Make | 4.3+ | Build automation | Makefile |
| GitHub Actions | (built-in) | CI/CD pipeline | .github/workflows/*.yml |

## Compatibility Matrix

| Go Version | PostgreSQL | Alpine | Status |
|---|---|---|---|
| 1.21.x | 13.0+ | 3.18+ | ✅ Fase 1 MVP |
| 1.22.x | 13.0+ | 3.19+ | ✅ Future |

## Example go.mod (Fase 1)

```
go 1.21

require (
  github.com/gin-gonic/gin v1.9.1
  github.com/golang-jwt/jwt/v5 v5.0.0
  github.com/lib/pq v1.10.9
  go.uber.org/zap v1.26.0
  github.com/openai/go-openai v1.18.0
  github.com/anthropics/sdk-go v0.1.0
)
```

## Fase 1 Stack Summary

**Minimal, production-ready**:
- Core: Gin, JWT, PostgreSQL
- SDKs: OpenAI, Anthropic, Google + local adapters
- Observability: Prometheus + structured logging
- Testing: stdlib + testify
- Security: stdlib crypto + go-jose
- Total: ~20 direct, ~80 transitive dependencies

**Build**: Single static binary (musl), ~50-60 MB compressed
**Runtime**: Go 1.21+ goroutines (no OS threads)

## References

- ADR-001: `docs/12-adr/ADR-001-backend-stack.md`
- Tech PRD: `docs/13-tech-prd/api-llm-gateway.md`
- Historias: HU-018, HU-024, HU-031
