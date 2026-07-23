# OpenAI SDK Setup with API LLM Gateway

Use the official OpenAI SDK pointing to the gateway instead of api.openai.com.

## Installation

**Python:**
```bash
pip install openai
```

**Node.js:**
```bash
npm install openai
```

**Go:**
```bash
go get github.com/sashabaranov/go-openai
```

## Configuration

### Python

```python
from openai import OpenAI

client = OpenAI(
    api_key="your-api-key",
    base_url="http://localhost:8080/v1"  # Point to gateway
)

response = client.chat.completions.create(
    model="gpt-4",
    messages=[{"role": "user", "content": "Hello"}],
    max_tokens=100
)
print(response.choices[0].message.content)
```

### Node.js

```javascript
const OpenAI = require("openai");

const openai = new OpenAI({
  apiKey: "your-api-key",
  baseURL: "http://localhost:8080/v1"  // Point to gateway
});

const response = await openai.chat.completions.create({
  model: "gpt-4",
  messages: [{ role: "user", content: "Hello" }],
  max_tokens: 100
});
console.log(response.choices[0].message.content);
```

### Go

```go
import "github.com/sashabaranov/go-openai"

config := openai.DefaultConfig("your-api-key")
config.BaseURL = "http://localhost:8080/v1"  // Point to gateway
client := openai.NewClientWithConfig(config)

resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
    Model: "gpt-4",
    Messages: []openai.ChatCompletionMessage{
        {Role: openai.ChatMessageRoleUser, Content: "Hello"},
    },
    MaxTokens: 100,
})
```

## Capability-Based Routing

Use `router:` prefixed model names for automatic provider selection:

```python
response = client.chat.completions.create(
    model="router:chat",      # Auto-select best chat model
    messages=[{"role": "user", "content": "Hello"}],
    max_tokens=100
)
```

Available capabilities:
- `router:chat` — conversational (default)
- `router:vision` — image understanding (auto-selected if images in request)
- `router:embedding` — text embeddings (use via `/embeddings` endpoint)
- `router:reasoning` — extended thinking

## Parameter Support

**Supported parameters:**
- `temperature` (0-2) — Will be clamped per provider (e.g., 0-1 for Anthropic)
- `top_p` (0-1)
- `max_tokens` — Required
- `stop` — Stop sequences
- `tools` — Function definitions
- `tool_choice` — "auto", "required", "none"
- `seed` — For reproducibility (provider-dependent)

**Unsupported parameters (silently filtered):**
- `response_format: json_object` — Use in request, manually parse response
- `presence_penalty`, `frequency_penalty` — Anthropic doesn't support
- `n` — Multiple completions not supported

## Error Handling

```python
from openai import OpenAI
from openai import APIError

client = OpenAI(base_url="http://localhost:8080/v1")

try:
    response = client.chat.completions.create(
        model="gpt-4",
        messages=[{"role": "user", "content": "Hello"}],
        max_tokens=100
    )
except APIError as e:
    if "model not found" in str(e):
        print("Model unavailable, check /v1/models")
    elif "max_tokens is required" in str(e):
        print("Provider (Anthropic) requires max_tokens")
    else:
        print(f"API error: {e}")
```

## Streaming

Streaming works as normal:

```python
response = client.chat.completions.create(
    model="gpt-4",
    messages=[{"role": "user", "content": "Hello"}],
    max_tokens=100,
    stream=True
)

for chunk in response:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")
```

## Vision (Images)

```python
response = client.chat.completions.create(
    model="router:vision",  # Auto-routes to vision-capable model
    messages=[
        {
            "role": "user",
            "content": [
                {"type": "text", "text": "What's in this image?"},
                {
                    "type": "image_url",
                    "image_url": {"url": "https://example.com/image.jpg"}
                }
            ]
        }
    ],
    max_tokens=100
)
```

## Caching (Recommended)

Cache responses to reduce costs:

```python
from functools import lru_cache

@lru_cache(maxsize=1000)
def chat_cached(model: str, message: str) -> str:
    response = client.chat.completions.create(
        model=model,
        messages=[{"role": "user", "content": message}],
        max_tokens=100
    )
    return response.choices[0].message.content
```

## Environment Variables

```bash
export OPENAI_API_KEY="your-api-key"
export OPENAI_BASE_URL="http://localhost:8080/v1"
```

Then use without explicit configuration:
```python
from openai import OpenAI
client = OpenAI()  # Picks up env vars
```
