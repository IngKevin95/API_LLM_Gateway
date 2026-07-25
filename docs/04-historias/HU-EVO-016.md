---
id: HU-EVO-016
title: Migrar Quota Manager a PostgreSQL Asíncrono
epic: EP-EVO-004
type: evolutivo
status: pending
---

# HU-EVO-016: Migrar Quota Manager a PostgreSQL Asíncrono

**Como** Gateway Core
**Quiero** delegar la persistencia de cuotas y saldos de memoria RAM hacia PostgreSQL usando un Sync Worker asíncrono
**Para** asegurar que los rate limits y cuotas de los clientes no se pierdan ante reinicios, manteniendo 0ms de latencia de red en la ruta crítica del request.

## Criterios de Aceptación

**Escenario 1: Deducción de cuota asíncrona exitosa**
- **Given** que el Gateway recibe un request válido y debita 1500 tokens de la caché L1 en RAM
- **When** el Sync Worker ejecuta su ciclo de flasheo cada 5 segundos
- **Then** se debe emitir un query `UPDATE quotas SET tokens = tokens - 1500` hacia PostgreSQL
- **And** la latencia HTTP de la respuesta al cliente no se ve afectada por esta inserción.

**Escenario 2: Recuperación desde Write-Ahead Log (WAL)**
- **Given** que el Gateway sufre una interrupción abrupta (crash) con transacciones en memoria pendientes
- **When** el nodo vuelve a arrancar
- **Then** debe leer el archivo WAL local y reproducir las deducciones hacia PostgreSQL antes de aceptar nuevo tráfico.
