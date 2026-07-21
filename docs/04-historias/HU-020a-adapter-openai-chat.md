---
id: HU-020a
titulo: Adapter OpenAI — chat y tool calling
epica: EP-002
prioridad: Must
complejidad: S
estado: lista
---

# Adapter OpenAI — chat y tool calling

Como **desarrollador de la plataforma**, quiero **un adapter que traduzca el formato interno al `/v1/chat/completions` de OpenAI incluyendo tool calling**, para **consumir los modelos GPT de forma nativa en peticiones de chat**.

Contexto: adapter base que heredarán proveedores compatibles con OpenAI (vLLM, LM Studio). Streaming y embeddings se cubren en HU-020b y HU-020c.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — chat básico | Dado que un payload de chat normal en la Gateway | Cuando el router selecciona un modelo de OpenAI | Entonces el adapter transforma la request al formato `/v1/chat/completions` y devuelve la respuesta normalizada |
| 2 | Edge — Tool Calling | Dado que el request incluye una definición de tools | Cuando se enruta a OpenAI | Entonces el adapter preserva intacto el schema de tools y el formato de llamada de función en la respuesta |
| 3 | Error — timeout o falla externa | Dado que OpenAI está caído (ej. 500) | Cuando el adapter intenta la llamada | Entonces captura el timeout/error y retorna el formato estandarizado de falla para que la Gateway inicie el failover |

## Checklist INVEST

- [x] Independent — no depende de otros adapters
- [x] Negotiable — detalles de mapeo abiertos
- [x] Valuable — integra el LLM líder del mercado para chat, requisito de v1.0
- [x] Estimable — API de OpenAI documentada
- [x] Small — solo chat + tool calling
- [x] Testable — se mockea el endpoint HTTP de OpenAI

## Notas técnicas

Modelo base para adapters OpenAI-compat. El streaming (020b) y embeddings (020c) reutilizan esta traducción.
