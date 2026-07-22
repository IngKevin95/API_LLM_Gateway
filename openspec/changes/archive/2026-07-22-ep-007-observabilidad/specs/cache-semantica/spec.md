## ADDED Requirements

### Requirement: Caché semántica con validación por umbral
El sistema SHALL interceptar prompts muy similares usando un umbral alto de similitud vectorial para devolver respuestas cacheadas y ahorrar tokens.

#### Scenario: Hit semántico exacto (similitud alta)
- **WHEN** un prompt tiene similitud mayor al 0.98 con uno en caché
- **THEN** devuelve la respuesta desde la memoria evitando el LLM proveedor

#### Scenario: Cache miss (similitud baja)
- **WHEN** la similitud está por debajo del umbral configurado
- **THEN** procesa la petición enviándola al LLM

#### Scenario: Bypass para prompts cortos
- **WHEN** el prompt enviado es muy corto
- **THEN** evita la búsqueda vectorial para prevenir coincidencias imprecisas

#### Scenario: Fallo asíncrono o lentitud
- **WHEN** la validación semántica toma mucho tiempo (timeout interno de <50ms)
- **THEN** el sistema interrumpe la caché y envía el request original sin bloqueo
