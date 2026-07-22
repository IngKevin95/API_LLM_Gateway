## ADDED Requirements

### Requirement: Drenado ante SIGTERM con flush garantizado
El Graceful Shutdown SHALL, ante SIGTERM, dejar de aceptar conexiones, drenar las requests en vuelo con timeout, y garantizar el flush del buffer del Sync Worker a DB (o al WAL si la DB no responde) antes de salir. (Traza: HU-040 AC1/AC2/AC3)

#### Scenario: Drenado feliz
- **WHEN** el Gateway recibe SIGTERM y cierra nuevas conexiones
- **THEN** espera (timeout <30s) a que se drenen las requests en vuelo y se flushee el buffer

#### Scenario: Timeout en el drain
- **WHEN** una request excede el timeout de drain configurado (p. ej. 25s)
- **THEN** cierra la conexión y registra el evento de timeout

#### Scenario: Flush garantizado del buffer
- **WHEN** hay eventos en el buffer del Sync Worker y llega SIGTERM
- **THEN** el Sync Worker flushea todo a DB (o al WAL si la DB no responde) antes de salir

### Requirement: Secuencia de boot con recuperación
El boot SHALL procesar el WAL residual e hidratar cachés antes de aceptar tráfico, sin deadlock entre los workers que compiten por la DB. (Traza: HU-040 AC4/AC5)

#### Scenario: Recuperación en el boot
- **WHEN** el Gateway bootea tras un graceful shutdown
- **THEN** el Recovery Worker procesa el WAL residual (si quedó) e hidrata las cachés de cuota/auth antes de aceptar tráfico

#### Scenario: Prevención de deadlock
- **WHEN** el Health Monitor y el Sync Worker compiten por la conexión a DB durante el flush de shutdown
- **THEN** no hay deadlock, porque el timeout de DB es menor que el timeout de shutdown
