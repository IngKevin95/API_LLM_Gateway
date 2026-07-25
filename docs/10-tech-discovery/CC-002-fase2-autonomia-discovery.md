# Technical Discovery Brief: CC-002 - Fase 2 (Autonomía y Gobernanza)

## 1. Contexto del Cambio
El API LLM Gateway pasará de su estado MVP (validación in-memory, configuración estática YAML) a una arquitectura escalable de Nivel Enterprise. Se requiere persistencia, caché inteligente y una UI administrativa para gestionar proveedores de forma autónoma.

## 2. Decisiones Arquitectónicas del Discovery
- **Stack Vectorial**: Se usará **`pgvector`** instalado como extensión sobre la instancia de PostgreSQL, consolidando la infraestructura de persistencia relacional y vectorial en un solo motor.
- **Autenticación UI**: El Dashboard de Administración se integrará con un servicio de autenticación externo (SSO/OAuth2) (ej. Auth0, Keycloak o Google Workspace), no almacenando contraseñas propias.
- **Tolerancia a Latencia (Quotas)**: La validación de límites de cuota será **asíncrona** (caché RAM en L1 sincronizada periódicamente contra PostgreSQL L2) para garantizar un overhead de 0ms adicionales en el enrutamiento.
- **Caché Semántica**: Diseñada para la **mínima latencia posible**, indexando los vectores mediante HNSW o IVFFlat en `pgvector` para respuestas en crudo de menos de 10-15ms.

## 3. Riesgos y Dependencias Detectadas
- **Dependencia Crítica**: La base de datos PostgreSQL se vuelve un Single Point of Failure (SPOF) para la Caché Semántica y persistencia asíncrona (WAL mitiga pérdida de auditoría, pero no caídas de Caché).
- **Riesgos de Concurrencia**: Al usar cuotas asíncronas, un pico de tráfico podría exceder momentáneamente la cuota estricta (over-provisioning temporal) hasta que el Sync Worker aplique el bloqueo. Se acepta este trade-off a favor de la latencia.

## 4. Usuarios Impactados
- **Administradores de Plataforma**: Quienes gobernarán las llaves (API Keys) de los proveedores desde el Dashboard UI, en lugar de reiniciar nodos modificando un YAML.
- **Desarrolladores/Agentes**: Beneficiados directamente por respuestas instantáneas de la Caché Semántica.
