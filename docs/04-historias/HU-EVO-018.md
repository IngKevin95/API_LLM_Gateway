---
id: HU-EVO-018
title: Integración SSO para Dashboard
epic: EP-EVO-005
type: evolutivo
status: pending
---

# HU-EVO-018: Integración SSO para Dashboard

**Como** Administrador de la Plataforma
**Quiero** acceder al Dashboard autenticándome a través del Identity Provider de mi organización (OAuth2)
**Para** evitar la gestión local de contraseñas y mantener la seguridad bajo estándares corporativos.

## Criterios de Aceptación

**Escenario 1: Login exitoso con OAuth2**
- **Given** que un administrador navega al UI del Dashboard sin sesión
- **When** hace clic en "Login with SSO" e ingresa credenciales válidas en el Identity Provider
- **Then** el sistema lo redirige al Dashboard
- **And** emite un JWT firmado que autoriza acceso total a las APIs de administración (Admin API).

**Escenario 2: Bloqueo de acceso sin privilegios**
- **Given** que un empleado autenticado vía SSO no tiene el rol de `gateway_admin`
- **When** intenta acceder al panel del Gateway
- **Then** debe ser bloqueado con un mensaje 403 Forbidden.
