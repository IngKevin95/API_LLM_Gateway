# CLAUDE.md

## Vertical documentaria Factory

Este proyecto usa la vertical Factory (`@factory/spec-driven-product`) para producir y mantener artefactos de discovery / producto:

```
PRD → User Story Map → Backlog → Historias (AC G/W/T) → Priorización → Flows
```

### Estructura

```
docs/
├── 01-prd/                 PRD (12 componentes obligatorios)
├── 02-user-story-map/      Mapa estilo Jeff Patton (backbone + ranking + releases)
├── 03-backlog/             epicas.md + backlog.md (tabla ordenada)
├── 04-historias/           HU-XXX.md con frontmatter YAML obligatorio + AC en G/W/T
├── 05-priorizacion/        Aplicación de framework (MoSCoW/RICE/Valor-Esfuerzo/Eisenhower)
└── 06-flows/               Flujos de navegación en Mermaid, un archivo por épica
```

### Reglas duras (no negociables)

1. Toda Historia de Usuario vive en `docs/04-historias/HU-XXX.md` con frontmatter YAML obligatorio (`id, titulo, epica, prioridad, complejidad, estado`).
2. Todo Acceptance Criterion se escribe en **Given/When/Then**; nada de prosa libre. 3-5 escenarios incluyendo happy / error / edge.
3. Toda Historia pasa los 6 criterios **INVEST** antes de marcarse `estado: lista`.
4. La metodología canónica vive en el `METODOLOGIA.md` del paquete `@factory/spec-driven-product` (ruta exacta depende de dónde npm haya instalado el paquete global; típicamente `$(npm root -g)/@factory/spec-driven-product/METODOLOGIA.md`). Si una skill contradice la metodología, **gana la metodología**.
5. Ningún artefacto se marca como "final" sin pasar por su agente revisor (manual con `/factory:revisar` o automático si `FACTORY_AUTO_AUDIT=true`).

### Slash commands disponibles

| Operativos | Por artefacto |
|---|---|
| `/factory:onboard` — parametriza el proyecto | `/factory:prd` — escribe PRD |
| `/factory:flujo` — pipeline completo | `/factory:epicas` — descompone PRD |
| `/factory:revisar` — auditoría global | `/factory:mapa` — User Story Map |
|  | `/factory:historia` — historia de usuario |
|  | `/factory:ac` — criterios de aceptación BDD |
|  | `/factory:invest` — valida INVEST |
|  | `/factory:backlog` — backlog consolidado |
|  | `/factory:priorizar` — priorización |
|  | `/factory:flows` — flujos de navegación (Mermaid, por épica) |

### Contexto del proyecto

- **Nombre**: API LLM Gateway
- **Dominio**: Gateway universal de LLMs con selección automática y failover
- **Stakeholders**: Kevin Beltrán, equipo, agentes consumidores
- **Framework de priorización**: Valor / Esfuerzo (default backlog)

### Convenciones de naming

- Épicas: `EP-001`, `EP-002`, … (3 dígitos)
- Historias: `HU-001`, `HU-002`, … (3 dígitos)
- Slugs en kebab-case-sin-acentos

### Hooks de calidad (opt-in)

Dos flags opcionales en `.claude/settings.local.json`:

```json
{
  "env": {
    "FACTORY_AUTO_AUDIT": "false",
    "FACTORY_REFLECT_SESSION": "false"
  }
}
```

- `FACTORY_AUTO_AUDIT=true` → tras cada Write/Edit en `docs/`, dispara el agente revisor correspondiente. Cuando esté off (default), usar `/factory:revisar` manualmente.
- `FACTORY_REFLECT_SESSION=true` → al cerrar sesión, escanea heurísticas y propone mejoras en `docs/.factory-suggestions.md`.

## Estado del repositorio

Proyecto **greenfield en Fase 1 de construcción**: documentación de discovery (PRD, épicas, historias, arquitectura) completada y traceada bidireccionalamente. Fase 1 MVP construyendo componentes básicos (Registry, Router, Health Monitor, Adapters). Idioma: **español**.

Documentos de referencia en orden de lectura:
- `docs/01-prd/api-llm-gateway.md` — PRD con visión y NFR de disponibilidad/latencia/throughput
- `docs/11-architecture/api-llm-gateway.md` — 18 componentes C4 y tabla de ubicación/latencia
- `docs/13-tech-prd/api-llm-gateway.md` — especificación técnica, YAML schema, SLA
- `docs/04-historias/HU-*.md` — 48 historias traceadas a épicas con AC en formato Given/When/Then

## Principio central de diseño

Los agentes **no consumen modelos, consumen capacidades**:
- Agente pide: `router.coding()`, `router.reasoning()`, `router.vision()`, `router.embedding()`
- Gateway resuelve: modelo + proveedor + API key en tiempo de ejecución
- Resultado: desacoplamiento total agente-proveedor

**Ningún código de agente debe referenciar OpenAI/Anthropic/Google o un modelo concreto.**
