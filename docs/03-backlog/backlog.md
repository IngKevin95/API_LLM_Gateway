# Backlog de API LLM Gateway

**Framework de priorización**: Valor / Esfuerzo · **Actualizado**: 2026-07-20

> El orden de las filas es la fuente de verdad de la priorización vigente. Los criterios que sustentan ese orden (cuadrantes Valor/Esfuerzo y su derivación a Must/Should/Could) están documentados en `docs/05-priorizacion/valor-esfuerzo-2026-07-19.md`, no aquí.

## Regla cuadrante Valor/Esfuerzo → prioridad (MoSCoW)

- **Must** = camino crítico del producto: sin ellas no hay sistema funcional (core de enrutamiento, adapters + endpoints OpenAI-compat, AuthN, failover, auditoría, CLI + Free Claude Code).
- **Should** = alto valor pero no bloquea la arquitectura base: gobernanza (cuota/costo), AuthZ, adapters adicionales, tools de CLI, seguridad avanzada, guardián de prompts.
- **Could** = valor diferido o complejidad alta: observabilidad, dashboard, learning engine, cache semántica, MCP, mTLS, degradación local dedicada.

## Tabla

| Orden | ID | Historia | Épica | Prioridad | Talla | Estado | # AC | Bloqueos |
|---|----|----------|-------|-----------|-------|--------|------|----------|
| 1 | HU-001 | Cargar providers/models/routing desde YAML (Registry) | EP-001 | Must | M | lista | 5 | — |
| 2 | HU-002a | Resolver capacidad a modelo por score (Router automático) | EP-001 | Must | S | lista | 3 | HU-001 |
| 3 | HU-002b | Manejo de errores y desempates en el enrutamiento | EP-001 | Must | S | lista | 3 | HU-002a |
| 4 | HU-003 | Forzar modelo explícito con política de fallback | EP-001 | Must | S | lista | 4 | HU-001 |
| 5 | HU-035 | Tokenizador de Contexto (Context Window) | EP-001 | Must | M | lista | 1 | — |
| 6 | HU-008 | Autenticar toda petición con API key | EP-004A | Must | S | lista | 4 | HU-001 |
| 7 | HU-020a | Adapter OpenAI — chat y tool calling | EP-002 | Must | S | lista | 3 | HU-002a |
| 8 | HU-029 | Integrar AIHubMix API (Local/Economía) | EP-008 | Must | S | lista | 3 | HU-002a |
| 9 | HU-024 | Adapter para modelos locales (Ollama / vLLM / LM Studio) | EP-002 | Must | M | lista | 2 | HU-002a |
| 10 | HU-012a | Endpoint OpenAI-compat de chat (sin streaming) | EP-005 | Must | M | lista | 4 | HU-002a, HU-020a |
| 11 | HU-020b | Adapter OpenAI — streaming SSE | EP-002 | Must | S | lista | 3 | HU-020a |
| 12 | HU-012b | Streaming SSE compatible OpenAI | EP-005 | Must | S | lista | 5 | HU-012a, HU-020b |
| 13 | HU-020c | Adapter OpenAI — embeddings | EP-002 | Must | S | lista | 3 | HU-020a |
| 14 | HU-012c | Endpoint OpenAI-compat de embeddings | EP-005 | Must | S | lista | 4 | HU-002a, HU-020c |
| 15 | HU-021a | Adapter Anthropic — chat, roles y tool calling | EP-002 | Must | M | lista | 5 | HU-002a |
| 16 | HU-021b | Adapter Anthropic — streaming | EP-002 | Must | S | lista | 3 | HU-021a |
| 17 | HU-013 | Endpoint Anthropic-compat para Free Claude Code | EP-005 | Must | M | lista | 4 | HU-002a, HU-021a |
| 18 | HU-004a | Failover básico de cadena con degradación a local | EP-002 | Must | M | lista | 5 | HU-002a |
| 19 | HU-004b | Circuit Breaker pasivo y Max In-Flight | EP-002 | Must | M | lista | 3 | HU-004a |
| 20 | HU-004c | Timeouts dinámicos por capacidad y Stream Idle Timeout | EP-002 | Must | M | lista | 3 | HU-004a |
| 21 | HU-010 | Guardar traza inmutable (Auditoría) | EP-004B | Must | M | lista | 3 | HU-001 |
| 22 | HU-026a | Redacción Síncrona de Secretos | EP-004B | Must | S | lista | 4 | HU-010 |
| 23 | HU-022 | Rate Limiting y protección de Payload | EP-004A | Must | M | lista | 2 | HU-001 |
| 24 | HU-034 | Protección TCP contra ataques Slowloris | EP-004B | Must | S | lista | 1 | — |
| 25 | HU-016 | Configurar Free Claude Code contra la Gateway | EP-005 | Must | S | lista | 4 | HU-013 |
| 26 | HU-005 | Health checks periódicos con retiro y reactivación | EP-002 | Should | S | lista | 5 | HU-001 |
| 27 | HU-009 | Autorizar por scope/RBAC y tenant | EP-004A | Should | M | lista | 5 | HU-008 |
| 28 | HU-011 | Gestionar y rotar secretos sin exponerlos | EP-004B | Should | M | lista | 4 | HU-001 |
| 29 | HU-006 | Contabilizar y limitar consumo por cuota | EP-003 | Should | M | lista | 5 | HU-001, HU-008 |
| 30 | HU-007 | Registrar costo por petición/agente/proveedor | EP-003 | Should | S | lista | 5 | — |
| 31 | HU-022b | Límite de concurrencia y enrutamiento de red para vision | EP-004A | Should | S | lista | 3 | HU-022 |
| 32 | HU-025a | Autenticación OAuth2/OIDC | EP-004A | Should | M | lista | 4 | HU-008 |
| 33 | HU-026b | Kill-Switch Asíncrono de PII | EP-004B | Should | M | lista | 4 | HU-026a |
| 34 | HU-028 | Cifrado de Sobre (Envelope) en BD de Auditoría | EP-004B | Should | M | lista | 5 | HU-010 |
| 35 | HU-030 | Adapter para Google (Gemini) | EP-008 | Should | M | lista | 4 | HU-002a |
| 36 | HU-031 | Integrar OpenRouter API (Fallback/Long-tail) | EP-008 | Should | S | lista | 3 | HU-002a |
| 37 | HU-027 | Guardián de Prompts (Seguridad y Jailbreak) | EP-004A | Should | M | lista | 5 | HU-002a |
| 38 | HU-017 | Exponer métricas por modelo y proveedor | EP-007 | Could | M | lista | 4 | HU-007, HU-010 |
| 39 | HU-023 | API de Métricas para Dashboard | EP-007 | Could | M | lista | 4 | HU-017 |
| 40 | HU-018 | Registrar histórico de peticiones para aprendizaje | EP-007 | Could | M | lista | 4 | HU-010 |
| 41 | HU-019 | Ajustar pesos del score con histórico (Fase 1) | EP-007 | Could | M | lista | 4 | HU-018 |
| 42 | HU-032 | Cache Exacta (Primera Fase) | EP-007 | Could | M | lista | 4 | HU-002a |
| 43 | HU-025b | Autenticación mTLS | EP-004A | Could | M | lista | 4 | HU-008 |
| 44 | HU-033 | Integración MCP para multi-agentes | EP-005 | Could | M | lista | 4 | HU-002a |

## Trazabilidad épica → historias

- **EP-001 · Registro y Router Core**: HU-001, HU-002a, HU-002b, HU-003, HU-035
- **EP-002 · Resiliencia y Conectividad Base**: HU-004a, HU-004b, HU-004c, HU-005, HU-020a, HU-020b, HU-020c, HU-021a, HU-021b, HU-024
- **EP-003 · Gobernanza (cuota/costo)**: HU-006, HU-007
- **EP-004A · Seguridad empresarial**: HU-008, HU-009, HU-022, HU-022b, HU-025a, HU-025b, HU-027
- **EP-004B · Protección de datos**: HU-010, HU-011, HU-026a, HU-026b, HU-028, HU-034
- **EP-005 · API universal compatible**: HU-012a, HU-012b, HU-012c, HU-013, HU-016, HU-033
- **EP-007 · Observabilidad y aprendizaje**: HU-017, HU-018, HU-019, HU-023, HU-032
- **EP-008 · Adaptadores Secundarios**: HU-029, HU-030, HU-031

## Conteo por estado

| Estado | Cantidad |
|---|---|
| lista | 44 |
| en-curso | 0 |
| hecha | 0 |
| draft | 0 |

Total: 44 historias (25 Must / 12 Should / 7 Could).

## Ideas sin desarrollar aún

- [ ] Generación de imágenes/vídeo (capacidad `image`/Seedance) como capability adicional — candidata a nueva HU
