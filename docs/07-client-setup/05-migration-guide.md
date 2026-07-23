# Migration Guide: Direct Provider → API LLM Gateway

Step-by-step guide to migrate from direct provider calls to the gateway.

## Before (Direct Provider)

```python
from openai import OpenAI

# Direct connection to OpenAI
client = OpenAI(api_key="sk-...")

response = client.chat.completions.create(
    model="gpt-4",
    messages=[{"role": "user", "content": "Hello"}]
)
```

## After (Via Gateway)

```python
from openai import OpenAI

# Connection to gateway (proxy)
client = OpenAI(
    api_key="your-api-key",
    base_url="http://localhost:8080/v1"
)

response = client.chat.completions.create(
    model="gpt-4",
    messages=[{"role": "user", "content": "Hello"}]
)
```

**That's it!** Your code works without changes. Gateway translates parameters automatically.

## Step 1: Update Base URL

Change your client initialization to point to the gateway:

**Before:**
```python
client = OpenAI()  # Uses api.openai.com
```

**After:**
```python
client = OpenAI(base_url="http://localhost:8080/v1")
```

## Step 2: Set API Key

Gateway accepts any API key (it validates and routes to real providers):

```bash
export GATEWAY_API_KEY="sk-proxy-key"
```

Or in code:
```python
client = OpenAI(api_key="sk-proxy-key", base_url="http://localhost:8080/v1")
```

## Step 3: Add Provider Credentials

Configure actual provider credentials in gateway:

**Environment variables:**
```bash
export OPENAI_API_KEY=sk-...
export ANTHROPIC_API_KEY=sk-ant-...
export GOOGLE_API_KEY=...
```

**Or gateway config.yaml:**
```yaml
providers:
  - id: openai
    api_key: ${OPENAI_API_KEY}
  - id: anthropic
    api_key: ${ANTHROPIC_API_KEY}
```

## Step 4: Handle New Features

### Automatic Model Selection

Instead of choosing models, use capabilities:

**Before:**
```python
response = client.chat.completions.create(
    model="gpt-4",  # Locked to GPT-4
    messages=[...]
)
```

**After:**
```python
response = client.chat.completions.create(
    model="router:chat",  # Best chat model automatically
    messages=[...]
)
```

Benefits:
- Cost optimization (use cheaper models when performance allows)
- Provider failover (automatic fallback if model unavailable)
- Future-proof (upgrade models without code changes)

### Parameter Translation

Gateway automatically handles parameter differences:

**Before (OpenAI only):**
```python
response = client.chat.completions.create(
    model="gpt-4",
    messages=[...],
    temperature=1.5,  # OpenAI range [0, 2]
    response_format="json_object"
)
```

**After (Works with any provider):**
```python
response = client.chat.completions.create(
    model="router:chat",  # Use Anthropic or OpenAI
    messages=[...],
    temperature=1.5,  # Auto-clamped to [0, 1] for Anthropic
    response_format="json_object"  # Silently filtered if not supported
)
```

## Step 5: Update Error Handling

Add provider-specific error messages:

**Before:**
```python
try:
    response = client.chat.completions.create(...)
except Exception as e:
    print(f"OpenAI error: {e}")
```

**After:**
```python
from openai import APIError

try:
    response = client.chat.completions.create(...)
except APIError as e:
    if "max_tokens is required" in str(e):
        print("Using Anthropic - max_tokens is required")
    elif "model not found" in str(e):
        print("Model not available - check /v1/models")
    else:
        print(f"API error: {e}")
```

## Step 6: Leverage New Capabilities

### Vision (Multi-Provider)

**Before (OpenAI only):**
```python
response = client.chat.completions.create(
    model="gpt-4-vision",
    messages=[...image_message...]
)
```

**After (OpenAI or Anthropic Claude 3):**
```python
response = client.chat.completions.create(
    model="router:vision",  # Auto-selects vision-capable model
    messages=[...image_message...]
)
```

### Streaming

Works as before, no changes needed:
```python
response = client.chat.completions.create(
    model="router:chat",
    messages=[...],
    stream=True
)
```

### Tool Use (Function Calling)

Works with any provider's implementation:
```python
response = client.chat.completions.create(
    model="router:chat",
    messages=[...],
    tools=[...]  # Gateway handles format translation
)
```

## Step 7: Monitor & Optimize

Use `/v1/models` to understand available models:

```bash
# List all models
curl http://localhost:8080/v1/models

# List vision models
curl http://localhost:8080/v1/models?capability=vision

# Include cost/latency metadata
curl http://localhost:8080/v1/models?include_metadata=true
```

## Common Issues During Migration

**Issue: "max_tokens is required"**
- Solution: Always pass max_tokens (required by Anthropic)
- Code: Add `max_tokens=2048` to all requests

**Issue: "model not found"**
- Solution: Check model availability via `/v1/models`
- Code: Use `router:chat` for automatic selection

**Issue: "response_format not supported"**
- Solution: Gracefully handle or parse response manually
- Code: Wrap json_object requests with try/except

**Issue: Provider keeps switching**
- Solution: Use explicit model names for consistency
- Code: Change `router:chat` to `gpt-4` or `claude-3-opus`

## Rollback Plan

If you need to revert:

```python
# Just change base_url back
client = OpenAI()  # Uses api.openai.com directly
```

No code changes needed - gateway acts as drop-in proxy.

## Validation Checklist

- [ ] Base URL changed to gateway
- [ ] API key configured
- [ ] Provider credentials set in gateway
- [ ] Tested with explicit model names
- [ ] Tested with router: prefixed models
- [ ] Error handling updated
- [ ] Streaming tested
- [ ] Tool use tested (if used)
- [ ] Vision tested (if used)
- [ ] Cost metrics monitored

## Next Steps

- [Best Practices](06-best-practices.md) — Caching, retries, error handling
- [Performance Tuning](09-performance.md) — Optimize throughput and latency
- [Troubleshooting](08-troubleshooting.md) — Common issues and solutions
