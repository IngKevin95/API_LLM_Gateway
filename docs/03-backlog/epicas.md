# Épicas de API LLM Gateway

Regla de trazabilidad bidireccional: toda épica cubre ≥ 1 objetivo del PRD; toda historia futura pertenece a ≥ 1 épica.

Fuente: `docs/01-prd/api-llm-gateway.md`. Objetivos del PRD:
1. **Obj. 1** — Desacople total (0 refs a proveedor/modelo en agentes; consumo por capacidad).
2. **Obj. 2** — Resiliencia (failover transparente; éxito ≥ 99% con ≥1 proveedor sano).
3. **Obj. 3** — Selección óptima por score (overhead de routing < 50 ms).
4. **Obj. 4** — Seguridad empresarial (100% auth + authz + auditoría; secretos protegidos).
5. **Obj. 5** — Compatibilidad universal (API OpenAI/Anthropic-compat; Free Claude Code).

---

## EP-001 · Enrutamiento por capacidad

| Campo | Valor |
|---|---|
| Objetivo(s) del PRD cubiertos | Obj. 1, Obj. 3 |
| Capa (build) | foundational |
| OpenSpec change | enrutamiento-por-capacidad |
| Métrica de éxito | Agente pide capacidad sin nombrar modelo; Router resuelve por score con overhead p95 < 100 ms (enrutamiento puro en RAM, excluye latencia externa del proveedor); 0 refs a proveedor/modelo en el código del consumidor |

**De qué se trata**: el corazón del Gateway. Registry declarativo (YAML) de providers, modelos y capacidades, más el Model Router que traduce una capacidad (`chat`, `reasoning`, `coding`, `vision`, `image`, `embedding`) al modelo óptimo según score = f(calidad, velocidad, disponibilidad, cuota restante, costo, latencia). Soporta modo automático (sin `model`) y explícito.

**Por qué existe**: sin esta capa, cada agente vuelve a acoplarse a un proveedor concreto — es la razón de ser del proyecto. Es prerequisito de todas las demás épicas.

**Capabilities que agrupa**:
- Registry / configuración declarativa YAML (providers, models, routing)
- Model Router con scoring y resolución por capacidad
- Validación de Contexto (Context Window) pre-score (Buffer estricto del 20%) con tokenizador por adapter
- Modo automático vs. modelo explícito
- Optimización Semántica / Guardián de Prompts

**Historias anticipadas**:
- HU-001 — Cargar providers/models/routing desde YAML (Registry)
- HU-002a — Resolver capacidad → modelo por score (Router automático)
- HU-002b — Manejo de errores y desempates en el enrutamiento
- HU-003 — Forzar modelo explícito vía parámetro `model` con política de fallback\n- HU-035 — Tokenizador de Contexto (Context Window)
- 

---

## EP-002 · Resiliencia y Conectividad Base

| Campo | Valor |
|---|---|
| Objetivo(s) del PRD cubiertos | Obj. 2 |
| Capa (build) | foundational |
| OpenSpec change | ep-002-resiliencia-conectividad |
| Métrica de éxito | Ante 429/500/timeout del primario la petición se completa por el siguiente de la cadena; éxito ≥ 99% con ≥1 proveedor sano |

**De qué se trata**: Failover transparente en cadena ordenada por capacidad, más Health Monitor que mide disponibilidad/latencia/throughput/errores y retira o reactiva proveedores automáticamente. Último eslabón: modelos locales (Ollama/vLLM/LM Studio).

**Por qué existe**: los modelos gratuitos y locales tienen cuotas y caídas frecuentes (429/500); sin failover automático la plataforma no es confiable para uso productivo.

**Capabilities que agrupa**:
- Adaptadores de Proveedores Base (OpenAI, Anthropic)
- Failover en cadena por capacidad
- Health Monitor (checks periódicos + reactivación)
- Circuit Breaker Preventivo (Max In-Flight configurable por proveedor en YAML) y Stream Idle Timeout
- TTFT dinámico (estricto 2.0s para chat/código, relajado para reasoning)
- Manejo de fallo mid-stream (no hay failover transparente; se aborta la conexión y se penaliza el score)
- Degradación a modelos locales

**Historias anticipadas**:
- HU-004a — Failover básico de cadena con degradación a local
- HU-004b — Circuit Breaker pasivo y Max In-Flight
- HU-004c — Timeouts dinámicos por capacidad y Stream Idle Timeout
- HU-005 — Health checks periódicos y retiro/reactivación de proveedor
- HU-020a — Adapter OpenAI (chat y tool calling)
- HU-020b — Adapter OpenAI (streaming SSE)
- HU-020c — Adapter OpenAI (embeddings)
- HU-021a — Adapter Anthropic (chat, roles y tool calling)
- HU-021b — Adapter Anthropic (streaming)\n- HU-024 — Adapter para modelos locales (Ollama / vLLM / LM Studio)

---

## EP-003 · Gobernanza de consumo (cuota y costo)

| Campo | Valor |
|---|---|
| Objetivo(s) del PRD cubiertos | Obj. 4 (gobernanza de cuotas) + KPI costo (Obj. 3) |
| Capa (build) | business |
| Métrica de éxito | Proveedor con cuota agotada se retira automáticamente; costo por petición/agente/proveedor registrado y consultable |

**De qué se trata**: Quota Manager que controla **cuota acumulada** (requests/tokens por ventana temporal: minuto/día/mes) por proveedor y por clave, y corta al agotarse; registra costo por petición, agente y proveedor. *Distinción con EP-004A*: EP-003 gestiona consumo acumulado por ventana; el **rate limiting** vive en EP-004A.

**Por qué existe**: aprovechar cuotas gratuitas sin excederlas (respetando ToS) y evitar gasto descontrolado exige control central de cuota y costo.

**Capabilities que agrupa**:
- Quota Manager (límites por clave/tenant/proveedor)
- Registro y agregación de costo
- Corte automático por agotamiento

**Historias anticipadas**:
- HU-006 — Contabilizar y limitar requests/tokens por cuota
- HU-007 — Registrar costo por petición/agente/proveedor

---

## EP-004A · Identidad y Accesos

| Campo | Valor |
|---|---|
| Objetivo(s) del PRD cubiertos | Obj. 4 |
| Capa (build) | foundational |
| OpenSpec change | ep-004a-identidad-accesos |
| Métrica de éxito | 100% de peticiones autenticadas + autorizadas por scope |

**De qué se trata**: AuthN (API key / OAuth2-OIDC / mTLS), AuthZ por scope/RBAC y por tenant con aislamiento multi-tenant, rate limiting, y prevención de abuso a nivel red.

**Por qué existe**: es requisito explícito y exhaustivo del sponsor para controlar quién y cómo ingresa al gateway.

**Capabilities que agrupa**:
- AuthN (API key / OAuth2-OIDC / mTLS)
- AuthZ (RBAC/scopes, multi-tenant)
- Rate limiting L7 (throttling por request/segundo contra abuso, Least Connections y límite de 2 peticiones para visión; cuota de EP-003)

**Historias anticipadas**:
- HU-008 — Autenticar toda petición (API key)
- HU-009 — Autorizar por scope/RBAC y tenant
- HU-022 — Rate Limiting y protección de Payload
- HU-022b — Límite de concurrencia y enrutamiento de red para capacidad vision
- HU-025a — Autenticación OAuth2/OIDC
- HU-025b — Autenticación mTLS\n- HU-027 — Guardián de Prompts (Seguridad y Jailbreak)

---

## EP-004B · Seguridad, Protección y Auditoría

| Campo | Valor |
|---|---|
| Objetivo(s) del PRD cubiertos | Obj. 4 |
| Capa (build) | business |
| Métrica de éxito | 0 secretos en logs/config versionada; auditoría inmutable completa |

**De qué se trata**: Gestión de secretos (env/secret manager, rotación, multi-key legítimo), guardrails contra exfiltración/DLP, y auditoría inmutable con redacción de PII/secretos. TLS obligatorio; cifrado en tránsito y reposo.

**Por qué existe**: Garantizar que los datos y secretos empresariales nunca queden expuestos ni lleguen al proveedor en texto claro.

**Capabilities que agrupa**:
- Secretos (rotación, sin exposición)
- Auditoría con redacción de PII/secretos; TLS y cifrado en reposo
- Redacción síncrona / DLP
- Protección TCP contra ataques Slowloris

**Historias anticipadas**:
- HU-010 — Auditar cada petición con redacción de secretos/PII
- HU-011 — Gestionar y rotar secretos sin exponerlos en logs/config
- HU-026a — Redacción Síncrona de Secretos
- HU-026b — Kill-Switch Asíncrono de PII
- HU-028 — Cifrado de Sobre (Envelope) en BD de Auditoría\n- HU-034 — Protección TCP contra ataques Slowloris

---

## EP-005 · API universal compatible (LLM universal)

| Campo | Valor |
|---|---|
| Objetivo(s) del PRD cubiertos | Obj. 5 |
| Capa (build) | business |
| Métrica de éxito | Un cliente OpenAI-compat y Free Claude Code (Anthropic-compat) funcionan apuntando a la Gateway; petición sin `model` enruta, con `model` usa el indicado |

**De qué se trata**: exponer endpoints compatibles OpenAI (`/v1/chat/completions`, `/v1/embeddings`, `/v1/models`) y Anthropic Messages, para que la Gateway sea un LLM universal consumible por herramientas existentes sin cambios.

**Por qué existe**: la compatibilidad de contrato es lo que permite adopción sin fricción (OpenCode, Free Claude Code, apps) y hace real el "LLM universal".

**Capabilities que agrupa**:
- Endpoint OpenAI-compat (chat, embeddings, models)
- Endpoint Anthropic-compat (Messages) para Free Claude Code
- Streaming de respuestas
- Soporte Multiagente (MCP)

**Historias anticipadas**:
- HU-012a — Endpoint OpenAI-compat de chat (sin streaming)
- HU-012b — Streaming SSE compatible OpenAI
- HU-012c — Endpoint OpenAI-compat de embeddings
- HU-013 — Endpoint Anthropic-compat para apuntar Free Claude Code
- HU-016 — Configurar Free Claude Code contra la Gateway
- HU-033 — Integración MCP para multi-agentes

---

## EP-006 (Eliminada / Deprecated)

| Campo | Valor |
|---|---|
| Objetivo(s) del PRD cubiertos | - |
| Métrica de éxito | - |

**De qué se trata**: Reservado por continuidad histórica.

---

## EP-007 · Observabilidad y aprendizaje

| Campo | Valor |
|---|---|
| Objetivo(s) del PRD cubiertos | Obj. 2, Obj. 3 (retroalimentan resiliencia y scoring) |
| Capa (build) | business |
| Métrica de éxito | Métricas por modelo/proveedor consultables (latencia, success_rate, tokens, quota, costo); ranking visible; pesos del score ajustables con histórico |

**De qué se trata**: telemetría estructurada, métricas dinámicas por modelo/proveedor, dashboard/endpoint de ranking, y el Learning Engine que ajusta el enrutamiento con datos históricos.

**Por qué existe**: sin métricas reales el score usa pesos fijos y no mejora; la observabilidad es condición para operar en producción y para el autoaprendizaje futuro.

**Capabilities que agrupa**:
- Telemetría y métricas dinámicas
- Dashboard/endpoint de ranking
- Learning Engine (ajuste de pesos con histórico, **Fase 3**, no MVP)
- Semantic Cache (Fase 2, exploratoria)

**Historias anticipadas**:
- HU-017 — Exponer métricas por modelo/proveedor
- HU-018 — Registrar histórico de peticiones para aprendizaje
- HU-019 — Ajustar pesos del score con datos históricos (Learning Engine)
- HU-023 — API de Métricas para Dashboard\n- HU-032 — Cache Exacta (Primera Fase)
- HU-032 — Cache semántica

---

## EP-008 · Ecosistema de Adaptadores Secundarios

| Campo | Valor |
|---|---|
| Objetivo(s) del PRD cubiertos | Obj. 2 |
| Capa (build) | business |
| Métrica de éxito | Proveedores de nicho integrados sin afectar el core |

**De qué se trata**: Integrar modelos y proveedores adicionales más allá de los base (OpenAI, Anthropic).

**Capabilities que agrupa**:
- Adaptadores adicionales

**Historias anticipadas**:
- HU-029 — Adapter para AIHubMix
- HU-030 — Adapter para Google (Gemini)
- HU-031 — Adapter para OpenRouter

---

## EP-009 · Sincronización Asincronista y Persistencia (Infraestructura Crítica)

| Campo | Valor |
|---|---|
| Objetivo(s) del PRD cubiertos | Obj. 4 (Auditoría y Seguridad), Obj. 2 (Resiliencia) |
| Capa (build) | foundational |
| Métrica de éxito | Logs de auditoría persisten de forma inmutable en PostgreSQL sin pérdida ante caída del Gateway; crash recovery restaura cuotas/auth desde WAL en < 5min; secretos nunca se loguean en plaintext |

**De qué se trata**: infraestructura asincronista de fondo (Sync Worker) que:
1. Escribe eventos de auditoría a DB vía KMS Envelope (cifrado local, llave maestra guardada en KMS)
2. Mantiene Write-Ahead Log (WAL) local para crash recovery (evita pérdida de eventos entre caídas)
3. Maneja Graceful Shutdown (flushing de buffers antes de salir)
4. Implementa Cache Invalidator (Fase 2+) para refresco dinámico de cuotas/auth cuando la DB cambia externamente

**Por qué existe**: sin persistencia asincronista, la ruta crítica se bloquea escribiendo a DB; sin WAL, cualquier caída pierde auditoría; sin Graceful Shutdown, se pierden eventos. Es prerequisito silencioso de EP-004B (Auditoría).

**Capabilities que agrupa**:
- Sync Worker (writer asincronista a DB vía channels de Go)
- KMS Envelope Encryption (DEK local, KEK en KMS)
- Write-Ahead Log (WAL) local con recovery
- Graceful Shutdown (flush antes de exit)
- Cache Invalidator (Fase 2+, webhook/polling)

**Historias anticipadas**:
- HU-038 — Implementar Sync Worker (persistencia asincronista vía channels)
- HU-039 — Write-Ahead Log (WAL) local con recuperación ante crash
- HU-040 — Graceful Shutdown con flush obligatorio de WAL/buffers
- HU-041 — Cache Invalidator (Fase 2+, polling/webhook)

---

## Resumen de cobertura

| Épica | Objetivo(s) del PRD |
|---|---|
| EP-001 Enrutamiento por capacidad | Obj. 1, Obj. 3 |
| EP-002 Resiliencia y Conectividad Base | Obj. 2 |
| EP-003 Gobernanza (cuota/costo) | Obj. 4 (gobernanza de cuotas) + KPI costo (Obj. 3) |
| EP-004A Identidad y Accesos | Obj. 4 |
| EP-004B Seguridad y Auditoría | Obj. 4 |
| EP-005 API universal compatible | Obj. 5 |
| EP-007 Observabilidad y aprendizaje | Obj. 2, Obj. 3 |
| EP-008 Adaptadores Secundarios | Obj. 2 |
| EP-009 Sincronización Asincronista y Persistencia | Obj. 4, Obj. 2 |

**Cobertura de objetivos** (ningún objetivo huérfano):
- Obj. 1 → EP-001 (directo)
- Obj. 2 → EP-002 (directo), EP-007 (retroalimenta), EP-008 (directo)
- Obj. 3 → EP-001, EP-007
- Obj. 4 → EP-004A, EP-004B (directo) y EP-003 (gobernanza de cuotas, per PRD §"Seguridad y gobernanza de acceso"). EP-003 también alimenta el KPI de costo (Obj. 3).
- Obj. 5 -> EP-005

**Épicas huérfanas** (sin objetivo): ninguna. Toda épica ata a ≥1 objetivo del PRD.
