# OpenHands con API LLM Gateway

## Configuración de Entorno para Agentes Autónomos

OpenHands es un agente autónomo de SWE (Software Engineering). Requiere alta compatibilidad con llamadas a herramientas (Tool Calling).

### `config.toml` o Variables de Entorno

```bash
export LLM_BASE_URL="http://localhost:8080/v1"
export LLM_API_KEY="tu_api_key"
# Puedes usar un modelo explícito o una capability
export LLM_MODEL="openai/router:capability:coding"
```

El Gateway garantiza que las funciones enviadas por OpenHands bajo el formato de OpenAI `tools` sean traducidas correctamente al formato nativo de Anthropic `tool_use` si el router elige a Claude 3.5 Sonnet, asegurando el éxito del agente.
