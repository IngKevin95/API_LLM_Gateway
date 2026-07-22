## ADDED Requirements

### Requirement: Optimización opt-in del prompt con bypass pasivo
El sistema SHALL, cuando el guardián está habilitado (opt-in), envolver el último mensaje `user` en un template de optimización preservando el texto original, sin alterar tool calling ni el streaming, y hacer bypass pasivo ante prompt inválido o si excede el presupuesto de tiempo. (Traza: HU-027)

#### Scenario: Prompt optimizado
- **WHEN** el guardián está habilitado y se recibe un prompt
- **THEN** el último mensaje con rol `user` queda envuelto en el template de optimización (instrucciones de sistema añadidas) y el texto original se preserva íntegro dentro del wrapper

#### Scenario: Prompt inválido hace bypass
- **WHEN** el guardián intenta reestructurar un prompt malformado
- **THEN** omite la optimización sin excepción y deja pasar el prompt original

#### Scenario: Tool calling intacto
- **WHEN** se aplica la optimización a un prompt con llamadas a funciones definidas
- **THEN** la sintaxis de tool calling se mantiene intacta

#### Scenario: Overhead excesivo
- **WHEN** la optimización excede 100ms
- **THEN** aborta la optimización y envía el prompt original

#### Scenario: Petición en streaming
- **WHEN** el payload original especifica `stream: true`
- **THEN** el guardián no altera la respuesta token a token (opera sobre el request, no la respuesta)
