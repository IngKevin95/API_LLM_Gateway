# Contratos de API con Newman

`api-contract-tester` usa esta guía en slices que exponen Route Handlers o Server Actions. Newman es
el runner CLI de colecciones Postman — no es un MCP, corre vía Bash.

## Layout de archivos

```
tests/postman/
├── <slice>.postman_collection.json    # requests + aserciones del slice
└── environment.json                   # baseUrl y variables (sin secretos reales)
```

## Diseñar la colección a partir de los AC

Cada endpoint del slice necesita, como mínimo, tres requests que reflejen sus escenarios
Given/When/Then:

- **Happy** — entrada válida, status 2xx, y la **forma de la respuesta** validada contra esquema (no
  basta con el status code).
- **Error** — entrada inválida, 4xx con mensaje útil, sin filtrar PII/datos sensibles ni detalles
  internos de implementación.
- **Edge** — límites reales del dominio: campo obligatorio ausente, respuesta parcial de un servicio
  externo, casos borde que el consumidor haya declarado.

Ejemplo de aserciones (`pm.test`):

```javascript
pm.test("status", () => pm.response.to.have.status(200));
pm.test("shape", () => {
  const b = pm.response.json();
  pm.expect(b).to.have.property("resultado");
  // sustituir por el contrato real del consumidor — las decisiones de alto
  // impacto del dominio se declaran en su domain-pack/PRD
  pm.expect(b.decision).to.be.oneOf(["DECISION_A","DECISION_B","DECISION_C"]);
});
```

## Cómo correrlo

```bash
# 1) levantar la app (si aplica)
npm run dev   # o: npm run build && npm run start
# 2) correr Newman cuando el server responda
npx --yes newman run tests/postman/<slice>.postman_collection.json \
  -e tests/postman/environment.json \
  --reporters cli,json \
  --reporter-json-export .claude/state/newman-<slice>.json
```

## Del resultado al gate

- 100% de requests y aserciones en verde → `gates.api: true`.
- Cualquier falla → `gates.api: false`, reportando request, aserción y causa raíz.
- Slice sin endpoints → `gates.api: null` (no bloquea el DoD).

> Datos siempre sintéticos: ni PII/datos sensibles reales ni claves de servicios externos
> server-side deben aparecer en la colección o el environment versionados.
