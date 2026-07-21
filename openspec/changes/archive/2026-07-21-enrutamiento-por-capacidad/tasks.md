## 1. Sub-slice 1 — Registry (HU-001)

- [x] 1.1 Añadir `gopkg.in/yaml.v3` a `go.mod` (justificado en design.md) y `go mod tidy`
- [x] 1.2 Definir structs del catálogo (`Provider`, `Model` con atributos de score, `Routing` por capacidad, `max_in_flight`, `stream_idle_timeout`) en `src/internal/registry`
- [x] 1.3 (test-first) Test de carga válida: `config.yaml` de ejemplo (Anexo A del PRD) → catálogo en RAM + modelos habilitados por capacidad + conteo en stdout (AC1)
- [x] 1.4 (test-first) Test fail-fast: YAML inválido/campo faltante → error con archivo+línea+campo, sin estado parcial (AC2)
- [x] 1.5 (test-first) Test secreto literal: `api_key` literal → rechazo, exige `${VAR}`, no imprime el valor (AC3)
- [x] 1.6 (test-first) Test capacidad sin modelos → marca no disponible + WARN, no aborta (AC4)
- [x] 1.7 (test-first) Test parámetros de red físicos expuestos (`max_in_flight`, `stream_idle_timeout`) (AC5)
- [x] 1.8 Implementar `Load(path)` que hace verde 1.3–1.7: parseo, validación estricta, resolución `${VAR}` desde entorno, logging sin secretos
- [x] 1.9 Cablear carga del Registry en boot en `src/cmd/gateway/main.go` (fallo de carga = fail-fast del proceso)
- [x] 1.10 journey_smoke sub-slice 1: `go build ./...` + arranque con `config.yaml` de ejemplo imprime conteo y `/health` sigue 200

## 2. Sub-slice 2 — Router por score + tokenizador de contexto (HU-002a, HU-035)

- [x] 2.1 (test-first) Interfaz `Tokenizer` + implementación heurística; tests happy/error/edge de ventana con buffer 20% (HU-035 AC1/AC2/AC3)
- [x] 2.2 Implementar tokenizador heurístico en `src/internal/tokenizer` que hace verde 2.1
- [x] 2.3 (test-first) Test función `Score(model, ctx)` determinista sobre las 6 variables normalizadas (tablas)
- [x] 2.4 (test-first) Test resolución óptima: 3 modelos "chat" → cadena de fallback ordenada por score desc (HU-002a AC1)
- [x] 2.5 (test-first) Test filtro por estado: top deshabilitado/unhealthy/sin cuota → excluido, segundo toma primer lugar (HU-002a AC2)
- [x] 2.6 (test-first) Test filtro pre-score por ventana de contexto descarta candidato antes del score (HU-002a AC3)
- [x] 2.7 Definir interfaces `HealthSource`/`QuotaSource` con stubs (inyección de estado vivo diferida a EP-002/003)
- [x] 2.8 Implementar `src/internal/router.Resolve(capacidad, ctx)` que hace verde 2.3–2.6
- [x] 2.9 journey_smoke sub-slice 2: resolución end-to-end en memoria contra `config.yaml` de ejemplo, verde

## 3. Sub-slice 3 — Errores/desempates + modelo explícito (HU-002b, HU-003)

- [x] 3.1 (test-first) Test capacidad desconocida → 400 (HU-002b AC1)
- [x] 3.2 (test-first) Test sin candidatos aptos → 503 sin failover inútil (HU-002b AC2)
- [x] 3.3 (test-first) Test desempate: menor costo, luego orden alfabético de ID (HU-002b AC3)
- [x] 3.4 (test-first) Test modelo explícito sano → usa ese modelo, sin scoring (HU-003 AC1)
- [x] 3.5 (test-first) Test modelo explícito inexistente → 404 listando válidos, sin sustituir (HU-003 AC2)
- [x] 3.6 (test-first) Test modelo caído con fallback permitido → aplica cadena + anota sustitución (HU-003 AC3)
- [x] 3.7 (test-first) Test modelo caído sin fallback → 503 sin sustituir (HU-003 AC4)
- [x] 3.8 Extender `src/internal/router` para hacer verde 3.1–3.7 (errores tipados, política de fallback configurable por request/global)
- [x] 3.9 journey_smoke sub-slice 3: recorrido completo automático + explícito + rutas de error, verde

## 4. Cierre de épica

- [x] 4.1 Coherencia triple AC↔specs↔tests (coherence-three-way) sin huecos
- [x] 4.2 Verificación adversarial de cableado (wiring-adversarial-verifier) → `wiring_verified=true`
- [x] 4.3 DoD reducido (dor-dod-gatekeeper) → `dod=true`
- [ ] 4.4 PR + `opsx:archive` del change en el mismo PR
