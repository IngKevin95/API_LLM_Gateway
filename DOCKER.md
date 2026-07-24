# Docker Deployment

Dos servicios independientes, misma red `gateway-network`.

## Startup

```bash
# 1. OmniRoute (proveedor local, gratuito)
docker-compose -f docker-compose.omniroute.yml up -d

# 2. Gateway (orquestador)
docker-compose up -d

# Verificar
docker ps
docker network ls
```

## URLs

- **Gateway**: `http://localhost:8080` (health: `/health`, API: `/v1/...`)
- **OmniRoute**: `http://localhost:20128` (dashboard + API `/v1`)

## Shutdown

```bash
docker-compose down
docker-compose -f docker-compose.omniroute.yml down
```

## Logs

```bash
docker logs api-llm-gateway -f
docker logs omniroute-provider -f
```

## Networking

Ambos servicios están en `gateway-network` (bridge).
- Gateway alcanza OmniRoute en `http://omniroute:20128/v1` (por DNS interno)
- Desde host: `localhost:8080`, `localhost:20128`

## Env vars

`.env` debe existir (ya lo creaste):
```
ANTHROPIC_API_KEY=...
OPENAI_API_KEY=...
GOOGLE_API_KEY=...
OPENROUTER_API_KEY=...
AIHUBMIX_API_KEY=...
```
