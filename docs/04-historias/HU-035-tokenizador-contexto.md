---
id: HU-035
titulo: Tokenizador de Contexto (Context Window)
epica: EP-001
prioridad: Must
complejidad: S
estado: lista
---

# HU-035: Tokenizador de Contexto y validación de buffer

## INVEST
- [x] Independent: lógica autocontenida que implementa una interfaz `ITokenizer`.
- [x] Negotiable: el algoritmo de conteo puede variar.
- [x] Valuable: evita envíos fallidos al proveedor si el texto excede el límite máximo del modelo.
- [x] Estimable: conteo de palabras heurístico o integraciones con librerías `tiktoken`.
- [x] Small: lógica de string manipulation / parsing.
- [x] Testable: contadores verificables mediante tests unitarios.

## Criterios de Aceptación (BDD)
| ID | Escenario | Dado (Given) | Cuando (When) | Entonces (Then) |
|---|---|---|---|---|
| 1 | Petición dentro de ventana (happy) | Un payload estima 80k tokens (crudo + buffer 20% = 96k) y el modelo soporta 100k | El router valida el contexto | La validación pasa y el modelo permanece candidato, avanzando al cálculo de score |
| 2 | Petición que excede ventana (error) | Un payload estima 120k tokens y el modelo soporta 100k | El router intenta validar el contexto | La validación falla, el router descarta este modelo del score y, si no hay fallback en la cadena, devuelve 400 Bad Request |
| 3 | Buffer 20% empuja fuera de ventana (edge) | Un payload estima 85k tokens crudos y el modelo soporta 100k, pero el buffer de seguridad del 20% eleva la estimación a 102k | El router valida el contexto aplicando el buffer | La validación falla por buffer: el modelo se descarta aunque el conteo crudo cabría, evitando un envío al borde del límite |
