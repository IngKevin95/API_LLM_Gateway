# Flujo Core de Petición (API LLM Gateway)

Este flujo describe el ciclo de vida completo de una petición que pasa por los componentes documentados en la arquitectura C4 (Capa Determinista y Capa I/O-Bound), asegurando alineación con los límites de contenedores y los Criterios de Aceptación (HU-006, HU-026, HU-032, etc.).

```mermaid
sequenceDiagram
    autonumber
    actor C as Cliente
    participant LB as Load Balancer L7
    box Dominio Gateway (Capa Determinista RAM)
        participant A as Auth & Rate Limit
        participant Q as Quota Manager
        participant S as Scanner Síncrono
        participant SC as Semantic Cache
        participant R as Model Router
    end
    box Dominio Gateway (Capa I/O-Bound)
        participant E as Embedding Engine
        participant AS as Scanner Asíncrono
        participant F as Failover Engine
    end
    participant V as Vector DB
    participant LLM as Proveedor LLM (OpenAI/etc)
    participant W as Sync Worker

    C->>LB: POST /v1/chat/completions (API Key)
    LB->>A: Enruta nodo sticky (Hash de Key)
    A->>A: Valida JWT/Key (Memoria)
    A->>Q: Verifica Rate Limit y Cuota (Memoria)
    alt Límite excedido / Cache Miss Cuota
        Q-->>C: 429 Too Many Requests (Fail-fast)
        Q-)W: Encola hidratación asíncrona
    end
    Q->>S: Pasa payload a redacción
    S->>S: Redacta PII (CGO/Regex <10ms)
    S-)AS: Dispara escaneo Base64 paralelo
    S->>SC: Pasa payload redactado
    SC->>E: Pide vector de embedding
    E-->>SC: Retorna vector (ONNX local)
    SC->>V: Busca similitud vectorial
    alt Similitud > 98% (Cache Hit)
        V-->>SC: Retorna respuesta cacheada
        SC-->>C: Retorna respuesta (Costo $0)
    else Similitud baja (Cache Miss)
        V-->>SC: No hay coincidencia exacta
        SC->>R: Continúa flujo
        R->>R: Calcula tokens y scores
        R->>F: Genera cadena de failover
        F->>LLM: Llama al LLM primario vía Adapter
        alt PII detectado en Base64 por AS
            AS-->>F: Context Cancel (Kill Switch)
            F-->>C: 400 Bad Request (Fuga prevenida)
        else Respuesta exitosa
            LLM-->>F: Retorna Completions (Streaming)
            F-->>C: Transmite JSON / Stream al cliente
            F-)W: Emite evento de auditoría/costo
        end
    end
```
