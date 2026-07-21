---
id: HU-029
titulo: Adapter para AIHubMix
epica: EP-008
prioridad: Must
complejidad: S
estado: lista
---

# Adapter para AIHubMix

Como **desarrollador de la plataforma**, quiero **un adapter que traduzca el formato interno al de AIHubMix**, para **poder usarlo como proveedor gratuito por defecto en el producto**.

Contexto: Requisito explícito del PRD. AIHubMix expone una API casi 100% compatible con OpenAI.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — chat básico | Dado que un payload de chat normal | Cuando el router selecciona AIHubMix | Entonces el adapter reenvía a AIHubMix y obtiene respuesta |
| 2 | Error — rate limit | Dado que aIHubMix devuelve 429 | Cuando el adapter procesa la respuesta | Entonces se emite error interno para disparar Failover |
| 3 | Sad path — Upstream 500/503 | Dado que el servicio upstream AIHubMix retorna 500 o 503 | Cuando el adapter recibe el error | Entonces lo traduce a un error estándar para que el Gateway inicie el failover |
| 4 | Edge — Parámetros no soportados | Dado que el request incluye un parámetro exótico de OpenAI no soportado por AIHubMix | Cuando el adapter lo procesa | Entonces ignora el parámetro de forma segura y procesa el resto |

## Checklist INVEST

- [x] Independent — Adaptador aislado.
- [x] Negotiable — Mapeo de errores específicos del proveedor.
- [x] Valuable — Respaldo de contingencia y modelos alternativos a bajo costo.
- [x] Estimable — Mapear JSON.
- [x] Small — Un solo proveedor.
- [x] Testable — Suite de contrato.

## Notas técnicas
- Asegurar alineación con NFRs de latencia y uso de caché si aplica.
