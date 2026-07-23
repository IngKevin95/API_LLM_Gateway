# context-validation Specification

## Purpose
TBD - created by archiving change enrutamiento-por-capacidad. Update Purpose after archive.
## Requirements
### Requirement: Validación de ventana de contexto con buffer
La validación de contexto SHALL estimar los tokens del request y descartar como candidato cualquier modelo cuya ventana sea excedida por la estimación aumentada en un buffer de seguridad del 20%, aplicada como filtro pre-score. (Traza: HU-035 AC1/AC2/AC3, HU-002a AC3)

#### Scenario: Payload dentro de ventana (happy)
- **WHEN** un payload estima 80k tokens (crudo + buffer 20% = 96k) y el modelo soporta 100k
- **THEN** la validación pasa y el modelo permanece candidato, avanzando al cálculo de score

#### Scenario: Payload excede ventana (error)
- **WHEN** un payload estima 120k tokens y el modelo soporta 100k
- **THEN** la validación falla, el router descarta el modelo del score y, si no hay fallback en la cadena, devuelve 400 Bad Request

#### Scenario: Buffer 20% empuja fuera de ventana (edge)
- **WHEN** un payload estima 85k tokens crudos para un modelo de 100k, pero el buffer del 20% eleva la estimación a 102k
- **THEN** la validación falla por buffer y el modelo se descarta aunque el conteo crudo cabría

### Requirement: Estimador de tokens intercambiable
La validación de contexto SHALL exponer la estimación de tokens tras una interfaz `Tokenizer`, con una implementación heurística por defecto, permitiendo sustituirla (p. ej. `tiktoken-go`) sin cambiar el router. (Traza: HU-035 INVEST Negotiable, notas técnicas HU-002a)

#### Scenario: Implementación por defecto
- **WHEN** el router valida el contexto sin un tokenizador específico configurado
- **THEN** usa la implementación heurística por defecto y produce una estimación con el buffer aplicado

