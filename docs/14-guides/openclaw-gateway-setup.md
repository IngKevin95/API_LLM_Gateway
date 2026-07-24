# OpenClaw con API LLM Gateway

## Configuración para Integración por Voz

OpenClaw se utiliza a menudo para asistentes de voz. Aquí la latencia es crítica.

```json
{
  "llm": {
    "provider": "openai",
    "base_url": "http://localhost:8080/v1",
    "api_key": "tu_api_key",
    "model": "router:capability:low-latency"
  }
}
```

Usando la capacidad `low-latency`, el Gateway enrutará preferentemente hacia modelos con bajo `latency_p95_ms` según su Health Monitor, ideal para respuestas por voz rápidas.
