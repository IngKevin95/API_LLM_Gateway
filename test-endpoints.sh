#!/bin/bash
# Automated test script for API LLM Gateway v1.0.0

set -e

BASE_URL="http://localhost:8080"
AUTH_TOKEN="gateway-secret-key"
PASSED=0
FAILED=0

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🧪 Testing API LLM Gateway v1.0.0"
echo "=================================="
echo ""

# Test 1: Health check
echo -n "Test 1: Health check... "
if curl -s "$BASE_URL/health" | grep -q "healthy\|ok"; then
  echo -e "${GREEN}✓ PASS${NC}"
  ((PASSED++))
else
  echo -e "${RED}✗ FAIL${NC}"
  ((FAILED++))
fi

# Test 2: OpenAI chat endpoint with auth
echo -n "Test 2: OpenAI /v1/chat/completions with auth... "
RESPONSE=$(curl -s -X POST "$BASE_URL/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "test"}],
    "max_tokens": 50
  }')

if echo "$RESPONSE" | grep -q "choices\|content"; then
  echo -e "${GREEN}✓ PASS${NC}"
  ((PASSED++))
else
  echo -e "${RED}✗ FAIL${NC}"
  echo "  Response: $RESPONSE"
  ((FAILED++))
fi

# Test 3: OpenAI chat without auth (should fail)
echo -n "Test 3: OpenAI /v1/chat/completions without auth (expect 401)... "
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "test"}]
  }')

if [ "$HTTP_CODE" = "401" ]; then
  echo -e "${GREEN}✓ PASS${NC}"
  ((PASSED++))
else
  echo -e "${RED}✗ FAIL (got $HTTP_CODE)${NC}"
  ((FAILED++))
fi

# Test 4: OpenAI chat malformed (missing messages)
echo -n "Test 4: OpenAI malformed payload (expect 400)... "
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -d '{"model": "gpt-4"}')

if [ "$HTTP_CODE" = "400" ]; then
  echo -e "${GREEN}✓ PASS${NC}"
  ((PASSED++))
else
  echo -e "${RED}✗ FAIL (got $HTTP_CODE)${NC}"
  ((FAILED++))
fi

# Test 5: OpenAI streaming
echo -n "Test 5: OpenAI /v1/chat/completions streaming... "
RESPONSE=$(curl -s -X POST "$BASE_URL/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "test"}],
    "stream": true
  }')

if echo "$RESPONSE" | grep -q "data:\|event:"; then
  echo -e "${GREEN}✓ PASS${NC}"
  ((PASSED++))
else
  echo -e "${RED}✗ FAIL${NC}"
  ((FAILED++))
fi

# Test 6: OpenAI embeddings
echo -n "Test 6: OpenAI /v1/embeddings... "
RESPONSE=$(curl -s -X POST "$BASE_URL/v1/embeddings" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -d '{
    "model": "text-embedding-3-small",
    "input": "test"
  }')

if echo "$RESPONSE" | grep -q "data\|embedding"; then
  echo -e "${GREEN}✓ PASS${NC}"
  ((PASSED++))
else
  echo -e "${RED}✗ FAIL${NC}"
  ((FAILED++))
fi

# Test 7: Anthropic messages
echo -n "Test 7: Anthropic /v1/messages... "
RESPONSE=$(curl -s -X POST "$BASE_URL/v1/messages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [{"role": "user", "content": "test"}],
    "max_tokens": 50
  }')

if echo "$RESPONSE" | grep -q "content\|type"; then
  echo -e "${GREEN}✓ PASS${NC}"
  ((PASSED++))
else
  echo -e "${RED}✗ FAIL${NC}"
  ((FAILED++))
fi

# Test 8: MCP discovery
echo -n "Test 8: MCP /mcp tools/list... "
RESPONSE=$(curl -s -X POST "$BASE_URL/mcp" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/list"
  }')

if echo "$RESPONSE" | grep -q "result\|jsonrpc"; then
  echo -e "${GREEN}✓ PASS${NC}"
  ((PASSED++))
else
  echo -e "${RED}✗ FAIL${NC}"
  ((FAILED++))
fi

# Summary
echo ""
echo "=================================="
echo -e "Results: ${GREEN}$PASSED passed${NC}, ${RED}$FAILED failed${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
  echo -e "${GREEN}✅ All tests passed!${NC}"
  exit 0
else
  echo -e "${RED}❌ Some tests failed.${NC}"
  exit 1
fi
