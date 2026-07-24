## Context

El Gateway implementa hoy un adapter Go dedicado por proveedor (`src/internal/adapter/openai`,
`anthropic`, `google`, `openrouter`, `omniroute`, `aihubmix`, `local`). Cada nuevo proveedor
gratuito prioritario (Groq, Cerebras, Mistral, Gemini, Cloudflare AI) es en la práctica
OpenAI-compatible o Claude-compatible en su wire format — agregar un paquete Go por proveedor
duplica lógica de traducción, serialización de streaming y manejo de errores ya resuelta. Registry
(`src/internal/registry/registry.go`), Health Monitor (`src/internal/health/health.go`) y Quota
Manager (`src/internal/quota/manager.go`) ya existen y este change los extiende sin reescribirlos.

## Goals / Non-Goals

**Goals:**
- Permitir declarar un proveedor OpenAI-compatible o Claude-compatible vía `ProviderSpec` (YAML),
  sin escribir un paquete Go nuevo.
- Cargar los 5 proveedores free-tier priorizados desde `config/providers/free-tier.yaml`.
- Validar automáticamente (conformance test) que cada `ProviderSpec` cumple el contrato
  Chat/Stream/Embed.
- Retirar temporalmente proveedores que devuelven 429, con reactivación automática y backoff.
- Inicializar la cuota estimada por proveedor desde `quota_hint`, con precedencia del valor
  aprendido en runtime/persistido.

**Non-Goals:**
- No se cubren proveedores con wire format propietario incompatible con OpenAI/Claude (quedan
  para adapters dedicados, como hoy).
- No se modifica el contrato público HTTP del Gateway (`/v1/chat/completions`, etc.).
- No se agrega UI ni endpoint nuevo — es una extensión interna de Registry/Adapter/Health/Quota.

## Decisions

### 1. Adapter genérico via `ProviderSpec` + strategy de formato
Se introduce `src/internal/adapter/generic` con un tipo `Adapter` parametrizado por `ProviderSpec`
(`BaseURL`, `AuthHeader`, `Format`, `Headers map[string]string`, `TimeoutMs`). El campo `Format`
selecciona en runtime una de dos estrategias de traducción ya probadas: reutiliza la lógica de
`adapter/openai` para `format: "openai"` y la de `adapter/anthropic` para `format: "claude"`,
extraída a funciones compartidas en vez de duplicarlas.
- Alternativa descartada: generar un paquete Go por proveedor nuevo (statu quo) — rechazada porque
  no escala a 20+ proveedores y duplica tests de conformance.
- Alternativa descartada: un único formato "universal" con traducción propia — rechazada porque
  ya existen dos adapters (openai, anthropic) con traducción correcta y probada; reformularla
  introduce riesgo sin beneficio.

### 2. `free-tier.yaml` como archivo de config adicional, no reemplazo de `config.yaml`
`Registry.Load()` se extiende para leer `config/providers/free-tier.yaml` después de
`config.yaml`, mergeando por `providerID`: si un proveedor aparece en ambos, la entrada de
`free-tier.yaml` gana (los `quota_hint` ahí son más realistas y curados). Mantiene la validación
fail-fast existente (`ErrInvalidConfig`) para el archivo nuevo.
- Alternativa descartada: fusionar todo en `config.yaml` — rechazada porque el catálogo free-tier
  se actualiza con más frecuencia y separarlo evita conflictos de merge en configuración de
  producción.

### 3. Conformance test data-driven sobre providers registrados
`conformance_test.go` (existente en `src/internal/adapter`) se extiende con un caso table-driven
que itera `Registry.AllProviderSpecs()` y ejecuta Chat/Stream/Embed contra un servidor de test
(`httptest.Server`) simulando cada `format`. Usa `t.Parallel()` por proveedor y `context.WithTimeout`
individual para no bloquear la suite ante un proveedor lento.

### 4. Retiro temporal en Health Monitor basado en tabla de blacklist en memoria
Se agrega a `src/internal/health/health.go` un mapa `providerID -> retiredUntil time.Time` y un
contador de fallos consecutivos por 429 para calcular backoff exponencial (30s/60s/120s, tope
configurable). El Router consulta este estado vía el mismo `HealthSource` que ya expone el Health
Monitor — no se agrega una interfaz nueva.

### 5. Quota Manager: precedencia init < learned < persisted
`manager.go` inicializa `remaining` con `quota_hint` en boot; el aprendizaje desde headers de
rate-limit (ya cubierto por HU-EVO-006/007 en el roadmap evolutivo, fuera de alcance de este
change salvo la precedencia) sobrescribe ese valor; y en boot, si existe un valor persistido en
PostgreSQL (WAL/tabla de cuota, EP-009), este tiene precedencia sobre `quota_hint` — se restaura
antes de aplicar el default de YAML.

## Risks / Trade-offs

- [Riesgo] Traducción compartida entre `adapter/openai` y `adapter/generic` introduce acoplamiento
  entre paquetes → Mitigación: extraer solo funciones puras de traducción a un paquete interno
  (`adapter/internal/translate`), sin exponer tipos internos de `openai`/`anthropic`.
- [Riesgo] Merge de `free-tier.yaml` sobre `config.yaml` puede ocultar configuración de producción
  si un operador edita el proveedor equivocado → Mitigación: log INFO explícito listando qué
  `providerID` fueron sobrescritos por `free-tier.yaml` en cada boot.
- [Riesgo] Backoff exponencial sin tope puede dejar un proveedor retirado indefinidamente ante
  fallos sostenidos → Mitigación: tope máximo configurable (por defecto 120s) definido en esta
  primera iteración; alertas de retiro prolongado quedan fuera de alcance (HU-EVO-012, fuera de
  este change).
- [Riesgo] Providers sin `quota_hint` compartiendo el mismo default de 1M puede sobreestimar
  cuota real de proveedores muy limitados → Mitigación: aceptado como comportamiento conservador
  documentado; el aprendizaje desde headers corrige rápido en el primer request real.

## Migration Plan

- Cambio aditivo: no rompe adapters ni providers existentes en `config.yaml`.
- Deploy: agregar `config/providers/free-tier.yaml` al build/release; sin migración de datos.
- Rollback: remover o vaciar `free-tier.yaml` — Registry vuelve al comportamiento actual basado
  solo en `config.yaml`.

## Open Questions

- Ninguna bloqueante para iniciar TDD; el tope de backoff exponencial y el default de quota (1M)
  quedan como constantes configurables a ajustar con datos reales post-release.
