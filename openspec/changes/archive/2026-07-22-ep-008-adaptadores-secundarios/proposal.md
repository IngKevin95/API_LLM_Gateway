## Why

Sin adaptadores para AIHubMix, Google Gemini, OpenRouter y OmniRoute, el Gateway solo accede a OpenAI y Anthropic directamente. Agregar estos proveedores amplía el ecosistema de modelos disponibles, incrementa la resiliencia del failover (más eslabones en la cadena) y habilita acceso a modelos de nicho sin modificar el core del router.

## What Changes

- Nuevo adapter `aihubmix` (OpenAI-compat 100%, proveedor gratuito por defecto en el PRD).
- Nuevo adapter `google` (Gemini — API propia con tradución de roles/herramientas).
- Nuevo adapter `openrouter` (agregador multi-modelo compatible con OpenAI, routing por modelo ID).
- Nuevo adapter `omniroute` (proxy/router externo configurable).
- Cada adapter implementa la interfaz `adapter.Adapter` (`Chat`, `Stream`, `Embed`) y emite `adapter.ProviderError` para que el Failover Engine los gestione sin conocer el proveedor.
- Sin cambios en contratos públicos ni en el core del router.

## Trazabilidad

- Épica: **EP-008** (`docs/03-backlog/epicas.md`)
- Historias: HU-029, HU-030, HU-031, HU-036

## Capabilities

### New Capabilities

- `adapter-aihubmix`: Adapter para AIHubMix. Reenvía al endpoint OpenAI-compatible de AIHubMix, ignora parámetros no soportados de forma segura, y emite `ProviderError` retryable en 429/5xx.
- `adapter-google`: Adapter para Google Gemini. Traduce mensajes internos al formato `contents`/`parts` de la API de Gemini, mapea tool calling y streaming SSE, y normaliza errores a `ProviderError`.
- `adapter-openrouter`: Adapter para OpenRouter. Wrapper OpenAI-compat con cabecera `HTTP-Referer` requerida, soporte de model IDs de OpenRouter y propagación de errores del upstream al failover.
- `adapter-omniroute`: Adapter para OmniRoute. Proxy configurable que delega a un endpoint externo; soporte de autenticación configurable y traducción de errores estándar.

### Modified Capabilities

_(ninguna — ningún spec existente cambia sus requisitos)_

## Impact

- Código: nuevos paquetes bajo `src/internal/adapter/aihubmix/`, `adapter/google/`, `adapter/openrouter/`, `adapter/omniroute/`.
- Dependencias: `golang.org/x/net` ya presente; las peticiones HTTP usan `net/http` estándar. Sin dependencias externas nuevas.
- Registry: para usar estos adapters en producción el usuario agrega providers en `config/providers.yaml`; no hay cambios en el código del Registry.
- Sin breaking changes en APIs públicas.
