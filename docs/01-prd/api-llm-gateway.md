---
id: PRD-001
titulo: API LLM Gateway
estado: draft
version: 0.1
sponsor: Kevin Beltrán
actualizado: 2026-07-19
para-agentes: true
---

# API LLM Gateway — Product Requirements Document

**Estado**: draft · **v0.1** · **Sponsor**: Kevin Beltrán · **Actualizado**: 2026-07-19

> Fuentes de verdad del dominio: `docs/REQUERIMIENTO_GENERAL.md`, `docs/ARQUITECTURA.md`,
> `docs/MODELS.md`. Este PRD las consolida en los 12 componentes de la metodología Factory.

## Resumen y objetivos

**Problema**: los agentes de IA y las aplicaciones actuales quedan acoplados a un proveedor
concreto (OpenAI, Anthropic, Google). Cambiar de modelo obliga a tocar código, no hay failover
transparente cuando un proveedor devuelve 429/500, no se controla cuota ni costo de forma central,
y no existe una capa única de seguridad empresarial sobre todos los proveedores.

**Por qué ahora**: proliferan modelos gratuitos y locales (AIHubMix, GLM, MiniMax, Ollama, vLLM)
con cuotas y latencias muy distintas. Aprovecharlos de forma resiliente exige una capa de
enrutamiento y gobernanza que hoy se resuelve a mano en cada proyecto. Un Gateway propio da control
total del routing, de los costos y de la seguridad que los gateways de terceros (LiteLLM, Portkey,
OpenRouter) no exponen completo.

Objetivos medibles (3-5):
1. **Desacople total**: 0 referencias a proveedor/modelo concreto en el código de agentes; los
   agentes solo piden capacidades (`coding`, `reasoning`, `vision`, `image`, `embedding`, `chat`).
2. **Resiliencia**: failover automático y transparente (previo a emitir el primer token) ante 429/500/timeout con éxito de la petición ≥ 99.9%. Si el error ocurre a mitad de la generación (mid-stream), la conexión se aborta reportando el error (no hay failover transparente mid-stream) y este fallo alimenta la heurística del Circuit Breaker para futuras peticiones.
3. **Selección óptima**: el Router elige por score compuesto (calidad, velocidad, disponibilidad,
   cuota restante, costo, latencia) en p95 < 100 ms de overhead sobre la latencia del proveedor.
4. **Seguridad empresarial**: 100% de las peticiones autenticadas, autorizadas por scope y
   auditadas; secretos nunca en claro en config ni logs.
5. **Compatibilidad universal**: exponer una API compatible OpenAI (`/v1/chat/completions`,
   `/v1/embeddings`) y una Anthropic-compatible que permita apuntar Free Claude Code a la Gateway.

**Fuera de alcance**:
- Fine-tuning o hosting propio de modelos (siempre se consumen vía adapters).
- RAG y memoria compartida entre agentes (candidato a extensión futura post-MCP).

## Stakeholders

| Rol | Persona / equipo | Responsabilidad |
|---|---|---|
| Sponsor | Kevin Beltrán | Visión, prioridades, aprobación de releases |
| Product Owner | Kevin Beltrán | Backlog, decisiones de alcance |
| Equipo técnico | Equipo de desarrollo | Diseño e implementación del Gateway y adapters |
| Usuario primario | Agentes de IA / desarrolladores vía OpenCode/Free Claude Code | Consumen capacidades a través de la Gateway |
| Usuario secundario | Aplicaciones/servicios que usan la API como LLM universal | Integración vía API compatible OpenAI/Anthropic |
| Soporte / operaciones | Equipo de desarrollo (rol Ops) | Health, cuotas, rotación de claves, observabilidad |

## Historias de usuario (resumen)

> Detalle con AC en G/W/T vive en `docs/04-historias/`. Aquí solo perfil + necesidad.

- Como **agente de código**, quiero solicitar una capacidad (`coding`) sin conocer el modelo, para obtener la mejor respuesta disponible sin acoplarme a un proveedor.
- Como **aplicación integradora**, quiero llamar a la Gateway como LLM universal (en automático o con `model` explícito), para recibir la respuesta del modelo adecuado con formato compatible.
- Como **operador**, quiero que un proveedor con 429/agotado se retire solo y se reactive al recuperarse, para tener alta disponibilidad sin intervención manual.
- Como **operador de seguridad**, quiero que toda petición sea autenticada, autorizada y auditada con secretos fuera de logs, para cumplir requisitos empresariales.
- Como **owner/finanzas**, quiero ver consumo, costo y cuota por agente/proveedor, para controlar el gasto.

## Arquitectura de alto nivel

```
   OpenCode / Free Claude Code / App integradora
                        │
                        ▼
        API compatible (OpenAI /v1, Anthropic)  ── AuthN/AuthZ · Rate limit · Audit
                        │
                        ▼
                   AI Gateway
   ┌─────────┬─────────┬───────────┬────────────┬───────────────┬─────────────┐
   ▼         ▼         ▼           ▼            ▼               ▼
 Registry  Model    Health      Quota       Failover        Learning
 (YAML)    Router   Monitor     Manager     (cadena)        Engine
          (score)                 │
             │                   │
     Cache   │           ┌──────┴──────┐
   Semántica  │           ▼              ▼
             │       PostgreSQL ◄───► KMS
             │      (audit/hist)    (cifrado)
             ▼
          Adapters                   Dashboard
   ┌──────────┬──────────┬──────────┬──────────┬──────────┬──────────┐
   ▼          ▼          ▼          ▼          ▼          ▼          ▼
 OpenAI   Anthropic   Google   OpenRouter  AIHubMix   Ollama     vLLM/LM Studio
```

Regla estructural: agregar un proveedor = crear Adapter + entrada YAML + registrar capacidades.
Modelos locales se tratan idénticamente a los remotos.

## Funcionalidades por capability

**Enrutamiento por capacidad (Model Router)**
- Resolución capacidad → modelo por score = `(w_calidad * norm(Q)) + (w_velocidad * norm(V)) + (w_disponibilidad * norm(D)) - (w_costo * norm(Costo)) - (w_latencia * norm(L)) - (w_penalizacion * P)`, donde `norm()` normaliza el valor a una escala 0-1, los pesos (w) se configuran en el YAML y P representa fallos recientes (timeout, abortos de stream).
- **Validación de Contexto (Context Window)**: Antes de calcular el score, el Router debe estimar los tokens del request usando el **tokenizador específico de cada adapter** (ej. `tiktoken-go` para OpenAI, tokenizador de Anthropic para Claude, sentencepiece para Llama/Google) y aplicar un **margen de seguridad del 20%** (buffer). Descarta modelos cuyo límite máximo no soporte el request. Los adjuntos multimodales (ej. imágenes) se descontarán del límite global.
- Modo automático (sin `model`) y modo explícito (con parámetro `model` → usa ese modelo si está sano; si no, política de fallback configurable).

**Resiliencia (Failover + Health Monitor)**
- Cadena de fallback ordenada por capacidad (YAML `routing`).
- **Circuit Breaker Pasivo**: Si un proveedor devuelve 429/5xx/timeout, el Gateway lo marca instantáneamente como "inalcanzable" por un periodo corto. Incorpora un límite de **peticiones en vuelo (Max In-Flight configurable por proveedor en YAML, ej. 300 a 500 peticiones para top tiers)** por proveedor; si se supera, el circuito se abre preventivamente (Fast Fail) sin esperar el timeout de 2.0s, evitando el *Failover Suicide* (donde cientos de peticiones agotarían el pool de conexiones). En modo Fast Fail, el Gateway rechaza la petición instantáneamente devolviendo un error HTTP al cliente sin intentar failover.
- Retiro automático de proveedor ante 429/500/timeout con un periodo de gracia (backoff fijo, ej. 30s) antes del siguiente health check. Fallback Exhaustion: Si toda la cadena de fallback falla, el Gateway retorna `503 Service Unavailable`.

**Gobernanza (Quota Manager)**
- Control de requests/tokens por minuto/día/mes por proveedor y por clave; corte al agotar cuota.
- Persistencia de Cuotas: Volcado asíncrono desde la RAM hacia PostgreSQL cada 5 segundos para minimizar el *drift* ante reinicios o despliegues (diferente del L7 Hash de Rate Limiting que sí garantiza 0 drift).
- Registro de costo por petición, agente y proveedor.

**API universal / compatibilidad**
- Endpoints compatibles OpenAI (`/v1/chat/completions`, `/v1/embeddings`, `/v1/models`).
- Endpoint compatible Anthropic Messages para apuntar Free Claude Code a la Gateway.
- Capacidades soportadas: `chat`, `reasoning`, `coding`, `vision`, `image`, `embedding`.
**Configuración declarativa**
- `providers`, `models` (atributos quality/coding/reasoning/speed/vision/cost/latency), `routing` en YAML; API keys por variable de entorno.

**Seguridad y gobernanza de acceso** (cubre Objetivo 4)
- AuthN (API key / OAuth2-OIDC / mTLS), AuthZ por scope/RBAC y por tenant, rate limiting y quotas por clave/tenant.
- Auditoría por petición con redacción de PII/secretos; secretos fuera de config y logs. Detalle en *Requisitos técnicos → Seguridad*.
- **Redacción Síncrona y Prevención**: Escaneo en memoria (< 10ms por cada 10k tokens) para garantizar que ningún secreto en texto plano llegue al modelo. **Límite de payload por tipo**: el cuerpo JSON/texto de la petición se limita a 10MB (peticiones textuales > 10MB se rechazan con `413 Payload Too Large`); el contenido binario de imagen para la capacidad `vision` elude el motor DLP pero exige validación estricta y síncrona de Magic Bytes (MIME real en memoria) y límites de resolución para prevenir DDoS y malware. La capacidad `vision` deshabilita el DLP sobre el contenido de la imagen (no hay OCR), por lo que DEBE restringirse a agentes confiables mediante AuthZ/Scopes (aislamiento de Tenant).
- **Guardián de Prompts**: Capacidad opt-in para optimizar semánticamente los prompts (técnicas de prompt engineering) antes de enrutarlos.


**Inteligencia y Optimización**
- **Caché Semántica**: Intercepta peticiones repetidas basándose en coincidencia exacta primero, para reducir latencia y costo.
- **Learning Engine**: Ajusta dinámicamente los pesos del score basándose en el historial de fallos y latencias reales observadas por proveedor.
- **Protocolo de Streaming SSE**: Las respuestas en stream deben apegarse estrictamente al formato Server-Sent Events (SSE).

## UX y diseño

| Aspecto | Definición |
|---|---|
| Principios UX | API-first; contratos compatibles OpenAI/Anthropic para adopción sin fricción. CLI con salida clara, streaming de tokens, mensajes de error accionables. Idioma de la doc: español. |
| Accesibilidad | CLI operable por teclado, salida legible por lectores de pantalla (texto plano, sin depender de color). |
| Marca | API-first. El Dashboard (HU-023) expone métricas vía endpoint JSON y una UI ligera de solo lectura; no es una aplicación web completa. |
| Referencias | LiteLLM, Portkey, OpenRouter, Vercel AI Gateway (benchmarks de contrato y features; no copiar). |

## Riesgos y mitigaciones

| Riesgo | Impacto | Mitigación |
|---|---|---|
| Cambios abruptos en las APIs de los LLMs | Los clientes fallan al enviar peticiones | Uso del patrón Adapter; actualización rápida de los adaptadores y abstracción del payload |
| Latencia inaceptable añadida por el Gateway | Los agentes timeout | Implementación en Go, límite de concurrencia y validación estricta del p95 < 100ms |
| Vulnerabilidad en la exposición de Secretos | Filtración de API Keys empresariales | KMS, redacción síncrona/asíncrona y despliegue en memoria segura |

## Requisitos técnicos

- **Stack**: backend de servicio API en Go (Golang) según [ADR-001]; config en YAML; almacenamiento de métricas/estado en PostgreSQL (con extensión pgvector para Caché Semántica) para soportar 43M logs/día; caché en Local RAM para Quotas. Modelos locales vía Ollama/vLLM/LM Studio.
- **Integraciones**: proveedores vía Adapters (OpenAI, Anthropic, Google, OpenRouter, AIHubMix, Ollama, vLLM, LM Studio); Free Claude Code y OpenCode como clientes; MCP.
- **Performance**: overhead de routing p95 < 100 ms sobre la latencia del proveedor; streaming obligatorio para payloads pesados; throughput objetivo de 500 **RPS continuos** para endpoints de texto (equivalente a ~2500 conexiones concurrentes en vuelo asumiendo TTFT y streaming prolongado), y un límite estricto de concurrencia para visión de **2 peticiones concurrentes** por nodo para modelos locales (VRAM), y un límite dinámico por Max In-Flight para modelos remotos (RAM).
- **Seguridad (exhaustiva, empresarial)**:
  - *AuthN*: API keys de cliente, OAuth2/OIDC y mTLS opcional para servicio-a-servicio.
  - *AuthZ*: RBAC/scopes por agente y por tenant; permisos por capacidad y por modelo; aislamiento multi-tenant.
  - *Secretos*: claves de proveedor solo por variable de entorno / secret manager; rotación de claves y soporte multi-key.
  - *Transporte y reposo*: TLS obligatorio; **Cifrado de Sobre (Envelope Encryption)** para proteger la auditoría en PostgreSQL (KMS cifra la llave AES maestra, y el texto se cifra en local ultrarrápido).
  - *Límites y abuso*: Rate limiting en RAM local apoyado por balanceo **L7 (Hash de API Key)** para garantizar 0 drift multi-nodo. Excepción: Las peticiones intensivas (`vision`) eluden el Hash L7 y usan *Least Connections* para evitar saturar un solo nodo; *Riesgo Asumido*: se acepta un leve "drift" de cuota en RAM local exclusivamente para visión, priorizando la estabilidad física del cluster. Protección TCP estricta contra ataques Slowloris.
  - *Prevención Síncrona y Guardrails*: Escaneo DLP 100% síncrono para el payload de salida (request) para evitar fuga de datos al proveedor, relajando el SLA de latencia para payloads masivos. Escaneo asíncrono reservado solo para respuestas (responses). Guardián de prompts opt-in.
  - *Auditoría*: log de auditoría inmutable por petición (el cifrado KMS Envelope y escritura en DB deben ser estrictamente asíncronos vía workers respaldados por un WAL local para garantizar 0 pérdida ante OOM; aplica política de retención TTL de 30 días para purgar métricas antiguas y evitar saturar PostgreSQL).
  - *Cumplimiento*: respetar ToS de proveedores.
- **Disponibilidad**: SLO objetivo ≥ 99.9% uptime mensual (coherente con la tabla de KPIs; exige diseño HA con failover multi-proveedor y degradación a modelos locales). **RTO (Recovery Time Objective)**: < 1 hora ante caída del Gateway (nodos pueden recuperarse vía WAL local). **RPO (Recovery Point Objective)**: < 15 minutos (máximo sin persistir = 15 min de buffer en WAL local). Umbral de timeout TTFT (Time-To-First-Token) **dinámico**: estricto de 2.0s para capacidades estándar (chat/código), **5.0s para visión (processing de imágenes)**, pero relajado y adaptativo para modelos de `reasoning` que piensan antes de emitir tokens (configurable, ej. < 30s). El TTFT externo no incluye el overhead del Guardián de Prompts. Adicionalmente, se aplica un **Stream Idle Timeout (configurable por modelo en YAML)**: 5s estándar, **10s para visión**, < indefinido para reasoning; si el proveedor deja de emitir tokens en medio del stream (TBT excede el timeout físico), el socket se cierra unilateralmente para liberar conexiones y el fallo penaliza el score del modelo.

- **Latencia de Autenticación (Auth & Rate Limiting)**: < 5ms p99 mediante lookup O(1) en RAM local, cero I/O a base de datos en camino crítico. Requisito: balanceo L7 con sticky sessions (Hash de API Key) para garantizar consistencia de cuota multi-nodo.

- **Capacidad de Auditoría**: soporte para ~43 millones de eventos de auditoria diarios (43M/día), con particionamiento mensual en PostgreSQL y retención TTL de 30 días. Cifrado de sobre (Envelope Encryption) con KMS para proteger eventos en reposo.

- **Max In-Flight (Concurrencia)**: límites configurables por proveedor en YAML: 300-500 para proveedores top-tier (OpenAI, Anthropic), **50 para modelos locales** (Ollama, vLLM, LM Studio) para evitar bufferbloat en nodos de desarrollo.
- **Observabilidad**: métricas por modelo/proveedor (latencia, success_rate, tokens, quota_remaining, availability, costo), logs estructurados, alertas proactivas ante caída de proveedor o agotamiento de cuota.

## Plan de entrega (Anexo A — fases para agentes)

> Estimaciones preliminares en semanas-persona, a refinar cuando se dimensione el equipo. El
> desarrollo es iterativo; las duraciones son rangos de planificación, no fechas comprometidas.
> Modo `para-agentes: true`: cada fase declara dependencias, alcance, tiempo estimado y un
> **resultado verificable** (criterio checable que marca la fase como completa).

| Fase | Duración (estimada) | Alcance | Depende de | Resultado verificable |
|---|---|---|---|---|
| Discovery | 1 semana | PRD, épicas, historias, priorización (este pipeline) + **ADR de stack backend** | — | Los 6 artefactos existen en `docs/` y pasan `/factory:revisar` + `/arch:revisar` sin bloqueantes |
| Construcción Core | 3-4 semanas | Gateway básico, Registry YAML, Router, Failover automático, Health Monitor (Circuit Breaker), AuthN, Adapters, Auditoría estática | Discovery | Un agente pide `coding` sin nombrar proveedor y recibe respuesta; ante 429 del primario la petición se completa vía el siguiente de la cadena; suite de contrato OpenAI verde |
| Gobernanza y Seguridad | 3-4 semanas | Quotas (PostgreSQL + RAM), AuthZ (RBAC), KMS, Dashboard, Redacción Síncrona | Construcción Core | Cuota agotada corta con 429; AuthZ por scope deniega con 403; ningún secreto en logs (auditoría redactada) |
| Inteligencia y Optimización | 4-6 semanas | Learning Engine + ajuste dinámico de pesos + MCP, Cache semántica | Gobernanza y Seguridad | El Learning Engine ajusta pesos del score con histórico y el cambio es auditable/reversible; cache semántica devuelve hit ≥ umbral de similitud |

## Definición de "hecho" (producto)

- [ ] Un agente solicita `coding`/`reasoning`/`vision`/`image`/`embedding`/`chat` sin nombrar proveedor y recibe respuesta.
- [ ] Petición sin `model` enruta por score; petición con `model` explícito usa ese modelo (o aplica fallback configurado).
- [ ] Ante 429/500/timeout del proveedor primario, la petición se completa vía el siguiente de la cadena sin error visible.
- [ ] Toda petición pasa AuthN + AuthZ por scope y queda en el log de auditoría; ningún secreto aparece en logs ni en config versionada.
- [ ] Free Claude Code y un cliente OpenAI-compat funcionan apuntando a la Gateway.
- [ ] Agregar un proveedor nuevo solo requiere Adapter + YAML + capacidades (sin tocar código de agentes).
- [ ] El Dashboard expone métricas y costos acumulados en tiempo real.
- [ ] La caché semántica intercepta repeticiones y reduce la latencia en al menos 20%.

## KPIs

> Baseline `N/A` = sistema no existe aún (greenfield); se mide desde el primer release.

| Métrica | Objetivo cubierto | Baseline | Objetivo | Frecuencia |
|---|---|---|---|---|
| Referencias a proveedor/modelo en código de agentes | Obj.1 Desacople | N/A | 0 referencias (escaneo estático en CI) | por release |
| Costo por 1k peticiones | Obj.3 Selección óptima | N/A | ≤ 70% del costo mono-proveedor equivalente | mensual |
| Tasa de Éxito de Failover | Obj.2 Resiliencia | N/A | ≥ 99% éxito global | continuo |
| Latencia Overhead (p95) | Obj.3 Selección óptima | N/A | < 100ms sobre LLM (exceptúa grandes payloads DLP) | continuo |
| Peticiones autenticadas + autorizadas + auditadas | Obj.4 Seguridad | N/A | 100% (0 accesos anónimos; 0 secretos en logs) | continuo |
| SLO Disponibilidad API | Obj.2 Resiliencia | N/A | ≥ 99.9% uptime | mensual |
| Clientes OpenAI-compat y Free Claude Code operativos | Obj.5 Compatibilidad | N/A | 2/2 contratos verdes en suite de contrato | por release |
| Tasa de aciertos en Cache Semántica (Hit Rate) | Obj.3 Selección óptima | N/A | ≥ 20% en entornos de alta repetición | mensual |

## Referencias

- Investigación de usuario: `docs/REQUERIMIENTO_GENERAL.md`, `docs/ARQUITECTURA.md`, `docs/MODELS.md`
- Mockups / wireframes: N/A (API-first)
- Benchmarks: LiteLLM, Portkey, OpenRouter, Vercel AI Gateway
- Docs técnicas adicionales: catálogo de modelos y proveedores en `docs/MODELS.md`

## Decisiones pendientes (open questions)

- (Ninguna) Todas las decisiones de infraestructura, base de datos (PostgreSQL), caché (Local RAM) y concurrencia (500 RPS) fueron resueltas y selladas en el Discovery y Arquitectura Técnica.
