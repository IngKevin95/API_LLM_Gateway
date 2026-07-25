---
id: HU-EVO-019
title: Interfaz de Gestión de Tokens y Cuotas de Proveedores
epic: EP-EVO-005
type: evolutivo
status: pending
---

# HU-EVO-019: Interfaz de Gestión de Tokens y Cuotas

**Como** Administrador de la Plataforma
**Quiero** visualizar las cuotas por modelo y poder añadir/revocar tokens de proveedores desde el Dashboard
**Para** gestionar los recursos del API Gateway de forma 100% autónoma, sin depender del despliegue de archivos YAML.

## Criterios de Aceptación

**Escenario 1: Inyección Segura de Tokens (Credentials Manager)**
- **Given** que el administrador ingresa un nuevo token para el proveedor OpenAI
- **When** guarda el formulario en el Dashboard
- **Then** el token viaja al Admin API
- **And** se cifra con KMS (Envelope Encryption) antes de persistirse en PostgreSQL
- **And** jamás es retornable o visible en texto plano en la UI posteriormente.

**Escenario 2: Visualización de Gasto y Cuota por Modelo**
- **Given** que existen varios modelos consumiendo saldo (GPT-4o, Claude 3.5 Sonnet)
- **When** el administrador navega a la sección de "Consumo"
- **Then** debe visualizar gráficos que muestren el costo total (USD) y la cuota consumida desagregada por modelo en tiempo real.
