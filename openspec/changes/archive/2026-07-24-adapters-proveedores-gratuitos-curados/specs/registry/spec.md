## ADDED Requirements

### Requirement: Carga del catálogo free-tier.yaml
El Registry SHALL cargar `config/providers/free-tier.yaml` con los proveedores gratuitos
priorizados (Groq, Cerebras, Mistral, Gemini, Cloudflare AI), registrando cada uno con su spec
completo y dejándolo disponible para el routing. Los proveedores declarados en `free-tier.yaml`
SHALL tener precedencia sobre entradas del mismo proveedor en `config.yaml`. (Traza: HU-EVO-002)

#### Scenario: Registry.Load carga free-tier.yaml
- **WHEN** `Registry.Load()` se ejecuta en boot y existe `config/providers/free-tier.yaml` con 5 proveedores
- **THEN** deserializa cada proveedor, lo registra con su spec completo, y queda disponible para routing

#### Scenario: free-tier.yaml sobrescribe config.yaml
- **WHEN** un proveedor (p. ej. Groq) está declarado tanto en `config.yaml` como en `free-tier.yaml`
- **THEN** la versión de `free-tier.yaml` (con `quota_hint` más realista) es la que queda activa tras la carga

#### Scenario: YAML malformado aborta el arranque
- **WHEN** `free-tier.yaml` tiene sintaxis inválida (p. ej. JSON en lugar de YAML)
- **THEN** `Registry.Load()` retorna `ErrInvalidConfig` y el Gateway no inicia (fail-fast)

#### Scenario: Proveedor sin modelo default excluido del scoring
- **WHEN** un proveedor en `free-tier.yaml` tiene `models: []` vacío
- **THEN** el Router lo excluye automáticamente del scoring de esa capacidad sin crashear

#### Scenario: quota_hint negativo tratado como agotado
- **WHEN** un proveedor en `free-tier.yaml` tiene `quota_hint: -100`
- **THEN** el Quota Manager lo trata como cuota agotada (`remaining = 0`) y el Router lo retira de la selección
