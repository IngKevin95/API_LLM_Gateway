## ADDED Requirements

### Requirement: Inicialización de cuota desde quota_hint del YAML
El Quota Manager SHALL inicializar `remaining` por proveedor desde el campo `quota_hint` de
`free-tier.yaml` en boot, tratando valores `<= 0` como cuota agotada y usando un default de 1M
tokens cuando el campo está ausente. El valor aprendido en runtime (desde headers de
rate-limit) o restaurado desde persistencia SHALL tener precedencia sobre el `quota_hint` inicial.
(Traza: HU-EVO-005)

#### Scenario: Init remaining desde quota_hint YAML
- **WHEN** Groq en `free-tier.yaml` tiene `quota_hint: 14400` y Quota Manager arranca
- **THEN** `Remaining("groq")` devuelve 14400 sin haber realizado ninguna request

#### Scenario: Header learned sobrescribe quota_hint
- **WHEN** el primer request a Groq devuelve header `X-RateLimit-Remaining: 14300`
- **THEN** Quota Manager aprende ese valor y actualiza `remaining` a 14300, con precedencia sobre el `quota_hint` inicial

#### Scenario: quota_hint <= 0 tratado como agotado
- **WHEN** un proveedor tiene `quota_hint: 0` (o negativo) en `free-tier.yaml`
- **THEN** Quota Manager lo carga como agotado (`remaining = 0`) y el Router lo excluye hasta el primer aprendizaje real

#### Scenario: Proveedor sin quota_hint usa default
- **WHEN** un proveedor nuevo no tiene `quota_hint` definido en el YAML
- **THEN** Quota Manager asume un default de 1M tokens como `remaining` inicial

#### Scenario: Reinicio restaura learned quota desde PostgreSQL
- **WHEN** antes de un reinicio se aprendió que un proveedor tiene 500M `remaining`, y el Gateway reinicia
- **THEN** Quota Manager lee PostgreSQL y restaura 500M como `remaining`, sin volver al `quota_hint` original del YAML
