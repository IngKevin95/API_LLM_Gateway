# Getting Started with API LLM Gateway

## Overview

API LLM Gateway provides a universal interface to multiple LLM providers:
- **OpenAI** (GPT-3.5, GPT-4, etc.)
- **Anthropic** (Claude family)
- **Google** (Gemini)
- **OpenRouter** (aggregated providers)
- **Local models** (Ollama, vLLM, LM Studio)

## Key Concepts

### Automatic Capability Routing
Instead of specifying a model name, specify a **capability**:
```json
{
  "model": "router:chat",      // Automatic model selection
  "messages": [...],
  "max_tokens": 1024
}
```

Supported capabilities:
- `router:chat` — conversational AI
- `router:vision` — image understanding
- `router:embedding` — text embeddings
- `router:reasoning` — extended thinking

### Universal Request Format
All requests use a common format, translated internally for each provider:
```json
{
  "messages": [{"role": "user", "content": "hello"}],
  "model": "gpt-4",             // or "router:chat"
  "max_tokens": 1024,
  "temperature": 0.7
}
```

### Automatic Parameter Translation
- OpenAI temperature range [0, 2] → Anthropic [0, 1] (auto-clamped)
- Unsupported parameters filtered silently
- Required fields enforced per provider (e.g., max_tokens for Anthropic)

## Quick Start

### 1. Set Up Environment
```bash
export GATEWAY_URL=http://localhost:8080
export OPENAI_API_KEY=sk-...
export ANTHROPIC_API_KEY=sk-ant-...
```

### 2. Make Your First Request

**Using Python + requests:**
```python
import requests

response = requests.post(
    "http://localhost:8080/v1/chat/completions",
    json={
        "model": "gpt-4",
        "messages": [{"role": "user", "content": "Say hello"}],
        "max_tokens": 100
    }
)
print(response.json())
```

**Using cURL:**
```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Say hello"}],
    "max_tokens": 100
  }'
```

### 3. Try Capability-Based Routing
```bash
# Let the gateway choose the best chat model
curl -X POST http://localhost:8080/responses \
  -H "Content-Type: application/json" \
  -d '{
    "model": "router:chat",
    "messages": [{"role": "user", "content": "Say hello"}],
    "max_tokens": 100
  }'
```

## Next Steps

- [OpenAI SDK Setup](02-openai-sdk-setup.md) — Use official OpenAI Python/Node.js SDK
- [Anthropic SDK Setup](03-anthropic-sdk-setup.md) — Use official Anthropic SDK
- [HTTP Client Setup](04-http-client-setup.md) — Raw HTTP requests
- [Migration Guide](05-migration-guide.md) — From direct provider to gateway
- [Best Practices](06-best-practices.md) — Error handling, caching, retries

## Endpoints

| Endpoint | Purpose | Format |
|----------|---------|--------|
| `POST /v1/chat/completions` | Chat completions (OpenAI-compatible) | OpenAI/Anthropic/Universal |
| `POST /responses` | Universal format (recommended) | Normalized JSON |
| `GET /v1/models` | List available models | Query params: ?capability=chat, ?provider=openai |

## Troubleshooting

**"Model not found"**
- Use `GET /v1/models` to list available models
- Ensure model name is spelled correctly
- Check provider credentials in environment variables

**"max_tokens is required"**
- Anthropic requires max_tokens in all requests
- Gateway enforces this and returns 400 if missing

**"Provider unavailable"**
- Gateway automatically falls back to next provider
- Check `GET /v1/models` to see availability

See [Troubleshooting Guide](08-troubleshooting.md) for more issues.
