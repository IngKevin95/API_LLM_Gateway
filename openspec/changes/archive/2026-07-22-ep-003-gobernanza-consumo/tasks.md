# Tareas de Implementación: EP-003

## Fase 1: Quota Manager (HU-006)
- [x] Implementar la interfaz `Manager` en `internal/quota/manager.go` con estado en memoria (`sync.RWMutex`).
- [x] Implementar la lógica para resetear contadores en cambio de ventana (`Window`).
- [x] Implementar `Reserve()` pre-descontando la estimación y retornando false en sobregiro.
- [x] Implementar `Commit()` para fijar el gasto real.
- [x] Escribir tests unitarios para concurrencia, sobregiros, y rotación de ventanas.

## Fase 2: Cost Tracker (HU-007)
- [x] Implementar el Tracker que reciba eventos pos-ejecución y calcule costos (multiplicando tokens reales por las tarifas de `Registry`).
- [x] Adaptar la captura de costos para escenarios donde se completó (HTTP 200) y cuando hay aborto de stream (contabilizar consumo parcial).
- [x] Escribir tests unitarios probando tarifas con valores válidos, cero, o no-declarados (desconocidos).

## Integración y Refactor
- [x] Conectar `QuotaManager` y `CostTracker` como middlewares lógicos alrededor del `Router` y el flujo de los adaptadores de LLM.
