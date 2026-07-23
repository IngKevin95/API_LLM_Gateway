---
id: HU-048
titulo: Documentación de configuración reproducible por herramienta (setup guides)
epica: EP-010
prioridad: Must
complejidad: S
estado: draft
---

# Documentación de configuración reproducible por herramienta (setup guides)

Como **usuario final o integrador**, quiero **guías claras y reproducibles para cada herramienta (OpenWebUI, OpenCode, Claude Code, OpenHands, OpenClaw, CrewAI, UI-TARS)**, para **configurar el Gateway sin adivinar o trial-and-error**.

Contexto: Cierra EP-010 documentando cómo configurar cada cliente. Actividad 7 de EP-010 (documentación final).

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — documento para OpenWebUI | Dado que usuario quiere usar OpenWebUI contra el Gateway | Cuando lee `docs/14-guides/openwebui-gateway-setup.md` | Entonces encuentra: config.yaml ejemplo, env vars necesarias, pasos exactos de setup, comando de test curl, troubleshooting |
| 2 | Happy — documento para OpenCode | Dado que usuario quiere usar OpenCode | Cuando lee `docs/14-guides/opencode-gateway-setup.md` | Entonces encuentra: endpoint /responses, OPENCODE_BASE_URL config, reasoning_effort usage, ejemplo completo |
| 3 | Happy — documento para Free Claude Code | Dado que usuario tiene Free Claude Code instalado | Cuando lee `docs/14-guides/free-claude-code-gateway-setup.md` | Entonces encuentra: ANTHROPIC_BASE_URL, API key setup, diferencia con Anthropic directo, troubleshooting |
| 4 | Happy — documento para CrewAI | Dado que usuario corre scripts CrewAI | Cuando lee `docs/14-guides/crewai-gateway-setup.md` | Entonces encuentra: LLM class instantiation, base_url config, Python example, multi-agent setup |
| 5 | Happy — documento master con todos | Dado que usuario quiere visión de alto nivel | Cuando lee `docs/14-guides/GATEWAY_CLIENTS.md` | Entonces encuentra: tabla comparativa de 8 herramientas, endpoints soportados, parámetros traducidos, matriz de compatibilidad |
| 6 | Happy — verificación de guías | Dado que técnico ejecuta scripts de verificación en cada guía | Cuando corre los comandos curl/Python de ejemplo | Entonces recibe respuestas exitosas sin cambios de código |

## Checklist INVEST

- [x] Independent — pura documentación, no código; no bloquea otras HUs
- [x] Negotiable — alcance: 7 guías por cliente + 1 master
- [x] Valuable — cierra brechas de integración
- [x] Estimable — escritura + verificación (~4-5 horas)
- [x] Small — un sprint
- [x] Testable — ejecutar curl de cada guía

## Notas técnicas

Documentación a crear en `docs/14-guides/`:

### 1. openwebui-gateway-setup.md
- Instalación de OpenWebUI (si no existe)
- Config del Gateway (config.yaml subset)
- Env vars (OPENAI_BASE_URL, OPENAI_API_KEY)
- Test: `curl -X POST http://localhost:8080/v1/chat/completions ...`
- Troubleshooting: modelo no disponible, 401, timeouts

### 2. opencode-gateway-setup.md
- OpenCode installation
- Endpoint `/responses` configuration
- Environment: OPENCODE_BASE_URL, OPENCODE_API_KEY
- Example: reasoning_effort usage
- Test: request Responses API ejemplo
- Troubleshooting: reasoning models down, fallback behavior

### 3. free-claude-code-gateway-setup.md
- Free Claude Code repo reference (OSS)
- ANTHROPIC_BASE_URL configuration
- API key from free provider (AIHubMix, etc.)
- Diferencia con Anthropic directo (parámetros soportados)
- Test: setup verification
- Troubleshooting: 401, unsupported models

### 4. claude-code-gateway-setup.md
- Similar a free-claude-code pero con Anthropic key oficial
- Extra: setup en VS Code / IDE
- Multi-tenant considerations

### 5. openhands-gateway-setup.md
- OpenHands installación
- LLM_MODEL = "openai/gpt-4o" (nota: prefijo para LiteLLM)
- LLM_BASE_URL = http://localhost:8080
- Config precedence (env > config file)
- Test: agente simple

### 6. openclaw-gateway-setup.md
- OpenClaw setup
- Voice + text integration
- Provider config en OpenClaw (pointing to Gateway)
- Test: voice request (si está soportado)

### 7. crewai-gateway-setup.md
- CrewAI installation (pip install crewai)
- Python code example:
  ```python
  llm = LLM(model="gpt-4o", base_url="http://localhost:8080", api_key="...")
  agent = Agent(llm=llm)
  ```
- Multi-agent example
- Test: agent.invoke()

### 8. ui-tars-gateway-setup.md
- UI-TARS local deployment
- Gateway pointing (OpenAI-compatible endpoint)
- Configuration as if UI-TARS is a consumer (no deployment needed if remote)
- Test: GUI automation request via Gateway

### 9. GATEWAY_CLIENTS.md (master)
Table:
| Cliente | Formato | Endpoint | Auto-Detect | Parameters | Status |
|---------|---------|----------|------------|------------|--------|
| OpenWebUI | OpenAI | /v1/chat/completions | ✓ | temp, top_p | ✓ |
| OpenCode | Responses | /responses | ✓ | reasoning_effort | ✓ |
| Free Claude Code | Anthropic | /v1/messages | ✓ | temp, top_k | ✓ |
| Claude Code | Anthropic | /v1/messages | ✓ | temp, thinking | ✓ |
| OpenHands | OpenAI | /v1/chat/completions | ✓ | all OpenAI | ✓ |
| OpenClaw | OpenAI | /v1/chat/completions | ✓ | temp, top_p | ✓ |
| CrewAI | OpenAI | /v1/chat/completions | ✓ | all OpenAI | ✓ |
| UI-TARS | OpenAI | /v1/chat/completions | ✓ | temp, top_p | ✓ |

Ubicación: `docs/14-guides/` (create if not exists)

Validación: cada guía debe incluir
- Instalación/setup exacto
- Configuración del Gateway (config.yaml snippet)
- Env vars necesarias
- Curl/Python test ejemplo funcional
- Troubleshooting section
- Link a HUs relevantes (HU-042, HU-043, etc.)
