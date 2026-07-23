# GitHub Flow estricto — construcción

`main` es siempre desplegable. Todo cambio nace en rama corta y solo entra por Pull Request con
checks verdes — nunca por push directo. `gitflow-guard.sh` hace cumplir esto de forma determinista,
no como convención de honor.

## Qué bloquea el hook

- `git commit` estando parado en `main`/`master` → primero crea una rama.
- `git commit` desde una rama que no sea `feature/*`, `fix/*` o `chore/*`.
- `git push` apuntando directo a `main`/`master` → el camino correcto es un PR.

## Ciclo típico de un slice

```bash
git switch main && git pull --ff-only          # arrancar desde main al día
git switch -c feature/<slug>                    # rama tipada
# ... TDD + gates, todo registrado en build-state.json ...
git add -A && git commit -m "<tipo>: <mensaje>" # siempre en la rama, nunca en main
git push -u origin feature/<slug>
gh pr create --base main --head feature/<slug> --fill
```

## Convenciones de nombre y mensaje

- **Ramas**: `feature/<slug-kebab>` para valor nuevo, `fix/<slug>` para corrección, `chore/<slug>`
  para infraestructura o docs. Una rama por épica — un slice, una rama —, p. ej.
  `feature/ep-003-pricing-engine`. Las variantes `fix/*`/`chore/*` son también el carril que usa
  `building-a-micro-change` para mantenimiento que no abre épica.
- **Commits**: Conventional Commits — `feat:`, `fix:`, `test:`, `refactor:`, `chore:`, `docs:`.
- **Pull Requests**: título descriptivo, cuerpo enlazando la épica, sus HU (`hus[]`) y el change de
  OpenSpec; se mergea con lint/types/tests/newman en verde. Squash es lo recomendado por defecto.

## Lo que este modelo NO es

No es GitFlow clásico — sin ramas `develop` ni `release/*`. Una única línea estable (`main`), ramas
cortas de vida efímera, integración exclusivamente vía PR.
