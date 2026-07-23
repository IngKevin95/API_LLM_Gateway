# EP-010: Compatibilidad Universal de Clientes (Master Plan)

**Estado**: 📋 Documentación Completa (Draft)  
**Fase**: Diseño e Especificación  
**Objetivo**: Soporte completo para 8 herramientas diferentes apuntando al Gateway

---

## 🎯 Visión

Transformar el Gateway de un LLM-switcher para 2 clientes (OpenWebUI, Free Claude Code) a una **solución universal** que habilita 8 herramientas diferentes para consumir LLMs vía una interfaz común, con **routing automático por capacidad** y **traducción de parámetros**.

```
ANTES (EP-005):
OpenWebUI        Free Claude Code
    ↓                 ↓
/v1/chat         /v1/messages
(OpenAI-compat)  (Anthropic)
    ↓                 ↓
         Gateway
            ↓
      LLM Pool

DESPUÉS (EP-010):
OpenWebUI  OpenCode  Claude Code  OpenHands  OpenClaw  CrewAI  UI-TARS
     ↓        ↓          ↓           ↓         ↓         ↓        ↓
  /v1/chat /responses /v1/messages /v1/chat  /v1/chat  /v1/chat /v1/chat
     ↓        ↓          ↓           ↓         ↓         ↓        ↓
Middleware (Format Auto-Detect + Parameter Normalization)
              ↓↓↓ Internal Normalized Format ↓↓↓
                    Gateway Router
                         ↓
                    LLM Pool
```

---

## 📊 Composición de la Épica

**7 Historias de Usuario** (5 Must + 2 Should)

| # | HU-ID | Título | Prioridad | Talla | Bloqueos | AC |
|---|-------|--------|-----------|-------|----------|-----|
| 1 | **HU-042** | Routing automático por capability (modelo implícito) | **Must** | M | HU-002a | 5 |
| 2 | **HU-043** | Endpoint /responses para OpenCode (Responses API) | **Must** | M | HU-002a, HU-020a | 5 |
| 3 | **HU-044** | Parámetros OpenAI completos (temperature, top_p, etc.) | **Must** | M | HU-012a | 6 |
| 4 | **HU-045** | Parámetros Anthropic completos (temperature, top_k, etc.) | **Must** | M | HU-013 | 6 |
| 5 | **HU-048** | Documentación configuración reproducible por herramienta | **Must** | S | HU-042-045 | 6 |
| 6 | **HU-046** | Endpoint /v1/models mejorado con metadata | Should | S | HU-002a | 6 |
| 7 | **HU-047** | Middleware de normalización automática de formatos | Should | M | HU-042-045 | 6 |

---

## 🚀 Roadmap de Implementación

### Fase 1: Núcleo + Documentación (2-3 semanas)
**Objetivo**: Todas las herramientas funcionales básicamente

- ✅ HU-042 — Routing automático (enable `model: "router:coding"` or no model)
- ✅ HU-043 — Endpoint /responses (OpenCode)
- ✅ HU-044 — Parámetros OpenAI (temperature, top_p, etc.)
- ✅ HU-045 — Parámetros Anthropic (temperature, top_k, thinking, etc.)
- ✅ HU-048 — Setup guides para todas las 8 herramientas

**Test coverage**:
- 8 clientes funcionan apuntando a `http://localhost:8080`
- Cada cliente soporta su formato nativo sin cambios
- Parámetros completos traducidos sin error

### Fase 2: Polish + Observabilidad (1 semana)
**Objetivo**: Experiencia completa y troubleshooting fácil

- ✅ HU-046 — `/v1/models` con metadata (debugging + eligibility checks)
- ✅ HU-047 — Middleware de normalización (tolerancia con variaciones de formato)

**Test coverage**:
- Clientes con pequeñas variaciones de formato funcionan (robustez)
- `/v1/models` consultable desde cada cliente
- Logs informativos para troubleshooting

---

## 📋 Criterios de Aceptación Consolidados

### Criterio 1: Routing Automático Funcional ✓
```
$ curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "router:coding", "messages": [...]}'
→ Gateway elige mejor modelo de cadena "coding" automáticamente
```

### Criterio 2: OpenCode Via /responses ✓
```
$ curl -X POST http://localhost:8080/responses \
  -d '{"model": "gpt-5", "input": [...], "reasoning_effort": "medium"}'
→ Respuesta en formato Responses API
```

### Criterio 3: Parámetros Completos ✓
```
Clientes pueden enviar: temperature, top_p, seed, tool_choice, etc.
→ Parámetros traducidos/mapeados a cada adapter sin error
```

### Criterio 4: 8 Clientes Funcionando ✓
| Cliente | Config | Endpoint | Test |
|---------|--------|----------|------|
| OpenWebUI | OPENAI_BASE_URL | /v1/chat/completions | ✓ |
| OpenCode | OPENCODE_BASE_URL | /responses | ✓ |
| Free Claude Code | ANTHROPIC_BASE_URL | /v1/messages | ✓ |
| Claude Code | ANTHROPIC_BASE_URL | /v1/messages | ✓ |
| OpenHands | LLM_BASE_URL | /v1/chat/completions | ✓ |
| OpenClaw | custom provider | /v1/chat/completions | ✓ |
| CrewAI | base_url param | /v1/chat/completions | ✓ |
| UI-TARS | deployed locally | /v1/chat/completions | ✓ |

### Criterio 5: Guías Reproducibles ✓
Cada herramienta tiene setup guide completo:
- Env vars exactas
- Config.yaml snippet
- Curl test ejemplo
- Troubleshooting section

---

## 🔧 Dependencias y Bloqueos

```
Bloqueadores de EP-010:
├── EP-001 (Router básico) — DONE ✓
├── EP-002 (Adapters OpenAI/Anthropic) — DONE ✓
├── EP-005 (Endpoints OpenAI + Anthropic) — DONE ✓
└── Nada más = 0 bloques externos
```

Bloqueadores INTERNOS:
```
HU-042 (Routing automático)
  ↓ (necesario para)
HU-043, HU-044, HU-045 (endpoints + parámetros)
  ↓ (juntos alimentan)
HU-048 (documentación)

HU-046, HU-047 (polish)
  ↓ (dependen de)
HU-044, HU-045 (parámetros resueltos)
```

---

## 📝 Estimación de Esfuerzo

| HU | Estimación | Ríes | Notas |
|----|-----------|-----|--------|
| HU-042 | 6 horas | Bajo | Handler + Router.Resolve, parsing de `router:` |
| HU-043 | 8 horas | Bajo | Nuevo endpoint handler, conversión Responses ↔ internal |
| HU-044 | 6 horas | Bajo | Ampliar tipos + validación (ranges, mapeos) |
| HU-045 | 6 horas | Bajo | Tipos Anthropic + mapeos a adapters |
| HU-048 | 5 horas | Muy bajo | Escritura de docs (no código, 8 guías) |
| HU-046 | 4 horas | Muy bajo | Ampliar handler existente + queries |
| HU-047 | 8 horas | Medio | Detector heurístico + mapeador de parámetros |

**Total**: ~43 horas = **5-6 días de un dev** (con testing)

---

## 🧪 Estrategia de Testing

### Test unitarios por HU
- HU-042: routing con `router:*`, fallback, capability inexistente
- HU-043: request Responses básica, streaming, reasoning_effort
- HU-044: validación de ranges (temperature 0-2, top_p 0-1), parámetros múltiples
- HU-045: temperatura, top_k, thinking, tool_use, fallback sin thinking
- HU-046: listar modelos, filtrar por capability, sorting
- HU-047: auto-detect OpenAI/Anthropic/Responses, mapeo de parámetros desconocidos
- HU-048: ejecutar curl/Python de cada guía

### Test de integración (end-to-end)
```bash
# Setup: config.yaml con providers
# Test 1: OpenWebUI cliente (OpenAI-format) → Anthropic backend
# Test 2: OpenCode cliente (Responses-format) → OpenAI backend
# Test 3: Cada cliente con 1-2 parámetros especiales
# Resultado: todos ✓ sin cambios de cliente code
```

### Test de conformidad (matriz de cliente × parámetro)
|  | temperature | top_p | seed | tool_choice | thinking |
|---|---|---|---|---|---|
| OpenWebUI | ✓ | ✓ | ✓ | ✓ | - |
| OpenCode | ✓ | ✓ | ✓ | ✓ | ✓ |
| Claude Code | ✓ | - | - | ✓ | ✓ |
| OpenHands | ✓ | ✓ | ✓ | ✓ | - |
| OpenClaw | ✓ | ✓ | ✓ | ✓ | - |
| CrewAI | ✓ | ✓ | ✓ | ✓ | - |
| UI-TARS | ✓ | ✓ | ✓ | ✓ | - |

(✓ = supported, - = client doesn't support natively)

---

## 📚 Documentación a Crear

En `docs/14-guides/`:
- `openwebui-gateway-setup.md` (HU-048)
- `opencode-gateway-setup.md` (HU-048)
- `free-claude-code-gateway-setup.md` (HU-048)
- `claude-code-gateway-setup.md` (HU-048)
- `openhands-gateway-setup.md` (HU-048)
- `openclaw-gateway-setup.md` (HU-048)
- `crewai-gateway-setup.md` (HU-048)
- `ui-tars-gateway-setup.md` (HU-048)
- `GATEWAY_CLIENTS.md` (comparativa master)

Cada guía incluye:
- Instalación/setup exacto
- Config.yaml snippet
- Env vars necesarias
- Curl/Python test ejemplo funcional
- Troubleshooting section
- Links a HUs relevantes

---

## 🎓 Success Criteria (Definition of Done)

✅ Todos los AC de HU-042 a HU-048 pasados  
✅ Tests de integración verde (8 clientes × 2 parámetros mín)  
✅ Documentación reproducible (ejecutar curl de cada guía = ✓)  
✅ No breaking changes en EP-005 (backwards compatible)  
✅ Logs informativos (WARN para parámetros ignorados, etc.)  
✅ Code review pasado (security + design patterns)  

---

## 📈 Métricas de Impacto

**Antes**: 2 clientes soportados (OpenWebUI, Free Claude Code)  
**Después**: 8 clientes soportados (↑300% expansión)

**Antes**: Parámetros básicos solamente  
**Después**: Parámetros completos (temperature, top_p, seed, tool_choice, thinking, etc.)

**Antes**: Cliente debe conocer modelo concreto  
**Después**: Cliente elige por capacidad (router:coding) = desacoplamiento completo

**Resultado Final**: Gateway verdaderamente universal, cero fricción de integración.

---

## 🚦 Próximos Pasos

1. ✅ Completar esta especificación (Done - you're here)
2. ⏭️ Revisión de especificación (AC checklist, bloqueos, estimación)
3. ⏭️ Construcción fase 1 (HU-042 → HU-048, todos Must)
4. ⏭️ Testing e2e (matrix de 8 clientes)
5. ⏭️ Construcción fase 2 (HU-046, HU-047 - polish)
6. ⏭️ Finalizar guides + PR

---

**Documento actualizado**: 2026-07-23  
**Status**: 📋 Listo para construcción  
**Next action**: Revisión de especificación + inicio de HU-042
