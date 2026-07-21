# SLA/NFR Validation Matrix

Vinculación de requisitos no-funcionales a historias, componentes y tests.

## ✅ NFR → Historias → Componentes → Tests

| NFR | Target | Historias Implementadoras | Componentes | Test Strategy | Observabilidad |
|---|---|---|---|---|---|
| **Disponibilidad** | ≥99.9% | HU-014 (Failover), HU-009 (Health), HU-040 (Graceful) | Failover, Circuit Breaker, Health Monitor | Synthetic probes (5min), uptime tracking | CloudWatch/Datadog SLO dashboard |
| **Router Latency** | <100ms p95 | HU-007 (Router), HU-008 (Scoring) | Model Router, Scoring Engine | Load test (100 req/s), p95 percentile | Request tracing (OpenTelemetry) |
| **Auth Latency** | <5ms p99 | HU-031 (Auth) | Auth & AuthZ | Unit test (O(1) JWT), e2e auth flow | Auth request metrics |
| **Vision TTFT** | 5.0s strict | HU-025 (Validation), HU-018 (Adapter) | Request Validator, Adapters | E2E vision request timing | Model-specific TTFT histogram |
| **Chat TTFT** | 2.0s strict | HU-007 (Router), HU-008 (Scoring) | Model Router, Scoring | E2E chat request timing | Prompt-type TTFT SLI |
| **Throughput** | 500 RPS | HU-012 (Rate Limiter), HU-013 (Quota) | Rate Limiter, Quota Manager | Load test (ramp 1-500 RPS) | RPS gauge + error rate |
| **PII Redaction** | 100% | HU-021 (Redactor) | PII Redactor | Regex + NER coverage tests | Redaction event log audit |
| **Audit Trail** | Immutable 30d | HU-010 (Logger), HU-039 (WAL) | Audit Logger, WAL | Append-only constraint test, WAL recovery | Event count gauge, WAL size monitor |
| **Encryption** | DEK/KEK | HU-020 (Envelope) | Envelope Encryption, KMS | Encrypt/decrypt round-trip test | Key rotation events |
| **Quota Enforcement** | Atomic | HU-013 (Quota) | Quota Manager | Atomic deduction test (race conditions) | Quota consumption gauge |
| **In-Flight Limit** | 50-500 per provider | HU-012 (Rate Limiter), HU-014 (Failover) | Rate Limiter, Circuit Breaker | Load test + saturation point | In-flight connection gauge |
| **WAL Durability** | fsync ≤1ms | HU-039 (WAL) | Write-Ahead Log | Latency histogram test | WAL flush latency SLI |
| **Graceful Shutdown** | 30s drain + flush | HU-040 (Graceful Shutdown) | Graceful Shutdown, Sync Worker | Integration test (SIGTERM + health checks) | Shutdown event log |
| **Stream Idle Timeout** | 5s std, 10s vision | HU-034 (Slowloris), HU-022 (Streaming) | Request Validator, Adapters | Long-lived stream + idle detection test | Stream idle counter |

## ✅ Fase 1 MVP SLA Coverage

**Critical SLIs for Release**:
- Availability ≥99.9% (HA multi-region)
- Router latency <100ms p95 (no scoring bottleneck)
- Auth latency <5ms p99 (O(1), zero I/O)
- TTFT 2.0s for chat, 5.0s for vision
- Throughput 500 RPS (burst handling)
- PII redaction 100% (regex + NER)
- Audit trail immutable + 30-day retention
- Graceful shutdown 30s (zero event loss)

**Non-Critical (Fase 2)**:
- Learning Engine (HU-011) → dynamic scoring weights
- Health Monitor enhancements (HU-009) → latency histograms
- Cache layer (HU-032) → reduced latency for repeated calls

## ✅ Test Strategy per NFR

### 1. Availability (≥99.9%)

**Unit**: N/A (system-level metric)

**Integration**:
```
Test: Failover chain on 429/500
Setup: Mock OpenAI → 503 error
Expect: Retry Claude 3.5 (next scorer)
Metric: Uptime SLO tracked externally
```

**E2E**:
- Synthetic probes every 5min (CloudWatch)
- Measure successful responses / total requests
- Alert if < 99.9% over rolling 30-day window

### 2. Router Latency (<100ms p95)

**Unit**:
```go
// Scoring algorithm complexity
func TestScoringO1(t *testing.T) {
  models := []Model{{name: "gpt-4"}, {name: "claude"}}
  start := time.Now()
  selected := scorer.Select(models, "coding")
  elapsed := time.Since(start)
  if elapsed > 10*time.Millisecond {
    t.Fatalf("Scoring took %v, expected <10ms", elapsed)
  }
}
```

**Load**:
```bash
# Apache Bench: 100 concurrent, 1000 requests
ab -c 100 -n 1000 -t 30 http://localhost:8080/v1/chat/completions

# Extract p95 latency from percentile distribution
```

### 3. Auth Latency (<5ms p99)

**Unit**:
```go
// JWT validation O(1)
func TestAuthO1(t *testing.T) {
  for i := 0; i < 1000; i++ {
    start := time.Now()
    valid, err := auth.ValidateToken(testToken)
    elapsed := time.Since(start).Microseconds()
    if elapsed > 5000 { // 5ms in microseconds
      t.Fatalf("Auth took %vμs at iteration %d", elapsed, i)
    }
  }
}
```

### 4. TTFT (2.0s chat, 5.0s vision)

**E2E**:
```bash
# Time from request to first token from LLM
time curl -N -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $KEY" \
  -d @chat-request.json | head -c 1 # First byte = time to first token
```

**Target**: 2.0s p95 for chat-40k, 5.0s p95 for vision-4v

### 5. PII Redaction (100%)

**Unit**:
```go
func TestPIIRedaction(t *testing.T) {
  cases := []struct {
    input    string
    expected string
  }{
    {"My SSN is 123-45-6789", "My SSN is [SSN]"},
    {"Email: john@example.com", "Email: [EMAIL]"},
  }
  for _, tc := range cases {
    result := redactor.Redact(tc.input)
    if result != tc.expected {
      t.Errorf("Expected %q, got %q", tc.expected, result)
    }
  }
}
```

**NER Coverage**: Test against typical customer data (names, addresses, phone numbers)

### 6. WAL Durability (fsync ≤1ms)

**Unit**:
```go
func TestWALLatency(t *testing.T) {
  for i := 0; i < 10000; i++ {
    event := Event{UserID: fmt.Sprintf("user_%d", i)}
    start := time.Now()
    err := wal.Write(event) // fsync inside
    elapsed := time.Since(start)
    if elapsed > 1*time.Millisecond {
      t.Fatalf("WAL write took %v at event %d", elapsed, i)
    }
  }
}
```

### 7. Graceful Shutdown (30s drain + flush)

**Integration**:
```bash
# 1. Start Gateway with in-flight requests
# 2. Send SIGTERM
# 3. Measure time to exit
# 4. Verify all events flushed to DB

bash -c '
  ./gateway &
  PID=$!
  sleep 2
  kill -TERM $PID
  wait $PID
  EXIT_CODE=$?
  if [ $EXIT_CODE -ne 0 ]; then echo "Ungraceful exit: $EXIT_CODE"; fi
'
```

## ✅ SLO Error Budget (Fase 1)

Assuming 99.9% SLO over 30 days:

```
Total minutes in 30 days: 30 × 24 × 60 = 43,200 min
99.9% uptime: 43,200 × 0.999 = 43,156.8 min available
Downtime allowed: 43,200 - 43,156.8 = 43.2 min / month

At current release cadence (weekly):
- 43.2 min ÷ 4 weeks ≈ 10.8 min downtime allowance per week
- Use sparingly: major incidents, emergency patches only
```

## ✅ Monitoring & Alerting

### Prometheus Metrics (per component)

| Component | Metric | Threshold | Alert |
|---|---|---|---|
| Router | request_latency_ms | p95 > 100 | CRITICAL |
| Auth | auth_latency_us | p99 > 5000 | CRITICAL |
| Scoring | scoring_latency_ms | p95 > 10 | WARNING |
| WAL | wal_write_latency_ms | p95 > 1 | WARNING |
| Quota | quota_depletion_percent | > 90% | WARNING |
| Audit Logger | log_events_dropped | > 0 | CRITICAL |

### CloudWatch Custom Dashboards

- TTFT by model type (histogram)
- Availability timeline (99.9% SLO band)
- P95/P99 latencies per component
- Event loss counter (audit trail)
- In-flight connection gauge (per provider)

## ✅ Fase 1 Test Coverage Target

| Category | Target | Estimated Test Count |
|---|---|---|
| Unit (NFR-specific) | >80% | 40+ tests |
| Integration (component ↔ component) | >70% | 30+ tests |
| E2E (SLA validation) | 100% | 10+ scenarios |
| Load (throughput, saturation) | >90% | 5+ profiles |

**Total**: ~85 tests for NFR validation before release

## References

- PRD: `docs/01-prd/api-llm-gateway.md` (L164-170: formal NFRs)
- Tech PRD: `docs/13-tech-prd/api-llm-gateway.md` (L11, L30: SLA specs)
- Architecture: `docs/11-architecture/api-llm-gateway.md` (L167-184: component latencies)
- Historias: `docs/04-historias/HU-*.md` (AC with test criteria)
- Flows: `docs/06-flows/EP-*.md` (SLA noted per flow)
