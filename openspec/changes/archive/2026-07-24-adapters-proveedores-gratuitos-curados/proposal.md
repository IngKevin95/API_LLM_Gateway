## Why

El Gateway hoy soporta proveedores mediante adapters Go escritos a mano (uno por proveedor), lo
que hace costoso agregar los proveedores gratuitos priorizados (Groq, Cerebras, Mistral, Gemini,
Cloudflare AI — 1.19B tokens/mes documentados) y no valida automáticamente que cada nuevo
proveedor cumpla el contrato Chat/Stream/Embed, ni retira proveedores que empiezan a devolver 429,
ni arranca con una estimación de cuota realista antes del primer aprendizaje desde headers.

## What Changes

- Adapter genérico data-driven: instancia un cliente `Chat/Stream/Embed` a partir de un
  `ProviderSpec` declarativo (baseURL, authHeader, format, headers extra, timeoutMs) sin código Go
  nuevo por proveedor; valida el spec y falla con `ErrInvalidProviderSpec` si es inválido.
- Registry carga `config/providers/free-tier.yaml` con los 5 proveedores gratuitos priorizados,
  sobrescribiendo entradas duplicadas de `config.yaml`; falla fail-fast (`ErrInvalidConfig`) ante
  YAML malformado; excluye del scoring providers sin modelo default.
- `conformance_test.go` se extiende para iterar todos los `ProviderSpec` registrados y validar
  Chat/Stream/Embed contra el contrato normalizado (Content/Model/Usage), en paralelo y con
  timeout por proveedor.
- Health Monitor detecta HTTP 429 (con o sin `Retry-After`), retira temporalmente al proveedor de
  la selección del Router, reactiva al vencer el retiro, aborta streams mid-stream sin failover
  transparente, y aplica backoff exponencial ante 429 repetidos.
- Quota Manager inicializa `remaining` por proveedor desde `quota_hint` de `free-tier.yaml` (0/negativo
  ⇒ agotado; ausente ⇒ default 1M), permite que el aprendizaje desde headers en runtime sobrescriba
  ese valor inicial, y restaura desde PostgreSQL el último valor aprendido tras un reinicio.

## Capabilities

### New Capabilities
(ninguna — el cambio extiende contratos existentes, no introduce una capacidad nueva)

### Modified Capabilities
- `provider-adapters`: se agrega un adapter genérico data-driven configurado por `ProviderSpec`
  (formatos openai/claude), reemplazando la necesidad de un adapter Go por proveedor; se extiende
  `conformance_test.go` para validar cada spec registrado.
- `registry`: `Registry.Load()` incorpora `config/providers/free-tier.yaml` como fuente adicional,
  con precedencia sobre `config.yaml` y validación fail-fast.
- `health-monitor`: se agrega detección de 429 con retiro temporal (respetando `Retry-After` o
  default 30s), reactivación automática y backoff exponencial ante 429 repetidos.
- `quota-manager`: se agrega inicialización de `remaining` desde `quota_hint` del YAML y precedencia
  del valor aprendido en runtime/persistido sobre el valor inicial.

## Impact

- Código: `src/internal/adapter/generic/*` (nuevo), `src/internal/registry/*`,
  `src/internal/health/*`, `src/internal/quota/*`, `src/internal/adapter/conformance_test.go`.
- Config: nuevo archivo `config/providers/free-tier.yaml`.
- Datos: tabla de cuota aprendida en PostgreSQL (lectura en boot, ya prevista en `quota-manager`).
- Dependencias: ninguna nueva (reutiliza `net/http`, `gopkg.in/yaml.v3` ya presentes en el stack).
- Sin impacto en UI (Fase 1 backend-puro, sin superficie visual).

## Trazabilidad

- Épica: EP-EVO-001
- Historias: HU-EVO-001, HU-EVO-002, HU-EVO-003, HU-EVO-004, HU-EVO-005
