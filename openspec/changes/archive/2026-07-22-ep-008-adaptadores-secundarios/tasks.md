## Tasks

### SS1 — AIHubMix & Google (HU-029, HU-030)

- [ ] **SS1-T1**: Crear `src/internal/adapter/aihubmix/aihubmix.go`
  - Wrapper sobre `openai.Adapter` con `BaseURL` apuntando a `https://aihubmix.com/v1`
  - `Chat`, `Stream`, `Embed` reutilizan lógica de openai adapter
  - Manejo de `Params` no soportados: omitir sin error
- [ ] **SS1-T2**: Crear `src/internal/adapter/aihubmix/aihubmix_test.go`
  - `httptest.NewServer` simulando respuestas AIHubMix
  - Escenarios: happy path, 429 retryable, 500 retryable, params ignorados
- [ ] **SS1-T3**: Crear `src/internal/adapter/google/google.go`
  - Traducción payload: `messages[system]` → `systemInstruction`, `user/assistant` → `contents`
  - Imágenes base64 → `inlineData`
  - Streaming SSE con parser de `candidates[0].content.parts[0].text`
  - Errores: 400 → `Retryable:false`, 429/5xx → `Retryable:true`
- [ ] **SS1-T4**: Crear `src/internal/adapter/google/google_test.go`
  - Escenarios: chat con imagen, system prompt extraído, 400 non-retryable, 429 retryable

### SS2 — OpenRouter & OmniRoute (HU-031, HU-036)

- [ ] **SS2-T1**: Crear `src/internal/adapter/openrouter/openrouter.go`
  - Wrapper OpenAI-compat con headers `HTTP-Referer` y `X-Title` inyectados
  - Propagar model ID completo (ej. `anthropic/claude-3-haiku`)
- [ ] **SS2-T2**: Crear `src/internal/adapter/openrouter/openrouter_test.go`
  - Verificar que headers están presentes en cada petición
  - Escenarios: happy path, 503 retryable, timeout/ctx cancelado
- [ ] **SS2-T3**: Crear `src/internal/adapter/omniroute/omniroute.go`
  - Adapter configurable (baseURL + auth desde config struct)
  - Contrato OpenAI-compat; traducción de errores upstream
- [ ] **SS2-T4**: Crear `src/internal/adapter/omniroute/omniroute_test.go`
  - Escenarios: petición exitosa, fallback activo, 429/500 retryable
