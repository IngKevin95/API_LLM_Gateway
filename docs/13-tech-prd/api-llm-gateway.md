---
status: "approved"
last_updated: "2026-07-19"
---

# Technical PRD: api-llm-gateway

Este documento actúa como el contrato técnico (`docs/13-tech-prd/api-llm-gateway.md`) que consumirá la fábrica de construcción, basado en el [Product PRD](../01-prd/api-llm-gateway.md). Consolida el discovery y la arquitectura en requerimientos accionables para desarrollo.

## 2. Componentes y Capas ([Ver Arquitectura](../11-architecture/api-llm-gateway.md))
- **Auth & Rate Limiting:** JWT/Keys (memoria, **< 5ms p99 via O(1) lookup, cero I/O en auth path crítico**), RBAC y multitenancy (aislamiento de datos por tenant), con Balanceo L7 (Hash de API Key).
- **Security & Redact:** CGO/Regex (Síncrono para texto, excluye Base64, con un límite estricto de 10MB limit para payloads de texto HTTP 413. La capacidad `vision` omite el OCR síncrono y acepta hasta 50MB estricto) + Kill-switch asíncrono profundo (PII). Protección contra Slowloris configurando `ReadHeaderTimeout` estricto en el servidor Go.
- **Model Router:** Calcula fallback usando scores heurísticos compuestos por 6 variables (calidad, velocidad, disponibilidad, cuota restante, costo, latencia). Estima tokens velozmente (`tiktoken-go` + 20% buffer de seguridad) para Validación de Context Window descontando adjuntos antes de rutear.
- **Prompt Guardian:** Optimización semántica y formateo (opt-in). Relaja el overhead SLA a < 1.5s cuando está activo.
- **Adapters (MVP/Fase 1):** OpenAI, Anthropic, Google Gemini, OpenRouter, AIHubMix, Modelos Locales (Ollama/vLLM/LM Studio). Traducen *System Prompts* y definiciones de *Tool Calling* unificadamente.
- **Quota Manager:** Cache L1 en RAM (con validación O(1)).
- **Registry (Config):** Carga declarativa de YAML a RAM en boot.
- **Failover Engine:** Controla el I/O respetando un umbral de TTFT dinámico (estricto de 2.0s para estándar, adaptativo para reasoning) y un Stream Idle Timeout (5s por defecto) mid-stream. Implementa un Circuit Breaker pasivo (Max In-Flight configurable en YAML (baseline de 300 para top tiers, 50 para locales) por proveedor) marcando nodos inalcanzables temporalmente para prevenir *Failover Suicide*. Si toda la cadena de fallback falla, retorna un `503 Service Unavailable` terminal al cliente.
- **Cache Semántica:** Almacena resultados previos para latencia 0ms (Fase 2, HU-032).
- **Learning Engine:** Ajuste dinámico de pesos del router basado en métricas (Fase 3).
- **Health Monitor:** Worker periódico que sondea proveedores para retirar/reactivar en el router.
- **Cache Invalidator:** Worker que refresca cuotas y llaves en memoria.
- **Sync Worker:** Sincroniza logs y métricas a DB, y encripta (KMS Envelope) asíncronamente en tiempo de ejecución.
- **Recovery Worker:** Lee el WAL local en el arranque para crash recovery e hidrata cachés, coordinándose con KMS.
- **CLI de Agente:** Agente local con tools nativas para (Archivos, Terminal, Git).

## 1. Capacidades Soportadas
El sistema enruta en base a capacidades abstractas solicitadas por el cliente: `coding`, `reasoning`, `vision`, `image`, `embedding` y `chat`.
## 1.1 Métricas de Rendimiento Obligatorias
- **Concurrencia y Carga**: Soporte para **500 RPS continuos (peticiones por segundo)** = suma de capacidad multi-proveedor, equivalente a **~2,500 conexiones simultáneas largas** en vuelo. Tráfico de visión estrictamente limitado a **2 peticiones concurrentes** por nodo usando política Least Connections (eludiendo el Hash L7) para evitar bufferbloat.
- **Volumetría de Datos**: PostgreSQL diseñado y particionado para ingerir ~**43 Millones de logs de auditoría diarios** (sujeto a un 30-day purge policy por partición mensual).
- **Memoria**: Uso de **Streaming Obligatorio** end-to-end. Los payloads grandes (ej. 50MB) nunca se cargan completos en memoria para evitar colapsos OOM (Out of Memory).
- **Recuperación**: RTO < 1h, RPO < 15m.

## 3. Datos Sensibles y Secretos ([Ver Discovery](../10-tech-discovery/api-llm-gateway.md))
- **Categorías de Datos Sensibles (`Prompts, Respuestas y Traza de Usuarios`)**:
  Pueden contener PII o propiedad intelectual del cliente. Estos campos no se persistirán en el `AuditLog` ni se expondrán en las salidas del logger general (stdout). El log de auditoría está restringido a metadata inmutable: (tokens, tiempo, status, costo, usuario, agente, herramientas, cache hit).

- **Secretos Server-Side (`API Keys y Secretos JWT`)**:
  Claves maestras de acceso a los LLMs. Restricción absoluta: jamás deben versionarse en el `config.yaml` ni volcarse en errores o logs. Se cargarán exclusivamente desde el entorno de ejecución (`env vars`) o Secret Manager.

## 4. Stack Tecnológico Base
- Go 1.22+ (Backend, [ADR-001](../12-adr/ADR-001-backend-stack.md))
- PostgreSQL 16+ (Quotas, Auditoría asíncrona inmutable vía KMS)
- Local RAM Cache (Manejo de estados, Rate Limiting y Quotas para < 100ms routing p95)

## 5. Decisiones Críticas (ADRs)
- **Decisiones de Alto Impacto (`ADR-001: Backend en Go`)**:
  Se desarrollará el Gateway enteramente en Go (Golang) para asegurar el throughput necesario (overhead p95 < 100ms con alta concurrencia). El diseño de adaptadores deberá ser idiomático y prescindir de frameworks pesados de IA.

## 6. Diseño e Interfaz y CLI
- **Interfaz API (OpenAI/Anthropic Compat)**:
  Soporte para Auth mTLS y OIDC. Streaming asíncrono para respuestas.
- **Arquitectura del CLI de Agente**:
  El sistema incluye un CLI que opera como agente local aislado. Ejecuta tools nativas de lectura/escritura de archivos y terminal interactiva encapsulada, actuando como cliente de facto de la Gateway simulando la experiencia de Claude Code y OpenCode.
- **Dashboard UI y Métricas**: Endpoint JSON (Fase 1) y Dashboard ligero de React/Next (Fase 2) para visualización.

## Anexo A: YAML Schema Canonical para `config.yaml` (Fase 1 MVP)

Estructura declarativa de proveedores, modelos y configuración de rutin:

```yaml
# config.yaml — API LLM Gateway MVP

gateway:
  port: 8080
  read_header_timeout_ms: 5000  # Slowloris protection

providers:
  - id: openai
    type: openai
    base_url: https://api.openai.com/v1
    api_key: ${OPENAI_API_KEY}
    max_in_flight: 300
    circuit_breaker:
      failure_threshold_percent: 50
      reset_timeout_sec: 30
    models:
      - name: gpt-4o
        capabilities: [chat, vision, reasoning]
        quality_score: 95
        latency_p50_ms: 800
        cost_per_1m_tokens: 15
      - name: gpt-4-turbo
        capabilities: [chat, vision]
        quality_score: 90
        latency_p50_ms: 600
        cost_per_1m_tokens: 10

  - id: anthropic
    type: anthropic
    base_url: https://api.anthropic.com
    api_key: ${ANTHROPIC_API_KEY}
    max_in_flight: 250
    models:
      - name: claude-opus-4
        capabilities: [chat, reasoning, vision]
        quality_score: 92
        latency_p50_ms: 900
        cost_per_1m_tokens: 20

  - id: local-ollama
    type: ollama
    base_url: http://localhost:11434
    max_in_flight: 50
    models:
      - name: mistral-7b
        capabilities: [chat, coding]
        quality_score: 70
        latency_p50_ms: 100
        cost_per_1m_tokens: 0

# Routing by capability (fallback chain)
routing:
  capabilities:
    chat:
      providers: [openai, anthropic, local-ollama]  # ordered fallback
      ttft_timeout_ms: 2000  # standard
      stream_idle_timeout_ms: 5000
      context_window_buffer_percent: 20
    
    reasoning:
      providers: [anthropic, openai]  # reasoning models first
      ttft_timeout_ms: 30000  # relaxed for thinking models
      stream_idle_timeout_ms: 60000
      context_window_buffer_percent: 20
    
    vision:
      providers: [openai, anthropic]
      ttft_timeout_ms: 5000
      stream_idle_timeout_ms: 10000
      max_concurrent_per_node: 2  # strict limit for vision payloads
      load_balancing: least_connections  # override L7 hash
    
    embedding:
      providers: [openai, anthropic]
      ttft_timeout_ms: 1000
      context_window_buffer_percent: 10

# Data persistence and encryption
persistence:
  database: postgresql://localhost/gateway
  audit_log_partition_by: month  # or week
  audit_log_ttl_days: 30
  kms_envelope:
    kms_endpoint: ${KMS_ENDPOINT}
    key_id: ${KMS_KEY_ID}
    
# Cache invalidation (Fase 2+)
cache_invalidator:
  enabled: false  # enable in Fase 2 when Cache Invalidator HU is implemented
  poll_interval_sec: 60
  webhook_url: ~  # optional: external trigger

# Semantic Cache (Fase 2, HU-032)
semantic_cache:
  enabled: false  # enable in Fase 2 when HU-032 is implemented
  similarity_threshold: 0.95
  embedding_model: all-MiniLM-L6-v2
  vector_db: pgvector

# Observability (Fase 1 basic, Fase 2+ advanced)
observability:
  metrics_endpoint: /metrics
  log_level: info
```

**Notas para desarrolladores:**
- `max_in_flight` por proveedor previene Failover Suicide
- `ttft_timeout_ms` varia por capacidad (reasoning es tolerante, estándar es estricto)
- `stream_idle_timeout_ms` corta streams colgados
- `context_window_buffer_percent: 20` aplicar a todas las estimaciones de tokens
- `max_concurrent_per_node: 2` solo para vision (evita bufferbloat)
- `load_balancing: least_connections` para vision (excepción a Hash L7)
