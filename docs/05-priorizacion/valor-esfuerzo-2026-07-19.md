# Sesión de priorización — API LLM Gateway

**Framework**: Valor / Esfuerzo · **Fecha inicial**: 2026-07-19 · **Re-priorización**: 2026-07-20 (tras auditoría conjunta + splits INVEST) · **Participantes**: Kevin Beltrán (owner/PO)

Priorización de las **41 historias** del backlog (incluye los splits de HU-004→a/b/c, HU-020→a/b/c,
HU-021→a/b y HU-022→+022b, y las historias antes ausentes del backlog: HU-023..HU-033, habiendo retirado las del CLI). Objetivo:
maximizar valor temprano respetando dependencias duras (p.ej. el Router necesita el Registry). Valor
= cuánto avanza los objetivos del PRD y el beneficio visible; Esfuerzo = talla S/M/L de la historia.

Resolución de esfuerzo: el cuadrante usa S = bajo esfuerzo y M/L = alto esfuerzo (proxy binario).

## Regla de derivación cuadrante → prioridad (MoSCoW)

El framework del proyecto es Valor/Esfuerzo; la columna Prioridad del backlog usa vocabulario MoSCoW.
La traducción es explícita y auditable:

- **Must** = historia en el **camino crítico del producto**: sin ella no hay sistema funcional
  (core de enrutamiento, adapters + endpoints OpenAI-compat, AuthN, failover, auditoría, CLI
  + Free Claude Code). Puede ser quick win o big bet; lo que la hace Must es ser bloqueante.
- **Should** = **alto valor pero no bloquea la arquitectura base**: gobernanza (cuota/costo), AuthZ,
  adapters adicionales, tools de CLI, seguridad avanzada, guardián de prompts.
- **Could** = **valor diferido o complejidad alta**: observabilidad, dashboard, learning
  engine, cache semántica, MCP, mTLS, degradación local dedicada.

Verificación de salud MoSCoW (regla local §2): 24 Must / 44 = 52% del conteo (< 60%). OK.

## Resultado — cuadrantes Valor / Esfuerzo

- **Quick wins** (alto valor, bajo esfuerzo — hacer primero cuando la dependencia lo permita):
  HU-002a, HU-002b, HU-003, HU-008, HU-020a, HU-029, HU-020b, HU-020c, HU-021b, HU-026a, HU-016,
  HU-012b, HU-012c, HU-007, HU-022b, HU-031.
- **Big bets** (alto valor, alto esfuerzo — núcleo, planificar con cuidado):
  HU-001, HU-012a, HU-021a, HU-013, HU-004a, HU-004b, HU-004c, HU-010, HU-022, HU-009,
  HU-011, HU-025a, HU-026b, HU-028, HU-030, HU-027.
- **Fill-ins** (bajo valor relativo, bajo esfuerzo — cuando sobre capacidad): HU-005 (parcial).
- **Postergables en flujo** (valor diferido / fases posteriores): HU-006, HU-017, HU-023, HU-018,
  HU-019, HU-032, HU-033, HU-025b.

## Orden final (respeta dependencias duras)

El orden de filas de `backlog.md` **es** la priorización vigente (fuente de verdad); esta sesión lo
justifica. La secuencia agrupa por tier MoSCoW y, dentro de cada tier, por dependencia de construcción.
Resumen por bloque:

1. **Bloque 1 — Núcleo (Must, órdenes 1-22)**: Registry → Router (002a/002b) → modelo explícito →
   AuthN → Adapter OpenAI (chat/streaming/embeddings) + AIHubMix → endpoints OpenAI-compat → Adapter
   Anthropic (chat/streaming) + endpoint → Failover (004a/b/c) → Auditoría + redacción síncrona →
   Rate limiting → Free Claude Code.
2. **Bloque 2 — Gobernanza y Seguridad (Should, órdenes 23-34)**: Health → AuthZ → secretos →
   cuota → costo → concurrencia vision → OAuth2 → kill-switch async → envelope → adapters Google/OpenRouter →
   guardián de prompts.
3. **Bloque 3 — Observabilidad e Inteligencia (Could, órdenes 35-41)**: métricas → dashboard →
   histórico → learning engine → cache semántica → mTLS → MCP.

> Nota: la dominancia de big bets al inicio (núcleo Registry/Router/API/Failover) es estructural del
> dominio, no una mala selección: son must-have arquitectónicos inevitables para un Gateway.

## Decisiones

1. **HU-001 y HU-008 primero absoluto**: HU-001 es prerequisito de casi todo; HU-008 (auth) es quick
   win y requisito de seguridad desde el día 1.
2. **Un solo camino de adapter+endpoint completo primero**: OpenAI (HU-020a/b/c) + AIHubMix como
   proveedor gratuito por defecto, más Anthropic para Free Claude Code. Google/OpenRouter se difieren
   a Should (redundancia, no bloqueante del núcleo).
3. **Failover dividido**: HU-004a (básico) es Must; 004b/004c (circuit breaker, timeouts
   dinámicos) siguen inmediatamente por dependencia pero también son Must por resiliencia ≥99%.
4. **HU-007 (costo) va después de seguridad**: Fill-in de bajo valor; no debe secuenciarse antes de
   historias de alto valor como HU-009 (AuthZ) o HU-011 (secretos).
6. **Learning Engine (HU-019) al final**: alto esfuerzo y su valor depende de acumular histórico
   (HU-018).

## Disidencias / preguntas abiertas

- ¿HU-016 (config Free Claude Code) debería fusionarse como AC de HU-013? La auditoría INVEST lo
  sugirió; se mantiene separada para entregar y documentar el flujo end-to-end de forma verificable.
- Confirmar en sprint planning que los splits de adapters (020a/b/c, 021a/b) no generan overhead de
  coordinación que aconseje reagruparlos; a la fecha se mantienen separados por INVEST-Small.

## Próxima sesión

**2026-08-17** (fecha concreta): revalidar valor/esfuerzo con velocidad real medida al cierre de la
Bloque 1 de construcción.
