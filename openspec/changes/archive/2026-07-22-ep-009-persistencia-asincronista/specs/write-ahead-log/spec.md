## ADDED Requirements

### Requirement: WAL append-only durable
El WAL SHALL registrar cada evento en un log append-only en disco antes del flush a DB, con overhead mínimo en el camino crítico. (Traza: HU-039 AC1/AC5)

#### Scenario: WAL registra eventos
- **WHEN** el Sync Worker bufferea 1000 eventos antes del flush a DB
- **THEN** el WAL contiene los 1000 registros serializados en disco (append-only)

#### Scenario: Overhead mínimo
- **WHEN** se escriben eventos en el WAL en el camino crítico
- **THEN** la latencia adicional es mínima (append-only, sin sync forzado en el camino crítico; el objetivo <1ms/evento se valida en load-test)

### Requirement: Rotación del WAL por tamaño
El WAL SHALL archivar el log activo al superar un umbral de tamaño y comenzar uno nuevo. (Traza: HU-039 AC3)

#### Scenario: Rotación al alcanzar el umbral
- **WHEN** el WAL alcanza su tamaño máximo y el Sync Worker ya flusheó a DB
- **THEN** archiva el WAL (p. ej. `wal-20260720-001.log`) y comienza uno nuevo

### Requirement: Lectura de recuperación
El WAL SHALL exponer la lectura de todos los registros no flusheados (activo + archivados) para el replay de recuperación ante crash. (Traza: HU-039 AC2/AC4)

#### Scenario: Recuperación ante crash
- **WHEN** el Gateway muere con eventos en el WAL no flusheados y reinicia
- **THEN** la lectura de recuperación devuelve esos eventos para que el Recovery Worker los inserte en DB antes de aceptar tráfico

#### Scenario: Durabilidad ante crash de OS
- **WHEN** el OS crashea mid-write y el sistema recupera
- **THEN** el journal del FS + la recuperación del WAL permiten recuperar hasta el último registro consistente (garantía absoluta de fsync configurable en despliegue)
