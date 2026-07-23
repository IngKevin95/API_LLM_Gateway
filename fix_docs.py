import os
import re

base_dir = r"E:\Datos\Documentos\GitHub Personal\API_LLM_Gateway"

# 1. Update PRD
prd_path = os.path.join(base_dir, "docs", "01-prd", "api-llm-gateway.md")
with open(prd_path, "r", encoding="utf-8") as f:
    prd = f.read()

prd = prd.replace("40 historias", "44 historias")
prd = prd.replace("norm(C))", "norm(Q))")
prd = prd.replace("AWS, Vault, Local", "AWS KMS o HashiCorp Vault")

if "**Inteligencia y Optimización**" not in prd:
    opt_section = """
**Inteligencia y Optimización**
- **Caché Semántica**: Intercepta peticiones repetidas basándose en coincidencia exacta primero, para reducir latencia y costo.
- **Learning Engine**: Ajusta dinámicamente los pesos del score basándose en el historial de fallos y latencias reales observadas por proveedor.
- **Protocolo de Streaming SSE**: Las respuestas en stream deben apegarse estrictamente al formato Server-Sent Events (SSE).
"""
    prd = prd.replace("## UX y diseño", opt_section + "\n## UX y diseño")

with open(prd_path, "w", encoding="utf-8") as f:
    f.write(prd)

# 2. Update valor-esfuerzo
ve_path = os.path.join(base_dir, "docs", "05-priorizacion", "valor-esfuerzo-2026-07-19.md")
with open(ve_path, "r", encoding="utf-8") as f:
    ve = f.read()

ve = ve.replace("22 Must / 42", "24 Must / 44")
ve = ve.replace("6. Todo lo que no sea", "5. Todo lo que no sea")
with open(ve_path, "w", encoding="utf-8") as f:
    f.write(ve)

# 3. Create missing stories
h24 = """---
id: HU-024
title: Adapter para modelos locales (Ollama / vLLM / LM Studio)
epic: EP-002
type: Must
---

# HU-024: Adapter para modelos locales (Ollama / vLLM / LM Studio)

## INVEST
- [x] Independent: asume que existe un stub/mock de enrutamiento (HU-001).
- [x] Negotiable: el detalle de la API local puede variar, pero debe ser OpenAI-compatible.
- [x] Valuable: permite failover a modelos gratuitos sin costo y offline.
- [x] Estimable: complejidad clara al ser compatible con OpenAI.
- [x] Small: un solo adapter.
- [x] Testable: se puede probar con un contenedor de Ollama.

## Criterios de Aceptación (BDD)
| ID | Escenario | Dado (Given) | Cuando (When) | Entonces (Then) |
|---|---|---|---|---|
| 1 | Petición exitosa | Un servidor Ollama está corriendo | La Gateway enruta una petición de chat a Ollama | El Adapter se comunica y devuelve la respuesta en el formato estándar del Gateway |
| 2 | Timeout local | El servidor local está colgado | Se enruta una petición a un modelo local | Se aplica el TTFT o timeout dinámico y falla si se supera |
"""

h34 = """---
id: HU-034
title: Protección TCP contra ataques Slowloris
epic: EP-004B
type: Must
---

# HU-034: Protección TCP contra ataques Slowloris

## INVEST
- [x] Independent: puede implementarse de forma independiente en la capa de red del framework.
- [x] Negotiable: los timeouts exactos (ReadHeaderTimeout, WriteTimeout) pueden configurarse en YAML.
- [x] Valuable: protege al servidor contra ataques de agotamiento de conexiones (DoS).
- [x] Estimable: usar timeouts estándar del servidor HTTP.
- [x] Small: requiere configuración en el servidor HTTP base.
- [x] Testable: simulable con clientes lentos (ej. slowhttptest).

## Criterios de Aceptación (BDD)
| ID | Escenario | Dado (Given) | Cuando (When) | Entonces (Then) |
|---|---|---|---|---|
| 1 | Ataque de cabeceras lentas | Un cliente abre una conexión y envía cabeceras a 1 byte por segundo | El tiempo excede el ReadHeaderTimeout | El servidor cierra el socket devolviendo un 408 Request Timeout |
"""

h35 = """---
id: HU-035
title: Tokenizador de Contexto (Context Window)
epic: EP-001
type: Must
---

# HU-035: Tokenizador de Contexto y validación de buffer

## INVEST
- [x] Independent: lógica autocontenida que implementa una interfaz `ITokenizer`.
- [x] Negotiable: el algoritmo de conteo puede variar.
- [x] Valuable: evita envíos fallidos al proveedor si el texto excede el límite máximo del modelo.
- [x] Estimable: conteo de palabras heurístico o integraciones con librerías `tiktoken`.
- [x] Small: lógica de string manipulation / parsing.
- [x] Testable: contadores verificables mediante tests unitarios.

## Criterios de Aceptación (BDD)
| ID | Escenario | Dado (Given) | Cuando (When) | Entonces (Then) |
|---|---|---|---|---|
| 1 | Petición que excede ventana | Un payload tiene 120k tokens y el modelo soporta 100k | El router intenta validar el contexto | La validación falla y el router descarta este modelo del score devolviendo 400 Bad Request si no hay fallback |
"""

stories_dir = os.path.join(base_dir, "docs", "04-historias")
with open(os.path.join(stories_dir, "HU-024-adapter-locales.md"), "w", encoding="utf-8") as f: f.write(h24)
with open(os.path.join(stories_dir, "HU-034-proteccion-tcp.md"), "w", encoding="utf-8") as f: f.write(h34)
with open(os.path.join(stories_dir, "HU-035-tokenizador-contexto.md"), "w", encoding="utf-8") as f: f.write(h35)


# 4. Update specific stories

# HU-002a
p = os.path.join(stories_dir, "HU-002a-router-por-score.md")
with open(p, "r", encoding="utf-8") as f: text = f.read()
text = text.replace("usa el tokenizador específico de cada adapter", "usa una interfaz genérica `ITokenizer` o la validación recae sobre la capa de Adapter")
with open(p, "w", encoding="utf-8") as f: f.write(text)

# HU-027
p = os.path.join(stories_dir, "HU-027-prompt-optimizer.md")
with open(p, "r", encoding="utf-8") as f: text = f.read()
text = text.replace("Guardián de Prompts (Optimización Semántica)", "Guardián de Prompts (Seguridad y Jailbreak)")
text = text.replace("falla grácilmente", "bypasses optimization without exception")
with open(p, "w", encoding="utf-8") as f: f.write(text)

# HU-032
p = os.path.join(stories_dir, "HU-032-cache-semantica.md")
with open(p, "r", encoding="utf-8") as f: text = f.read()
text = text.replace("Cache semántica", "Cache Exacta (Primera Fase)")
text = text.replace("el cliente invisible", "el cliente")
with open(p, "w", encoding="utf-8") as f: f.write(text)

# HU-010
p = os.path.join(stories_dir, "HU-010-auditoria.md")
with open(p, "r", encoding="utf-8") as f: text = f.read()
lines = text.split("\\n")
lines = [l for l in lines if not ("Omisión de Prompts masivos" in l and "| 4 |" in l)]
text = "\\n".join(lines)
text = text.replace("comportamiento es explícito y probado", "devuelve HTTP 500 para fail-closed o encola con flag de riesgo para fail-open")
with open(p, "w", encoding="utf-8") as f: f.write(text)

# HU-006
p = os.path.join(stories_dir, "HU-006-cuota.md")
with open(p, "r", encoding="utf-8") as f: text = f.read()
if "| 5 | Race condition" in text:
    text = text.replace("| 5 | Race condition", "| 5 | Race condition initial |\\n| 6 | Post-generation token overshoot |")
with open(p, "w", encoding="utf-8") as f: f.write(text)

# HU-022, HU-026a, HU-012b, HU-001, HU-004c
replacements = [
    ("HU-022-rate-limiting.md", [("pasa sin demora", "routes request to provider"), ("| 5 | Sad path — Ataque Slowloris", "| 6 | Sad path — Ataque Slowloris")]),
    ("HU-026a-redaccion-sincrona.md", [("aborta por seguridad", "devuelve un error HTTP 500")]),
    ("HU-012b-openai-streaming.md", [("limpiamente", "emitiendo un evento de error SSE y cerrando el socket TCP")]),
    ("HU-001-registry-yaml.md", [("lo advierte en el log", "escribe un WARN level en el log")]),
    ("HU-004c-timeouts-dinamicos.md", [("umbral se relaja", "applies configured timeout_reasoning")])
]
for fname, reps in replacements:
    p = os.path.join(stories_dir, fname)
    if os.path.exists(p):
        with open(p, "r", encoding="utf-8") as f: text = f.read()
        for old, new in reps: text = text.replace(old, new)
        with open(p, "w", encoding="utf-8") as f: f.write(text)

print("Historias actualizadas.")

# 5. Update backlog and epicas (rewriting them securely to inject the 3 new stories and change priorities)
backlog_path = os.path.join(base_dir, "docs", "03-backlog", "backlog.md")
with open(backlog_path, "r", encoding="utf-8") as f: backlog_text = f.read()

# I will write a completely new string for backlog table to insert 24, 34, 35 in the MUST section, and move 29 up.
backlog_new = """# Backlog de API LLM Gateway

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
| 37 | HU-036 | Integrar OmniRoute API | EP-008 | Should | S | lista | 3 | HU-002a |
| 38 | HU-027 | Guardián de Prompts (Seguridad y Jailbreak) | EP-004A | Should | M | lista | 5 | HU-002a |
| 39 | HU-017 | Exponer métricas por modelo y proveedor | EP-007 | Could | M | lista | 4 | HU-007, HU-010 |
| 40 | HU-023 | API de Métricas para Dashboard | EP-007 | Could | M | lista | 4 | HU-017 |
| 41 | HU-018 | Registrar histórico de peticiones para aprendizaje | EP-007 | Could | M | lista | 4 | HU-010 |
| 42 | HU-019 | Ajustar pesos del score con histórico (Fase 1) | EP-007 | Could | M | lista | 4 | HU-018 |
| 43 | HU-032 | Cache Exacta (Primera Fase) | EP-007 | Could | M | lista | 4 | HU-002a |
| 44 | HU-025b | Autenticación mTLS | EP-004A | Could | M | lista | 4 | HU-008 |
| 45 | HU-033 | Integración MCP para multi-agentes | EP-005 | Could | M | lista | 4 | HU-002a |

## Trazabilidad épica → historias

- **EP-001 · Registro y Router Core**: HU-001, HU-002a, HU-002b, HU-003, HU-035
- **EP-002 · Resiliencia y Conectividad Base**: HU-004a, HU-004b, HU-004c, HU-005, HU-020a, HU-020b, HU-020c, HU-021a, HU-021b, HU-024
- **EP-003 · Gobernanza (cuota/costo)**: HU-006, HU-007
- **EP-004A · Seguridad empresarial**: HU-008, HU-009, HU-022, HU-022b, HU-025a, HU-025b, HU-027
- **EP-004B · Protección de datos**: HU-010, HU-011, HU-026a, HU-026b, HU-028, HU-034
- **EP-005 · API universal compatible**: HU-012a, HU-012b, HU-012c, HU-013, HU-016, HU-033
- **EP-007 · Observabilidad y aprendizaje**: HU-017, HU-018, HU-019, HU-023, HU-032
- **EP-008 · Adaptadores Secundarios**: HU-029, HU-030, HU-031, HU-036

## Conteo por estado

| Estado | Cantidad |
|---|---|
| lista | 45 |
| en-curso | 0 |
| hecha | 0 |
| draft | 0 |

Total: 45 historias (25 Must / 13 Should / 7 Could).

## Ideas sin desarrollar aún

- [ ] Generación de imágenes/vídeo (capacidad `image`/Seedance) como capability adicional — candidata a nueva HU
"""
with open(backlog_path, "w", encoding="utf-8") as f: f.write(backlog_new)

# 6. Update epicas.md
epicas_path = os.path.join(base_dir, "docs", "03-backlog", "epicas.md")
with open(epicas_path, "r", encoding="utf-8") as f: epicas_text = f.read()
epicas_text = epicas_text.replace("HU-027 — Guardián de Prompts (Optimización Semántica)", "")
if "HU-035 — Tokenizador" not in epicas_text:
    epicas_text = epicas_text.replace("HU-003 — Forzar modelo explícito vía parámetro `model` con política de fallback", "HU-003 — Forzar modelo explícito vía parámetro `model` con política de fallback\\n- HU-035 — Tokenizador de Contexto (Context Window)")

epicas_text = epicas_text.replace("HU-021b — Adapter Anthropic (streaming)", "HU-021b — Adapter Anthropic (streaming)\\n- HU-024 — Adapter para modelos locales (Ollama / vLLM / LM Studio)")
epicas_text = epicas_text.replace("HU-025b — Autenticación mTLS", "HU-025b — Autenticación mTLS\\n- HU-027 — Guardián de Prompts (Seguridad y Jailbreak)")
epicas_text = epicas_text.replace("HU-028 — Cifrado de Sobre (Envelope) en BD de Auditoría", "HU-028 — Cifrado de Sobre (Envelope) en BD de Auditoría\\n- HU-034 — Protección TCP contra ataques Slowloris")
epicas_text = epicas_text.replace("HU-023 — API de Métricas para Dashboard", "HU-023 — API de Métricas para Dashboard\\n- HU-032 — Cache Exacta (Primera Fase)")
epicas_text = epicas_text.replace("HU-032 — Cache semántica\\n", "")

with open(epicas_path, "w", encoding="utf-8") as f: f.write(epicas_text)
print("Archivos de épicas y backlog actualizados.")
