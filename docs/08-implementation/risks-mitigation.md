# Risk & Mitigation Matrix

Identificación de riesgos técnicos Fase 1 y estrategias de mitigación.

## ✅ Critical Risks (Fase 1)

| Risk | Impact | Likelihood | Mitigation | Owner | Deadline |
|---|---|---|---|---|---|
| **Provider API Outage (OpenAI/Anthropic)** | 🔴 HIGH | 🟡 MEDIUM | Failover chain (HU-014), 3+ providers active, 30s switchover | Router | Week 2 |
| **Auth Token Expiry Mid-Request** | 🔴 HIGH | 🟡 MEDIUM | Token refresh logic (HU-031), 5min TTL buffer | Auth | Week 1 |
| **WAL Corruption on Crash** | 🔴 HIGH | 🟢 LOW | fsync + checksum validation (HU-039), recovery test | Sync Worker | Week 3 |
| **Quota Manager Race Condition** | 🔴 HIGH | 🟡 MEDIUM | Atomic DB transactions (HU-013), isolation level SERIALIZABLE | Quota | Week 2 |
| **PII Leak in Logs** | 🔴 HIGH | 🟡 MEDIUM | Redactor (HU-021) + audit (HU-010), 100% coverage test | Audit | Week 1 |

## ⚠️ High Risks (Fase 1)

| Risk | Impact | Likelihood | Mitigation | Owner |
|---|---|---|---|---|
| **Router Scoring Timeout** | 🟠 MEDIUM | 🟡 MEDIUM | Circuit breaker (HU-012), <10ms timeout, fallback to simple score | Router |
| **Stream Slowloris Attack** | 🟠 MEDIUM | 🟡 MEDIUM | Read/write timeouts (HU-034), 5s idle cutoff | Network |
| **KMS Service Unavailable** | 🟠 MEDIUM | 🟢 LOW | Offline DEK cache, 24h rotation, fallback to local key | Encryption |
| **Graceful Shutdown Timeout** | 🟠 MEDIUM | 🟢 LOW | SIGTERM handler (HU-040), 30s drain, force exit on timeout | Infra |
| **Database Replication Lag** | 🟠 MEDIUM | 🟡 MEDIUM | Async batch every 5s (HU-038), accept <5s lag for audit | Sync Worker |

## 🟡 Medium Risks (Fase 1)

| Risk | Impact | Likelihood | Mitigation | Owner |
|---|---|---|---|---|
| **Scoring Model Staleness** | 🟡 LOW | 🟡 MEDIUM | Health Monitor (HU-009) every 60s, refresh on error | Health |
| **Rate Limit Header Parsing** | 🟡 LOW | 🟢 LOW | Strict parsing (HU-012), default to conservative limit | Rate Limiter |
| **Request Validation Bypass** | 🟡 LOW | 🟢 LOW | Multi-stage validation (HU-025): schema + size + auth | Validator |
| **Adapter Protocol Mismatch** | 🟡 LOW | 🟢 LOW | Interface-based design (HU-018/024), test per adapter | Adapters |

## ✅ Detection & Testing

### Per-Risk Test Strategy

**Provider API Outage (Failover)**
```go
// Test failover chain: OpenAI 503 → Claude 3.5 → local Ollama
func TestFailoverChain(t *testing.T) {
  mockOpenAI := mockProvider("503 Service Unavailable")
  mockClaude := mockProvider("200 OK")
  
  result, err := router.Route("coding", "test prompt")
  if err != nil { t.Fatalf("Failover failed: %v", err) }
  if result.Model != "claude-3.5-sonnet" { t.Errorf("Wrong model: %s", result.Model) }
}
```

**Auth Token Expiry**
```go
// Test token refresh during long-lived request
func TestTokenRefreshMidRequest(t *testing.T) {
  oldToken := genToken(1*time.Second) // Expire in 1s
  // Start request at T=0.5s
  // Token expires at T=1.5s
  // Check refresh happens automatically
  result, err := makeRequest(oldToken, 2*time.Second)
  if err != nil { t.Fatalf("Should auto-refresh: %v", err) }
}
```

**WAL Corruption**
```go
// Test recovery from corrupted WAL
func TestWALRecovery(t *testing.T) {
  wal.Write(event1)
  wal.Write(event2)
  // Simulate crash: truncate last N bytes
  corruptWAL(10)
  // On restart: recover valid events, discard corrupted tail
  events, err := wal.Recover()
  if len(events) != 2 { t.Errorf("Expected 2 events, got %d", len(events)) }
}
```

**Quota Manager Race**
```go
// Concurrent deduction test
func TestQuotaRace(t *testing.T) {
  quota.Set("user_1", 100)
  var wg sync.WaitGroup
  for i := 0; i < 10; i++ {
    wg.Add(1)
    go func() {
      defer wg.Done()
      quota.Deduct("user_1", 10)
    }()
  }
  wg.Wait()
  if quota.Get("user_1") != 0 { t.Fatal("Race condition detected") }
}
```

**PII Redaction Coverage**
```go
// Comprehensive PII test
func TestPIIRedaction(t *testing.T) {
  cases := []struct { input, expected string }{
    {"SSN: 123-45-6789", "SSN: [SSN]"},
    {"Email: john@example.com", "Email: [EMAIL]"},
    {"Phone: +1-555-123-4567", "Phone: [PHONE]"},
    {"Credit card: 4111-1111-1111-1111", "Credit card: [CARD]"},
  }
  for _, tc := range cases {
    if got := redact(tc.input); got != tc.expected {
      t.Errorf("Expected %q, got %q", tc.expected, got)
    }
  }
}
```

## ✅ Monitoring & Alerting

### Prometheus Metrics per Risk

| Risk | Metric | Threshold | Alert |
|---|---|---|---|
| Provider outage | provider_errors_total | > 10/min | CRITICAL |
| Token expiry | auth_refresh_total | > 100/hour | WARNING |
| WAL corruption | wal_checksum_failures | > 0 | CRITICAL |
| Quota race | quota_deduction_conflicts | > 0 | CRITICAL |
| PII leak | pii_redaction_failures | > 0 | CRITICAL |
| Graceful shutdown timeout | shutdown_timeout_total | > 0 | WARNING |

### CloudWatch Alarms

```
Availability SLO: < 99.9% → PagerDuty
Auth failure rate: > 1% → Slack #infra
Event loss counter: > 0 → PagerDuty CRITICAL
```

## ✅ Mitigation Timeline

**Week 1 (Auth + PII)**
- HU-031: Auth token validation + refresh
- HU-021: PII redactor (regex + NER)
- HU-010: Audit logger setup

**Week 2 (Router + Quota)**
- HU-007: Model router with O(1) lookup
- HU-012: Rate limiter + circuit breaker
- HU-013: Quota manager (atomic deduction)
- HU-014: Failover chain (MVP: 2 providers)

**Week 3 (Persistence)**
- HU-039: Write-Ahead Log (crash recovery)
- HU-038: Sync Worker (async persistence)
- HU-040: Graceful Shutdown handler

**Week 4 (Validation + Hardening)**
- HU-025: Request validator (size, auth, schema)
- HU-009: Health monitor (periodic checks)
- Full test suite (85+ tests)

## ✅ Escalation Path

1. **Development Phase**: Detected in unit/integration tests → fix before merge
2. **Staging**: Detected in load/chaos tests → hotfix in 24h
3. **Production**: Detected via monitoring → incident response playbook (< 15min MTTR target)

## ✅ Known Limitations (Fase 1)

| Limitation | Rationale | Fase 2+ Solution |
|---|---|---|
| No Learning Engine (HU-011) | Scoring weights fixed | Dynamic weight adjustment based on histórico |
| Single DB replica | High availability | Multi-region replication + failover |
| No cache layer | Redundant calls to LLMs | Redis cache + invalidation (HU-041) |
| Local KMS fallback only | Offline resilience | AWS KMS + EKS-managed encryption |
| Manual quota reset | No automatic recovery | Webhook-based invalidation (HU-041) |

## References

- ADR-001: `docs/12-adr/ADR-001-backend-stack.md` (Go selection + tradeoffs)
- Tech PRD: `docs/13-tech-prd/api-llm-gateway.md` (SLA targets)
- Flows: `docs/06-flows/EP-*.md` (failure scenarios per flow)
- SLA/NFR: `docs/07-trazabilidad/sla-nfr-validation.md` (test strategies)
