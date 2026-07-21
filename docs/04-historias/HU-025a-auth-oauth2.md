---
id: HU-025a
titulo: Autenticación OAuth2/OIDC
epica: EP-004A
prioridad: Should
complejidad: M
estado: lista
---

# Autenticación OAuth2/OIDC

Como **ingeniero de seguridad**, quiero **soportar autenticación mediante OAuth2/OIDC**, para **integrar el Gateway con el IdP corporativo**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — OIDC auth | Dado que un cliente con token JWT válido de un IdP corporativo | Cuando envía la petición | Entonces la gateway valida la firma, aud/iss y autoriza el acceso |
| 2 | Error — Token expirado | Dado que un cliente con token JWT expirado | Cuando envía la petición | Entonces el Gateway rechaza con HTTP 401 Unauthorized |
| 3 | Error — Firma inválida | Dado que un cliente con token JWT adulterado | Cuando envía la petición | Entonces el Gateway rechaza con HTTP 401 y loguea intento de intrusión |
| 4 | Sad path — JWT válido pero claims incorrectos | Dado que un cliente envía un JWT criptográficamente válido pero sin el claim de audience (aud) correcto | Cuando el middleware evalúa el JWT | Entonces se rechaza la petición con un 403 Forbidden |

## Checklist INVEST

- [x] Independent — AuthN es un middleware puro que no toca lógica de routing.
- [x] Negotiable — Proveedor inicial puede ser Okta/Auth0; flujos avanzados se pueden omitir en v1.
- [x] Valuable — Habilita la adopción corporativa sin compartir API keys estáticas.
- [x] Estimable — Patrón estándar OIDC con librerías maduras.
- [x] Small — Solo valida JWTs contra JWKS, sin manejar login en sí.
- [x] Testable — Verificable mockeando el endpoint JWKS y pasando tokens vencidos/válidos.

## Notas técnicas
- Asegurar alineación con NFRs de latencia y uso de caché si aplica.
