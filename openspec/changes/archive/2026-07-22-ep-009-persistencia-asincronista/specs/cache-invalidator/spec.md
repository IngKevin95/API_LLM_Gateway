## ADDED Requirements

### Requirement: Cache Invalidator no-op en Fase 1 con fallback a poll/retry
El Cache Invalidator SHALL ser no-op cuando su feature flag está desactivado (Fase 1), aplicando únicamente el patrón poll/retry ante miss para no dejar cuotas stale permanentemente. El worker de polling/webhook completo es Fase 2. (Traza: HU-041 AC4/AC2)

#### Scenario: Deshabilitado en MVP
- **WHEN** el feature flag está en OFF (Fase 1)
- **THEN** el Cache Invalidator es no-op; solo aplica poll/retry ante miss

#### Scenario: Fallback a poll ante fallo de webhook
- **WHEN** un webhook hace timeout o falla de red
- **THEN** se hace fallback a polling en la siguiente iteración, sin dejar cuotas stale permanentemente

### Requirement: Invalidación por cambio detectado (Fase 2)
Cuando el flag está activo (Fase 2), el Cache Invalidator SHALL detectar cambios en DB e invalidar la caché en RAM encolando re-hidratación asíncrona, resolviendo la carrera con escrituras concurrentes vía fail-fast + retry. (Traza: HU-041 AC1/AC3/AC5)

#### Scenario: Detección por polling (Fase 2)
- **WHEN** el flag está activo y un cambio de cuota se detecta comparando timestamps de última sincronización
- **THEN** invalida la caché en RAM y encola la hidratación asíncrona

#### Scenario: Carrera con escritura concurrente (Fase 2)
- **WHEN** el Quota Manager en RAM y una escritura externa ocurren simultáneamente y el invalidador invalida
- **THEN** la siguiente solicitud encola la re-hidratación (fail-fast + retry)
