# Runner fuera-de-chat: `integration-check`

Este runner cierra `journey_smoke` por **ejecución real**, no por lectura de código — el mismo
patrón que un gate de CI fuera-de-chat (piensa en algo como `tools/loop/gateway-check.sh`), llevado
también al inner-loop manual, no solo a corridas autónomas. No introduce un gate nuevo: simplemente
alimenta el `journey_smoke` que ya existe en el slice. El gate `integration` con dependencias reales
del proyecto sigue siendo terreno del Release Gate (outer loop).

## Por qué correrlo fuera de chat, en sesión virgen

- Cablear de punta a punta hay que **ejecutarlo**, no razonarlo desde el diff. Un script lo hace
  reproducible y no quema contexto de conversación.
- Quien escribió el código es mal juez de su propio cableado — por eso este runner corre en una
  sesión/contexto **virgen** (o como paso de CI): corre, lee el reporte, arregla, repite hasta verde.

## Puesta en marcha

1. Copia `templates/integration-check.sh.template` al repo consumidor, p. ej. a
   `tools/loop/integration-check.sh`, y dale permisos de ejecución (`chmod +x`).
2. Ajusta los comandos marcados `# ADAPTA` al stack real (build, suite de tests, e2e). El template
   viene con defaults de Node como punto de partida — el arnés en sí es agnóstico al stack.
3. Ejecútalo en la fase `smoke`. Código de salida `0` = verde (se puede marcar
   `journey_smoke: true`); cualquier otro código = rojo. El reporte queda en
   `.claude/state/integration-report.txt`.
4. Mientras la suite no cierre en verde y el journey no camine de punta a punta, `journey_smoke`
   permanece en `false`.

## Cómo se conecta con el cableado fino

Cada paso verde de este runner es la evidencia que permite mover ítems de `wiring_checklist[]` de
`failing` a `passing` (protocolo completo en `state-protocol.md`). El
agente wiring-adversarial-verifier, en la fase `dod`, revisa después que esa evidencia sea genuina y no
un `passing` prematuro.
