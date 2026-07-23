---
id: HU-016
titulo: Configurar Free Claude Code contra la Gateway
epica: EP-005
prioridad: Must
complejidad: S
estado: lista
---

# Configurar Free Claude Code contra la Gateway

Como **desarrollador**, quiero **apuntar Free Claude Code a la Gateway configurando `ANTHROPIC_BASE_URL`**, para **usar la experiencia de Claude Code enrutada por la Gateway**.

Contexto: cierra el uso de Free Claude Code sobre el endpoint Anthropic-compat (HU-013). Actividad 6.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — apuntado correcto | Dado que free Claude Code ya lanzado con `ANTHROPIC_BASE_URL` apuntando a la Gateway y credencial válida | Cuando el usuario envía un prompt | Entonces la petición pasa por la Gateway, se enruta y la respuesta se muestra en el cliente en el mismo formato que al usar Anthropic directamente |
| 2 | Error — base URL mal configurada | Dado que `ANTHROPIC_BASE_URL` con una URL inválida | Cuando el usuario lanza el cliente | Entonces el cliente falla con un error claro y la doc indica cómo corregir la variable |
| 3 | Edge — credencial de Gateway ausente | Dado que base URL correcta pero sin API key de la Gateway | Cuando el usuario envía un prompt | Entonces la Gateway responde 401 y la guía documenta el paso de credencial |
| 4 | Edge — funcionalidad no soportada | Dado que el cliente usa una función Anthropic no cubierta por HU-013 | Cuando se dispara esa función | Entonces la Gateway responde un error claro de "no soportado" en vez de fallar de forma opaca |

## Checklist INVEST

- [x] Independent — depende de HU-013 (endpoint) entregable
- [x] Negotiable — mecanismo de config documentado
- [x] Valuable — habilita Free Claude Code end-to-end
- [x] Estimable — configuración + documentación
- [x] Small — 1-2 días
- [x] Testable — arranque del cliente contra Gateway

## Notas técnicas

Entregable incluye guía de configuración reproducible. Depende de la cobertura de contrato de HU-013.
