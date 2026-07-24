# UI Tars con API LLM Gateway

## Automatización de UI

UI Tars automatiza interfaces de usuario y depende enormemente del soporte para *Vision*.

```bash
export OPENAI_API_BASE="http://localhost:8080/v1"
export OPENAI_API_KEY="tu_api_key"
# Forzamos al Gateway a buscar el mejor modelo con capacidad de Visión
export MODEL_NAME="router:capability:vision"
```

El Gateway interceptará las imágenes en base64 de UI Tars, validará la cuota y normalizará el payload para asegurar compatibilidad con modelos como Gemini 1.5 Pro o GPT-4o, devolviendo siempre una estructura OpenAI compatible.
