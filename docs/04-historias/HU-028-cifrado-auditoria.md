---
id: HU-028
titulo: Cifrado de Sobre (Envelope) en BD de Auditoría
epica: EP-004B
prioridad: Should
complejidad: M
estado: lista
---

# Cifrado de Sobre (Envelope) en BD de Auditoría

Como **ingeniero de seguridad**, quiero **que los payloads almacenados en la base de datos de PostgreSQL estén cifrados del lado de la aplicación**, para **que un compromiso de la BD no exponga información sensible de los usuarios**.

Contexto: Client-Side Encryption usando KMS (Key Management Service). Si la redacción en memoria falla por un falso negativo, el dato queda cifrado y solo es legible con la llave KMS que no reside en la base de datos.

## Criterios de aceptación (Given/When/Then)

| # | Escenario | Given | When | Then |
|---|-----------|-------|------|------|
| 1 | Happy — inserción cifrada | Dado que una petición finalizada y lista para auditar | Cuando el background worker procesa el log | Entonces el worker cifra el payload en local con una DEK (Data Encryption Key) provista por KMS y luego inserta el texto cifrado en PostgreSQL |
| 2 | Edge — lectura autorizada | Dado que un admin quiere revisar el log | Cuando consulta el log específico en el Dashboard | Entonces el servicio del Dashboard llama a KMS para descifrar la DEK (si tiene el rol KMS_READER) y descifra el log localmente |
| 3 | Error — KMS inaccesible | Dado que el KMS externo está caído | Cuando el worker intenta cifrar | Entonces el payload se desecha y el log de auditoría se inserta indicando que el texto fue descartado por seguridad |
| 4 | Error — No autorizado | Dado que operador sin rol `KMS_READER` quiere leer | Cuando intenta visualizar el log | Entonces obtiene acceso denegado localmente sin llamar a KMS |
| 5 | Sad path — Fallo inserción DB | Dado que KMS generó la Data Key correctamente pero la inserción asíncrona en PostgreSQL falla | Cuando se intenta persistir | Entonces el log se guarda encriptado en un archivo local temporal (fallback) para reintento posterior |

## Checklist INVEST

- [x] Independent — Lógica de persistencia aislada de la API L7.
- [x] Negotiable — Algoritmo local (AES-GCM) y KMS provider (AWS KMS vs Vault) son intercambiables.
- [x] Valuable — Cumple normativas estrictas de seguridad (GDPR/HIPAA) al proteger audit logs.
- [x] Estimable — Patrón de arquitectura clásico de cifrado de sobre.
- [x] Small — Un middleware en el worker de base de datos.
- [x] Testable — Escribir log, hacer query directa a base de datos y verificar que el texto es ininteligible.

## Notas técnicas
- Asegurar alineación con NFRs de latencia y uso de caché si aplica.
