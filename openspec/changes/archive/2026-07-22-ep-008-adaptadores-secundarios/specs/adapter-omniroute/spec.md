## ADDED Requirements

### Requirement: Adapter OmniRoute (HU-036)
El sistema SHALL tener un adapter configurable para OmniRoute que implemente `adapter.Adapter`, delegando al endpoint externo configurado en el registry YAML.

#### Scenario: Petición exitosa vía OmniRoute
- **GIVEN** un provider OmniRoute configurado en el registry con endpoint base y credenciales
- **WHEN** el Router selecciona OmniRoute para la capacidad solicitada
- **THEN** el adapter reenvía el request transformado al endpoint externo y retorna respuesta normalizada

#### Scenario: OmniRoute como fallback en cadena
- **GIVEN** que el proveedor principal ha fallado y la política de failover incluye OmniRoute
- **WHEN** el Failover Engine activa el siguiente proveedor
- **THEN** el adapter OmniRoute completa el request exitosamente

#### Scenario: Error de capacidad/cuota (429/500)
- **GIVEN** que OmniRoute responde con 429 o 500
- **WHEN** el adapter recibe el error
- **THEN** emite `*adapter.ProviderError{Retryable: true}` para que el failover continúe al siguiente proveedor
