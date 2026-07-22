---
status: drafting
---

# EP-007: Observabilidad y aprendizaje

## 1. Por qué
Sin métricas reales el score usa pesos fijos y no mejora; la observabilidad es condición para operar en producción y para el autoaprendizaje futuro. Permite también la reducción de costos mediante caché semántico (Obj. 2, Obj. 3 del PRD).

## 2. Qué cambia
- **Exposición de Métricas**: Endpoints para consultar métricas agregadas y ranking por modelo/proveedor. API para dashboard visual independiente.
- **Histórico de Peticiones**: Registro asíncrono y detallado (modelo, costo, tokens, latencia, resultado) de cada petición procesada para auditoría y aprendizaje.
- **Learning Engine**: Ciclo heurístico que ajusta dinámicamente los pesos del router con base en latencias y success rate históricos.
- **Caché Semántica**: Middleware ligero de vector search en memoria (<50ms overhead) para responder peticiones similares desde caché local antes de despachar al proveedor.

## 3. Capacidades
- exposicion-metricas
- historico-peticiones
- learning-engine
- cache-semantica

## 4. Impacto
- Incremento de rendimiento y ahorro de tokens mediante hit semántico local (caché).
- Adaptación autónoma de la calidad del enrutamiento.
- Alta visibilidad para decisiones operativas del equipo de infraestructura.

## Trazabilidad
- Épica: EP-007
- Sub-slice 1 (Exposicion de Metricas): HU-017, HU-023
- Sub-slice 2 (Historico de Peticiones): HU-018
- Sub-slice 3 (Ajuste de Pesos / Learning Engine): HU-019
- Sub-slice 4 (Cache Semantica): HU-032
