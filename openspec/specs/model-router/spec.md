# model-router Specification — Delta

## MODIFIED Requirements

### Requirement: Resolución automática por score (con penalización de cuota baja)
El Model Router SHALL resolver una capacidad solicitada al modelo óptimo por un score determinista de 7 variables (calidad, velocidad, disponibilidad, cuota restante, costo, latencia, penalización por cuota baja), retornando la cadena de fallback ordenada por score descendente.

FROM: Score determinista de 6 variables (sin cuota)

TO: Score ahora incluye penalización cuando `remaining < limit * 0.2`. Si remaining ≤ 20% del límite, score decrece 50% (multiplicación por 0.5).

#### Scenario: Penalización aplicada cuando remaining < 20%
- **WHEN** Groq tiene `limit: 100, remaining: 15` (15%) y OpenAI tiene `limit: 1000, remaining: 500` (50%)
- **THEN** Groq's score se multiplica por 0.5 (penalización), OpenAI sin penalización; OpenAI se elige primero

#### Scenario: Sin penalización si remaining > 20%
- **WHEN** Groq tiene `limit: 100, remaining: 25` (25%)
- **THEN** Groq sin penalización; score se calcula normalmente

#### Scenario: Proveedor exhausto excluido
- **WHEN** Cerebras tiene `remaining: 0` (0%)
- **THEN** score se reduce significativamente; Router elige otro proveedor sano primero

#### Scenario: Competencia entre penalizados — mejor latencia gana
- **WHEN** 3 proveedores están todos <20% cuota pero Mistral tiene latencia más baja
- **THEN** Router elige Mistral entre los penalizados

### Requirement: Failover respeta penalización
El Model Router SHALL generar cadena de fallback ordenada por score + penalización. Si proveedor 1 falla, failover intenta proveedor 2 (mejor score), pero puede caer a penalizados si necesario.

FROM: Failover por score sin considerar penalización de cuota

TO: Failover chain respeta penalización: intenta primero no-penalizados, luego penalizados, como fallback

#### Scenario: Failover prefiere no-penalizados
- **WHEN** proveedor 1 (no-penalizado) falla
- **THEN** failover intenta proveedor 2 (no-penalizado) antes que proveedor 3 (penalizado por cuota baja)
