## Why

Los agentes consumidores no deben acoplarse a un proveedor ni a un modelo concreto: piden una **capacidad** (`chat`, `reasoning`, `coding`, `vision`, `image`, `embedding`) y el Gateway resuelve el modelo óptimo en tiempo de ejecución. Sin esta capa de enrutamiento por capacidad no existe el desacople que da razón de ser al proyecto, y es prerequisito de todas las demás épicas. Es la primera épica `foundational` (EP-001).

## What Changes

- Nuevo **Registry**: carga declarativa de `config.yaml` (providers, models, routing) a memoria RAM en el arranque, con validación estricta y resolución de secretos por `${VAR}` (nunca literales).
- Nuevo **Model Router**: resuelve `capacidad → modelo` por un score heurístico compuesto de 6 variables (calidad, velocidad, disponibilidad, cuota restante, costo, latencia), en modo automático (sin `model`) y explícito (con `model` + política de fallback).
- Nueva **validación de contexto** (Context Window): estimación de tokens con buffer de seguridad del 20% que descarta candidatos cuyo payload excede la ventana antes de calcular el score.
- Manejo determinista de errores y desempates del enrutamiento (empates por score, capacidad sin modelos, cadena agotada).
- NO se exponen endpoints HTTP nuevos en esta épica: la API universal OpenAI/Anthropic-compat es EP-005. El enrutamiento se consume como librería interna del Gateway.

## Capabilities

### New Capabilities
- `registry`: carga y validación declarativa de providers/models/routing desde `config.yaml` a RAM; expone modelos habilitados por capacidad y parámetros de red físicos (`max_in_flight`, `stream_idle_timeout`) al resto del Gateway. Secretos solo vía `${VAR}`.
- `model-router`: resolución de capacidad a modelo por score heurístico de 6 variables; modo automático y explícito con política de fallback; manejo determinista de errores y desempates.
- `context-validation`: tokenizador de contexto y validación de ventana con buffer del 20%, aplicado como filtro pre-score de candidatos.

### Modified Capabilities
<!-- Ninguna: no existen specs previos en openspec/specs/ (primera épica del proyecto). -->

## Impact

- Código nuevo (Go): paquetes `internal/registry`, `internal/router`, `internal/tokenizer` (nombres tentativos, se fijan en design.md). Sin cambios al entrypoint `cmd/gateway` en esta épica salvo cableado de carga del Registry en boot.
- Dependencias: parser YAML (a evaluar en design.md contra `stack-allowlist.json`; el scaffold es stdlib-only). Tokenizador: heurístico o `tiktoken-go` según HU-035.
- Configuración: contrato `config.yaml` per Anexo A del PRD técnico (`docs/13-tech-prd/api-llm-gateway.md`).
- Sin impacto en APIs públicas (no hay endpoints nuevos); sin breaking changes (greenfield).

## Trazabilidad

- **Épica**: EP-001 · Enrutamiento por capacidad (`layer: foundational`) — objetivos del PRD: Obj. 1 (desacople total), Obj. 3 (selección óptima por score).
- **Historias cubiertas**:
  - HU-001 — Cargar providers/models/routing desde YAML (Registry) → capacidad `registry`
  - HU-002a — Resolver capacidad → modelo por score (Router automático) → capacidad `model-router`
  - HU-002b — Manejo de errores y desempates en el enrutamiento → capacidad `model-router`
  - HU-003 — Forzar modelo explícito vía parámetro `model` con política de fallback → capacidad `model-router`
  - HU-035 — Tokenizador de Contexto (Context Window) → capacidad `context-validation`
