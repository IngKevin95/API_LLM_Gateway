# Proposal: Protección de Datos (EP-004B)

## Por qué
El API Gateway procesa peticiones y respuestas que pueden contener información sensible y PII (Personal Identifiable Information). Además de proteger el acceso, es indispensable garantizar que:
1. Las operaciones sean inmutables y auditables (HU-010).
2. Los datos persistidos estén cifrados de manera segura (HU-028).
3. Los secretos no se expongan ni accidental ni estáticamente (HU-011).
4. El PII se oculte en memoria antes de ir a terceros (HU-026a).
5. Se interrumpan transmisiones que revelen PII de manera asíncrona (HU-026b).
6. Los ataques TCP de agotamiento de recursos (Slowloris) sean mitigados (HU-034).

Esta épica asegura el cumplimiento regulatorio, mitiga el riesgo de filtraciones y robustece la resiliencia en la capa de transporte.

## Qué cambia
Se añadirán componentes independientes integrados en el core:
- **Cifrado KMS e Inmutabilidad**: Expansión del Sync Worker para integrar Client-Side Encryption de los registros de auditoría y políticas restrictivas sobre la base de datos subyacente.
- **Gestor de Secretos**: Refactorización de la inyección de credenciales para utilizar resolución dinámica por variable de entorno o bóveda de secretos con soporte para rotación en caliente.
- **DLP y Redacción (Síncrono/Asíncrono)**: Motores de regex/procesamiento rápido para redacción (en-vuelo) e interrupción asincrónica (Kill-Switch) conectados en el flujo de I/O de la petición.
- **Defensas de Transporte TCP**: Endurecimiento del servidor HTTP base configurando timeouts agresivos contra clientes lentos.

## Capacidades (Capacities) afectadas / nuevas
- `audit-trail` (Nueva): Registro inmutable y cifrado (Client-Side KMS) de las transacciones.
- `secret-manager` (Nueva): Resolución y rotación segura de credenciales.
- `dlp-engine` (Nueva): Redactor síncrono y Kill-Switch asíncrono.
- `tcp-shield` (Nueva): Timeouts preventivos en el socket (ReadHeaderTimeout).

## Impacto
Estos cambios afectarán la pipeline de middlewares y configuraciones core. Aunque se espera impacto mínimo en latencia gracias a la asincronía del Kill-Switch y el DLP `< 10ms`, se requiere una configuración fina para evitar degradación. 

## Trazabilidad
Épica: EP-004B
Historias: HU-010, HU-011, HU-026a, HU-026b, HU-028, HU-034
