## Decisiones de diseño

### Patrón de implementación

Todos los adapters secundarios siguen el mismo patrón de EP-002:
- Implementan `adapter.Adapter` (`Chat`, `Stream`, `Embed`)
- Emiten `*adapter.ProviderError` con `Retryable=true` para 429/5xx, `false` para 4xx cliente
- Usan `net/http` estándar sin dependencias externas nuevas
- Se prueban con un servidor HTTP de prueba (`httptest.NewServer`) — sin red real

### AIHubMix (HU-029)

API 100% compatible con OpenAI. El adapter es un thin wrapper sobre el adapter OpenAI, configurando el `BaseURL` al endpoint de AIHubMix (`https://aihubmix.com/v1`). Solo override del host; toda la lógica de marshaling la hereda de `adapter/openai`.

### Google Gemini (HU-030)

API propia: requiere traducción de payload.

| Interno | Gemini |
|---|---|
| `messages[role=system]` | `systemInstruction.parts[0].text` |
| `messages[role=user/assistant]` | `contents[].role=user/model` |
| `content` string | `parts[0].text` |
| imagen base64 | `parts[0].inlineData.{mimeType,data}` |

Streaming: SSE con prefijo `data:` + JSON `candidates[0].content.parts[0].text`.

Errores:
- HTTP 400 → `ProviderError{Retryable: false}`
- HTTP 429/500/503 → `ProviderError{Retryable: true}`

### OpenRouter (HU-031)

OpenAI-compat. Solo requiere dos headers adicionales:
- `HTTP-Referer: https://api-llm-gateway`
- `X-Title: API LLM Gateway`

El model ID viaja tal cual (ej. `anthropic/claude-3-haiku`). No requiere adapter de traducción de payload; se reutiliza la lógica OpenAI con headers customizados.

### OmniRoute (HU-036)

Proxy configurable. El endpoint base y credenciales se inyectan desde registry YAML. Contrato OpenAI-compat. Traduce errores upstream a `ProviderError`.

### Estructura de paquetes

```
src/internal/adapter/
  aihubmix/
    aihubmix.go       // thin wrapper sobre openai.Adapter con BaseURL override
    aihubmix_test.go
  google/
    google.go         // traducción payload + streaming SSE
    google_test.go
  openrouter/
    openrouter.go     // wrapper OpenAI-compat + headers requeridos
    openrouter_test.go
  omniroute/
    omniroute.go      // proxy configurable
    omniroute_test.go
```

### Testing

Cada adapter se prueba con `httptest.NewServer` que simula respuestas del proveedor. Sin dependencias de red. Conformance test heredado de `adapter/conformance_test.go` garantiza que implementan la interfaz correctamente.
