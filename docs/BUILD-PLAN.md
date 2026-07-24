# Plan de Construcción — Fase 1 MVP Fixes

**Fecha**: 2026-07-23  
**Estado**: Planning  
**Objetivo**: Hacer el Gateway funcional end-to-end (health → chat/completions → OmniRoute)

---

## Mapeo Issues → Épicas

| # | Issue | Épica | Prioridad | Esfuerzo | AC |
|---|---|---|---|---|---|
| 1 | Debug `/v1/chat/completions` (HTTP 500) | **EP-011** | CRÍTICA | 2h | ✓ Responde 200 con choice valido |
| 2 | Crear adaptador OmniRoute | **EP-012** | CRÍTICA | 1h | ✓ OmniRoute aparece en adapters map |
| 3 | Normalizar IDs proveedores | **EP-013** | IMPORTANTE | 30m | ✓ config.yaml usa IDs que buildAdapters entiende |
| 4 | Fix `/v1/embeddings` | **EP-014** | IMPORTANTE | 1.5h | ✓ Responde 200 con embeddings array |
| 5 | Fix `/v1/messages` (Anthropic) | **EP-015** | IMPORTANTE | 1h | ✓ Responde 200 con content |
| 6 | Añadir logging de debug en handlers | **EP-016** | IMPORTANTE | 1h | ✓ Errores loguean stack trace |
| 7 | Implementar `/metrics` real | **EP-017** | Opcional | 2h | ✓ Devuelve latency, tokens, providers |

---

## Slice #1: EP-011 — Debug `/v1/chat/completions`

### Descripción
El endpoint `/v1/chat/completions` devuelve HTTP 500. Necesario debuguear, loguear, y fijar el error interno.

### Historias (HU)
- **HU-050** Añadir logging a OpenAI handler  
- **HU-051** Debug y fix del GatewayProcessor.Handle()  
- **HU-052** Validar que Router.Route() retorna proveedor válido

### Criterios de Aceptación
```gherkin
Given: POST /v1/chat/completions con modelo "gpt-4o"
When: envío request válido OpenAI
Then: recibo HTTP 200 con choice.message.content no vacío

And: logs muestran qué proveedor se usó
And: latencia < 5s
```

### Archivos a Modificar
- `src/internal/api/openai/handler.go` — añadir logging
- `src/cmd/gateway/main.go` — GatewayProcessor.Handle()
- `src/internal/router/router.go` — logging en Route()

### Comandos para Testing
```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"test"}]}'
```

---

## Slice #2: EP-012 — Adaptador OmniRoute

### Descripción
OmniRoute existe en config.yaml pero no tiene adaptador implementado. Crear adaptador compatible con OpenAI API.

### Historias (HU)
- **HU-053** Crear `internal/adapter/omniroute/omniroute.go`  
- **HU-054** Registrar adaptador en buildAdapters()  
- **HU-055** Test de conectividad OmniRoute → Gateway

### Criterios de Aceptación
```gherkin
Given: adaptador OmniRoute registrado
When: Router elige OmniRoute para "chat"
Then: llamada llega a http://omniroute:20128/v1/chat/completions
And: respuesta se parsea correctamente
```

### Archivos a Crear/Modificar
- `src/internal/adapter/omniroute/omniroute.go` (new)
- `src/cmd/gateway/main.go` buildAdapters()

---

## Slice #3: EP-013 — Normalizar IDs Proveedores

### Descripción
Alinear IDs en config.yaml con lo que buildAdapters() espera:
- `google-gemini` → `google`
- `local-ollama` → `local`
- Remover proveedores sin adaptador (openrouter, etc.)

### Criterios de Aceptación
```gherkin
Given: config.yaml con IDs normalizados
When: Gateway carga config
Then: todos los proveedores en config tienen adaptador en buildAdapters()
And: no hay WARNING "sin adapter para proveedor"
```

### Archivos a Modificar
- `src/config.yaml`

---

## Slice #4: EP-014 — Implementar `/v1/embeddings`

### Descripción
Endpoint está registrado pero no funciona (HTTP 503). Implementar usando adapters.

### Criterios de Aceptación
```gherkin
Given: POST /v1/embeddings con modelo "text-embedding-3-small"
When: envío input "test"
Then: recibo HTTP 200
And: response.data es array de embeddings
And: cada embedding.vector.length == 1536 (dimensiones esperadas)
```

---

## Slice #5: EP-015 — Fix `/v1/messages` (Anthropic)

### Descripción
Handler Anthropic devuelve HTTP 400. Fix en validación y mapeo.

### Criterios de Aceptación
```gherkin
Given: POST /v1/messages con modelo "claude-opus-4"
When: envío {"messages": [{"role":"user","content":"test"}]}
Then: recibo HTTP 200
And: response.content[0].text no vacío
```

---

## Slice #6: EP-016 — Logging Estructurado

### Descripción
Añadir logging de debug en todos los handlers para facilitar troubleshooting.

### Criterios de Aceptación
- Logs muestran: entrada request, proveedor elegido, latencia, salida
- Errores muestran stack trace completo
- JSON structured logs (compatible con ELK)

---

## Slice #7: EP-017 — Métricas Reales (Opcional)

### Descripción
`/metrics` devuelve `{}`. Implementar métricas operacionales.

---

## Cronograma

| Slice | Esfuerzo | Inicio | Fin | Gate |
|---|---|---|---|---|
| EP-011 (chat fix) | 2h | HOY | +2h | ✓ Test OK |
| EP-012 (OmniRoute) | 1h | +2h | +3h | ✓ Test OK |
| EP-013 (IDs) | 30m | +3h | +3.5h | ✓ No warnings |
| EP-014 (embeddings) | 1.5h | +3.5h | +5h | ✓ Test OK |
| EP-015 (messages) | 1h | +5h | +6h | ✓ Test OK |
| EP-016 (logging) | 1h | +6h | +7h | ✓ Logs visible |
| EP-017 (metrics) | 2h | Opcional | | ✓ Test OK |

**Total**: ~9.5h (1 día de trabajo)

---

## Documentación Lineal

Cada slice generará:
1. Código en `src/`
2. Tests en `*_test.go`
3. Entry en `docs/04-historias/HU-0XX.md` (traceado a épica)
4. Commit message linealizado (issue → HU → código)
5. Build-state.json actualizado

**Referencia**: `.claude/state/build-state.json` mantiene estado pipeline.

---

## GO/NO-GO Criteria para MVP

- ✓ `/health` devuelve 200
- ✓ `/v1/chat/completions` devuelve 200 con choice válido
- ✓ Gateway enruta a OmniRoute cuando está disponible
- ✓ Failover a AIHubMix cuando OmniRoute cae
- ✓ Logs muestran decisión de routing
- ✓ config.yaml es válido (sin warnings)

---

## Stack Actual

| Componente | Tech | Status |
|---|---|---|
| Gateway | Go 1.22 | Partial (health OK, handlers en construcción) |
| OmniRoute | Node.js Docker | Up |
| Config | YAML | Valid (pero IDs desalineados) |
| Tests | Bash + curl | Ready |
| Docs | Markdown | Ready |

