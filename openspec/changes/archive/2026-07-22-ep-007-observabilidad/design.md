## Context
El sistema requiere observabilidad para exponer latencias, costos y success rates (EP-007). Estos datos deben persistirse (Histórico) para posibilitar el auto-ajuste de pesos (Learning Engine) y deben posibilitar que la plataforma devuelva información de uso agregada (Métricas). Además se incorpora una caché semántica local ligera (Fase 2) para ahorrar peticiones.

## Goals / Non-Goals
**Goals:**
- Exponer un endpoint de métricas agregadas `/v1/metrics/dashboard`.
- Persistir asíncronamente cada petición (modelo, costo, latencia, resultado).
- Ajustar pesos del router automáticamente mediante heurística.
- Cache semántica rápida (<50ms) en memoria sin dependencias externas pesadas.

**Non-Goals:**
- Desarrollar la UI del dashboard (vive en otro repo).
- Uso de bases de datos vectoriales externas complejas en este MVP.

## Decisions
- **Decisión 1**: Histórico basado en el Store existente.
  - *Razón*: Reduce dependencias operacionales; el volumen inicial puede ser manejado con la persistencia establecida en EP-009.
- **Decisión 2**: Librería zero-deps local para Vector Search.
  - *Razón*: Evitar la latencia y complejidad de una DB externa hasta validar valor real de la caché.
- **Decisión 3**: Heurística simple para Learning Engine.
  - *Razón*: Más explicable, evitable el sobreajuste inicial, y más fácil para aplicar rollbacks a fallas.

## Risks / Trade-offs
- [Risk] Crecimiento desmedido de BD histórica → Mitigation: Retención configurable con purga automática.
- [Risk] Falsos positivos en caché semántica → Mitigation: Umbral conservador (>0.98) y by-pass de prompts cortos.
- [Risk] Ciclo de aprendizaje daña performance → Mitigation: Guardrails topados y rollback inmediato de pesos.
