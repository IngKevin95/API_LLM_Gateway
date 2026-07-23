---
id: HU-008
titulo: Autenticar toda petición con API key
epica: EP-004A
prioridad: Must
complejidad: S
estado: lista
---

# Autenticar toda petición con API key

Como **operador de seguridad**, quiero **que la Gateway exija una API key válida en cada petición**, para **que ningún consumidor anónimo acceda a los modelos**.

Contexto: AuthN base del producto; base de la seguridad empresarial. Actividad 1 del journey.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — key válida | Dado que una API key de cliente activa | Cuando llega una petición con esa key en el header | Entonces la Gateway la acepta y procesa; el consumidor queda identificado para auditoría |
| 2 | Error — sin key | Dado que una petición sin header de autenticación | Cuando llega a la Gateway | Entonces responde 401 sin procesar la petición ni exponer detalles internos |
| 3 | Error — key inválida o revocada | Dado que una API key inexistente o revocada | Cuando llega la petición con esa key | Entonces responde 401 y registra el intento fallido (sin loguear la key completa) |
| 4 | Edge — key en cuerpo o URL | Dado que una petición que pone la key en la query string en vez del header | Cuando llega a la Gateway | Entonces la Gateway no acepta la key por canal inseguro y responde 401, evitando exponer secretos en URLs/logs |

## Checklist INVEST

- [x] Independent — entregable sin dependencias
- [x] Negotiable — formato de key abierto
- [x] Valuable — cierra el acceso anónimo
- [x] Estimable — middleware de auth
- [x] Small — 1-2 días
- [x] Testable — keys válidas/ inválidas

## Notas técnicas

Comparación de keys en tiempo constante; almacenar hash, no la key en claro; nunca loguear la key.

> **OpenSpec change**: `ep-004a-identidad-accesos` (EP-004A)
