# Deployment & Release Runbook

Guía paso a paso para Fase 1 MVP: deploy, smoke tests, rollback.

## Pre-Release Checklist

### Code Readiness (Day -1)

- All 28 Fase 1 historias merged to main
- All 85+ tests passing (unit, integration, e2e, load)
- Test coverage > 80%
- No security issues (dependabot, SAST, code review)
- Changelog updated with version, date, features
- Docker image built & pushed

### Configuration Readiness

- YAML config template ready (config/gateway.yaml.template)
- Environment variables documented (.env.example)
- Secrets stored in AWS Secrets Manager
- Database migrations ready
- PostgreSQL 13+ instance created & tested

### Infrastructure Readiness

- ECS/K8s cluster provisioned
- Load balancer configured (ALB + auto-scale)
- CloudWatch dashboards created
- PagerDuty integration configured
- Rollback plan documented

---

## Release Steps (Day 0)

### Phase 1: Staging Deployment (08:00 - 09:00)

1. Pull latest main branch
2. Tag release version (v1.0.0)
3. Build Docker image and push to ECR
4. Deploy to staging (ECS/K8s)
5. Wait for rollout to complete (5-10 min)
6. Verify staging health check

**Expected**: HTTP 200 from /v1/health with all components "ok"

### Phase 2: Smoke Tests in Staging (09:00 - 10:00)

- Basic connectivity (/v1/health)
- Provider routing (model selection)
- Auth flow (JWT validation)
- Audit trail (events persisted)
- Rate limiting (quota enforcement)
- Failover (provider outage simulation)

**All pass?** → Proceed to Phase 3

### Phase 3: Load Test (10:00 - 11:00)

- Ramp from 10 RPS → 500 RPS over 10 min
- Validate p50, p95, p99 latencies
- Target: router <100ms p95, auth <5ms p99
- Target: 500 RPS throughput, <0.1% error rate

**Load test passed?** → Proceed to Phase 4

### Phase 4: Production Deployment (11:30 - 12:30)

1. Deploy to production (canary: 10% traffic first)
2. Monitor canary for 10 min (error rate, latency, CPU, memory)
3. If healthy: shift 100% traffic to v1.0.0
4. Run subset of smoke tests in production
5. Announce go-live

**Production deployment complete!**

---

## Monitoring & Alerting (Post-Deploy)

### First 24 Hours

- Monitor SLO dashboard every 30 min
- Alert thresholds:
  - Error rate > 1% → Slack + PagerDuty
  - Latency p95 > 150ms → Slack #gateway-team
  - Event loss > 0 → PagerDuty CRITICAL
  - Availability < 99.9% → PagerDuty CRITICAL

### First Week

- Daily SLO review
- Gather user feedback
- Document any issues / post-mortems

---

## Rollback Plan

### Automatic Rollback

If canary error rate > 5% after 5 min, Kubernetes automatically rolls back to previous stable version.

### Manual Rollback

1. Identify stable version from git tags
2. Rollout previous version via kubectl
3. Verify rollback complete
4. Post-mortem (24h after incident)

**Rollback success criteria**:
- Previous version running
- Traffic normalized (latency <100ms p95)
- Error rate <0.1%
- No event loss (audit_events count increasing)
- All health checks passing

---

## Post-Release (Week 1+)

### Day 1

- Verify all historical events in audit table
- Confirm quota enforcement working
- Test one failover scenario manually

### Week 1

- Collect performance baseline
- Analyze logs for errors/warnings
- Gather internal team feedback
- Schedule post-launch retrospective

### Metrics to Track

| Metric | Target |
|---|---|
| Availability | 99.9% |
| TTFT (chat) | 2.0s p95 |
| Router latency | <100ms p95 |
| Error rate | <0.1% |
| Events persisted | 100% |

---

## FAQ

**Q: What if smoke tests fail?**
A: Do NOT proceed. Investigate in staging, fix code, re-test. Tag hotfix and restart.

**Q: What if load test shows p95 > 100ms?**
A: Investigate router bottleneck. Fix before production or document as limitation.

**Q: How long does full deployment take?**
A: ~90 min (smoke 1h, load 1h, prod deploy 30min including canary)

**Q: Can we skip load testing?**
A: No. Non-negotiable for SLO validation.

**Q: What's the SLO for this release?**
A: 99.9% availability (43.2 min downtime allowed per month)

---

## References

- Deployment: kubernetes/deployment.yaml (canary config)
- Smoke tests: test/smoke/*.sh (automated validation)
- Load test: test/load/load-test.sh
- Alerting: observability/prometheus-rules.yaml
- Runbooks: docs/09-runbooks/*.md (incident response)
