---
id: HU-EVO-016
titulo: Toggle de tema claro/oscuro en el Dashboard
epica: EP-EVO-003
prioridad: Should
complejidad: S
estado: lista
---

# Toggle de tema claro/oscuro en el Dashboard

Como **operador del Gateway**, quiero **alternar entre modo claro y modo oscuro en el dashboard**, para **usar la interfaz cómodamente tanto en un ambiente de oficina de día como en monitoreo nocturno**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — cambiar a modo claro | Dado que el dashboard está en modo oscuro (default) | Cuando el usuario hace click en el toggle de tema del header | Entonces la interfaz completa (las 4 tabs existentes, header, tablas, badges) cambia a la paleta clara sin recargar la página |
| 2 | Happy — cambiar a modo oscuro | Dado que el dashboard está en modo claro | Cuando el usuario hace click en el toggle de tema | Entonces vuelve a la paleta oscura original |
| 3 | Happy — persistencia | Dado que el usuario eligió modo claro | Cuando cierra y vuelve a abrir el navegador | Entonces el dashboard carga directamente en modo claro (preferencia guardada en localStorage) |
| 4 | Edge — preferencia del sistema operativo | Dado que el usuario nunca tocó el toggle (no hay preferencia guardada) | Cuando el dashboard carga | Entonces respeta `prefers-color-scheme` del navegador/SO como default |
| 5 | Edge — contraste de badges y gráficos | Dado que hay alertas críticas o gráficos de latencia visibles | Cuando se cambia de tema | Entonces los colores semáforo (verde/ámbar/rojo) y las series de los gráficos recharts mantienen suficiente contraste en ambos modos |

## Checklist INVEST

- [x] Independent — solo toca CSS/variables de tema del dashboard ya construido (HU-EVO-014), no requiere cambios de backend
- [x] Negotiable — mecanismo de persistencia (localStorage vs cookie) es detalle de implementación
- [x] Valuable — comodidad de uso en distintos ambientes, reduce fatiga visual
- [x] Estimable — CSS variables + toggle de estado + localStorage, sin lógica de negocio nueva
- [x] Small — 1 día
- [x] Testable — test de componente verifica cambio de clase/atributo de tema y persistencia en localStorage

## Notas técnicas

Basado en el prototipo Stitch "Gateway Ops Light" (variante del design system "Gateway Ops Dark" ya usado en HU-EVO-014), mismo layout y jerarquía, paleta invertida.

Implementación sugerida: variable CSS `data-theme="dark"|"light"` en la raíz de `gateway-ops-dark.css` (renombrar a `gateway-ops-theme.css` si aplica), toggle en el header junto a los íconos de refresh/settings ya existentes, persistencia en la misma clave de `authConfig.js` usada para otras preferencias del dashboard.

---

## Relación con existentes

- Extiende: HU-EVO-014 (Dashboard React con tabs) — mismo componente `Dashboard.jsx`
- Fuente de diseño: Stitch project 12981760791975432480, design system "Gateway Ops Light" (assets/14531664396944343373), pantalla `screens/837ce97fb506483a84f57de15441f1e0`
