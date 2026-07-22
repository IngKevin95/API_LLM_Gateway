# Tasks: Protección de Datos (EP-004B)

## Preparación (Config y Tests)
- [ ] 1. Configurar timeouts en el servidor HTTP base (`ReadHeaderTimeout`, `WriteTimeout`) - HU-034.
- [ ] 2. Crear pruebas TDD iniciales para la resolución de variables de entorno y fallos de secretos ausentes - HU-011.

## Fase 1: SecretManager y Servidor (Sub-slice 1)
- [ ] 3. Implementar `internal/secrets/manager.go` que resuelva strings estilo `${VAR_NAME}` desde el entorno.
- [ ] 4. Implementar soporte para rotación (función `Reload()`) sin reiniciar servidor.
- [ ] 5. Ajustar el factory de providers para validar dependencias de secretos al arranque y en recargas.

## Fase 2: Motor DLP Síncrono (Sub-slice 2)
- [ ] 6. Implementar `internal/dlp/engine.go` con análisis de expresiones regulares básicas (emails, tarjetas).
- [ ] 7. Inyectar middleware `dlp.Middleware` antes del router que modifique el payload en vuelo (redacción).
- [ ] 8. Asegurar límite de tiempo (< 50ms) o devolver HTTP 500 en la redacción síncrona.

## Fase 3: Auditoría y Cifrado KMS (Sub-slice 3)
- [ ] 9. Crear esquema de PostgreSQL en `sql/04_auditlog.sql` con tabla `AuditLog` particionada y trigger para evitar UPDATE/DELETE.
- [ ] 10. Implementar abstracción `internal/audit/kms.go` (Client-Side Encryption en memoria).
- [ ] 11. Adaptar el Sync Worker para que cifre la carga útil con KMS antes de insertar en la base de datos.

## Fase 4: DLP Asíncrono (Sub-slice 4)
- [ ] 12. Extender el engine DLP para permitir lectura streaming asíncrona.
- [ ] 13. Modificar el adaptador upstream para envolver el body response y ejecutar kill-switch (cancel context) si detecta PII en vuelo.
- [ ] 14. Realizar pruebas integrales para validar que la conexión TCP de cliente se corta abruptamente.

## Cierre y Documentación
- [ ] 15. Actualizar la documentación Swagger/OpenAPI con los códigos de error relacionados con DLP y timeouts (HTTP 408, 500).
- [ ] 16. Generar el PR de EP-004B completamente documentado.
