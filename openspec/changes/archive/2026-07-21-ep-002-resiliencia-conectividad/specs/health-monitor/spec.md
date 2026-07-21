## ADDED Requirements

### Requirement: Sondeo periódico con retiro y reactivación
El Health Monitor SHALL sondear periódicamente cada proveedor, retirarlo al fallar los checks, reactivarlo al recuperarse, y exponer el estado de salud vivo como `HealthSource` del router. (Traza: HU-005)

#### Scenario: Proveedor sano se mantiene
- **WHEN** un proveedor responde OK a los health checks durante el ciclo
- **THEN** permanece marcado sano y en la rotación

#### Scenario: Caída detectada y retiro
- **WHEN** un proveedor empieza a devolver errores en los health checks
- **THEN** se marca no-sano y deja de recibir peticiones nuevas hasta recuperarse

#### Scenario: Reactivación automática
- **WHEN** un proveedor no-sano vuelve a responder OK en el siguiente check
- **THEN** se marca sano y vuelve a la rotación automáticamente

#### Scenario: Todos no-sanos
- **WHEN** todos los proveedores de una capacidad fallan los checks y llega una petición de esa capacidad
- **THEN** la Gateway responde 503 y el estado refleja la capacidad como no disponible

#### Scenario: Proveedor intermitente con histéresis
- **WHEN** un proveedor alterna OK/fallo entre checks durante varios ciclos
- **THEN** se aplica histéresis (N fallos para retirar, M éxitos para reactivar) para evitar oscilación
