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
| 1 | Petición que excede ventana | Un payload tiene 120k tokens y el modelo soporta 100k | El router intenta validar el contexto | La validación falla y el router descarta este modelo del score devolviendo 400 Bad Request si no hay fallback |
