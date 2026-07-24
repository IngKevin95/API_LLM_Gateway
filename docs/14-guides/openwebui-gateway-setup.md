# OpenWebUI con API LLM Gateway

## Configuración Paso a Paso

### 1. Variables de Entorno
Inicia tu contenedor de OpenWebUI con las siguientes variables apuntando al Gateway:

```bash
docker run -d -p 3000:8080 \
  -e OPENAI_API_BASE_URL=http://tu-gateway:8080/v1 \
  -e OPENAI_API_KEY=tu_api_key_del_gateway \
  -v open-webui:/app/backend/data \
  --name open-webui \
  ghcr.io/open-webui/open-webui:main
```

### 2. Magia en la Interfaz
OpenWebUI hará fetch a `/v1/models`. El Gateway le devolverá automáticamente tanto los modelos de OpenAI como los de Anthropic y Google. Podrás seleccionarlos desde el dropdown de modelos.
