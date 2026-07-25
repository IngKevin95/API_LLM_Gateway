# Épicas Evolutivas — CC-001 Proveedores Gratuitos + Aprendizaje de Cuota + Dashboard de Alertas

**Cambio:** CC-001 — Ampliar Gateway con proveedores gratuitos curados, aprendizaje de cuota real-time desde headers, dashboard con alertas por tenant.

**Trazabilidad:** Extienden EP-001 (Enrutamiento), EP-003 (Gobernanza), EP-004A (Seguridad). Impactan HU-006 (Cuota), HU-060 (Métricas), HU-009 (RBAC).

**Fecha estimada:** 2026-08-30 (Fase 1: primeros 5 proveedores + aprendizaje + alertas básicas)

---

## EP-EVO-001 · Adapters de Proveedores Gratuitos Curados

| Campo | Valor |
|---|---|
| **Tipo** | evolutivo |
| **Extiende** | EP-002 (Resiliencia y Conectividad) |
| **Impacta** | HU-024 (Adapters locales), HU-020* (OpenAI), HU-021* (Anthropic) |
| **Objetivo(s) del PRD cubiertos** | Obj. 1 (Desacople total), Obj. 3 (Selección óptima) |
| **Riesgo** | bajo (adapters = config-driven, sin cambio de infraestructura) |
| **Métrica de éxito** | Agregar ≥5 nuevos proveedores en Fase 1 sin refactorizar code existente; cada proveedor pasa conformance_test.go |

### Descripción

Extender el catálogo de proveedores del Gateway incorporando **5 nuevos proveedores gratuitos de OmniRoute** (Groq, Cerebras, Mistral, Gemini, Cloudflare AI) en Fase 1, y definir el patrón para agregar el resto (~15 más) en Fase 2.

**Por qué existe:** Los proveedores gratuitos (sin tarjeta, sin OAuth) diversifican el failover y extienden la capacidad del Gateway a equipos con presupuesto cero. Los 5 prioritarios suman ~1.19B tokens/mes documentados (Mistral 1.0B + Gemini 60M + Cerebras 30M + Cloudflare 30M + Groq 15M).

**Approach:** Cada proveedor es una configuración declarativa en `config/providers/free-tier.yaml` + un adapter data-driven genérico que reutiliza la lógica probada de `openai.Adapter` (wrapping, como hace `omniroute.go` hoy). Sin código nuevo por proveedor.

### Capabilities

- Registry ampliado: cargar `config/providers/free-tier.yaml` con 5 nuevos proveedores
- Adapter genérico data-driven (`src/internal/adapter/generic/adapter.go`) que implementa `Chat/Stream/Embed` desde `ProviderSpec`
- Conformance tests extendidos para validar cada `ProviderSpec`
- Health Monitor: detectar 429 de cada proveedor y retirar temporalmente
- Quota Manager: contadores iniciales por proveedor (hint desde YAML)

### Historias anticipadas

- **HU-EVO-001** — Crear adapter genérico data-driven (baseURL, authHeader, formato openai|claude)
- **HU-EVO-002** — Cargar catálogo `free-tier.yaml` en Registry (5 nuevos proveedores)
- **HU-EVO-003** — Extender conformance_test.go para validar cada ProviderSpec
- **HU-EVO-004** — Health Monitor: detectar y retirar temporalmente proveedores 429
- **HU-EVO-005** — Quota Manager: inicializar contadores y hints por proveedor gratuito

### Change

Implementada por el openspec change `adapters-proveedores-gratuitos-curados` (branch `feature/ep-evo-001-adapters-gratuitos`). Ver `openspec/changes/adapters-proveedores-gratuitos-curados/`.

---

## EP-EVO-002 · Aprendizaje de Cuota desde Headers HTTP + Persistencia

| Campo | Valor |
|---|---|
| **Tipo** | evolutivo |
| **Extiende** | EP-003 (Gobernanza de Consumo) |
| **Impacta** | HU-006 (Contabilizar cuota), HU-017 (Métricas por modelo/proveedor) |
| **Objetivo(s) del PRD cubiertos** | Obj. 4 (Seguridad empresarial), KPI (Gobernanza de cuota) |
| **Riesgo** | medio (cambio en quota.Manager pero aislado; no toca auth/failover) |
| **Métrica de éxito** | Quota Manager aprende límites reales de cada proveedor tras primer request; persiste en PostgreSQL; Router penaliza automáticamente cuando remaining < 20%; Fase 2 agrega predicción de reset |

### Descripción

Extender `quota.Manager` para aprender automáticamente los límites reales de cada proveedor desde headers HTTP estándar (`X-RateLimit-Limit/Remaining/Reset` + variantes por proveedor: Anthropic, OpenAI, Groq, etc.), sin auditoría cron externa.

**Por qué existe:** Cada proveedor devuelve su cuota real en headers de respuesta. Hoy HU-006 arranca con `quota_hint` del YAML (valor estático); aprendiendo desde headers obtenemos cuota dinámica y exacta, sin esperar auditoría programada.

**Approach:** Nueva función `LearnFromHeaders(providerID, modelID, headers, status)` invocada después de cada respuesta (en el point donde hoy se llama `quota.Manager.Commit`). Parsea headers, actualiza RAM + persiste en PostgreSQL, y penaliza score del Router cuando remaining < 20%.

### Capabilities

- Parseo de headers estándar (`X-RateLimit-*`) + variantes (Anthropic `anthropic-ratelimit-*`, OpenAI, Groq, etc.)
- Actualización dinámica de `Remaining(providerID, modelID)` en memoria
- Persistencia asíncrona en PostgreSQL (tabla `provider_quotas_learned`)
- Penalización automática del score cuando remaining < 20%
- Manejo de 429 (exhausted) con `reset_at` derivado del header `Retry-After` o `X-RateLimit-Reset`
- Historial de aprendizaje (auditoría: cuota aprendida, timestamp, modelo)

### Historias anticipadas

- **HU-EVO-006** — Parsear headers estándar (`X-RateLimit-*`) por adapter
- **HU-EVO-007** — Implement `LearnFromHeaders()` en quota.Manager con actualización RAM
- **HU-EVO-008** — Persistencia asíncrona en PostgreSQL (`provider_quotas_learned`)
- **HU-EVO-009** — Router penaliza score cuando remaining < 20%
- **HU-EVO-010** — Manejo de 429: retirar proveedor hasta `reset_at`

---

## EP-EVO-003 · Dashboard Ampliado + Alertas por Tenant + UI React

| Campo | Valor |
|---|---|
| **Tipo** | evolutivo |
| **Extiende** | EP-001 (Enrutamiento) + EP-004A (Seguridad), HU-060 (Métricas) |
| **Impacta** | HU-017 (Métricas por modelo/proveedor), HU-023 (Dashboard), HU-009 (RBAC) |
| **Objetivo(s) del PRD cubiertos** | Obj. 4 (Seguridad empresarial), Obj. 5 (Compatibilidad) |
| **Riesgo** | bajo (extiende endpoint existente, new UI no toca API core) |
| **Métrica de éxito** | GET `/metrics` devuelve bloque `quota[provider, model, remaining, reset_at, status]` + alertas filtradas por tenant/scope; UI React muestra dashboard con visualización real-time; alertas enviadas cuando remaining < 10% umbral configurable |

### Descripción

Ampliar `/metrics` para exponer desglose de cuota por proveedor y modelo (desde EP-EVO-002), y construir una UI React que visualice métricas + alertas en tiempo real, respetando RBAC (cada usuario ve solo su tenant + capacidades autorizadas).

**Por qué existe:** Hoy HU-060 devuelve `/metrics` genérico (uptime, requests totales, latency); falta visibilidad de cuota por proveedor/modelo y alertas cuando se acerca el límite. Operadores necesitan dashboards para anticipar fallos y usuarios necesitan notificaciones cuando su cuota está baja.

**Approach:** 
1. Extender `metrics.Handler` + `metrics.Store` con bloque `quota` (alimentado por `quota.Manager.Snapshot()`)
2. Implementar `alert.Manager` que genera alertas cuando remaining < umbral (default 10%, configurable)
3. Filtrar alertas según RBAC: admin ve todas, usuarios ven solo su tenant + capacidades autorizadas
4. UI React que consume `/metrics` y `/alerts` cada 5 segundos, renderiza tablas + gráficos, notificaciones browser

### Capabilities

- Endpoint `/metrics` ampliado: bloque `quota[{provider, model, limit, remaining, reset_at, healthy}]`
- Alert Manager: genera alertas en `provider_alerts` (PostgreSQL) cuando remaining < umbral
- Filtrado RBAC: alerts respetan scopes + tenant del usuario
- GET `/alerts?tenant=T1&scope=coding` — devuelve alertas filtradas por tenant + capacidades
- UI React: dashboard con tabs (Overview, Quotas, Alerts, Providers)
  - Real-time graphs: cuota consumida por proveedor (últimas 24h)
  - Alert notifications: popups cuando umbral alcanzado
  - Admin panel: ver todas las alertas, configurar umbrales por proveedor
- Persistencia: histórico de alertas en `provider_alerts` para auditoría

### Historias anticipadas

- **HU-EVO-011** — Extender `metrics.Store` con snapshot de cuota por proveedor/modelo
- **HU-EVO-012** — Implementar `alert.Manager` que genera alertas en PostgreSQL
- **HU-EVO-013** — Filtrado RBAC en GET `/alerts` (respeta tenant + scopes)
- **HU-EVO-014** — UI React: dashboard de métricas + alertas (tabs Overview, Quotas, Alerts)
- **HU-EVO-015** — Notificaciones browser cuando remaining < umbral

### Change (SS1: HU-EVO-011/012/013)

Implementado en `openspec/changes/metrics-quota-alertas-rbac` (rama
`feature/ep-evo-003-ss1-metrics-alertas-rbac`). SS2 (HU-EVO-014/015, UI React)
implementado en `feature/ep-evo-003-ss2-ui-react-dashboard`. HU-EVO-016
(toggle claro/oscuro) extiende el mismo Dashboard.jsx de SS2, sin backend
nuevo.

---

## EP-EVO-004 · Gestión de Usuarios, Roles y Seguridad de Cuenta

| Campo | Valor |
|---|---|
| **Tipo** | evolutivo |
| **Extiende** | EP-004A (Seguridad), HU-009 (RBAC), EP-EVO-003 (Dashboard React) |
| **Impacta** | `internal/auth/apikey` (hoy in-memory, seedeado por env var, sin CRUD) |
| **Objetivo(s) del PRD cubiertos** | Obj. 4 (Seguridad empresarial) |
| **Riesgo** | alto (toca autenticación, sesiones, secretos — requiere `security-reviewer` antes de release) |
| **Métrica de éxito** | Un admin puede invitar/suspender miembros del equipo con rol y scopes desde el dashboard; cada usuario puede rotar su propia API key, activar 2FA y cerrar sesiones activas sin intervención manual en la base de datos |

### Descripción

Reemplaza el `apikey.Store` in-memory sembrado por `GATEWAY_API_KEYS` (env var, sin
persistencia, sin gestión desde UI) por un store persistente en PostgreSQL con
CRUD real de usuarios, roles (`admin`/`operator`/`viewer`), scopes/tenants
asignados, y las pantallas "Team & Roles" y "Profile & Security" del dashboard
React (HU-EVO-014).

**Por qué existe:** hoy la única forma de dar de alta un usuario es editar la
variable de entorno `GATEWAY_API_KEYS` y redeployar — no hay UI, no hay
revocación individual, no hay sesiones, no hay 2FA. A medida que el Gateway
suma operadores reales esto se vuelve un cuello de botella operativo y un
riesgo de seguridad (keys que nunca se rotan).

**Approach:**
1. Store de usuarios persistente en PostgreSQL (tabla `users`, hash de
   password con bcrypt/argon2, nunca texto plano — mismo criterio que
   `apikey.Store` de no loguear secretos).
2. CRUD de usuarios + roles vía endpoints REST protegidos (solo `admin`
   puede invitar/suspender/cambiar rol).
3. Gestión de API keys por usuario: generar, listar (prefijo enmascarado),
   revocar — reemplaza el seed manual por `GATEWAY_API_KEYS`.
4. Sesiones: listar sesiones activas por dispositivo/IP aproximada, cerrar
   sesión individual o todas.
5. 2FA/MFA: TOTP estándar (compatible con Google Authenticator/Authy),
   opcional por usuario.
6. UI: pantallas "Team & Roles" y "Profile & Security" del Dashboard React
   (HU-EVO-014), consumiendo los endpoints anteriores.

### Capabilities

- `POST /users` (invitar), `PATCH /users/:id` (rol/scopes/estado), `DELETE /users/:id` (suspender) — solo admin
- `GET /users` — lista filtrada por tenant si no es admin global
- `POST /users/:id/api-keys` (generar), `DELETE /users/:id/api-keys/:keyId` (revocar)
- `GET /sessions`, `DELETE /sessions/:id` — sesiones activas del usuario autenticado
- `POST /auth/mfa/enroll`, `POST /auth/mfa/verify` — alta y verificación de TOTP
- UI React: pantallas "Team & Roles" (tabla de usuarios, invitar, badges de rol/estado) y "Profile & Security" (API keys, 2FA, sesiones, preferencias de notificación)

### Historias anticipadas

- **HU-EVO-017** — Store de usuarios persistente en PostgreSQL + CRUD admin (`users` table, roles, scopes)
- **HU-EVO-018** — Gestión de API keys por usuario (generar/listar enmascarada/revocar), reemplaza seed por env var
- **HU-EVO-019** — Sesiones activas: listar y cerrar (individual/todas)
- **HU-EVO-020** — 2FA/TOTP opcional por usuario
- **HU-EVO-021** — UI React "Team & Roles" (consume HU-EVO-017)
- **HU-EVO-022** — UI React "Profile & Security" (consume HU-EVO-018/019/020)

### Fuente de diseño

Stitch project `12981760791975432480` ("API LLM Gateway - Dashboard de
Métricas"), design system "Gateway Ops Dark": pantallas
`screens/29f991058b5b4db7876c6d11ef699810` (Team & Roles) y
`screens/67a1f823cd9c4b97828785f4e37b5bbc` (Profile & Security).

### Nota de secuenciación

Por la regla del arnés "primero cimientos, después negocio, en trozos
chicos" (CLAUDE.md), esta épica toca autenticación — capa fundacional — y
excede el umbral de 3 HU. Se construye en sub-slices: primero el store de
usuarios + API keys (HU-EVO-017/018, sin UI), después sesiones + 2FA
(HU-EVO-019/020, sin UI), y por último las dos pantallas (HU-EVO-021/022)
una vez los endpoints existen y están probados. `security-reviewer` corre
sobre esta épica antes de cualquier release que la incluya.

### Build harness — referencia cruzada

- `EP-EVO-004-SS1` (usuarios + API keys, sin UI) — openspec change `usuarios-postgresql-api-keys`, PR #48 mergeado.
- `EP-EVO-004-SS2` (sesiones + 2FA, sin UI) — openspec change `sesiones-2fa-totp`, PR #49 mergeado.
- `EP-EVO-004-SS3` (UI Team & Roles + Profile & Security, HU-EVO-021/022) — openspec change
  `team-roles-profile-security` (`openspec/changes/team-roles-profile-security/`), rama
  `feature/ep-evo-004-ss3-team-roles-profile-security`.

---

## Trazabilidad y Validación

### Bidireccional: Épicas ↔ Objetivos PRD

| Objetivo PRD | EP-EVO-001 | EP-EVO-002 | EP-EVO-003 |
|---|---|---|---|
| Obj. 1 (Desacople) | ✓ | — | — |
| Obj. 3 (Selección óptima) | ✓ | ✓ | — |
| Obj. 4 (Seguridad) | — | ✓ | ✓ |
| Obj. 5 (Compatibilidad) | — | — | ✓ |

### Cadena de Dependencias

```
EP-EVO-001 (Adapters)
    ↓
EP-EVO-002 (Aprendizaje de cuota)
    ↓
EP-EVO-003 (Dashboard + Alertas)
```

**Orden de construcción:** Slice 1 (EP-EVO-001), Slice 2 (EP-EVO-002), Slice 3 (EP-EVO-003). Ningún paralelismo: cada épica requiere la anterior.

### Riesgos Identificados

1. **EP-EVO-001:** Nuevo proveedor tiene bug en headers → LearnFromHeaders() parsea mal → Mitigación: AC en HU-EVO-003 valida parseo con mock servers
2. **EP-EVO-002:** 429 mid-stream (cuota agotada a mitad de generación) → No hay failover transparente → Mitigación: AC documenta comportamiento esperado (abort + error al cliente)
3. **EP-EVO-003:** UI React se desincroniza con backend → Métricas stale → Mitigación: refresh cada 5s, mostrar timestamp último update, warning si staleness > 1 min

---

## Siguiente: Discovery Iterativo

¿Estos 3 EP-EVO alinean con tu visión? ¿Ajustes antes de generar 15 HU-EVO con AC en G/W/T?

Opciones:
- **Seguir** → Genero HU-EVO por épica (15 historias con AC)
- **Refinar** → Cambios a los EP-EVO (scope, riesgo, dependencies)
- **Corregir** → Revisar algún aspecto específico
