# browser-notifications Specification

## Purpose
TBD - created by archiving change ui-react-dashboard-notificaciones. Update Purpose after archive.
## Requirements
### Requirement: Notificación toast ante cuota baja
El dashboard SHALL mostrar un toast (esquina inferior derecha, auto-close) cuando `GET /alerts`
devuelve una alerta nueva, con severidad `warning` mostrando el mensaje del proveedor y su
remaining%, y desapareciendo automáticamente a los 5 segundos.
(Traza: HU-EVO-015)

#### Scenario: Toast ante cuota baja
- **WHEN** el remaining de un proveedor cae por debajo del umbral configurado y `GET /alerts` lo refleja
- **THEN** el dashboard muestra un toast en la esquina inferior derecha con el mensaje del proveedor, con fade-out a los 5s

### Requirement: Notificación sonora en alertas críticas
El dashboard SHALL reproducir un sonido corto cuando se genera una alerta `severity=critical` y el
usuario tiene el sonido habilitado en su configuración local.
(Traza: HU-EVO-015)

#### Scenario: Sonido en alerta crítica
- **WHEN** el usuario tiene sonido habilitado y se genera una alerta `critical`
- **THEN** se reproduce un beep corto

### Requirement: Notificación del sistema operativo vía Notification API
El dashboard SHALL solicitar permiso de notificaciones del navegador mediante un gesto explícito
del usuario, y SHALL mostrar una notificación del sistema (fuera del navegador) cuando se genera
una alerta `critical` y el permiso fue otorgado.
(Traza: HU-EVO-015)

#### Scenario: Notificación del sistema en alerta crítica
- **WHEN** el usuario otorgó permiso de notificaciones y se genera una alerta `critical`
- **THEN** el navegador dispara `Notification.requestPermission()`-gated y muestra una notificación del sistema operativo

### Requirement: Deduplicación de toasts activos
El dashboard SHALL NOT duplicar un toast para la misma combinación proveedor+modelo si ya existe
uno visible dentro de una ventana de 10 segundos; en su lugar, SHALL actualizar el toast existente.
(Traza: HU-EVO-015)

#### Scenario: Deduplicación de toast repetido
- **WHEN** ya hay un toast visible de Groq y llega otra alerta de Groq antes de 10 segundos
- **THEN** no se crea un segundo toast; el toast existente se actualiza

### Requirement: Umbral de notificación configurable sin redeploy
El dashboard SHALL permitir configurar el umbral de notificación (por defecto 10%) desde una
pantalla de configuración persistida en `localStorage`, sin requerir redeploy del dashboard.
(Traza: HU-EVO-015)

#### Scenario: Cambiar umbral de notificación
- **WHEN** el usuario edita el umbral a 25% en settings
- **THEN** las siguientes evaluaciones de alertas usan 25% como umbral de notificación, sin recompilar ni redesplegar el dashboard

