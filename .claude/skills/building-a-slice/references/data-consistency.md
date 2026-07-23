# Consistencia de datos: frontera externa + capa de decisión

`data-consistency-checker` aplica esta guía cuando un slice toca datos. Hay dos dominios distintos a
vigilar: el **JSON de esquema fijo** que entrega el servicio externo/IA en la frontera declarada, y
el **cálculo determinista** de la capa de decisión del dominio.

Nombres de campos, decisiones de alto impacto, umbrales y casos borde concretos viven en el bloque
de dominio del CLAUDE.md del consumidor y en su PRD (ruta en `stack-allowlist.json#source`). Esta
guía documenta **patrones** de invariante, agnósticos al dominio — no valores concretos. Para ver una
instanciación real, `docs/examples/reference/data-consistency.example.md`.

## Invariantes en la capa de decisión (determinista, sin IA)

- **Determinismo puro**: misma entrada → mismo resultado y misma clasificación, corrido N veces. Sin
  `Math.random`, sin reloj de sistema, sin llamada a IA dentro del cálculo.
- **Rango acotado + constantes versionadas**: el resultado cae dentro de un rango conocido y su
  clasificación pertenece al conjunto cerrado de decisiones de alto impacto del dominio, gobernado
  por umbrales versionados en configuración — nunca literales sueltos en el código.
- **Explicabilidad**: sumar los drivers/factores individuales reconstruye exactamente el total
  (cuadra al céntimo o al punto); todo resultado es trazable a lo que lo produjo.
- **Consistencia cruzada**: cuando se comparan campos entre fuentes o registros, cualquier
  discrepancia se atribuye a la fuente correcta — sin ambigüedad.

## Invariantes en la frontera (servicio externo/IA → JSON)

- Toda salida de un servicio externo/IA se trata como **entrada no confiable**: pasa por validación
  dura de esquema (p. ej. zod) antes de tocar la capa de decisión. Malformado = rechazo controlado,
  jamás propagación silenciosa.
- Tipos y unidades normalizados según las reglas del dominio (montos numéricos, fechas ISO 8601,
  identificadores consistentes).
- Ausencia de campo = `null`/opcional explícito. Nunca un valor fantasma o un default silencioso que
  disfrace un dato faltante.
- Idempotencia: reprocesar la misma entrada produce el mismo JSON, salvo que el prompt/contrato del
  servicio haya cambiado de versión.

## Batería de pruebas esperada

1. **Schema tests** — fixtures válidos pasan; inválidos (campo faltante, tipo erróneo, valor negativo
   donde no corresponde) fallan con un error legible.
2. **Determinismo** — correr la capa de decisión K veces sobre el mismo input y comparar que el
   resultado sea idéntico en las K corridas.
3. **Property-based** (recomendado) — sobre inputs generados, el resultado se mantiene en rango y la
   suma de drivers sigue cuadrando.
4. **Casos límite** — entradas vacías, negativos inválidos, variantes con acentos/typos, ceros,
   entradas ilegibles; los casos concretos salen del PRD/CLAUDE.md del consumidor.
5. **No-persistencia de datos regulados** — confirmar que el cálculo no escribe PII/datos sensibles
   crudos (los declarados por el consumidor) en persistencia, `localStorage` o logs.

## Veredicto

Toda invariante verificada con evidencia (test o `archivo:línea`) → `gates.data: true`. Cualquier
invariante sin cubrir → `false`, señalando el test o fix que falta.
