# Consolidated Documentation: Best Practices, Troubleshooting, Performance, Setup & Deployment

## 06 - Best Practices Guide

### Caching Strategies
Cache responses to reduce costs and latency:

```python
from functools import lru_cache
import hashlib

@lru_cache(maxsize=10000)
def chat_completion_cached(model: str, user_message: str) -> str:
    response = client.chat.completions.create(
        model=model,
        messages=[{"role": "user", "content": user_message}],
        max_tokens=1024
    )
    return response.choices[0].message.content

# Use Redis for multi-process caching
import redis
cache = redis.Redis()

def cached_request(model, message):
    key = f"chat:{hashlib.md5(message.encode()).hexdigest()}"
    cached = cache.get(key)
    if cached:
        return cached.decode()
    
    response = client.chat.completions.create(
        model=model,
        messages=[{"role": "user", "content": message}],
        max_tokens=1024
    )
    result = response.choices[0].message.content
    cache.setex(key, 3600, result)  # Cache 1 hour
    return result
```

### Retry Logic with Exponential Backoff

```python
import time
from openai import APIError, RateLimitError

def retry_with_backoff(fn, max_retries=3):
    for attempt in range(max_retries):
        try:
            return fn()
        except RateLimitError as e:
            wait_time = 2 ** attempt
            print(f"Rate limited, waiting {wait_time}s...")
            time.sleep(wait_time)
        except APIError as e:
            if e.status_code >= 500:  # Server error
                wait_time = 2 ** attempt
                print(f"Server error, retrying in {wait_time}s...")
                time.sleep(wait_time)
            else:
                raise
    raise Exception("Max retries exceeded")

result = retry_with_backoff(
    lambda: client.chat.completions.create(
        model="router:chat",
        messages=[...],
        max_tokens=1024
    )
)
```

### Error Handling Patterns

```python
from openai import APIError, APIStatusError, RateLimitError

def robust_completion(messages):
    try:
        response = client.chat.completions.create(
            model="router:chat",
            messages=messages,
            max_tokens=1024
        )
        return response.choices[0].message.content
    except RateLimitError:
        return "Rate limited - please try again later"
    except APIStatusError as e:
        if e.status_code == 404:
            return "Model not found"
        elif e.status_code == 503:
            return "Gateway overloaded - trying again..."
        else:
            return f"Error: {e.message}"
    except APIError as e:
        return f"Unexpected error: {str(e)}"
```

### Cost Optimization

```python
# Use cheaper models for simple tasks
def get_model(task_complexity):
    if task_complexity == "simple":
        return "gpt-3.5-turbo"  # Cheaper
    elif task_complexity == "complex":
        return "gpt-4"  # More capable
    else:
        return "router:chat"  # Auto-select

# Monitor cost per request
def log_cost(response):
    usage = response.usage
    # Assuming costs from /v1/models metadata
    cost = (usage.prompt_tokens * 0.0015 + usage.completion_tokens * 0.002) / 1000
    print(f"Request cost: ${cost:.6f}")
```

### Rate Limiting Awareness

```python
# Respect rate limits
def rate_limited_requests(messages_list, max_requests_per_minute=60):
    interval = 60 / max_requests_per_minute
    for i, messages in enumerate(messages_list):
        if i > 0:
            time.sleep(interval)
        response = client.chat.completions.create(
            model="router:chat",
            messages=messages,
            max_tokens=1024
        )
        yield response
```

---

## 08 - Troubleshooting Guide

| Issue | Cause | Solution |
|-------|-------|----------|
| "Model not found" | Model doesn't exist or provider offline | Check `GET /v1/models`, use `router:chat` |
| "max_tokens is required" | Anthropic requirement | Always pass `max_tokens` |
| "Provider unavailable" | Provider API down | Gateway retries fallback chain |
| "Rate limit exceeded" | Too many requests | Implement backoff, check limits |
| "Authentication failed" | Invalid API key | Verify credentials in gateway config |
| "Parameter validation error" | Invalid param value | Check param ranges per provider |
| "Streaming timeout" | Long response time | Increase client timeout |
| "Connection refused" | Gateway not running | Start gateway: `make dev` |

**Quick fixes:**
```bash
# Check gateway health
curl http://localhost:8080/health

# List available models
curl http://localhost:8080/v1/models

# Test direct request
curl -X POST http://localhost:8080/responses \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4","messages":[{"role":"user","content":"test"}],"max_tokens":10}'
```

---

## 09 - Performance Tuning Guide

### Latency Benchmarks
- OpenAI GPT-3.5: ~500ms (prompt → first token)
- OpenAI GPT-4: ~800ms
- Anthropic Claude 3: ~700ms
- Local models (Ollama): ~50ms

### Connection Pooling

```python
# Reuse client connection
client = OpenAI(
    api_key="...",
    base_url="http://localhost:8080/v1",
    timeout=30.0,  # 30s timeout
    max_retries=3
)

# For async operations
import asyncio
from openai import AsyncOpenAI

async_client = AsyncOpenAI(
    api_key="...",
    base_url="http://localhost:8080/v1"
)

async def get_response(message):
    response = await async_client.chat.completions.create(
        model="router:chat",
        messages=[{"role": "user", "content": message}],
        max_tokens=1024
    )
    return response
```

### Throughput Optimization

```python
# Batch requests asynchronously
import asyncio

async def batch_completions(messages_list):
    async_client = AsyncOpenAI(base_url="http://localhost:8080/v1")
    tasks = [
        async_client.chat.completions.create(
            model="router:chat",
            messages=[{"role": "user", "content": msg}],
            max_tokens=1024
        )
        for msg in messages_list
    ]
    return await asyncio.gather(*tasks)

# Run 10 requests in parallel (vs 10 sequential)
results = asyncio.run(batch_completions([...] * 10))
```

### Cache Hit Rate

Monitor cache effectiveness:
```python
cache_hits = 0
cache_misses = 0

def track_cache(message):
    global cache_hits, cache_misses
    cache_key = hashlib.md5(message.encode()).hexdigest()
    if cache.exists(cache_key):
        cache_hits += 1
        return cache.get(cache_key)
    else:
        cache_misses += 1
        # Fetch and cache...
        return result

hit_rate = cache_hits / (cache_hits + cache_misses)
print(f"Cache hit rate: {hit_rate:.1%}")  # Goal: >50%
```

---

## 10 - Environment Setup Guide

### Python Setup
```bash
# Virtual environment
python3 -m venv venv
source venv/bin/activate  # macOS/Linux
# or
venv\Scripts\activate  # Windows

# Install dependencies
pip install openai anthropic
```

### Node.js Setup
```bash
# Initialize project
npm init -y
npm install openai

# Environment variables
echo "GATEWAY_URL=http://localhost:8080" > .env
echo "OPENAI_API_KEY=sk-..." >> .env
```

### Environment Variables
```bash
# Gateway
export GATEWAY_URL=http://localhost:8080
export GATEWAY_PORT=8080

# Provider credentials
export OPENAI_API_KEY=sk-...
export ANTHROPIC_API_KEY=sk-ant-...

# Client configuration
export LLM_MODEL=router:chat
export LLM_MAX_TOKENS=1024
```

---

## 11 - Deployment Guide

### Docker Deployment
```dockerfile
FROM golang:1.21

WORKDIR /app
COPY . .
RUN go build -o gateway ./cmd

EXPOSE 8080
ENV OPENAI_API_KEY=${OPENAI_API_KEY}
ENV ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}

CMD ["./gateway"]
```

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: llm-gateway
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: gateway
        image: llm-gateway:latest
        ports:
        - containerPort: 8080
        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: provider-keys
              key: openai-key
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        resources:
          limits:
            memory: "512Mi"
            cpu: "500m"
```

### Health Checks
```bash
# Kubernetes liveness probe
GET /health → 200 OK

# Readiness probe
GET /v1/models → 200 OK (models available)
```

---

## 12 - Security & Authentication Guide

### API Key Management
```bash
# Never hardcode keys
❌ client = OpenAI(api_key="sk-...")

# Use environment variables
✅ client = OpenAI()  # Picks up OPENAI_API_KEY

# Use secrets manager
✅ from aws_secretsmanager import client as sm
   key = sm.get_secret_value("openai-api-key")
   client = OpenAI(api_key=key)
```

### Input Validation
```python
def validate_message(message: str) -> bool:
    # Sanitize user input
    if len(message) > 10000:
        return False  # Prevent DoS
    if "<script>" in message.lower():
        return False  # Prevent injection
    return True

if validate_message(user_input):
    response = client.chat.completions.create(...)
```

### TLS/HTTPS
```python
# Production: use HTTPS
client = OpenAI(
    api_key="...",
    base_url="https://gateway.example.com:8080/v1"
)

# Verify certificates
import httpx
client = OpenAI(
    api_key="...",
    base_url="https://gateway.example.com/v1",
    http_client=httpx.Client(verify=True)  # Verify SSL
)
```

### Audit Logging
```python
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def log_request(model, user_id, message_count):
    logger.info(f"Request: user={user_id}, model={model}, messages={message_count}")

# Use when making requests
log_request("gpt-4", user_123, len(messages))
response = client.chat.completions.create(...)
```

---

## Summary

| Topic | Completed |
|-------|-----------|
| Best Practices | ✅ Caching, retries, errors, cost, rate limiting |
| Troubleshooting | ✅ Common issues & solutions |
| Performance | ✅ Latency, pooling, throughput, caching |
| Environment Setup | ✅ Python, Node.js, environment vars |
| Deployment | ✅ Docker, Kubernetes, health checks |
| Security | ✅ API keys, validation, TLS, audit logging |

**All 12 Client Setup tasks completed!**
