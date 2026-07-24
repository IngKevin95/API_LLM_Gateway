## ADDED Requirements

### Requirement: Adapter genérico data-driven
El Gateway SHALL soportar un adapter genérico configurado íntegramente por un `ProviderSpec`
declarativo (`baseURL`, `authHeader`, `format` [`openai`|`claude`], `headers` extra opcionales,
`timeoutMs` opcional), implementando `Chat/Stream/Embed` sin requerir código Go nuevo por
proveedor. (Traza: HU-EVO-001)

#### Scenario: Adapter OpenAI-compatible (Groq)
- **WHEN** se instancia el adapter genérico con `ProviderSpec{baseURL: "https://api.groq.com/openai/v1", authHeader: "Authorization", format: "openai"}`
- **THEN** el adapter reenvía requests a Groq con los headers correctos y devuelve la respuesta normalizada en formato OpenAI-compatible

#### Scenario: Adapter Claude-compatible
- **WHEN** se instancia el adapter genérico con `ProviderSpec{baseURL: "https://api.anthropic.com", authHeader: "x-api-key", format: "claude"}`
- **THEN** el adapter traduce el request de formato OpenAI al formato Claude y devuelve la respuesta normalizada

#### Scenario: Spec inválido rechazado
- **WHEN** se intenta instanciar el adapter genérico con un `ProviderSpec` de `baseURL` vacía o `format` desconocido
- **THEN** retorna error `ErrInvalidProviderSpec` sin realizar ninguna request

#### Scenario: Headers extra por proveedor
- **WHEN** el spec declara `headers: {"X-Custom": "value"}`
- **THEN** el adapter inyecta esos headers extra en cada request sin sobrescribir el header de autenticación

#### Scenario: Timeout por adapter distinto del global
- **WHEN** el spec declara `timeoutMs: 5000` distinto del timeout global del Gateway
- **THEN** el adapter respeta el timeout del spec, no el timeout global, al ejecutar la request

### Requirement: Conformance test extendido por ProviderSpec
`conformance_test.go` SHALL iterar automáticamente todos los `ProviderSpec` registrados y validar
que cada uno implementa correctamente `Chat/Stream/Embed` contra el contrato normalizado
(`Content/Model/Usage`), en paralelo y con timeout individual por proveedor. (Traza: HU-EVO-003)

#### Scenario: Itera todos los providers
- **WHEN** corre `go test ./...`
- **THEN** `conformance_test.go` itera todos los `ProviderSpec` registrados, invoca `Chat/Stream/Embed` en cada uno y reporta el resultado por proveedor

#### Scenario: Respuesta normalizada
- **WHEN** un provider devuelve respuesta en su formato nativo (p. ej. Groq OpenAI-compatible)
- **THEN** el adapter la normaliza al schema interno con `Content`, `Model` y `Usage` poblados, sin error

#### Scenario: Spec sin modelo default
- **WHEN** un `ProviderSpec` no tiene ningún modelo disponible para seleccionar
- **THEN** el conformance test falla ese caso con `ErrNoModelAvailable` y reporta el detalle sin abortar la suite completa

#### Scenario: Timeout en test individual
- **WHEN** un provider tarda más de 10s y el test define timeout de 5s
- **THEN** el conformance test aborta ese caso por timeout y reporta el resultado sin bloquear el resto de la suite

#### Scenario: Ejecución en paralelo sin race conditions
- **WHEN** el conformance test corre validando 5 providers en paralelo (`t.Parallel()`)
- **THEN** todos completan sin condiciones de carrera ni conflictos de recursos compartidos
