---
status: "approved"
last_updated: "2026-07-19"
---

# Technical Discovery Brief: API LLM Gateway

## 1. Requisitos No Funcionales (NFRs)

> **Stack**: Backend seleccionado en [ADR-001: Backend en Go (Golang)](../12-adr/ADR-001-backend-stack.md). Los NFRs abajo están optimizados para esa arquitectura.
- **Disponibilidad**: SLO objetivo de éxito ≥ 99.9% uptime mensual para el Gateway (diseño HA con failover multi-proveedor y degradación a modelos locales).
- **Rendimiento**: Overhead estricto p95 < 100 ms (enrutamiento puro en RAM, excluye latencia externa del proveedor). TTFT dinámico (mayor para modelos de reasoning). El tiempo de TTFT externo (del proveedor) no cuenta el procesamiento interno. Redacción síncrona en memoria (< 10ms por cada 10k tokens) excluyendo bloques Base64/imágenes para prevenir falsos timeouts. Escaneo PII asíncrono.
- **Latencia de Auth y Rate Limiting**: Validación de JWT/API Key y evaluación de rate limits **< 5ms mediante RAM cache** (in-memory). Cero I/O a base de datos en camino crítico.
- **Validación de Contexto**: Estimación de tokens mediante `tiktoken-go` aplicando un margen de seguridad (buffer) del 20% para absorber variaciones propias de Anthropic/Llama, descartando peticiones que excedan la ventana antes de enrutar.
- **Guardián de Prompts (Opt-in)**: Si la optimización está activa, el SLA de overhead se relaja a < 1.5s (sin penalizar el límite TTFT hacia el proveedor).
- **Escalabilidad**: Throughput de 500 RPS continuos (~2500 peticiones concurrentes) para endpoints de texto. Payloads: cuerpo JSON/texto limitado a 10MB (rechazo inmediato `413 Payload Too Large`); contenido binario para capacidad `vision` soporta hasta 50MB vía ruta separada no-síncrona. Tráfico de visión limitado a 2 peticiones concurrentes (asumiendo red interna de 10Gbps para evitar bufferbloat por saturación física).
- **Recuperación**: RTO < 1h, RPO < 15m. Latencia máxima permitida del Quota Manager y Auth: < 5ms mediante in-memory cache. Política de retención de base de datos: Logs purgados cada 30 días (soporte para ~43M logs/día). Se exige compresión pre-cifrado (ej. zstd) de los payloads antes de persistirlos, reduciendo el footprint de disco de PostgreSQL. Exige obligatoriamente particionamiento de tablas (Table Partitioning) por rango temporal (diario/semanal) para usar DROP TABLE y evitar degradación de I/O.

## 2. Integraciones y Servicios Externos
- **Clientes (Entrada)**: OpenCode, Free Claude Code, y aplicaciones integradoras vía API compatible (OpenAI y Anthropic). Auth: API Key, OAuth2/OIDC (agnóstico), mTLS.
- **Proveedores LLM (Salida)**: OpenAI, Anthropic, Google, OpenRouter, AIHubMix, Ollama, vLLM, LM Studio. Auth: Bearer/API Keys en headers nativos por adapter.
- **Observabilidad**: Exportación de métricas para Prometheus y Grafana; logs de auditoría en formato estructurado (JSON/stdout).

## 3. Datos y Seguridad
- **Datos Sensibles**: Prompts y respuestas. Pueden contener PII o datos empresariales. Obligatoria su redacción o exclusión de los logs de auditoría y trazas de error.
- **Secretos del Servidor**: Claves (API Keys) de los proveedores LLM. Estrictamente prohibido su versionado en archivos YAML o inclusión en logs. Gestión mediante variables de entorno o Secret Manager.
- **Protección**: Rate Limiting y Quota Manager por clave/tenant para evitar abusos y sobrecostos.
