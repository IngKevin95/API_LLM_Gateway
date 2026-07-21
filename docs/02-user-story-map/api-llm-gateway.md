# Story Map de API LLM Gateway

Formato Jeff Patton: columnas = actividades del usuario en orden cronológico (backbone); filas = corte por prioridad/release.

Fuentes: `docs/01-prd/api-llm-gateway.md`, `docs/03-backlog/epicas.md`. Los IDs HU-XXX están
**reservados/anticipados** aquí; el detalle con AC en G/W/T se escribe en `docs/04-historias/`.
Nota: se desglosan 40 historias de usuario organizadas por épicas y releases, diseñadas 100% bajo metodología BDD e INVEST.

## Backbone

`1. Conectarse y autenticarse → 2. Enviar petición (auto / modelo) → 3. Resolver modelo por capacidad → 4. Ejecutar con resiliencia → 5. Registrar, gobernar y observar → 6. Operar como agente (CLI / Free Claude Code)`

## Tablero

| Release | 1. Conectarse y autenticarse | 2. Enviar petición | 3. Resolver modelo | 4. Ejecutar con resiliencia | 5. Registrar, gobernar y observar | 6. Operar como agente |
|---|---|---|---|---|---|---|
| **Release 1.0 (Producto Completo)** | HU-008 Auth API key, HU-022 Rate Limiting, HU-022b Visión Concurrencia, HU-009 AuthZ, HU-025a OAuth2, HU-025b mTLS | HU-012a chat, HU-012b streaming, HU-012c embeddings, HU-013 Anthropic-compat | HU-001 Registry YAML, HU-002a Router, HU-002b Errores, HU-003 explícito, HU-027 Guardián Prompts, HU-032 Cache semántica | HU-020a/b/c OpenAI, HU-021a/b Anthropic, HU-029 AIHubMix, HU-030 Google, HU-031 OpenRouter, HU-004a/b/c Failover, HU-024 Locales, HU-005 Health Monitor | HU-010 Auditoría, HU-026a Redacción Síncrona, HU-006 Cuota, HU-007 Costo, HU-011 Secretos, HU-026b Kill-Switch PII, HU-028 KMS, HU-017 Métricas, HU-018 Histórico, HU-019 Learning Engine, HU-023 Dashboard | HU-016 Free Claude Code, HU-033 Integración MCP |

## Alcance del Entregable Único (Release 1.0)

- **Slice vertical completo**: El producto recorre de forma madura las 6 actividades del backbone. Un consumidor se autentica (API Key, OAuth2, mTLS) (HU-008/009/025a/025b), envía peticiones protegidas con Rate Limiting (HU-022), el Router elige el modelo óptimo con soporte de Cache Semántica y Guardián (HU-001/002/027/032), se ejecuta con failover avanzado y degradación (HU-004a/b/c/005), se aplican controles de cuota/costo, KMS y redacción de secretos (HU-006/007/010/011/026/028), y se visualiza a través de un Dashboard y Learning Engine (HU-017-019/023). Finalmente se opera desde entornos multi-agente vía MCP e IDEs (HU-016/033).
- **Prueba el corazón del proyecto**: Valida al 100% el desacople total, resiliencia estricta y seguridad exhaustiva, que son los objetivos núcleo del PRD. 
- **Entrega de valor total**: Se ha consolidado la iteración para que el primer release productivo sea un API LLM Gateway de nivel empresarial (enterprise-ready) que no requiera parches de seguridad posteriores.

1. **Conectarse y autenticarse** — el consumidor (agente/app/CLI) alcanza la Gateway y prueba su identidad mediante API key, OAuth2/OIDC, mTLS, operando bajo validación estricta de scopes/RBAC, rate limit y multi-tenant.
2. **Enviar petición** — el consumidor manda su prompt con contrato compatible (OpenAI o Anthropic). Puede no indicar modelo (modo automático) o forzar uno (`model`).
3. **Resolver modelo por capacidad** — el Router traduce la capacidad requerida al modelo óptimo por score, ejecutando el guardián de prompts o consultando la caché semántica si aplica.
4. **Ejecutar con resiliencia** — el Adapter llama al proveedor; si falla (429/500/timeout) el failover proactivo (Health Monitor) pasa al siguiente de la cadena, terminando en la degradación a modelos locales si es necesario.
5. **Registrar, gobernar y observar** — cada petición se audita cifrada con KMS (con redacción de secretos asíncrona y síncrona), se contabiliza contra cuota y costo, y alimenta la telemetría (Dashboard) y el Learning Engine para futuros enrutamientos.
6. **Operar como agente** — el usuario trabaja a través de un CLI de agente (con acceso a archivos, bash y git), de Free Claude Code y de integraciones de orquestación multi-agente (MCP).

- **Actividad 3 (Resolver modelo)**: Al consolidar las fases, el Router nace desde el día 1 con la capacidad de usar métricas/histórico (Learning Engine), eliminando el hueco del score fijo.
- **Actividad 4 (Resiliencia)**: El Health Monitor proactivo (HU-005) se implementa en la construcción del núcleo, garantizando que el failover dependa de un estado proactivo.
- **Actividad 5 (Gobernar)**: La validación atómica en memoria (Quota Manager) entra al alcance base, por lo que no existe riesgo de exceder cuotas gratuitas en la primera versión.
