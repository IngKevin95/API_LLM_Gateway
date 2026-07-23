# Trazabilidad change ↔ épica (bloque `## Trazabilidad`)

Un OpenSpec change no flota solo: apunta siempre a **una épica** — la unidad de construcción — y
declara qué historias de esa épica cubre. Ese puente entre `docs/` (discovery) y `openspec/`
(construcción) lo audita el agente `change-epic-coherence`.

## Regla central: exactamente una épica por change

Un change nunca cruza dos épicas. Si el trabajo real abarca `EP-003` y `EP-004`, son **dos** changes
(dos slices), no uno con alcance mezclado.

Las historias declaradas tienen que ser las mismas que viven en `active_slice.hus[]`: cada `HU-XXX`
referenciado existe en `docs/04-historias/` y su campo `epica:` coincide con la EP del change. No se
listan HU "por si acaso" — solo las que entran en el alcance real de este slice.

## Cómo se declara

Al final del **cuerpo markdown** de `openspec/changes/<name>/proposal.md`, agrega:

```markdown
## Trazabilidad
- Épica: EP-003
- Historias: HU-010, HU-011, HU-012
- Discovery: docs/03-backlog/epicas.md#ep-003
```

**Nunca en el frontmatter YAML.** `openspec validate --strict` es estricto con la forma del
frontmatter; un campo extra ahí rompe la validación. El bloque markdown, en cambio, es inerte para
el validador y seguro de usar.

## Convención de nombre

`kebab-case` derivado de la épica: `ep-003-pricing-engine`. Así el nombre del change ya delata a qué
épica pertenece sin tener que abrir el archivo.

## Cierre del enlace: back-reference al archivar

Cuando el change se archiva, el enlace se vuelve bidireccional: anota
`> OpenSpec change: ep-003-pricing-engine` tanto en la épica (`docs/03-backlog/epicas.md`) como en
cada HU listada en `hus[]`. Sin esta nota de vuelta, alguien que lea la épica no puede encontrar qué
change la implementó.

## Verificación

```bash
openspec validate "<name>" --type change --strict --json
```

`change-epic-coherence` corre esta validación y además confirma: la épica existe, cada HU existe y
pertenece a esa épica, y la lista `Historias:` coincide exactamente con `hus[]`.
