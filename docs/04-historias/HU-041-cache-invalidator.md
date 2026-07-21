---
id: HU-041
titulo: Cache Invalidator (polling/webhook para refresco dinámico de cuotas/auth)
epica: EP-009
prioridad: Should
complejidad: M
estado: lista
---

# Cache Invalidator (polling/webhook para refresco dinámico de cuotas/auth)

Como **administrador de la plataforma**, quiero **invalidar y refrescar cuotas y keys en memoria cuando la base de datos cambia externamente**, para **soportar modificaciones dinámicas sin requerir reinicios de nodos**.

Contexto: Fase 2. DBA ajusta cuotas en PostgreSQL → Cache Invalidator detecta cambio → invalida en RAM de todos los nodos. Implementado como worker background con polling (30s default) o webhook.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — polling por cuota | Dado que DBA actualiza cuota de tenant X en DB | Cuando Cache Invalidator compara timestamp de última sincronización | Entonces detecta cambio, invalida cache en RAM, y encola hidratación asíncrona |
| 2 | Error — webhook perdido | Dado que webhook timeout o falla de red | Cuando Cache Invalidator reintenta | Entonces fallback a polling en la siguiente iteración (no quedan cuotas stale permanentemente) |
| 3 | Edge — race condition | Dado que Quota Manager en RAM y DBA escriben simultáneamente | Cuando Cache Invalidator invalida | Entonces la solicitud next encola re-hidratación (fail-fast + retry pattern) |
| 4 | Fase 2 — no en MVP | Dado que estamos en Fase 1 | Cuando Feature flag = OFF | Entonces Cache Invalidator es no-op; solo poll/retry en miss aplica |
| 5 | Performance — sincronización | Dado que 1000 tenants con cuotas actualizadas | Cuando Cache Invalidator procesa | Entonces latencia de invalidación < 5s (polling 30s, webhook < 1s si disponible) |

## Checklist INVEST

- [x] Independent — worker background, no bloquea camino crítico
- [x] Negotiable — poll interval, webhook URL, retry policy configurables
- [x] Valuable — habilita admin sin reinicios (Fase 2+)
- [x] Estimable — worker pattern + DB polling estándar
- [x] Small — integración con Quota Manager
- [x] Testable — mock DB changes, verificar cache invalidada

## Notas técnicas

Worker 1 — polling: `SELECT updated_at FROM Quota WHERE updated_at > last_sync`; 2 — webhook: POST /gateway/cache-invalidate (secured vía API key interna). Feature flag: `cache_invalidator.enabled` en YAML. Fase 2+, no en MVP.
