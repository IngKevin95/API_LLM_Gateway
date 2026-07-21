# model-router Specification

## Purpose
TBD - created by archiving change enrutamiento-por-capacidad. Update Purpose after archive.
## Requirements
### Requirement: Resolución automática por score
El Model Router SHALL resolver una capacidad solicitada al modelo óptimo por un score determinista de 6 variables (calidad, velocidad, disponibilidad, cuota restante, costo, latencia), retornando la cadena de fallback ordenada por score descendente. (Traza: HU-002a AC1)

#### Scenario: Elección óptima entre candidatos
- **WHEN** el cliente pide la capacidad "chat" y hay 3 modelos disponibles con distintos pesos y cuota > 0
- **THEN** el Router retorna la lista de fallbacks con el modelo de mayor score total en primer lugar

### Requirement: Filtro por estado antes del ranking
El Model Router SHALL excluir de la resolución todo modelo deshabilitado en YAML, unhealthy o sin cuota, aunque su score teórico sea el mayor. (Traza: HU-002a AC2)

#### Scenario: Modelo top deshabilitado o unhealthy
- **WHEN** el modelo de mayor score teórico está deshabilitado, unhealthy o sin cuota
- **THEN** el Router lo excluye y el de segundo mejor score toma el primer lugar

### Requirement: Capacidad desconocida rechazada
El Model Router SHALL rechazar con 400 Bad Request cualquier capacidad no definida. (Traza: HU-002b AC1)

#### Scenario: Capacidad no soportada
- **WHEN** el cliente solicita una capacidad no definida (p. ej. "quantum_computing")
- **THEN** el Router retorna 400 Bad Request indicando capacidad no soportada

### Requirement: Sin candidatos aptos retorna 503
El Model Router SHALL retornar 503 Service Unavailable cuando ningún modelo de la capacidad esté apto, sin iniciar failover inútil. (Traza: HU-002b AC2)

#### Scenario: Todos los modelos caídos o sin cuota
- **WHEN** el cliente pide "vision" pero todos los modelos de esa capacidad tienen cuota agotada o health=false
- **THEN** el Router retorna 503 Service Unavailable y no inicia failover

### Requirement: Desempate determinista
El Model Router SHALL desempatar scores idénticos de forma consistente: primero por menor costo, y en su defecto por orden alfabético del ID del modelo. (Traza: HU-002b AC3)

#### Scenario: Empate exacto de score
- **WHEN** dos modelos aptos obtienen exactamente el mismo score
- **THEN** el Router ordena priorizando el de menor costo, o por orden alfabético del ID si el costo también empata

### Requirement: Modo explícito usa el modelo pedido
El Model Router SHALL usar exactamente el `model` explícito indicado cuando existe y está sano, sin aplicar scoring automático. (Traza: HU-003 AC1)

#### Scenario: Modelo explícito disponible
- **WHEN** llega una petición con un `model` explícito que existe y está sano en el Registry
- **THEN** el Router usa exactamente ese modelo sin aplicar scoring automático

### Requirement: Modelo explícito inexistente rechazado
El Model Router SHALL responder 404 cuando el `model` explícito no exista en el Registry, listando modelos válidos, sin caer a otro modelo silenciosamente. (Traza: HU-003 AC2)

#### Scenario: Modelo inexistente
- **WHEN** llega una petición con un `model` que no existe en el Registry
- **THEN** el Router responde 404 "modelo no encontrado" listando modelos válidos, sin sustituir silenciosamente

### Requirement: Política de fallback en modo explícito
El Model Router SHALL aplicar la cadena de fallback de la capacidad cuando el `model` explícito esté no-sano y la política permita fallback, anotando la sustitución; y SHALL responder 503 sin sustituir cuando la política lo prohíba. (Traza: HU-003 AC3, AC4)

#### Scenario: Modelo caído con fallback permitido
- **WHEN** el `model` explícito existe pero está no-sano y la política permite fallback
- **THEN** el Router aplica la cadena de fallback de esa capacidad y anota en respuesta/log que sustituyó el modelo pedido

#### Scenario: Modelo caído sin fallback permitido
- **WHEN** el `model` explícito existe pero está no-sano y la política prohíbe fallback
- **THEN** el Router responde 503 "modelo solicitado no disponible" sin sustituir

