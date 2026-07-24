---
id: HU-EVO-022
titulo: UI React - Profile & Security (API keys, sesiones, 2FA)
epica: EP-EVO-004
prioridad: Must
complejidad: M
estado: lista
---

# UI React - Profile & Security (API keys, sesiones, 2FA)

Como **usuario del Gateway**, quiero **gestionar mis propias credenciales y seguridad desde el dashboard**, para **rotar API keys, activar 2FA y revisar mis sesiones sin depender de un administrador**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — ver perfil y API keys | Dado que entro a la tab "Profile & Security" | Cuando la pantalla carga | Entonces veo mi tarjeta de perfil (avatar, nombre, email, rol) y la tabla de mis API keys (nombre, prefijo enmascarado, scopes, creación, último uso), consumiendo `GET /users/:id/api-keys` (HU-EVO-018) |
| 2 | Happy — generar nueva key | Dado que estoy en la sección API Keys | Cuando hago click en "Generate new key", le pongo nombre y confirmo | Entonces se llama `POST /users/:id/api-keys`, la key en texto plano se muestra **una sola vez** en un modal con advertencia de copiarla ahora, y luego solo aparece enmascarada en la tabla |
| 3 | Happy — revocar key y cerrar sesión | Dado que tengo una key activa y una sesión activa | Cuando reboco la key o cierro la sesión desde sus respectivos botones | Entonces se llaman `DELETE /users/:id/api-keys/:keyId` y `DELETE /sessions/:id`, y la fila correspondiente desaparece de la tabla sin recargar toda la página |
| 4 | Happy — activar 2FA | Dado que no tengo 2FA activo | Cuando activo el toggle de 2FA, escaneo el QR mostrado y confirmo el código | Entonces se llaman `POST /auth/mfa/enroll` y `POST /auth/mfa/verify` (HU-EVO-020), el toggle queda en estado activo |
| 5 | Error — código 2FA incorrecto al activar | Dado que estoy confirmando el enrolamiento de 2FA | Cuando ingreso un código incorrecto | Entonces la UI muestra el error del backend sin activar el toggle, permitiendo reintentar |

## Checklist INVEST

- [x] Independent — depende de HU-EVO-018/019/020 (endpoints) para datos reales
- [x] Negotiable — diseño exacto de la pantalla es detalle de implementación (fuente: Stitch)
- [x] Valuable — autoservicio de seguridad, reduce carga operativa sobre admins
- [x] Estimable — nuevo componente `ProfileSecurity.jsx`, reutiliza patrón de tabs existente
- [x] Small/Medium — 2-3 días
- [x] Testable — tests de componente con mocks de los 3 endpoints, verifica flujos de generar/revocar/activar-2FA

## Notas técnicas

Fuente de diseño: Stitch project `12981760791975432480`, pantalla `screens/67a1f823cd9c4b97828785f4e37b5bbc` ("Profile & Security"), mismo design system "Gateway Ops Dark". Nuevo componente `src/ui/dashboard/ProfileSecurity.jsx`, 6ta tab en `Dashboard.jsx`, visible para cualquier usuario autenticado (a diferencia de Team & Roles, que es admin-only). Sección "Notification preferences" reutiliza `authConfig.js` (umbral y sonido) ya construido en HU-EVO-015 — no duplicar ese estado.

---

## Relación con existentes

- Depende de: HU-EVO-018 (API keys), HU-EVO-019 (sesiones), HU-EVO-020 (2FA)
- Integra: Dashboard React (HU-EVO-014), `authConfig.js` (HU-EVO-015)
