# alert-manager Specification

## Purpose
TBD - created by archiving change metrics-quota-alertas-rbac. Update Purpose after archive.
## Requirements
### Requirement: Generación periódica de alertas de cuota baja
El Alert Manager SHALL revisar `quota.Manager.Snapshot()` cada 1 minuto (intervalo configurable) y
generar una alerta en PostgreSQL (`provider_alerts`) por cada proveedor/modelo cuyo `remaining` sea
menor al umbral configurable (default 10% del `limit`), con `severity: "warning"`, y una alerta
`severity: "critical"` cuando `remaining == 0`.
(Traza: HU-EVO-012)

#### Scenario: Alerta warning generada bajo umbral
- **WHEN** Groq tiene `remaining=1200 limit=14400` (8.3%) y el Alert Manager corre
- **THEN** inserta en `provider_alerts` una fila con `severity: "warning"`, `message: "Groq remaining < 10%"`, `alert_time: now`

#### Scenario: No duplica alertas activas
- **WHEN** Groq ya tiene una alerta activa generada hace 5 minutos y el Alert Manager corre de nuevo con la misma condición
- **THEN** no genera una alerta nueva; solo actualiza `updated_at` de la existente

#### Scenario: Alerta critical ante agotamiento total
- **WHEN** Cerebras está agotado (`remaining=0`) y el Alert Manager procesa
- **THEN** genera una alerta con `severity: "critical"` y `message: "Cerebras EXHAUSTED"`

#### Scenario: Umbral configurable sin redeploy
- **WHEN** el operador configura `GATEWAY_ALERT_THRESHOLD=0.30` en lugar del default 0.10
- **THEN** el Alert Manager respeta el nuevo umbral en la siguiente corrida sin requerir recompilar/redeploy

#### Scenario: Alertas granulares por modelo, no por proveedor
- **WHEN** Mistral tiene 3 modelos y 2 están bajo el umbral
- **THEN** el Alert Manager genera 2 alertas (una por modelo bajo umbral), no una única alerta agregada por proveedor

