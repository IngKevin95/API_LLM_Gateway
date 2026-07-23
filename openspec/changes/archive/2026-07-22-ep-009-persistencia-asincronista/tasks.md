## 1. Sub-slice 1 — Write-Ahead Log (HU-039)

- [x] 1.1 Definir tipos de dominio en `src/internal/audit`: `Event` (metadata inmutable), interfaces `Store` y `Encryptor` (KMS Envelope)
- [x] 1.2 (test-first) WAL append-only: Append de N eventos → el archivo contiene los N registros serializados y legibles (HU-039 AC1) con t.TempDir
- [x] 1.3 Implementar `src/internal/wal` (append-only, serialización longitud-prefijada) que hace verde 1.2
- [x] 1.4 (test-first) Rotación por tamaño: al superar maxBytes archiva `wal-<ts>-NNN.log` y abre uno nuevo (HU-039 AC3)
- [x] 1.5 Implementar la rotación que hace verde 1.4
- [x] 1.6 (test-first) Recover(): lee activo + archivados y devuelve los eventos no flusheados para replay (HU-039 AC2); durabilidad AC4 cubierta como smoke (journal FS es NFR de OS, diferido a despliegue)
- [x] 1.7 Implementar Recover() que hace verde 1.6
- [x] 1.8 journey_smoke SS1: escribir eventos, rotar, y recuperarlos end-to-end desde disco temporal; suite verde. (AC5 overhead <1ms = NFR de load-test, diferido)

## 2. Sub-slice 2 — Sync Worker (HU-038)

- [x] 2.1 (test-first) Batching: encolar eventos → flush por 1000 o por 1s (lo que ocurra primero) vía Store, sin bloquear el productor (HU-038 AC1) con -race
- [x] 2.2 Implementar `src/internal/syncworker` (channel + batching por tamaño/tiempo, escribe al WAL antes del flush) que hace verde 2.1
- [x] 2.3 (test-first) KMS Envelope: cada evento se cifra vía Encryptor (mock DEK) antes de escribir al Store (HU-038 AC4)
- [x] 2.4 Implementar el cifrado en el pipeline del worker que hace verde 2.3
- [x] 2.5 (test-first) Backpressure: channel saturado → retry con jitter o drop de baja prioridad, sin bloquear ni fallar crítico (HU-038 AC2) con -race
- [x] 2.6 Implementar el manejo de backpressure que hace verde 2.5. AC3 (pérdida→WAL) se cubre por la escritura al WAL de 2.2
- [x] 2.7 journey_smoke SS2: worker real + WAL + Store/Encryptor mock; batch cifrado persiste; suite verde con -race. (AC5 throughput = NFR de load-test, diferido)

## 3. Sub-slice 3 — Graceful Shutdown (HU-040)

- [x] 3.1 (test-first) Drain + flush: Shutdown(ctx) espera in-flight (timeout) y flushea el buffer del Sync Worker a Store (o WAL si Store falla) antes de salir (HU-040 AC1/AC3) con -race
- [x] 3.2 Implementar `src/internal/shutdown` (drain con timeout + Flush del worker) que hace verde 3.1
- [x] 3.3 (test-first) Timeout en drain: request que excede el timeout → cierra y registra el evento (HU-040 AC2)
- [x] 3.4 Implementar el timeout de drain que hace verde 3.3
- [x] 3.5 (test-first) Boot recovery: al arrancar procesa el WAL residual y lo replica al Store antes de aceptar tráfico; sin deadlock (timeout DB < shutdown) (HU-040 AC4/AC5)
- [x] 3.6 Implementar Recover() de boot que hace verde 3.5
- [x] 3.7 journey_smoke SS3: ciclo shutdown (flush) → boot (recover del WAL) end-to-end; suite verde con -race

## 4. Sub-slice 4 — Cache Invalidator (Fase 2, alcance reducido) (HU-041)

- [x] 4.1 (test-first) MVP no-op: con flag OFF, Invalidate es no-op; ante miss aplica poll/retry (HU-041 AC4/AC2)
- [x] 4.2 Implementar `src/internal/cacheinval` (flag + no-op + fallback poll/retry) que hace verde 4.1. Worker de polling/webhook completo (AC1/AC3/AC5) diferido a Fase 2, documentado
- [x] 4.3 journey_smoke SS4: con flag OFF el invalidador no interfiere; el patrón poll/retry ante miss funciona; suite verde

## 5. Cierre de épica

- [x] 5.1 Coherencia triple AC↔specs↔tests (coherence-three-way) sin huecos
- [x] 5.2 Verificación adversarial de cableado (wiring-adversarial-verifier) → wiring_verified
- [x] 5.3 DoD reducido (dor-dod-gatekeeper) → dod
- [x] 5.4 PR + opsx:archive del change en el mismo PR
