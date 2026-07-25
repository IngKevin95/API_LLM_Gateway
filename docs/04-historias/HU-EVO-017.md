---
id: HU-EVO-017
title: Caché Semántica con pgvector
epic: EP-EVO-004
type: evolutivo
status: pending
---

# HU-EVO-017: Caché Semántica con pgvector

**Como** Router de Modelos
**Quiero** consultar una base de datos vectorial (pgvector) antes de enviar el prompt al proveedor externo
**Para** devolver la respuesta desde la caché si la similitud coseno (Cosine Similarity) del prompt es > 0.95, ahorrando tiempo y costos.

## Criterios de Aceptación

**Escenario 1: Cache Hit Exitoso**
- **Given** que entra un prompt cuya versión vectorizada tiene un 97% de similitud con un registro previo en pgvector
- **When** el Router consulta la caché semántica
- **Then** el Gateway debe devolver la respuesta guardada
- **And** la respuesta debe incluir un header `X-Cache: HIT` y consumirse a costo $0 en la cuota externa.

**Escenario 2: Cache Miss e Inserción**
- **Given** que entra un prompt totalmente nuevo
- **When** el Router no encuentra coincidencias (> 0.95) en pgvector
- **Then** el Gateway enruta normalmente al proveedor LLM
- **And** de forma asíncrona guarda el nuevo prompt, su embedding y la respuesta en pgvector para futuras consultas.
