# Free Claude Code con API LLM Gateway

## Configuración

Para utilizar clientes alternativos o herramientas CLI diseñadas para el ecosistema Anthropic, el Gateway provee un mock 100% compatible.

### Variables de Entorno

```bash
export ANTHROPIC_BASE_URL="http://localhost:8080"
export ANTHROPIC_API_KEY="tu_api_key_del_gateway"

# Iniciar la CLI
claude-code
```

El Gateway traducirá cualquier parámetro específico de Claude (como `thinking` o `top_k`) en caso de que el enrutador decida usar un modelo de OpenAI por debajo, evitando que el CLI falle.
