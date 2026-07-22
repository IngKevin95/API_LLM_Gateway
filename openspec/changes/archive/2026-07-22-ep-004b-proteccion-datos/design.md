# Design: Protección de Datos (EP-004B)

## Arquitectura

La épica introduce 4 pilares en la capa de protección de datos:

1. **Gestor de Secretos (SecretManager)**:
   - Abstracción que resuelve credenciales (ej. `api_key: ${OPENAI_KEY}`).
   - Provee un método para recargar (rotación en caliente) sin reiniciar el gateway.
   - Enmascara los secretos en logs.

2. **Motor DLP (Data Loss Prevention)**:
   - **Síncrono (Redactor)**: Middleware que inspecciona el payload de entrada, aplica expresiones regulares (O(N)) y censura datos sensibles (ej. emails, tarjetas de crédito) sustituyéndolos por `***`. Timeout máximo 50ms para evitar degradación.
   - **Asíncrono (Kill-Switch)**: Goroutine secundaria que inspecciona el payload en tránsito (streaming). Si detecta PII grave de forma tardía, interrumpe el socket TCP subyacente.

3. **Auditoría Inmutable & KMS (Client-Side Encryption)**:
   - El Sync Worker (HU-038 implementada en EP-009) se extiende. Antes de flushear el AuditLog a la base de datos (PostgreSQL), un middleware de cifrado solicita una Data Encryption Key (DEK) al KMS (local o AWS) y cifra el campo `payload` usando AES-GCM.
   - La tabla en BD implementará triggers `append-only` para rechazar operaciones UPDATE o DELETE (exceptuando la política de purga particionada de 30 días).

4. **Protección TCP (Slowloris)**:
   - A nivel del servidor base `http.Server`, se configura estrictamente el `ReadHeaderTimeout` y `WriteTimeout` según la directriz del YAML (por defecto ej. 5s y 30s respectivamente) cerrando automáticamente sockets que no envían datos a un ritmo razonable.

## Interfaces y Estructuras (Go)

```go
// internal/secrets/manager.go
type SecretManager interface {
    Resolve(value string) (string, error)
    Reload() error
}

// internal/dlp/engine.go
type DLPEngine interface {
    Redact(payload []byte) ([]byte, error)
    ScanAsync(ctx context.Context, payloadStream io.Reader, cancelFunc context.CancelFunc)
}

// internal/audit/kms.go
type KMSClient interface {
    GenerateDataKey(ctx context.Context) (DEK, error)
    DecryptDataKey(ctx context.Context, encryptedDEK []byte) (DEK, error)
}
```

## Flujo (Sequence)

1. Cliente inicia petición (HTTP/TCP).
2. Servidor base valida `ReadHeaderTimeout` (Slowloris).
3. Middleware de DLP inspecciona el request body sincrónicamente.
4. El enrutador invoca a `SecretManager` para resolver la API Key real.
5. El gateway llama al proveedor upstream (con DLP Kill-switch asíncrono leyendo la respuesta).
6. Al finalizar, se encola evento en Sync Worker.
7. El Sync Worker invoca al `KMSClient` para cifrar la data.
8. Sync Worker flushea a PostgreSQL (tabla append-only particionada).

## Trade-offs
- **DLP Asíncrono**: Minimiza impacto en latencia pero introduce riesgo teórico de que algunos bytes escapen antes de que el kill-switch interrumpa el socket si el procesamiento es lento.
- **Client-Side Encryption**: Aumenta el uso de CPU local en el worker de auditoría por cada evento procesado, pero elimina por completo la exposición en caso de inyección SQL o volcado de base de datos.
