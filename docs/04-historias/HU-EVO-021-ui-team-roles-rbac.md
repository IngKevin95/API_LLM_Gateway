---
id: HU-EVO-021
titulo: UI React - Team & Roles (gestion RBAC de usuarios)
epica: EP-EVO-004
prioridad: Must
complejidad: M
estado: lista
---

# UI React - Team & Roles (gestión RBAC de usuarios)

Como **administrador del Gateway**, quiero **una pantalla en el dashboard para gestionar el equipo y sus roles**, para **no depender de llamar a la API a mano ni de pedirle a otra persona que edite configuración**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — ver tabla de equipo | Dado que soy admin y entro a la tab "Team" | Cuando la pantalla carga | Entonces veo una tabla con usuario (avatar+nombre+email), rol (badge Admin/Operator/Viewer), scopes/tenants, estado (badge Active/Invited/Suspended) y última actividad, consumiendo `GET /users` (HU-EVO-017) |
| 2 | Happy — invitar miembro | Dado que estoy en la tab "Team" | Cuando hago click en "Invite member", completo email+rol+scopes y confirmo | Entonces se llama `POST /users`, la tabla se actualiza mostrando el nuevo usuario en estado "Invited" |
| 3 | Happy — cambiar rol/suspender | Dado que veo un usuario activo en la tabla | Cuando uso la acción de edición y cambio su rol o lo suspendo | Entonces se llama `PATCH /users/:id` y el badge de la fila se actualiza sin recargar toda la página |
| 4 | Error — usuario no-admin accede a la tab | Dado que mi rol es "Operator" | Cuando intento acceder a la tab "Team" | Entonces la tab no aparece en la navegación (o redirige con mensaje de permiso insuficiente) — el frontend respeta el mismo RBAC que el backend, sin exponer controles inertes |
| 5 | Edge — invitar email duplicado | Dado que intento invitar un email ya existente | Cuando el backend devuelve 409 | Entonces la UI muestra un mensaje de error claro sin romper el formulario ni perder los datos ya tipeados |

## Checklist INVEST

- [x] Independent — depende de HU-EVO-017 (endpoint `/users`) para tener datos reales
- [x] Negotiable — diseño exacto de la tabla es detalle de implementación (fuente: Stitch)
- [x] Valuable — autoservicio para admins, elimina fricción operativa
- [x] Estimable — nuevo componente `TeamRoles.jsx` + nueva tab en `Dashboard.jsx`, mismo patrón que Overview/Quotas/Alerts/Providers
- [x] Small/Medium — 2 días
- [x] Testable — tests de componente con mock de `/users`, verifica render de tabla y flujo de invitar

## Notas técnicas

Fuente de diseño: Stitch project `12981760791975432480`, pantalla `screens/29f991058b5b4db7876c6d11ef699810` ("Team & Roles"), mismo design system "Gateway Ops Dark" ya usado en HU-EVO-014. Nuevo componente `src/ui/dashboard/TeamRoles.jsx`, agregado como 5ta tab en `Dashboard.jsx`, visible solo si el usuario autenticado tiene rol admin (mismo criterio de "no mostrar controles que el backend igual va a rechazar").

---

## Relación con existentes

- Depende de: HU-EVO-017 (`GET/POST/PATCH /users`)
- Integra: Dashboard React (HU-EVO-014), mismo patrón de tabs

## Build harness — referencia cruzada

Implementada en sub-slice `EP-EVO-004-SS3`, openspec change `team-roles-profile-security`
(`openspec/changes/team-roles-profile-security/`), rama `feature/ep-evo-004-ss3-team-roles-profile-security`.
