# failover-engine Specification — Delta

## ADDED Requirements

### Requirement: Extract and respect Retry-After header
Failover Engine SHALL extract `Retry-After` header from 429 responses, parse in seconds or RFC1123 date format, and pass to Health Monitor for temporary provider retirement. (HU-EVO-010 AC1, AC2)

#### Scenario: Retry-After in seconds
- **WHEN** adapter returns `Status: 429, Retry-After: 60`
- **THEN** failover extracts 60s, retires provider for 60 seconds, failovers to next

#### Scenario: Retry-After in RFC1123 date
- **WHEN** adapter returns `Status: 429, Retry-After: Wed, 23 Jul 2026 19:00:00 GMT`
- **THEN** failover parses date, calculates delta, retires provider until that time

### Requirement: Default retry delay if Retry-After absent
When 429 lacks `Retry-After`, failover SHALL default to 30 seconds and retire provider. (HU-EVO-010 AC3)

#### Scenario: 429 without Retry-After defaults to 30s
- **WHEN** adapter returns `Status: 429` without `Retry-After` header
- **THEN** failover defaults to 30s retirement and failovers immediately

### Requirement: Abort mid-stream on 429 without failover
If 429 occurs during stream transmission (partial response already sent), failover SHALL abort stream, return error to client, and retire provider. No transparent failover mid-stream. (HU-EVO-010 AC4)

#### Scenario: 429 mid-stream aborts
- **WHEN** stream started successfully but proveedor devuelve 429 mid-chunk
- **THEN** failover aborta socket, retorna error (no retry transparente mid-flight)

## MODIFIED Requirements

### Requirement: Retry condicional (429) con backoff exponencial
El Failover Engine SHALL aplicar failover en lugar de retry ante 429, y Health Monitor aplicará backoff exponencial en retiros consecutivos (30s → 60s → 120s → tope 120s).

FROM: `Retry condicional (429) — el engine aplica failover en lugar de retry`

TO: `429 handling: extract Retry-After (o default 30s), retire proveedor, failover a siguiente. Múltiples 429 consecutivos aplican backoff exponencial: 30s → 60s → 120s → cap 120s.`

#### Scenario: Múltiples 429 consecutivos trigger backoff exponencial
- **WHEN** Cerebras devuelve 5 × 429 en 2 minutos
- **THEN** Health Monitor: 1st 429 → retire 30s, 2nd → 60s, 3rd → 120s, 4th+ → 120s (capped)

#### Scenario: Reset tras tiempo de recuperación
- **WHEN** proveedor retirado por 30s, Health Monitor hace health check y obtiene 200 OK
- **THEN** provider reactivado inmediatamente; backoff counter reset
