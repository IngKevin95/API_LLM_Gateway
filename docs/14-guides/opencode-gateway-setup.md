# OpenCode con API LLM Gateway

## Configuración

OpenCode utiliza la nueva **Responses API** (`/responses`) del Gateway.

### En el archivo de configuración `config.json` de OpenCode:

```json
{
  "api_endpoint": "http://localhost:8080/responses",
  "api_key": "tu_api_key_del_gateway",
  "default_model": "router:capability:reasoning",
  "reasoning_effort": "high"
}
```

El Gateway interceptará la petición, la normalizará y la enviará al proveedor óptimo que tenga capacidades de razonamiento (`reasoning`).
