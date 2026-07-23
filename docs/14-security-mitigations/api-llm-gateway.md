# Mitigaciones de Seguridad (Zero Latency Impact)

Las siguientes estrategias permiten cerrar las brechas de seguridad del API LLM Gateway manteniendo la latencia y el rendimiento intactos (overhead < 50ms, cero red sincrónica).

## 1. Fuga de PII y Prompt Injection (Detección paralela)
**Brecha:** Payload sin escanear sincrónicamente.
**Solución: Escaneo Paralelo (Kill-Switch)**
- **Mecanismo:** El cliente envía el payload. El Gateway abre la conexión con el LLM externo inmediatamente, y **al mismo tiempo** (en una *goroutine* asíncrona) lanza un escáner heurístico local súper rápido (ej. detectores de PII en memoria).
- **Acción:** Si el escáner detecta un ataque 200ms después de iniciada la petición, el Gateway **aborta el stream TCP abruptamente** y devuelve un error de seguridad, cortando la fuga de datos en curso.
- **Impacto Rendimiento:** 0ms overhead inicial. El escáner usa CPU local concurrente, no red.

## 2. Rate Limiting Drift (Desfase Multi-Nodo)
**Brecha:** Nodos desincronizados al guardar cuotas en memoria RAM individual.
**Solución: Sticky Sessions por API Key en Load Balancer**
- **Mecanismo:** Configurar el Load Balancer (Nginx/HAProxy/ALB) para hacer enrutamiento basado en el hash del header `Authorization` (API Key).
- **Acción:** Todas las peticiones de la "Empresa A" caerán siempre en el "Nodo 1". El Nodo 1 tendrá la contabilidad de RAM exacta y atómica. Ningún nodo necesita comunicarse con otro.
- **Impacto Rendimiento:** 0ms overhead. El balanceador hace el hash a velocidad de red nativa.

## 3. Fugas en Base de Datos de Auditoría
**Brecha:** Fallos en la redacción guardan datos sensibles en PostgreSQL.
**Solución: Cifrado de Sobre (Envelope Encryption) simétrico con KMS** (contrato en PRD, arquitectura y tech-prd)
- **Mecanismo:** Antes de enviar el log a PostgreSQL, el worker asíncrono de Go cifra el campo `payload` con una **llave de datos AES (DEK) simétrica** generada en local (cifrado ultrarrápido). Esa DEK se **envuelve** (wrap) con la llave maestra (KEK) custodiada por el KMS; en la fila se persiste el ciphertext + la DEK envuelta. La KEK nunca sale del KMS.
- **Acción:** Si la redacción falla y un atacante roba la base de datos de PostgreSQL, solo verá payloads cifrados y DEKs envueltas inútiles sin acceso al KMS para desenvolverlas. El descifrado exige una llamada `Decrypt` al KMS (fuera del clúster de DB), auditada y revocable.
- **Impacto Rendimiento:** 0ms sobre la respuesta al usuario (el cifrado simétrico + wrap KMS lo hace un worker en background tras enviar la respuesta HTTP).

## 4. Agotamiento de Conexiones (Ataques Slowloris)
**Brecha:** Clientes abren miles de streams y no los cierran.
**Solución: Límite de Concurrencia + Idle Timeouts**
- **Mecanismo:** 
  1. En la RAM local se añade un contador atómico de *conexiones activas* por API Key (límite hardcodeado, ej. 10 streams paralelos).
  2. Se configuran *Idle Timeouts* agresivos a nivel TCP en Go (ej. si el cliente no lee/escribe en 2 segundos, se le mata la conexión).
- **Acción:** Si el atacante intenta abrir la conexión 11, se rechaza instantáneamente (429). Si abre 10 y se queda callado, se cierran en 2 segundos.
- **Impacto Rendimiento:** Validación atómica O(1) en RAM. Ahorra descriptores de archivo y previene caídas del servidor.
