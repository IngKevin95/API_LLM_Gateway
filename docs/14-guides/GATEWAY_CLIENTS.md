# Índice de Clientes Compatibles (Universal Compatibility)

El **API LLM Gateway** expone interfaces estándar (OpenAI, Anthropic, Responses) permitiendo que cualquier cliente moderno se conecte a él y disfrute de enrutamiento dinámico, cuotas centralizadas, resiliencia (failovers) y telemetría, **sin importar qué modelo real esté respondiendo por debajo**.

## Matriz de Compatibilidad

| Cliente / SDK | Interfaz Principal | Caso de Uso | Guía de Configuración |
|---------------|--------------------|-------------|-----------------------|
| **OpenWebUI** | OpenAI `/v1` | Interfaz gráfica y Chat UI | [Guía OpenWebUI](./openwebui-gateway-setup.md) |
| **OpenCode** | Responses API | Asistente de programación IDE | [Guía OpenCode](./opencode-gateway-setup.md) |
| **Claude Code (Free)** | Anthropic | Programación en terminal | [Guía Claude Code Free](./free-claude-code-gateway-setup.md) |
| **Claude Code (Official)** | Anthropic | Programación en terminal | [Guía Claude Code Oficial](./claude-code-gateway-setup.md) |
| **OpenHands** | OpenAI `/v1` | Agente Autónomo (SWE) | [Guía OpenHands](./openhands-gateway-setup.md) |
| **OpenClaw** | OpenAI `/v1` | Integración por Voz / TTS | [Guía OpenClaw](./openclaw-gateway-setup.md) |
| **CrewAI** | OpenAI `/v1` | Framework multi-agente | [Guía CrewAI](./crewai-gateway-setup.md) |
| **UI Tars** | OpenAI `/v1` | Automatización RPA UI | [Guía UI Tars](./ui-tars-gateway-setup.md) |

## Características Transversales
Todos los clientes soportan:
- **Magia del Router:** Usar `router:capability:vision` o `router:capability:chat` en el campo del modelo para que el Gateway elija automáticamente el mejor modelo disponible.
- **Traducción de Parámetros:** Si un cliente como CrewAI envía parámetros OpenAI (`seed`, `response_format`) pero el Gateway enruta a Anthropic (Claude), el Gateway traduce y limpia los parámetros automáticamente sin que el cliente falle.
- **Métricas:** Todo el uso, latencias y tokens de cualquiera de estos clientes se centralizan en Prometheus/Grafana.
