# Configurar Free Claude Code contra API LLM Gateway

Esta guía permite a desarrolladores usar Free Claude Code (IDE extension) apuntando a la Gateway en lugar de conectarse directamente a Anthropic.

## AC1 — Configuración correcta

### Requisitos previos
- Gateway corriendo en `http://localhost:8080` (o URL remota accesible)
- API key válida de la Gateway (obtener en documentación de autenticación)

### Pasos de configuración

#### 1. Establecer `ANTHROPIC_BASE_URL`

En tu terminal o `.env` local:

```bash
export ANTHROPIC_BASE_URL=http://localhost:8080
export ANTHROPIC_API_KEY=your-gateway-api-key
```

#### 2. Verificar conectividad

```bash
curl -X POST http://localhost:8080/v1/messages \
  -H "Authorization: Bearer your-gateway-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [{"role": "user", "content": "test"}],
    "max_tokens": 100
  }'
```

Respuesta esperada: JSON con `content` array (formato Anthropic).

#### 3. Lanzar Claude Code

Con las variables de entorno configuradas, inicia Claude Code. Debería conectar a la Gateway automáticamente.

### Comportamiento esperado (AC1)

- Prompts se envían a la Gateway
- Gateway enruta al modelo seleccionado (OpenAI, Anthropic, etc.)
- Respuestas se reciben en formato Anthropic
- UX idéntica a conectar directamente a Anthropic

---

## AC2 — URL mal configurada

Si `ANTHROPIC_BASE_URL` apunta a una URL inválida:

```bash
export ANTHROPIC_BASE_URL=http://invalid-gateway:9999
```

### Error esperado

```
Error: Failed to connect to ANTHROPIC_BASE_URL (http://invalid-gateway:9999)
Details: connection refused or DNS resolution failed
```

### Resolución

1. Verificar que la Gateway esté corriendo: `curl http://localhost:8080/health`
2. Verificar URL sin typos
3. Si remoto: verificar conectividad de red y firewall

---

## AC3 — Credencial ausente

Si `ANTHROPIC_BASE_URL` es correcto pero sin API key:

```bash
export ANTHROPIC_BASE_URL=http://localhost:8080
# ⚠️ ANTHROPIC_API_KEY no establecida
```

### Error esperado

Cuando envíes un prompt:

```json
{
  "type": "error",
  "error": {
    "type": "authentication_error",
    "message": "Missing or invalid API key. Set ANTHROPIC_API_KEY environment variable."
  }
}
HTTP 401 Unauthorized
```

### Resolución

```bash
export ANTHROPIC_API_KEY=your-gateway-api-key
# Relanza Claude Code
```

---

## AC4 — Función no soportada

Si Claude Code intenta usar una función Anthropic no cubierta por HU-013:

### Error esperado

```json
{
  "type": "error",
  "error": {
    "type": "not_implemented_error",
    "message": "Feature not yet supported by Gateway. Check https://github.com/IngKevin95/API_LLM_Gateway/issues for status."
  }
}
HTTP 501 Not Implemented
```

### Soporte actual (HU-013 cobertura)

✅ Messages API (chat)  
✅ Streaming (`stream: true`)  
✅ Tool use (`tools` array)  
❌ Vision (image inputs) — Fase 2  
❌ File inputs — Fase 2  
❌ Batch API — Fase 3  

---

## Troubleshooting

| Síntoma | Causa | Solución |
|---------|-------|----------|
| "Connection refused" | Gateway no levantada | `docker run ...` o `go run ./src/cmd/gateway` |
| "401 Unauthorized" | API key inválida/ausente | `echo $ANTHROPIC_API_KEY` → reestablecerla |
| "502 Bad Gateway" | Gateway error interno | Revisar logs: `docker logs gateway-container` |
| "Prompt truncated" | Límite de tokens | Revisar AC de longitud máxima en documentación |

---

## Verificación end-to-end

Ejecutar este script para validar la configuración:

```bash
#!/bin/bash
set -e

echo "1. Verificando conectividad..."
curl -s http://localhost:8080/health || exit 1

echo "2. Verificando API key..."
[[ -z "$ANTHROPIC_API_KEY" ]] && { echo "ANTHROPIC_API_KEY no establecida"; exit 1; }

echo "3. Enviando test request..."
RESPONSE=$(curl -s -X POST http://localhost:8080/v1/messages \
  -H "Authorization: Bearer $ANTHROPIC_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-haiku",
    "messages": [{"role": "user", "content": "Hello"}],
    "max_tokens": 50
  }')

echo "Respuesta: $RESPONSE"
echo "✓ Configuración válida. Lanza Claude Code."
```

---

## Notas de seguridad

⚠️ **Nunca commits tu `ANTHROPIC_API_KEY` a Git.** Usar `.env.local` o variables de entorno.

```bash
# ✅ Seguro
export ANTHROPIC_API_KEY="secret-key"

# ❌ Inseguro
echo 'ANTHROPIC_API_KEY="secret-key"' > .env && git add .env
```

---

## Soporte

- Documentación técnica: `/docs/13-tech-prd/`
- Issues: https://github.com/IngKevin95/API_LLM_Gateway/issues
- Logs de Gateway: `docker logs -f gateway-container`
