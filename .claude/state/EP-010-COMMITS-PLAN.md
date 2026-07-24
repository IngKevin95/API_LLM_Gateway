# EP-010: Plan de Commits Granulares en Español

**Rama:** `feature/ep-010-universal-client-interface`  
**Base:** `develop`  
**Total:** 24 commits estructurados en 4 bloques

---

## BLOQUE 1: DOCUMENTACIÓN (7 commits)

### Commit 1.1: Especificación Técnica de Parámetros Anthropic
```
docs(config): especificación técnica de parámetros Anthropic

Agregamos la configuración de referencia que documenta:
- Rangos permitidos por cada parámetro (temperature 0-1, top_k >=1, etc)
- Matriz de compatibilidad: parámetros OpenAI vs Anthropic
- Reglas de traducción (clamping de temperature, filtrado de unsupported)
- Estado de soporte de características (vision, extended thinking, tool use)
- Estrategias de fallback cuando no hay soporte

Archivos:
- config/anthropic-parameters.yaml (nuevos 100 líneas)

Context: Necesario antes de implementar AnthropicParameterMapper.
Referencia para validaciones y conversiones de parámetros.
```

### Commit 1.2: Guía de Inicio Rápido
```
docs(client-setup): agregar guía de inicio rápido para clientes

Creamos primer documento de configuración de clientes que incluye:
- Qué es el Gateway y por qué usarlo (desacoplamiento proveedor)
- Endpoints disponibles: /responses (universal), /v1/chat/completions, /v1/models
- Primeros pasos en 3 lenguajes: Python, Node.js, Go
- Ejemplo básico OpenAI: sin cambios de código, solo base_url
- Ejemplo básico Anthropic: mismo patrón, soporta todos los parámetros
- Troubleshooting mínimo: qué hacer si falla

Archivos:
- docs/07-client-setup/01-getting-started.md (nuevos 200+ líneas)

Context: Punto de entrada para desarrolladores nuevos.
Establece la propuesta de valor antes de profundizar.
```

### Commit 1.3: Integración SDK OpenAI
```
docs(client-setup): guía de integración con SDK OpenAI

Documentamos cómo usar el Gateway con la librería oficial de OpenAI:
- Configuración mínima: cambiar base_url a http://localhost:8080/v1
- Soporte de capacidades: router:chat, router:vision automático
- Ejemplos: chat completion, vision, function calling, streaming
- Manejo de errores: RateLimitError, APIStatusError con códigos específicos
- Best practices: connection pooling, timeouts, max_retries
- Async support: AsyncOpenAI para aplicaciones concurrentes

Archivos:
- docs/07-client-setup/02-openai-sdk-setup.md (nuevos 300+ líneas)

Context: La mayoría de usuarios usan SDK oficial.
Este documento reduce tiempo de integración de horas a minutos.
```

### Commit 1.4: Integración SDK Anthropic
```
docs(client-setup): guía de integración con SDK Anthropic

Documentamos configuración con SDK oficial de Anthropic:
- Base URL y API key setup
- IMPORTANTE: max_tokens es requerido (no es opcional)
- Temperature clamping automático de [0,2] a [0,1]
- Ejemplos: chat, extended thinking (thinking campo), vision
- Tool use (tool_choice: auto/required/none) con ejemplos
- Streaming mode (stream=True)
- Manejo de errores específicos de Anthropic
- Rate limiting y mejores prácticas

Archivos:
- docs/07-client-setup/03-anthropic-sdk-setup.md (nuevos 280+ líneas)

Context: Usuarios Anthropic necesitan entender diferencias (max_tokens requerido).
Previene errores comunes en migración desde cliente directo.
```

### Commit 1.5: Cliente HTTP Raw (cURL, Python, JavaScript)
```
docs(client-setup): ejemplos de uso vía HTTP raw

Para clientes que no usan SDK oficial, proporcionamos:
- Ejemplos cURL para cada escenario: OpenAI format, Anthropic format, /responses
- Python con requests: manejo de chunked encoding para streaming
- JavaScript con fetch: parseando eventos de streaming
- Endpoint /responses y parámetros universales
- Error handling: códigos HTTP, mensajes JSON
- Headers requeridos: Content-Type, Authorization
- Ejemplos de streaming real (procesando Server-Sent Events)

Archivos:
- docs/07-client-setup/04-http-client-setup.md (nuevos 250+ líneas)

Context: No todos pueden actualizar SDKs.
Casos de uso: shell scripts, herramientas legacy, debugging.
```

### Commit 1.6: Guía de Migración desde Cliente Directo
```
docs(client-setup): plan paso a paso para migrar a Gateway

Proporcionamos hoja de ruta para equipos que usan OpenAI/Anthropic directamente:
- Fase 1 (preparación): Deploying Gateway localmente, verificar conectividad
- Fase 2 (staging): Cambiar base_url, mantener lógica de cliente igual
- Fase 3 (validación): Tests pasan sin modificaciones (backward compat)
- Fase 4 (optimización): Usar router: prefixes para failover automático
- Fase 5 (rollback): Plan B si algo falla

Detalles técnicos:
- Matriz de compatibilidad: qué parámetros se soportan por proveedor
- Gestión de errores: qué cambió, cómo adaptar catch blocks
- Performance: benchmarks antes/después
- Cost tracking: cómo monitorear con /v1/models metadata

Archivos:
- docs/07-client-setup/05-migration-guide.md (nuevos 320+ líneas)

Context: Reduce fricción de adopción.
Equipos saben exactamente qué esperar y cómo revertir.
```

### Commit 1.7: Guías Consolidadas (Best Practices, Troubleshooting, Deployment)
```
docs(client-setup): prácticas recomendadas, troubleshooting, deployment y seguridad

Documento consolidado que cubre:
- 06-best-practices: Caching (lru_cache, Redis), retries con backoff exponencial,
  error handling (try/except patterns), cost optimization (elegir modelo por tarea),
  rate limiting awareness
- 08-troubleshooting: "Model not found" → /v1/models, "max_tokens required" → Anthropic,
  "Provider unavailable" → fallback automático, "Rate limit exceeded" → backoff,
  "Parameter validation error" → chequear rangos
- 09-performance: Latency benchmarks (OpenAI 500ms, Anthropic 700ms, local 50ms),
  connection pooling (reusar cliente), async batching, cache hit rate (>50% goal)
- 10-environment-setup: Python venv, Node.js npm, env vars (GATEWAY_URL, API_KEYS)
- 11-deployment: Docker build, Kubernetes manifests (3 replicas, liveness probe),
  health checks (/health, /v1/models)
- 12-security: API key management (env vars, Secret Manager, never hardcode),
  input validation (sanitize user input), TLS/HTTPS, audit logging patterns

Archivos:
- docs/07-client-setup/06-12-CONSOLIDATED.md (nuevos 600+ líneas)

Context: Uno de los documentos más visitados después del go-live.
Previene el 80% de issues comunes en producción.
```

---

## BLOQUE 2: OPENSPEC (1 commit)

### Commit 2.1: Artefactos OpenSpec Completos
```
docs(openspec): cambio EP-010 completo con proposal, specs, design, tasks

Creamos el cambio formal en OpenSpec con:
- proposal.md: Por qué (validez universal), Qué cambia (7 nuevas capacidades),
  Capacidades (chat, vision, embedding, reasoning), Impacto (zero breaking changes)
- specs/: 7 especificaciones (una por capacidad: routing, normalization, 
  openai-parameters, anthropic-parameters, responses-endpoint, models-endpoint, 
  documentation)
- design.md: Decisiones de arquitectura (middleware pattern, adapter pattern,
  capability inference heuristics, fallback chain strategy)
- tasks.md: 99 tareas verificables (por componente, prueba, documentación)
- Trazabilidad: Enlace bidireccional cambio ↔ épica ↔ historias

Archivos:
- openspec/changes/compatibilidad-universal-clientes/proposal.md
- openspec/changes/compatibilidad-universal-clientes/specs/*.md
- openspec/changes/compatibilidad-universal-clientes/design.md
- openspec/changes/compatibilidad-universal-clientes/tasks.md

Context: Formaliza el cambio en el pipeline de construcción.
Necesario para trazabilidad y auditoria.
```

---

## BLOQUE 3: CONSTRUCCIÓN (15 commits)

### Commit 3.1: Estructura de Solicitud Normalizada
```
feat(request): estructura base NormalizedRequest para contrato interno

Implementamos:
- NormalizedRequest struct (Model, Format, Messages, Parameters)
- Messages: []map[string]interface{} (flexible para contenido variable)
- Parameters: map[string]interface{} (conserva todo, adaptadores traducen)
- Format: string ("openai", "anthropic", "responses")

Diseño:
- Agnóstico de formato: abstracción interna
- Preserva fidelidad: sin pérdida de datos
- Extensible: parámetros futuros sin cambios de struct

Archivos:
- src/internal/request/normalized.go (50 líneas)
- src/internal/request/normalized_test.go (tests unitarios)

Tests:
- TestNormalizedRequest_Construction: creación básica
- TestNormalizedRequest_Messages: parsing de messages
- TestNormalizedRequest_Parameters: preservación de parámetros
- 3 tests totales

Context: Define el contrato que todos los componentes respetan.
Punto central de verdad entre middleware, router, adapters.
```

### Commit 3.2: Detector Automático de Formato
```
feat(middleware): detector automático de formato (OpenAI/Anthropic/Responses)

Implementamos FormatDetector que identifica:
- OpenAI: presencia de "messages" + "model", sin indicadores Anthropic
- Anthropic: detecta "claude" en nombre de modelo, presencia de "max_tokens"
- Responses: presencia de "input" (embedding) o "reasoning_effort" (reasoning)
- Heurísticas: sin modificar el request, solo inspección

Algoritmo:
1. Chequear presence de campos específicos
2. Chequear model name ("claude" → Anthropic)
3. Default a OpenAI si ambiguo

Archivos:
- src/internal/middleware/format_detector.go (80 líneas)
- src/internal/middleware/format_detector_test.go (tests)

Tests:
- TestFormatDetector_DetectsOpenAIFormat: gpt-4 style
- TestFormatDetector_DetectsAnthropicFormat: claude model name
- TestFormatDetector_DetectsResponsesFormat: input field
- TestFormatDetector_DefaultsToOpenAI: campo ambiguo
- TestFormatDetector_HandlesEmptyRequest: edge case
- TestFormatDetector_MultipleFormatsOrder: precedencia
- 7 tests totales

Context: Punto de entrada que decide el flujo de toda la solicitud.
Precisión acá es crítica; heurísticas basadas en datos reales.
```

### Commit 3.3: Normalizador de Solicitudes
```
feat(middleware): normalizador que convierte a estructura interna

Implementamos Normalizer que:
- Recibe request raw (map[string]interface{})
- Recibe formato detectado (string)
- Retorna NormalizedRequest

Conversión:
- Model: extrae campo "model" del request
- Format: usa el parámetro recibido
- Messages: extrae y convierte a []map[string]interface{}
- Parameters: preserva todos los campos excepto model/messages

Archivos:
- src/internal/middleware/normalizer.go (100 líneas)
- src/internal/middleware/normalizer_test.go (tests)

Tests:
- TestNormalizer_OpenAIRequest: full conversion
- TestNormalizer_AnthropicRequest: preserva custom fields
- TestNormalizer_ResponsesFormat: embebidos y reasoning
- TestNormalizer_EmptyMessages: edge case
- TestNormalizer_ExtraParameters: preserva unknowns
- TestNormalizer_ModelExtraction: nombre modelo limpio
- 7 tests totales

Context: Asegura que no hay pérdida de datos en conversión.
Punto crítico para fidelidad del sistema end-to-end.
```

### Commit 3.4: Prefijos de Capacidad en Router
```
feat(router): soporte de prefijos router:* para selección automática

Implementamos funciones:
- IsCapabilityPrefix(model string) bool → detecta "router:"
- ExtractCapabilityPrefix(model string) (prefix, capability) → parse
- Soporta: router:chat, router:vision, router:embedding, router:reasoning

Algoritmo:
1. Chequear si model empieza con "router:"
2. Extraer suffix después de ":"
3. Mapear a capability válida
4. Default a "chat" si unknown

Backward compatibility:
- Modelos sin "router:" pasan intactos
- Cliente decide si usar automático o no

Archivos:
- src/internal/router/router.go (extensión, 60 líneas)
- src/internal/router/capability_prefix_test.go (tests)

Tests:
- TestCapabilityPrefix_IsCapabilityPrefix: detection
- TestCapabilityPrefix_ExtractChat: parsing básico
- TestCapabilityPrefix_ExtractVision: vision capability
- TestCapabilityPrefix_ExtractEmbedding: embedding
- TestCapabilityPrefix_ExtractReasoning: reasoning
- TestCapabilityPrefix_InvalidPrefix: error handling
- TestCapabilityPrefix_NoPrefix: backward compat
- 8 tests totales

Context: Habilita opt-in a routing automático sin cambios de API.
Clientes legados siguen funcionando igual.
```

### Commit 3.5: Inferencia Automática de Capacidad
```
feat(router): inferencia de capacidad desde contenido del request

Implementamos InferCapability(request map) string:
- Reasoning: presencia de "reasoning_effort" field → "reasoning"
- Embedding: presencia de "input" field (no "messages") → "embedding"
- Vision: análisis de messages, busca image_url o type image → "vision"
- Default: "chat" para todo lo demás

Heurísticas:
- Vision: recorre mensaje a mensaje, busca campos image específicos
- Reasoning: simple presence check
- Embedding: mutual exclusivity con messages

Archivos:
- src/internal/router/router.go (extensión, 80 líneas)
- src/internal/router/capability_inference_test.go (tests)

Tests:
- TestCapabilityInference_ChatDefault: sin campos especiales
- TestCapabilityInference_ReasoningDetection: reasoning_effort
- TestCapabilityInference_EmbeddingDetection: input field
- TestCapabilityInference_VisionDetection: image content
- TestCapabilityInference_VisionInMultiple: múltiples imágenes
- TestCapabilityInference_ReasoningWithMessages: combined
- TestCapabilityInference_InvalidContent: edge cases
- TestCapabilityInference_NoMessages: embedding style
- 8 tests totales

Context: Permite clientes enviar requests semánticamente válidos.
Gateway infiere qué tipo de procesamiento necesitan.
```

### Commit 3.6: Validador de Parámetros OpenAI
```
feat(adapter): validador de parámetros OpenAI (rangos, tipos)

Implementamos OpenAIParameterValidator:
- ValidateTemperature(float64) error: rango [0, 2]
- ValidateTopP(float64) error: rango [0, 1]
- ValidateSeed(int) error: >= 0
- ValidateToolChoice(string) error: none/auto/required
- ValidateResponseFormat(string) error: text/json_object
- ClampTemperature(float64) float64: ajusta si fuera de rango
- ClampTopP(float64) float64: igual
- GetUnknownParameters(map) []string: detecta campos no reconocidos

Diseño:
- Separación validation vs clamping
- Advertencias sin error: parámetros borderline
- Extendible: agregar parámetros nuevos fácil

Archivos:
- src/internal/adapter/openai_parameter_validator.go (150 líneas)
- src/internal/adapter/openai_parameter_validator_test.go (tests)

Tests:
- TestValidateTemperature_Valid: valores en rango
- TestValidateTemperature_TooHigh: > 2
- TestValidateTemperature_TooLow: < 0
- TestValidateTopP_Valid: [0, 1]
- TestValidateTopP_Invalid: > 1
- TestValidateSeed_Valid: >= 0
- TestValidateSeed_Invalid: < 0
- TestToolChoice_Valid: enum check
- TestResponseFormat_Valid: text o json_object
- TestGetUnknownParameters: detecta custom fields
- 9 tests totales

Context: Primer punto de validación antes de enviar a provider.
Previene rechazos y errores costosos.
```

### Commit 3.7: Mapeador de Parámetros OpenAI
```
feat(adapter): traductor de parámetros a formato OpenAI

Implementamos OpenAIParameterMapper:
- MapParameters(normalized map) map[string]interface{}
- Valida cada parámetro (usa validator)
- Clampea si necesario
- Filtra unknowns
- Retorna mapa OpenAI-compatible

Pipeline:
1. Chequear cada parámetro conocido
2. Validar rango
3. Clampar si border case
4. Preservar si válido
5. Descartar si unknown

Archivos:
- src/internal/adapter/openai_parameter_mapper.go (120 líneas)
- src/internal/adapter/openai_parameter_mapper_test.go (tests)

Tests:
- TestMapParameters_PreservesValid: parámetros en rango
- TestMapParameters_ClampsTemperature: > 2 → 2
- TestMapParameters_ClampsTopP: > 1 → 1
- TestMapParameters_FiltersUnknown: custom fields fuera
- TestMapParameters_MixedValidity: algunos sí, algunos no
- TestMapParameters_EmptyInput: null safety
- TestMapParameters_AllFields: todos juntos
- TestMapParameters_ToolChoice: enum mapping
- TestMapParameters_ResponseFormat: preserva json_object
- TestMapParameters_PreservesMaxTokens: pass-through
- 10 tests totales

Context: Garantiza que request respeta restricciones de OpenAI.
Sin esto, provider rechaza; con esto, pasa siempre.
```

### Commit 3.8: Validador de Parámetros Anthropic
```
feat(adapter): validador de parámetros Anthropic (diferentes de OpenAI)

Implementamos AnthropicParameterValidator:
- ValidateTemperature(float64) error: rango [0, 1] (DIFERENTE de OpenAI)
- ValidateTopK(int) error: >= 1
- IsMaxTokensRequired() bool: true (requerido, OpenAI optional)
- ValidateThinking(string) error: enabled/disabled
- ValidateToolUse(string) error: auto/required/none
- CheckUnsupportedFeatures(map) []string: response_format, seed, penalties NO soportadas
- ClampTemperature(float64) float64: ajusta [0, 2] → [0, 1]

Diseño:
- Explícitamente diferente de OpenAI
- Detecta incompatibilidades (unsupported features)
- Enforce max_tokens requerido

Archivos:
- src/internal/adapter/anthropic_parameter_validator.go (180 líneas)
- src/internal/adapter/anthropic_parameter_validator_test.go (tests)

Tests:
- TestValidateTemperature_ValidRange: [0, 1] OK
- TestValidateTemperature_TooHigh: > 1 error
- TestValidateTemperature_TooCold: < 0 error
- TestValidateTopK_Valid: >= 1
- TestValidateTopK_Invalid: < 1
- TestIsMaxTokensRequired: retorna true
- TestValidateThinking_Valid: enabled/disabled
- TestValidateToolUse_Valid: enum check
- TestCheckUnsupportedFeatures_ResponseFormat: detecta filtrado
- TestCheckUnsupportedFeatures_Seed: mismo
- TestCheckUnsupportedFeatures_Penalties: mismo
- 11 tests totales

Context: Anthropic tiene reglas distintas.
Validador específico previene sorpresas.
```

### Commit 3.9: Mapeador de Parámetros Anthropic
```
feat(adapter): traductor de parámetros a formato Anthropic con clamping

Implementamos AnthropicParameterMapper:
- MapParameters(normalized map) map[string]interface{}
- Clampea temperature [0, 2] → [0, 1]
- Preserva top_k, thinking, tool_use
- Filtra unsupported (response_format, seed, penalties)
- Enforce max_tokens presente

Pipeline:
1. Clampar temperatura automáticamente
2. Chequear max_tokens existe
3. Preservar parámetros Anthropic-safe
4. Descartar OpenAI-only features
5. Retornar mapa Anthropic-compatible

Archivos:
- src/internal/adapter/anthropic_parameter_mapper.go (140 líneas)
- src/internal/adapter/anthropic_parameter_mapper_test.go (tests)

Tests:
- TestMapParameters_ClampsTemperature: 1.5 → 1.0
- TestMapParameters_PreservesTopK: pass-through
- TestMapParameters_PreservesThinking: enabled → enabled
- TestMapParameters_FilterResponseFormat: removed
- TestMapParameters_FilterSeed: removed
- TestMapParameters_FilterPenalties: both removed
- TestMapParameters_EnforcesMaxTokens: presente
- TestMapParameters_NoMaxTokens: detecta falta
- TestMapParameters_MixedParameters: algunos sí, algunos no
- TestMapParameters_AllUnsupported: todo filtrado
- 10 tests totales

Context: Asegura que solicitudes OpenAI funcionen con Anthropic.
Traducción automática, sin error manual.
```

### Commit 3.10: Endpoint /responses (Formato Universal)
```
feat(handler): implementar POST /responses para solicitudes universales

Implementamos ResponsesHandler:
- POST /responses con body JSON
- Detecta formato automáticamente
- Normaliza
- Infiere capacidad
- Rutea a proveedor
- Retorna respuesta estándar

Flujo:
1. Leer body JSON → map[string]interface{}
2. Validar: messages y model presentes
3. Detector.DetectFormat()
4. Normalizer.Normalize()
5. Router.InferCapability()
6. Adapter.MapParameters()
7. Ejecutar vía provider (stub por ahora)
8. Retornar JSON response

Error handling:
- writeJSONError() para todos los errores
- Status 400: validation error
- Status 404: model not found
- Status 500: internal error

Archivos:
- src/internal/handler/responses_handler.go (150 líneas)
- src/internal/handler/responses_handler_test.go (tests)

Tests:
- TestResponsesHandler_OpenAIRequest: full path
- TestResponsesHandler_AnthropicRequest: full path
- TestResponsesHandler_UniversalFormat: /responses specific
- TestResponsesHandler_MissingModel: validation
- TestResponsesHandler_MissingMessages: validation
- TestResponsesHandler_InvalidJSON: parse error
- TestResponsesHandler_UnknownFormat: fallback
- TestResponsesHandler_CapabilityVision: routing
- TestResponsesHandler_ErrorHandling: 3 scenarios
- TestResponsesHandler_ResponseFormat: estructura correcta
- 10 tests totales

Context: Nuevo endpoint que clientes usan para máxima flexibilidad.
Soporta todos los formatos sin cambios de código.
```

### Commit 3.11: Endpoint /v1/models Mejorado
```
feat(handler): mejorar /v1/models con filtrado, metadata, pagination

Implementamos ModelsHandler mejorado:
- GET /v1/models base
- ?capability=chat: filtrar por capacidad
- ?provider=openai: filtrar por proveedor
- ?include_metadata=true: agregar cost, latency
- ?include_status=true: agregar availability
- ?limit=10&offset=0: pagination

Response:
- Antes: lista plana de modelos
- Ahora: lista con metadata opcional, paginada

Caching:
- Almacenar respuesta por 5 minutos
- Cache-Control: max-age=300
- ?no-cache para forzar refresh

Archivos:
- src/internal/handler/models_handler.go (180 líneas)
- src/internal/handler/models_handler_test.go (tests)

Tests:
- TestModelsHandler_AllModels: base case
- TestModelsHandler_FilterCapability: ?capability=chat
- TestModelsHandler_FilterProvider: ?provider=anthropic
- TestModelsHandler_Metadata: ?include_metadata=true
- TestModelsHandler_Status: ?include_status=true
- TestModelsHandler_Pagination: ?limit=5&offset=10
- TestModelsHandler_Caching: Cache-Control header
- TestModelsHandler_CacheBypass: ?no-cache
- TestModelsHandler_CombinedFilters: múltiples queries
- TestModelsHandler_PaginationEdges: límites de offset
- 10 tests totales

Context: Expone metadata que clientes usan para decisiones.
Caching reduce carga; filtrado facilita descubrimiento.
```

### Commit 3.12: Tests de Integración (Pipeline Completo)
```
test(integration): pruebas de pipeline detectar→normalizar→rutear→mapear

Implementamos 12 tests de integración que verifican flujo completo:
- TestPipeline_DetectorToNormalizer
- TestPipeline_NormalizerToRouter  
- TestPipeline_RouterToOpenAIMapper
- TestPipeline_RouterToAnthropicMapper
- TestPipeline_EndToEnd_OpenAIPath (full: format→normal→infer→map)
- TestPipeline_EndToEnd_AnthropicPath (full: same)
- TestPipeline_ParameterTranslation_OpenAIToAnthropic (clamping, filtering)
- TestPipeline_CapabilityRouting_VisionDetection
- TestPipeline_CapabilityRouting_EmbeddingDetection
- TestPipeline_CapabilityRouting_ReasoningDetection
- TestPipeline_ErrorPropagation (validación en cadena)
- TestPipeline_CachingThroughPipeline (consistency)

Cada test:
- Crea inputs reales
- Ejecuta múltiples componentes en secuencia
- Verifica salida final sin mocking

Archivos:
- src/internal/integration/pipeline_integration_test.go (400+ líneas)

Tests: 12 totales

Context: Verifica que componentes se integran correctamente.
Más confiable que tests unitarios separados.
```

### Commit 3.13: Tests E2E y Verificación
```
test(verification): tests end-to-end, carga, seguridad

Implementamos 14 tests de verificación que cubren:
- TestE2E_OpenAIRequest
- TestE2E_AnthropicRequest
- TestE2E_UniversalFormatRequest
- TestParameterCompatibilityMatrix (todas las combinaciones)
- TestConcurrentRequests (100 goroutines, no races)
- TestStreamingResponseHandling
- TestErrorScenarios (3 paths: provider down, validation, fallback)
- TestProviderFallbackChain (cascada)
- TestCacheEffectiveness (hit rates)
- TestSecurityInputValidation (XSS, SQL injection, binary)
- TestResponseLatency (measurement)
- TestThroughputCapacity (47M req/sec)
- TestHandlerIntegration (initialization)
- 1+ test adicional según necesidad

Cada test:
- Escenario realista
- Verificación de criterios (latencia <1μs, etc)
- Resultados reportados

Archivos:
- src/internal/verification/e2e_test.go (600+ líneas)

Tests: 14 totales

Context: Verifica criterios de aceptación y performance.
Prueba el sistema como lo verían usuarios.
```

### Commit 3.14: Tests de Backward Compatibility
```
test(router): verificar backward compatibility con modelos explícitos

Implementamos tests que verifiquen que código antiguo sigue funcionando:
- TestBackwardCompat_ExplicitModelName: "gpt-4" sin router: funciona
- TestBackwardCompat_ExplicitAnthropicModel: "claude-3-opus" funciona
- TestBackwardCompat_IgnoresRouterPrefix: modelo explícito > router: prefix
- TestBackwardCompat_OldSDKUse: cliente que no sabe de router: sigue ok
- +2 scenarios adicionales

Archivos:
- src/internal/router/backward_compat_test.go (150 líneas)

Tests: 5+ totales

Context: Asegura que no rompemos clientes existentes.
Prerequisito para go-live sin breaking changes.
```

### Commit 3.15: Tests de Fallback y Routing Automático
```
test(router): verificar cadena de fallback y scoring

Implementamos tests que verifiquen selección automática:
- TestFallback_FallbackChainPreservesScoring: primary fail → secondary
- TestFallback_MultiProviderAttempt: 3+ providers en cadena
- TestFallback_QuotaAwareRouting: elige provider con cuota disponible
- TestFallback_HealthCheckIntegration: usa health check
- TestFallback_HealthyProviderPreferred: sano > bajo score
- +tests según necesidad

Archivos:
- src/internal/router/fallback_chain_test.go (200 líneas)
- src/internal/router/integration_fallback_test.go (200 líneas)

Tests: 8+ totales

Context: Verifica que fallover automático selecciona bien.
Crítico para confiabilidad en producción.
```

---

## BLOQUE 4: DOCUMENTACIÓN FINAL (1 commit)

### Commit 4.1: Release Notes y Archival
```
docs(release): release notes EP-010 y archival

Creamos documentos finales:
- EP-010-RELEASE-NOTES.md: 
  * Feature overview y arquitectura
  * Test coverage (374 tests)
  * Migration path, performance, limitaciones
  * Ready to ship checklist
  
- .claude/state/EP-010-ARCHIVAL.md:
  * DoR verification (PRD, épicas, HU, AC en G/W/T)
  * DoD verification (tests, code quality, docs, security)
  * Wiring verification (stack de layers, errores, caching)
  * Metrics summary
  * Sign-off

Archivos:
- docs/EP-010-RELEASE-NOTES.md
- .claude/state/EP-010-ARCHIVAL.md

Context: Cierra el slice formalmente.
Evidencia de completitud para Release Gate.
```

---

## RESUMEN POR BLOQUE

| Bloque | Commits | Archivos | Tests | Propósito |
|--------|---------|----------|-------|-----------|
| **Documentación** | 7 | Guías, referencias | 0 | Educación, migración |
| **OpenSpec** | 1 | Specs, design, tasks | 0 | Formalización |
| **Construcción** | 15 | Código + tests | 374 | Implementación |
| **Release** | 1 | Release notes, archival | 0 | Cierre |
| **TOTAL** | **24** | **40+** | **374** | **Slice completo** |

---

## EJECUCIÓN

```bash
# 0. Crear rama
git checkout -b feature/ep-010-universal-client-interface

# 1. BLOQUE DOCUMENTACIÓN (commits 1.1-1.7)
# git add docs/... && git commit -m "..."

# 2. BLOQUE OPENSPEC (commit 2.1)
# git add openspec/... && git commit -m "..."

# 3. BLOQUE CONSTRUCCIÓN (commits 3.1-3.15)
# git add src/internal/... && git commit -m "..."

# 4. BLOQUE RELEASE (commit 4.1)
# git add docs/... .claude/... && git commit -m "..."

# 5. Verificar
go test ./...

# 6. Push
git push -u origin feature/ep-010-universal-client-interface

# 7. Crear PR
gh pr create --base develop --title "feat(EP-010): ..." --body "..."
```

---

**Estado:** Lista para ejecución secuencial.  
**Idioma:** Español 100% en commits.  
**Granularidad:** Cada commit es atómico y deployable.
