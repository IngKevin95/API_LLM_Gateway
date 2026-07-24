---
id: HU-EVO-018
titulo: Gestion de API keys por usuario (generar/listar/revocar)
epica: EP-EVO-004
prioridad: Must
complejidad: M
estado: lista
---

# Gestión de API keys por usuario (generar/listar/revocar)

Como **usuario del Gateway**, quiero **generar y revocar mis propias API keys desde el dashboard**, para **rotar credenciales sin pedirle a un admin que edite un archivo de configuración**.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — generar key | Dado que estoy autenticado | Cuando hago `POST /users/:id/api-keys` con un nombre descriptivo | Entonces recibo la key en texto plano **una sola vez** en la respuesta, y se persiste solo su hash (mismo criterio que `apikey.Store` actual: nunca se guarda ni se loguea en claro) |
| 2 | Happy — listar keys enmascaradas | Dado que tengo 3 keys generadas | Cuando hago `GET /users/:id/api-keys` | Entonces veo nombre, prefijo enmascarado (ej. `sk-***4f2a`), scopes, fecha de creación y último uso — nunca la key completa |
| 3 | Happy — revocar key | Dado que tengo una key activa | Cuando hago `DELETE /users/:id/api-keys/:keyId` | Entonces esa key deja de autenticar inmediatamente (siguiente request con ella devuelve 401) |
| 4 | Error — revocar key de otro usuario | Dado que soy usuario no-admin | Cuando intento revocar una key de otro usuario | Entonces recibo 403 Forbidden |
| 5 | Edge — último uso se actualiza | Dado que una key nunca fue usada (`last_used_at` null) | Cuando se autentica una request con ella | Entonces `last_used_at` se actualiza a la fecha/hora de esa request |

## Checklist INVEST

- [x] Independent — depende de HU-EVO-017 (existencia de `users`) pero es un módulo propio (`api_keys` table)
- [x] Negotiable — algoritmo de hash y formato del prefijo son detalle de implementación
- [x] Valuable — autoservicio de rotación de credenciales, reduce fricción operativa y riesgo de keys viejas sin rotar
- [x] Estimable — tabla `api_keys` + endpoints CRUD + migrar `apikey.Store.lookup` a consultar esta tabla en vez del mapa in-memory
- [x] Small/Medium — 2 días
- [x] Testable — test de integración: generar, verificar que autentica, revocar, verificar que deja de autenticar

## Notas técnicas

Este slice reemplaza el seed de `GATEWAY_API_KEYS` (env var) por lectura desde PostgreSQL en boot + refresco periódico o invalidación activa al revocar. Mantener el mismo criterio de seguridad ya vigente en `apikey.go`: comparación en tiempo constante (`subtle.ConstantTimeCompare`), nunca loguear la key completa.

---

## Relación con existentes

- Depende de: HU-EVO-017 (tabla `users`)
- Reemplaza: seed manual vía `GATEWAY_API_KEYS` en `cmd/gateway/main.go`
- Requisito para: HU-EVO-022 (UI Profile & Security, sección API Keys)
