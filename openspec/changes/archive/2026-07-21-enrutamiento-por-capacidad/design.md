## Context

Primera épica `foundational` del Gateway (greenfield Go 1.22+). El scaffold actual (`cmd/gateway/main.go`) es stdlib-only y solo levanta `/health` + `/metrics`. Esta épica introduce la capa de enrutamiento por capacidad como librería interna (sin endpoints HTTP nuevos; la API universal es EP-005). Contrato de datos: `config.yaml` según Anexo A del PRD técnico (`docs/13-tech-prd/api-llm-gateway.md`). ADR-001 obliga backend Go idiomático sin frameworks pesados de IA.

## Goals / Non-Goals

**Goals:**
- Cargar `config.yaml` (providers/models/routing) a RAM en boot con validación estricta y fail-fast.
- Resolver `capacidad → modelo` por score determinista de 6 variables, con modo automático y explícito.
- Filtrar candidatos por ventana de contexto (buffer 20%) antes de calcular score.
- Resolución de secretos por `${VAR}`; jamás persistir/loguear valores literales.
- Manejo determinista de errores y desempates (reproducible, testeable).

**Non-Goals:**
- Endpoints HTTP OpenAI/Anthropic-compat (EP-005).
- Llamadas reales a proveedores / adapters de red (EP-002).
- Health Monitor, Quota Manager, persistencia (EP-002/003/009). Aquí solo se **leen** `disponibilidad` y `cuota restante` como inputs del score; su fuente viva llega en otras épicas — en esta se inyectan vía interfaz.

## Decisions

- **Estructura de paquetes**: `internal/registry` (carga+validación YAML→structs), `internal/router` (scoring+resolución), `internal/tokenizer` (estimación+validación de ventana). `internal/` fuerza encapsulamiento (no importable fuera del módulo). Alternativa descartada: un solo paquete `gateway` — viola responsabilidad única y encarece el testeo aislado.
- **Parser YAML**: `gopkg.in/yaml.v3`. La stdlib de Go no parsea YAML y el contrato del PRD (Anexo A) es YAML. Alternativa descartada: convertir a JSON con stdlib `encoding/json` — obligaría a mantener el `config.yaml` como JSON, contradiciendo el PRD. **Acción**: registrar `gopkg.in/yaml.v3` como dependencia justificada (fuente: PRD §Anexo A). Nota: `stack-allowlist.json` está orientado a npm y `stack-guard.sh` solo dispara sobre `package.json`; el `go.mod` no queda cubierto por el hook, así que la disciplina de stack para Go se documenta aquí y se revisa en el Release Gate (`stack-guardian`).
- **Tokenizador**: interfaz `Tokenizer` con implementación heurística por defecto (conteo de palabras × factor) + buffer 20% (HU-035). `tiktoken-go` se deja como implementación intercambiable futura (Negotiable en INVEST de HU-035), no se compromete ahora para no arrastrar dependencia pesada en el primer sub-slice.
- **Score**: función pura `Score(model, ctx) float64` sobre las 6 variables normalizadas; pesos configurables (fijos en esta épica; el ajuste dinámico es EP-007 Learning Engine). Determinismo total → testeable con tablas.
- **Inyección de estado vivo**: `disponibilidad` y `cuota restante` entran al router vía interfaces (`HealthSource`, `QuotaSource`) con implementaciones stub en esta épica. Evita acoplar EP-001 a EP-002/003 y respeta "fundación antes que negocio".
- **Secretos**: el Registry rechaza en validación cualquier `api_key` con valor literal; exige patrón `${VAR}` y resuelve contra el entorno en carga. El valor resuelto nunca se imprime (ni en errores de validación).

## Risks / Trade-offs

- **Score con inputs stub (disponibilidad/cuota fijos)** → el ranking en esta épica no refleja salud real. Mitigación: interfaces listas para que EP-002/003 inyecten fuentes vivas sin tocar el router.
- **Heurística de tokens imprecisa** vs. conteo real del proveedor → riesgo de falso-positivo/negativo en el borde de la ventana. Mitigación: buffer 20% (conservador) + interfaz `Tokenizer` intercambiable por `tiktoken-go`.
- **`go.mod` fuera del guard de stack** → una dependencia no justificada podría colarse sin que el hook la frene. Mitigación: decisiones de dependencia documentadas en este design y revisadas por `stack-guardian` en el Release Gate.

## Migration Plan

Greenfield: sin migración ni rollback de datos. Despliegue = merge del PR de la épica. El `config.yaml` de ejemplo (Anexo A) sirve de fixture de arranque y de tests.

## Open Questions

- Ninguna bloqueante. `tiktoken-go` queda diferido a decisión de sub-slice 2 (HU-035) según precisión medida de la heurística.
