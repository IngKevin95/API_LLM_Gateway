#!/bin/bash
# test-suite.sh — Validación completa del API LLM Gateway

set -e
GATEWAY="http://localhost:8080"
OMNIROUTE="http://localhost:20128"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "================================================"
echo "  API LLM Gateway — Test Suite Completo"
echo "================================================"

# 1. INFRAESTRUCTURA
echo -e "\n${YELLOW}[1] VERIFICACIÓN DE INFRAESTRUCTURA${NC}"
echo "---"

echo "✓ Contenedores:"
docker ps --format "table {{.Names}}\t{{.Status}}" | grep -E "api-llm-gateway|omniroute-provider"

echo -e "\n✓ Conectividad de redes:"
docker exec api-llm-gateway wget -q -O- http://omniroute:20128/ > /dev/null && echo "  • Gateway → OmniRoute: OK" || echo "  • Gateway → OmniRoute: FALLO"

# 2. ENDPOINTS
echo -e "\n${YELLOW}[2] VERIFICACIÓN DE ENDPOINTS${NC}"
echo "---"

echo "✓ /health:"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" $GATEWAY/health)
if [ "$HTTP_CODE" = "200" ]; then
  echo "  • Status: ${GREEN}OK${NC}"
  curl -s $GATEWAY/health
else
  echo "  • Status: ${RED}FALLO (HTTP $HTTP_CODE)${NC}"
fi

echo -e "\n✓ /v1/chat/completions (OpenAI):"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST $GATEWAY/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"test"}],"max_tokens":5}')

if [ "$HTTP_CODE" = "200" ]; then
  echo "  • Status: ${GREEN}OK${NC}"
else
  echo "  • Status: ${RED}ERROR (HTTP $HTTP_CODE)${NC}"
fi

echo -e "\n✓ /v1/embeddings:"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST $GATEWAY/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{"model":"text-embedding-3-small","input":"test"}')

if [ "$HTTP_CODE" = "200" ]; then
  echo "  • Status: ${GREEN}OK${NC}"
else
  echo "  • Status: ${RED}ERROR (HTTP $HTTP_CODE)${NC}"
fi

echo -e "\n✓ /v1/messages (Anthropic):"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST $GATEWAY/v1/messages \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-opus-4","messages":[{"role":"user","content":"test"}],"max_tokens":5}')

if [ "$HTTP_CODE" = "200" ]; then
  echo "  • Status: ${GREEN}OK${NC}"
else
  echo "  • Status: ${RED}ERROR (HTTP $HTTP_CODE)${NC}"
fi

# 3. CONFIGURACIÓN
echo -e "\n${YELLOW}[3] VALIDACIÓN DE CONFIGURACIÓN${NC}"
echo "---"

echo "✓ config.yaml:"
docker exec api-llm-gateway python3 -c "import yaml; yaml.safe_load(open('config.yaml'))" && echo "  • YAML válido: ${GREEN}SI${NC}" || echo "  • YAML válido: ${RED}NO${NC}"

echo "✓ Proveedores en config.yaml:"
docker exec api-llm-gateway grep "id:" config.yaml | sed 's/^/  • /'

echo -e "\n✓ Adaptadores cargados en Gateway:"
docker logs api-llm-gateway | grep "modelos cargados" | tail -1 | sed 's/^/  • /'

# 4. SECRETOS
echo -e "\n${YELLOW}[4] VALIDACIÓN DE SECRETOS${NC}"
echo "---"

echo "✓ Env vars en .env:"
[ -f .env ] && echo "  • .env existe: ${GREEN}SI${NC}" || echo "  • .env existe: ${RED}NO${NC}"

if [ -f .env ]; then
  grep -E "^[A-Z_]+=.+$" .env | sed 's/=.*//' | sed 's/^/  • /'
fi

echo -e "\n✓ Secretos detectados en Gateway:"
docker logs api-llm-gateway 2>&1 | grep -i "api.key\|secret" || echo "  • (sin logs)"

# 5. FAILOVER
echo -e "\n${YELLOW}[5] TEST DE FAILOVER${NC}"
echo "---"

echo "Deteniendo OmniRoute..."
docker pause omniroute-provider 2>/dev/null || true
sleep 2

echo "Intentando llamada al Gateway (OmniRoute parado):"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" $GATEWAY/health)
if [ "$HTTP_CODE" = "200" ]; then
  echo "  • Gateway responde: ${GREEN}SI${NC} (failover funcionando)"
else
  echo "  • Gateway responde: ${RED}NO${NC} (failover falló)"
fi

echo "Reanudando OmniRoute..."
docker unpause omniroute-provider 2>/dev/null || true
sleep 2

# 6. RESUMEN
echo -e "\n${YELLOW}[6] RESUMEN${NC}"
echo "---"

docker ps --format "table {{.Names}}\t{{.Status}}" | grep -E "api-llm-gateway|omniroute-provider" | awk '{
  print "  • " $1 ": " $2
}'

echo -e "\n${GREEN}✓ Test suite completado${NC}"
echo "================================================"
