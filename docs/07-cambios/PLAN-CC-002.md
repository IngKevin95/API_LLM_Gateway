# Plan de Construcción: CC-002 (Fase 2 - Autonomía y Gobernanza)

## 1. Secuencia de Slices (Critical Path)

### Slice 1: Motores Asíncronos (Refactor Core)
**Épica**: EP-EVO-004
**Historias**: HU-EVO-016
- **Tareas**: 
  - Inicializar migraciones en PostgreSQL para la tabla `quotas`.
  - Crear el script de **Bootstrap Idempotente** (lee `config.yaml`, cifra llaves, inserta en DB).
  - Implementar el `Sync Worker` asíncrono con **Bounded Queues** y sentencias SQL atómicas (`UPDATE quotas SET tokens = tokens - X`).
- **Verificación**: Unit tests validando que no haya OOM si se simula un retardo largo en Postgres. Verificación de lectura atómica en concurrencia (goroutines).
- **Estimación**: 5 días.

### Slice 2: Semantic Cache Vectorial
**Épica**: EP-EVO-004
**Historias**: HU-EVO-017
- **Tareas**:
  - Habilitar extensión `pgvector` vía migración SQL.
  - Implementar la lógica del `Router` para interceptar requests, generar el Embedding y calcular `Cosine Similarity > 0.95`.
  - Desplegar la respuesta cacheada con cabecera HTTP `X-Cache: HIT`.
- **Verificación**: Tests de integración mockeando la base de datos vectorial para validar hit rates.
- **Estimación**: 4 días.
- **Dependencia**: Depende del éxito del Slice 1 (conexión PostgreSQL consolidada).

### Slice 3: Interfaz Administrativa y Gobernanza
**Épica**: EP-EVO-005
**Historias**: HU-EVO-018, HU-EVO-019
- **Tareas**:
  - Crear proyecto Frontend (Dashboard UI) en React/Vite.
  - Implementar integración SSO (OAuth2).
  - Crear `Admin API` (Go) para exponer métricas y crear Tokens.
  - Implementar visualización de gráficos de consumo en Dashboard.
- **Verificación**: Pruebas End-to-End de inicio de sesión, denegación 403 sin rol, y comprobación visual de la gestión de tokens.
- **Estimación**: 6 días.
- **Dependencia**: Puede paralelizarse parcialmente con el Slice 2 (construcción del UI mientras se hace el backend vector), pero requiere el Admin API finalizado.

## 2. Resumen y Riesgos del Plan
- **Duración Estimada Total**: 15 días laborables (~3 semanas).
- **Paralelización**: El desarrollo del frontend de Dashboard (Slice 3) puede arrancar en paralelo a la Caché Vectorial (Slice 2).
- **Riesgos de Cronograma**: Ajustar los parámetros óptimos del índice `pgvector` (ej. HNSW) para mantener latencias < 20ms podría tomar ciclos extra de testing (buffer de 2 días).
