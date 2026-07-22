# Proposal: EP-003 Gobernanza de consumo (cuota y costo)

## 1. Por qué
La Gateway necesita llevar la cuenta de los tokens/requests consumidos por proveedor y clave para garantizar que no se excedan los límites predefinidos (cuotas), y además registrar el costo de estas peticiones para permitir visibilidad sobre el gasto financiero (cost tracking). Esto forma parte del Objetivo 3 (visibilidad de costos) y Objetivo 4 (gobernanza) del PRD.

## 2. Qué cambia
- Se implementará un **Quota Manager** (gestor de cuota) en memoria (síncrono/atómico) para autorizar peticiones antes de rutearlas, validando saldos por ventanas de tiempo (ej: diario).
- Se implementará la atribución de costo en base al modelo y al proveedor que finalmente responde la petición (después de cualquier failover potencial), integrando estos registros para consulta de métricas.

## 3. Capacidades (Nuevas o Afectadas)
- `quota-manager`: Control y bloqueo atómico de cuotas por proveedor/clave/ventana.
- `cost-tracker`: Atribución y registro de costo (tokens * tarifa) por request.

## 4. Impacto en Integraciones
- Se integra con la arquitectura de resiliencia (failover de EP-002), ya que la atribución de costo aplica solo al proveedor exitoso, mientras que la validación de cuota puede excluir proveedores antes del intento.
- La persistencia asíncrona depende del WAL (implementado en EP-009) para no bloquear la latencia del proxy con bases de datos relacionales en la ruta crítica.

## Trazabilidad
**Épica:** EP-003

**Historias:**
- HU-006: Contabilizar y limitar consumo por cuota
- HU-007: Registrar costo por petición, agente y proveedor
