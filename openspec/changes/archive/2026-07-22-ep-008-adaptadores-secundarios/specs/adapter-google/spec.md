## ADDED Requirements

### Requirement: Adapter Google Gemini (HU-030)
El sistema SHALL tener un adapter que implemente `adapter.Adapter` para la API nativa de Google Gemini, traduciendo el formato interno al formato `contents/parts` de Gemini.

#### Scenario: Chat con visión (imagen base64)
- **GIVEN** un request con un mensaje que contiene imagen en base64
- **WHEN** el adapter lo traduce para Gemini
- **THEN** coloca la imagen en `parts[0].inlineData.{mimeType, data}` y retorna respuesta de texto normalizada

#### Scenario: System prompt extraído a systemInstruction
- **GIVEN** un request con un mensaje `role=system` en el array de messages
- **WHEN** el adapter construye el payload para Gemini
- **THEN** extrae el contenido y lo coloca en `systemInstruction.parts[0].text`, sin incluirlo en `contents`

#### Scenario: Payload demasiado grande (400 pre-vuelo)
- **GIVEN** que la imagen supera el límite de Gemini
- **WHEN** Gemini retorna HTTP 400
- **THEN** el adapter emite `*adapter.ProviderError{Retryable: false}` para abortar sin iniciar failover

#### Scenario: Cuota concurrente agotada (429)
- **GIVEN** que la cuota concurrente de Google está saturada
- **WHEN** Gemini retorna HTTP 429
- **THEN** el adapter emite `*adapter.ProviderError{Retryable: true}` para disparar failover
