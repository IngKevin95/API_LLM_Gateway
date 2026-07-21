## Context

EP-001 dejó el Router resolviendo una cadena de modelos por capacidad, con `HealthSource`/`QuotaSource` como stubs (`StaticHealth`/`StaticQuota`) y sin nadie que llame al proveedor. EP-002 añade la capa de Adapters (frontera externa, único I/O no determinista) y la orquestación de resiliencia por encima. ADR-001: Go idiomático, sin librerías de proveedor — solo `net/http` stdlib. Tests con `httptest.Server` (sin red real). Sin endpoints HTTP nuevos (EP-005).

## Goals / Non-Goals

**Goals:**
- Contrato `Adapter` estable que traduzca request/response y normalice errores de proveedor.
- Adapters OpenAI (chat/tools, streaming SSE, embeddings), Anthropic (Messages, streaming) y local OpenAI-compat.
- Failover Engine que recorra la cadena del Router ante 5xx/429/timeout, degrade a local, retorne 502 al agotar el pool, NO haga failover ante 400, y distinga retry vs failover.
- Timeouts dinámicos por capacidad (TTFT) + Stream Idle Timeout mid-stream.
- Circuit Breaker pasivo + Max In-Flight por proveedor.
- Health Monitor con histéresis que implemente el `HealthSource` vivo del Router.

**Non-Goals:**
- Endpoints OpenAI/Anthropic-compat (EP-005). El Failover Engine se consume como librería.
- Cache semántica (EP-007), Quota Manager real (EP-003), persistencia (EP-009).
- Learning Engine / ajuste de pesos (EP-007).

## Decisions

- **Contrato Adapter** (`src/internal/adapter`): interfaz mínima
  `type Adapter interface { Chat(ctx, Request) (Response, error); Stream(ctx, Request) (TokenStream, error); Embed(ctx, Request) (Embedding, error) }`.
  Los errores se normalizan a un `*ProviderError{ Status int; Retryable bool; Provider string }` para que el Failover decida. Alternativa descartada: devolver el `*http.Response` crudo — filtraría el formato de cada proveedor a la capa de orquestación.
- **Cliente HTTP**: `net/http` con `http.Client{Timeout}` por request; el `base_url`/`api_key` salen del Registry (EP-001). Sin SDKs.
- **OpenAI-compat como base**: el adapter local (Ollama/vLLM/LM Studio) reutiliza el adapter OpenAI cambiando `base_url`; el edge de "respuesta no compatible" (HU-024 AC3) se cubre validando el esquema al normalizar.
- **Anthropic**: traducción de roles (extrae `system` del array de messages), tools OpenAI→`tool_use`, `max_tokens` default 4096 si falta, parámetros no soportados (`seed`) se ignoran con WARN. Streaming: eventos nativos → chunks SSE OpenAI.
- **Streaming**: `TokenStream` sobre `io.Pipe`; el Stream Idle Timeout se implementa con un `time.Timer` reseteado por token. Regla dura: **no hay failover mid-stream** — si cae tras el primer token, se cierra el socket y se penaliza el score; el failover transparente solo aplica pre-primer-token.
- **Failover Engine** (`src/internal/failover`): recibe la cadena de `router.Resolve(capability, tokens)` y prueba cada eslabón vía su Adapter. Política por código: `400`→abortar sin failover; `429/5xx/timeout`→siguiente eslabón (429 nunca hace retry al mismo, protege cuota); pool agotado→`502` con el último error. Límite de intentos = largo de la cadena (evita bucles).
- **Timeouts dinámicos** (`src/internal/failover`): TTFT por capacidad leído del routing (`ttft_timeout_ms`), estricto para chat/código, relajado para reasoning; Stream Idle Timeout del `stream_idle_timeout_ms` que ya expone el Registry.
- **Circuit Breaker pasivo** (`src/internal/breaker`): contador atómico de in-flight por proveedor; si supera `max_in_flight` → fast-fail 0-I/O; ante fallo/superación se marca inalcanzable durante `cooldown_ms`; reactiva tras backoff + health check sano. No es un breaker activo (no sondea): es pasivo, reacciona al tráfico real.
- **Health Monitor** (`src/internal/health`): goroutine con `time.Ticker` que sondea cada proveedor; histéresis (N fallos para retirar, M éxitos para reactivar) evita oscilación. Expone `Healthy(providerID, model) bool` → **implementa `router.HealthSource`**, reemplazando `StaticHealth` en el cableado. Estado en un `map` protegido por `sync.RWMutex`.
- **Integración EP-001**: el Router no cambia su contrato; solo se le inyecta el Health Monitor real en lugar del stub. El Failover consume `router.Resolve`/`ResolveExplicit`.

## Risks / Trade-offs

- **Tests de streaming/timeout son sensibles a timing** → usar `httptest.Server` con handlers que controlan el ritmo de emisión y timeouts cortos deterministas; nunca `time.Sleep` largos en tests.
- **Circuit Breaker pasivo con contadores concurrentes** → race conditions. Mitigación: `sync/atomic` para in-flight, `-race` en CI, tests concurrentes.
- **Failover mid-stream imposible por diseño** → el cliente puede recibir un stream cortado. Es una decisión explícita del PRD (no se puede rebobinar tokens ya emitidos); se penaliza el score para desfavorecer al proveedor inestable.
- **Health Monitor sondea proveedores reales en runtime** → en tests se mockea; en producción respeta intervalos configurables para no gastar cuota.

## Migration Plan

Aditivo sobre EP-001 (ya en develop). El cableado en `cmd/gateway` inyecta el Health Monitor real donde hoy va `StaticHealth`. Sin migración de datos. Rollback = revertir el PR.

## Open Questions

- Ninguna bloqueante. Los defaults (cooldown 30s, histéresis N=?/M=?, max_tokens 4096) se fijan por sub-slice según los AC y se exponen en `config.yaml` donde el PRD lo pida.
