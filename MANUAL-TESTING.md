# Manual Testing Guide — API LLM Gateway v1.0.0

Guía para validar la aplicación antes de taggear v1.0.0.

## Opción 1: Test Local (sin Docker)

### 1. Compilar

```bash
cd src
go build -o gateway ./cmd/gateway
```

### 2. Configurar API Keys (opcional)

```bash
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."
export GOOGLE_API_KEY="..."
export GATEWAY_AUTH_TOKEN="gateway-secret-key"
```

### 3. Ejecutar Gateway

```bash
./gateway
# Output esperado:
# INFO: Gateway running on http://localhost:8080
# INFO: Health check: /health
# INFO: Metrics: /metrics
```

### 4. Test Endpoints (en otra terminal)

#### 4a. Health check

```bash
curl http://localhost:8080/health
# Esperado: 200 OK
```

#### 4b. OpenAI chat endpoint

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer gateway-secret-key" \
  -d '{
    "model": "gpt-4",
    "messages": [
      {"role": "user", "content": "Hello, say hi back"}
    ],
    "max_tokens": 100
  }'
# Esperado: 200 OK + JSON con response
```

#### 4c. OpenAI streaming

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer gateway-secret-key" \
  -d '{
    "model": "gpt-4",
    "messages": [
      {"role": "user", "content": "Count to 5"}
    ],
    "stream": true
  }'
# Esperado: SSE stream con eventos data: [JSON] y data: [DONE]
```

#### 4d. OpenAI embeddings

```bash
curl -X POST http://localhost:8080/v1/embeddings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer gateway-secret-key" \
  -d '{
    "model": "text-embedding-3-small",
    "input": "Hello world"
  }'
# Esperado: 200 OK + JSON con data array de embeddings
```

#### 4e. Anthropic messages endpoint

```bash
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer gateway-secret-key" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [
      {"role": "user", "content": "Hello"}
    ],
    "max_tokens": 100
  }'
# Esperado: 200 OK + JSON en formato Anthropic
```

#### 4f. MCP discovery

```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer gateway-secret-key" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/list"
  }'
# Esperado: 200 OK + JSON-RPC response con tools
```

#### 4g. Error handling — missing auth

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hi"}]
  }'
# Esperado: 401 Unauthorized
```

#### 4h. Error handling — malformed payload

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer gateway-secret-key" \
  -d '{
    "model": "gpt-4"
  }'
# Esperado: 400 Bad Request (missing messages field)
```

---

## Opción 2: Test Docker

### 1. Construir imagen

```bash
docker build -t api-llm-gateway:v1.0.0 .
```

### 2. Configurar .env (opcional)

```bash
cat > .env <<EOF
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
GOOGLE_API_KEY=...
GATEWAY_AUTH_TOKEN=gateway-secret-key
EOF
```

### 3. Levantar contenedor

```bash
docker-compose up -d gateway
# Esperar ~5s para que inicie
docker logs -f api-llm-gateway
```

### 4. Ejecutar pruebas (igual que Opción 1, paso 4)

```bash
curl http://localhost:8080/health
```

### 5. Detener

```bash
docker-compose down
```

---

## Checklist de Validación

- [ ] `go build ./cmd/gateway` compila sin errores
- [ ] `./gateway` inicia sin panics
- [ ] `/health` retorna 200
- [ ] `/v1/chat/completions` con auth → 200
- [ ] `/v1/chat/completions` sin auth → 401
- [ ] `/v1/chat/completions` malformado → 400
- [ ] `/v1/chat/completions?stream=true` retorna SSE
- [ ] `/v1/embeddings` retorna embeddings
- [ ] `/v1/messages` (Anthropic) retorna en formato Anthropic
- [ ] `/mcp` retorna JSON-RPC response
- [ ] Docker build exitoso (si aplica)
- [ ] Docker container inicia sin errores (si aplica)

---

## Notas

- **Config**: `src/config.yaml` define providers, modelos, y routing
- **Auth**: `GATEWAY_AUTH_TOKEN` controla autenticación por Bearer token
- **Logs**: Salida a stdout en nivel INFO
- **Health**: `/health` siempre retorna 200 si el proceso corre
- **Metrics**: `/metrics` disponible para Prometheus

## Siguiente paso

Después de pasar todas las pruebas:

```bash
git tag v1.0.0 -m "Release v1.0.0 MVP — 9 épicas, 48 historias, fully tested"
git push origin v1.0.0
```
