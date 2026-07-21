# EP-004A: Flujo de Auditoría, Compliance & Seguridad

Trazabilidad de requests, redacción de PII, encriptación de datos sensibles.

```mermaid
graph TD
    A["🔐 Request<br/>POST /v1/chat/completions"] -->|metadata + payload| B["Auth & AuthZ<br/>(HU-031)<br/>JWT / OAuth2"]
    B -->|token válido| C["PII Redactor<br/>(HU-021)<br/>regex + NER"]
    B -->|token inválido| Z1["❌ 401 Unauthorized"]
    C -->|redact: SSN, email, etc| D["Encryption Layer<br/>(HU-020)<br/>DEK + KMS"]
    D -->|encript payload| E["Audit Logger<br/>(HU-010)<br/>async event stream"]
    E -->|register: user_id, model,<br/>tokens_in, latency,<br/>cost, model_choice_reason| F["Local WAL<br/>(HU-039)<br/>crash recovery"]
    F -->|eventos en disco| G["Quota Deduction<br/>(HU-013)<br/>atomic update"]
    G -->|ok| H["Request → Router<br/>(EP-001)"]
    H -->|response| I["Decrypt Output<br/>(HU-020)<br/>DEK unwrap"]
    I -->|plaintext| J["Log Response<br/>(HU-010)<br/>modelo_actual, tokens_out"]
    J -->|✅ Response to Agent"]
    G -->|QUOTA_EXCEEDED| K["Rate Limit Response<br/>429 Too Many Requests"]
```

## Historias Críticas

| Historia | Fase | Rol |
|----------|------|-----|
| HU-010 | 1 | Audit Logger (persistencia eventos) |
| HU-020 | 1 | Envelope Encryption (DEK/KEK) |
| HU-021 | 1 | PII Redactor (antes de log) |
| HU-031 | 1 | Auth & AuthZ (JWT validation) |
| HU-013 | 1 | Quota Manager (deduction + enforcement) |
| HU-039 | 1 | Write-Ahead Log (crash recovery) |

## Compliance Requirements

- **GDPR**: PII redactado en logs + derecho al olvido implementado (HU-030)
- **SOC2**: Audit trail inmutable en WAL local + DB async (RTO < 1h, RPO < 15min)
- **Auth latency**: < 5ms p99 (O(1) JWT lookup, sin I/O en path crítico)

## Flujo de Redacción (HU-021)

```
Input: "My SSN is 123-45-6789 and email john@example.com"
Redactors:
  - SSN regex: 123-45-6789 → [SSN]
  - Email NER: john@example.com → [EMAIL]
Output: "My SSN is [SSN] and email [EMAIL]"
Logged to WAL (redacted) + audit table
```

## Encriptación (HU-020)

```
DEK (Data Encryption Key): AES-256 (rotated cada 90 días)
KEK (Key Encryption Key): KMS remote (offline resilience)

On encrypt:
  1. Generar DEK (o usar cached)
  2. Encriptar payload con DEK
  3. Wrappear DEK con KEK (KMS)
  4. Guardar (DEK_wrapped, payload_encripted)

On decrypt:
  1. Desencriptar DEK_wrapped via KMS
  2. Desencriptar payload con DEK
  3. Return plaintext to agent
```

## SLA Asociado
- **Auth latency**: < 5ms p99
- **Audit logging**: < 1ms latency (async, no blocking)
- **PII redaction**: 100% coverage (regex + NER)
- **Audit trail**: 30-day retention, immutable (append-only WAL)
