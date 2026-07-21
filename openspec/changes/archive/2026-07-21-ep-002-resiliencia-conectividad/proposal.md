## Why

EP-001 resuelve *qué* modelo usar (cadena ordenada por score) pero nadie *llama* al proveedor todavía. Sin adapters que traduzcan al formato de cada LLM ni un motor de failover que recorra la cadena ante fallos, la Gateway no completa una petición real ni cumple el Obj. 2 (resiliencia, éxito ≥99% con ≥1 proveedor sano). Es la segunda épica `foundational`: convierte la decisión de routing en una respuesta entregada, con degradación transparente a modelos locales.

## What Changes

- Nuevo **contrato Adapter**: interfaz que traduce el request interno al formato de cada proveedor, envía la llamada y normaliza respuesta y errores (para que el failover decida). Frontera de aislamiento con todo I/O externo no determinista.
- **Adapter OpenAI** (base OpenAI-compat): chat + tool calling, streaming SSE, embeddings.
- **Adapter Anthropic**: Messages (chat, roles, tool calling) + streaming.
- **Adapter local** OpenAI-compat (Ollama/vLLM/LM Studio) como último eslabón de degradación.
- **Failover Engine**: recorre la cadena del router ante 5xx/429/timeout, degrada a local, retorna 502 al agotar el pool; NO hace failover ante 400 (cliente); distingue retry vs failover (429 → failover para proteger cuota).
- **Timeouts dinámicos** por capacidad (TTFT estricto para chat/código, relajado para reasoning) + Stream Idle Timeout mid-stream.
- **Circuit Breaker pasivo** + Max In-Flight configurable por proveedor (previene Failover Suicide).
- **Health Monitor**: worker periódico que sondea proveedores y los retira/reactiva. Provee la fuente viva de `HealthSource` que EP-001 dejó stub.
- Sin endpoints HTTP nuevos: la orquestación es interna; los endpoints OpenAI/Anthropic-compat son EP-005.

## Capabilities

### New Capabilities
- `provider-adapters`: contrato Adapter + implementaciones OpenAI (chat/streaming/embeddings), Anthropic (chat/streaming) y local OpenAI-compat. Traduce request/response y normaliza errores de proveedor.
- `failover-engine`: recorrido de la cadena de fallback con degradación a local, política retry-vs-failover por código HTTP, límite de intentos, 502 al agotar el pool, y timeouts dinámicos + stream idle timeout por capacidad.
- `circuit-breaker`: breaker pasivo + Max In-Flight por proveedor que marca nodos inalcanzables temporalmente para prevenir Failover Suicide.
- `health-monitor`: worker de sondeo periódico que retira/reactiva proveedores y expone el estado de salud vivo consumible por el router (implementa `HealthSource`).

### Modified Capabilities
<!-- Ninguna spec previa cambia sus requisitos. El router (model-router) consumirá la fuente viva de salud vía la interfaz HealthSource ya existente, sin cambiar su contrato. -->

## Impact

- Código nuevo (Go): `src/internal/adapter` (contrato + openai/anthropic/local), `src/internal/failover`, `src/internal/breaker`, `src/internal/health`. Nombres tentativos, se fijan en design.md.
- Integra con EP-001: el Failover Engine consume `router.Resolve`/`ResolveExplicit` (cadena) y los adapters; el Health Monitor implementa `router.HealthSource` (hoy `StaticHealth`).
- Dependencias: cliente HTTP stdlib (`net/http`) — sin librerías de proveedor. Tests con `httptest.Server` (sin red real).
- Sin breaking changes; sin endpoints nuevos (EP-005). Config: usa `providers`/`routing` de `config.yaml` (Anexo A) ya cargados por el Registry.

## Trazabilidad

- **Épica**: EP-002 · Resiliencia y Conectividad Base (`layer: foundational`) — objetivo del PRD: Obj. 2 (resiliencia; failover transparente; éxito ≥99% con ≥1 proveedor sano).
- **Historias cubiertas** (por sub-slice):
  - SS1 — `provider-adapters` (OpenAI): HU-020a (chat + tool calling), HU-020b (streaming SSE), HU-020c (embeddings)
  - SS2 — `provider-adapters` (Anthropic + local): HU-021a (chat/roles/tools), HU-021b (streaming), HU-024 (local OpenAI-compat)
  - SS3 — `failover-engine`: HU-004a (failover en cadena + degradación local), HU-004c (timeouts dinámicos + stream idle)
  - SS4 — `circuit-breaker` + `health-monitor`: HU-004b (breaker pasivo + Max In-Flight), HU-005 (health checks + retiro/reactivación)
