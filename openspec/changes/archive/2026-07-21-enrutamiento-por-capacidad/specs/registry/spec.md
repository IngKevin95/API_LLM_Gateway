## ADDED Requirements

### Requirement: Carga declarativa del catálogo desde YAML
El Registry SHALL cargar `config.yaml` (providers, models con atributos de score, y routing por capacidad) a memoria RAM durante el arranque y exponer los modelos habilitados por capacidad. (Traza: HU-001 AC1)

#### Scenario: Carga válida
- **WHEN** la Gateway inicia con un `config.yaml` que declara providers, models (quality/coding/reasoning/speed/vision/cost/latency) y routing por capacidad
- **THEN** el catálogo queda en memoria, los modelos habilitados quedan expuestos por capacidad y el conteo final se imprime en stdout

### Requirement: Fail-fast ante YAML inválido
El Registry SHALL abortar el arranque cuando el `config.yaml` tenga sintaxis inválida o falte un campo obligatorio, sin dejar el sistema en estado parcial. (Traza: HU-001 AC2)

#### Scenario: YAML inválido o campo faltante
- **WHEN** la Gateway inicia con un `config.yaml` sintácticamente inválido o con un campo obligatorio ausente
- **THEN** el arranque falla con un error que nombra archivo, línea y campo problemático, y el sistema no arranca en estado parcial

### Requirement: Prohibición de secretos literales
El Registry SHALL rechazar valores de secreto literales (p. ej. `api_key`) y exigir referencia por `${VAR}` resuelta desde el entorno, sin imprimir nunca el valor. (Traza: HU-001 AC3)

#### Scenario: API key literal embebida
- **WHEN** el `config.yaml` declara una API key literal en vez de `${VAR}`
- **THEN** el Registry rechaza la carga (o la clave), exige la referencia a variable de entorno y registra la violación sin imprimir el valor

### Requirement: Capacidad sin modelos no aborta la carga
El Registry SHALL marcar como no disponible cualquier capacidad de routing sin modelos habilitados, emitiendo un WARN, sin abortar el resto de la carga. (Traza: HU-001 AC4)

#### Scenario: Capacidad vacía
- **WHEN** una capacidad en `routing` no lista ningún modelo habilitado
- **THEN** el Registry marca esa capacidad como no disponible, escribe un log WARN y continúa cargando el resto del catálogo

### Requirement: Exposición de parámetros de red físicos
El Registry SHALL exponer `max_in_flight` y `stream_idle_timeout` de cada provider al resto del Gateway (Failover Engine y Adapters). (Traza: HU-001 AC5)

#### Scenario: Parámetros de red presentes
- **WHEN** el `config.yaml` declara `max_in_flight` y `stream_idle_timeout`
- **THEN** el Registry los expone a los consumidores (Failover Engine y Adapters)
