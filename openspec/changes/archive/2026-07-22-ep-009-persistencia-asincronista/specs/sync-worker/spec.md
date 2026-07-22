## ADDED Requirements

### Requirement: Persistencia asincronista por batching
El Sync Worker SHALL consumir eventos de un channel y persistirlos en batches (cada 1s o 1000 eventos) vía el Store, sin bloquear al handler. (Traza: HU-038 AC1/AC5)

#### Scenario: Eventos persistidos en batch
- **WHEN** el handler encola un evento y el Sync Worker procesa el batch (cada 1s o 1000 eventos)
- **THEN** inserta el lote en el Store sin bloquear al handler

#### Scenario: Throughput sostenido
- **WHEN** ingresan grandes volúmenes de eventos
- **THEN** el Sync Worker procesa batches de 1000+ eventos sin lag (el objetivo de 43M/día se valida en load-test)

### Requirement: Cifrado KMS Envelope antes de persistir
El Sync Worker SHALL cifrar cada evento con KMS Envelope (DEK local, KEK en KMS) mediante el Encryptor antes de escribir al Store. (Traza: HU-038 AC4)

#### Scenario: Cifrado local con DEK
- **WHEN** el Sync Worker procesa eventos y obtiene la llave maestra vía el Encryptor
- **THEN** cifra localmente con la DEK antes de escribir al Store

### Requirement: Manejo de backpressure sin bloquear
El Sync Worker SHALL manejar la saturación del channel con retry con jitter o descarte de eventos de baja prioridad, sin fallar críticamente ni bloquear al handler. (Traza: HU-038 AC2/AC3)

#### Scenario: Channel saturado
- **WHEN** el channel está lleno y el handler intenta encolar
- **THEN** aplica retry con jitter o dropea eventos de baja prioridad, sin fallar críticamente

#### Scenario: Pérdida ante caída delega en el WAL
- **WHEN** el Sync Worker muere sin flushear y el Gateway reinicia
- **THEN** el WAL local retiene los eventos no persistidos para la recuperación (ver write-ahead-log)
