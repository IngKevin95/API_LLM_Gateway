# Technical Discovery Brief: Fase 2 - Autonomía y Gobernanza

## 1. Contexto de Negocio
El MVP (R1) del API LLM Gateway está en producción con routing in-memory y autenticación por yaml. El objetivo de la Fase 2 es dotar a la plataforma de autonomía total, persistencia robusta, caché semántica y una interfaz de gobernanza.

## 2. Requerimientos No Funcionales (NFRs)
- **Persistencia (PostgreSQL)**: RPO < 1 minuto, RTO < 5 minutos.
- **Semantic Cache (Vector DB)**: Latencia p95 < 50ms para cache hits.
- **UI Dashboard**: Latencia p95 < 200ms para dashboards analíticos.
- **Escalabilidad**: Soportar hasta 10,000 requests por minuto concurrentes.

## 3. Seguridad y Cumplimiento
- **Gestión de Tokens**: Las API Keys creadas por administradores en el Dashboard deben cifrarse usando KMS Envelope Encryption antes de persistir en PostgreSQL.
- **Redacción de PII**: Ya implementado en Fase 1, debe mantenerse en logs hacia PostgreSQL.

## 4. Integraciones
- **PostgreSQL**: Para Quota Manager, Credentials y Auditoría.
- **Vector DB (ej. pgvector/Milvus)**: Para Semantic Cache.
- **Proveedor LLMs**: Integración actual mantenida, pero las llaves se consumen desde la DB.
