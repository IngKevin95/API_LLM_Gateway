---
id: ADR-001
titulo: Backend en Go (Golang)
status: accepted
deciders: Kevin Beltrán, Equipo de Arquitectura
date: 2026-07-19
---

# ADR-001: Backend en Go (Golang)

## Context and Problem Statement

El API LLM Gateway actuará como un enrutador de red de altísimo rendimiento (I/O intensivo) para interceptar, inspeccionar y redirigir tráfico hacia múltiples proveedores LLM. Las latencias bajas (overhead < 100ms) y la alta concurrencia (soporte para 500 RPS continuos con ~2500 conexiones concurrentes en streaming) son críticas para la viabilidad del proyecto. ¿Qué stack tecnológico de backend debemos utilizar para garantizar este throughput y resiliencia sin introducir latencias artificiales?

## Decision Drivers

- Necesidad de overhead de enrutamiento p95 < 100ms sobre la latencia del proveedor.
- Soporte nativo para altísima concurrencia en I/O y streaming de red.
- Facilidad de despliegue y empaquetamiento (idealmente binario único).
- Fuerte tipado para el contrato de la API.

## Considered Options

- Go (Golang)
- Python + FastAPI
- Node.js + TypeScript (Fastify)

## Decision Outcome

Chosen option: "Go (Golang)", because it offers stellar performance, native concurrency handling (goroutines) ideal for a network proxy, and compiles to a single static binary.

### Positive Consequences

- Rendimiento estelar y bajísima latencia en el routing y streaming.
- Manejo de concurrencia nativo eficiente (goroutines), mitigando cuellos de botella de red.
- Tipado fuerte estático y compilación a binario único (facilita el despliegue).

### Negative Consequences

- El ecosistema de SDKs de IA es más pequeño comparado con Python, obligando a implementar adaptadores manuales mediante structs.
- Para evitar vulnerabilidades de ReDoS en el escaneo por regex, se requiere una librería de regex de tiempo lineal (re2), cuyo binding en Go exige CGO (librerías en C). El uso de CGO complica el tooling y complica ligeramente la compilación, forzando el uso de musl (Alpine) para mantener la promesa del binario estático con cross-compilation.

## Pros and Cons of the Options

### Go (Golang)

- Good, because it handles thousands of concurrent I/O connections efficiently natively.
- Good, because it compiles to a single static binary, simplifying deployment.
- Bad, because the AI ecosystem is smaller.
- Bad, because security (regex) requires CGO.

### Python + FastAPI

- Good, because it has a massive native AI ecosystem (LangChain, LlamaIndex, SDKs).
- Good, because data validation with Pydantic is robust and standard in AI.
- Bad, because higher memory footprint and the GIL can limit high concurrent I/O throughput without heavy scaling.

### Node.js + TypeScript (Fastify)

- Good, because it offers high async throughput and a massive developer ecosystem.
- Bad, because CPU-bound performance is lower than Go.
- Bad, because runtime validation (Zod/TypeBox) requires manual mapping compared to struct unmarshaling.

## Links

- [PRD Product](../01-prd/api-llm-gateway.md) — objetivos de producto y NFR de disponibilidad
- [Technical Discovery Brief](../10-tech-discovery/api-llm-gateway.md) — fuente de los drivers (overhead p95 < 100ms, 500 RPS, redacción síncrona)
- [Documento de Arquitectura](../11-architecture/api-llm-gateway.md) — selección de stack y trade-offs
- [Technical PRD](../13-tech-prd/api-llm-gateway.md) — contrato técnico consumido por construcción
- [Stack allowlist](../11-architecture/stack-allowlist.json) — backend: Go
