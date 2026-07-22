# Tareas de Implementación: EP-003

## Fase 1: Quota Manager (HU-006)
- [ ] Implementar la interfaz `Manager` en `internal/quota/manager.go` con estado en memoria (`sync.RWMutex`).
- [ ] Implementar la lógica para resetear contadores en cambio de ventana (`Window`).
- [ ] Implementar `Reserve()` pre-descontando la estimación y retornando false en sobregiro.
- [ ] Implementar `Commit()` para fijar el gasto real.
- [ ] Escribir tests unitarios para concurrencia, sobregiros, y rotación de ventanas.

## Fase 2: Cost Tracker (HU-007)
- [ ] Implementar el Tracker que reciba eventos pos-ejecución y calcule costos (multiplicando tokens reales por las tarifas de `Registry`).
- [ ] Adaptar la captura de costos para escenarios donde se completó (HTTP 200) y cuando hay aborto de stream (contabilizar consumo parcial).
- [ ] Escribir tests unitarios probando tarifas con valores válidos, cero, o no-declarados (desconocidos).

## Integración y Refactor
- [ ] Conectar `QuotaManager` y `CostTracker` como middlewares lógicos alrededor del `Router` y el flujo de los adaptadores de LLM.
