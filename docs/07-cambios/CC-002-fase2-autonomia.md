# Control de Cambios: CC-002 (Fase 2 - Autonomía y Gobernanza)

## 1. Relaciones e Impacto Directo
- **Épicas Afectadas**: 
  - `EP-007` (Limitador de Tasa y Costos): Se altera drásticamente al pasar de In-Memory total a In-Memory + Asíncrono DB.
  - `EP-009` (Persistencia Base): Se debe expandir el esquema base para soportar pgvector y credenciales encriptadas.
  - `EP-004A` (Caché): Se depreca o complementa la caché en RAM base por una caché persistida en base de datos.
- **Historias Impactadas**: `HU-038`, `HU-032`, `HU-061`, `HU-062`. Los tests E2E actuales sobre cuotas que asumen reinicio de estado podrían fallar al introducir persistencia.

## 2. Dependencias Técnicas Nuevas
- **PostgreSQL Extension**: Se requiere que la base de datos tenga `CREATE EXTENSION vector;`. Si el DBaaS del cliente (ej. AWS RDS) no lo soporta o tiene una versión desactualizada, la caché fallará en arranque.
- **Identidad (OIDC/OAuth2)**: El Dashboard requiere un `clientId` y `clientSecret` de un proveedor externo.

## 3. Riesgos de Cambios en Producción (🚨 ALTO RIESGO)
1. **Caída de Latencia Crítica / Riesgo OOM**: Si la base de datos presenta un cuello de botella, la memoria RAM podría saturarse encolando transacciones.
2. **Race Conditions en Facturación**: Al ser cuota asíncrona en despliegue multi-nodo, los deductos de un nodo pueden sobreescribir los de otro nodo.
3. **Pérdida de Configuración (Backward Incompatibility)**: Administradores usando YAML en CI/CD quedarán sin servicio al migrar a DB.

## 4. Estrategias de Mitigación Aprobadas (Requisitos de Construcción)
1. **OOM & Latencia (Bounded Queues + WAL en Disco)**: El Sync Worker de Go utilizará un `channel` con límite estricto de capacidad (ej. 5000) e implementará backpressure devolviendo 429. El WAL se volcará a disco en lugar de RAM pura.
2. **Race Conditions (Atomic SQL Updates)**: La persistencia de cuotas se hará estrictamente vía SQL atómico `UPDATE quotas SET tokens = tokens - $1` y nunca mediante un read-modify-write.
3. **Retrocompatibilidad (Bootstrap Idempotente)**: Al inicializar el Gateway, el sistema ejecutará un pipeline que leerá los tokens del YAML legado, los cifrará con KMS y los insertará a PostgreSQL (`ON CONFLICT DO NOTHING`) para no quebrar despliegues de CI/CD.
