# learning-engine Specification

## Purpose
TBD - created by archiving change ep-007-observabilidad. Update Purpose after archive.
## Requirements
### Requirement: Motor de Auto-Ajuste (Learning Engine)
El sistema SHALL ajustar dinámicamente los pesos de enrutamiento basado en histórico reciente.

#### Scenario: Ajuste con evidencia suficiente
- **WHEN** un modelo demuestra consistentemente latencia menor que el resto y hay >100 peticiones de muestra
- **THEN** ajusta sus pesos a favor para ser priorizado por el router

#### Scenario: Datos insuficientes para ajustar
- **WHEN** la muestra es muy pequeña
- **THEN** conserva los pesos originales por defecto evitando sobreajuste

#### Scenario: Guardrails de diversificación
- **WHEN** el motor intenta dar todo el tráfico a un único modelo
- **THEN** topa la asignación para preservar diversidad de proveedores

#### Scenario: Reversión (Rollback) por degradación
- **WHEN** un ajuste degrada el success rate observado
- **THEN** el sistema revierte automáticamente a los pesos anteriores

