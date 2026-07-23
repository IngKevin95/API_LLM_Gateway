# HTTP Client Setup (Raw API Requests)

Direct HTTP requests without SDK. Use the `/responses` endpoint for universal format.

## Endpoints

| Endpoint | Purpose | Format |
|----------|---------|--------|
| `POST /responses` | Universal format (recommended) | Any format, auto-detected |
| `POST /v1/chat/completions` | OpenAI-compatible | OpenAI/Anthropic format |
| `POST /v1/embeddings` | Embeddings | OpenAI format |
| `GET /v1/models` | List models | Query params |

## Basic Request

**cURL:**
```bash
curl -X POST http://localhost:8080/responses \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello"}],
    "max_tokens": 100
  }'
```

**Python (requests):**
```python
import requests

response = requests.post(
    "http://localhost:8080/responses",
    headers={"Content-Type": "application/json"},
    json={
        "model": "gpt-4",
        "messages": [{"role": "user", "content": "Hello"}],
        "max_tokens": 100
    }
)
data = response.json()
print(data["choices"][0]["message"]["content"])
```

**JavaScript (fetch):**
```javascript
const response = await fetch("http://localhost:8080/responses", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    model: "gpt-4",
    messages: [{ role: "user", content: "Hello" }],
    max_tokens: 100
  })
});
const data = await response.json();
console.log(data.choices[0].message.content);
```

## Capability-Based Routing

```bash
curl -X POST http://localhost:8080/responses \
  -H "Content-Type: application/json" \
  -d '{
    "model": "router:chat",
    "messages": [{"role": "user", "content": "Hello"}],
    "max_tokens": 100
  }'
```

## Streaming Responses

Add `stream: true` to enable streaming:

```bash
curl -N -X POST http://localhost:8080/responses \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello"}],
    "max_tokens": 100,
    "stream": true
  }'
```

Response is newline-delimited JSON (NDJSON):
```
{"choices":[{"delta":{"content":"Hello"}}]}
{"choices":[{"delta":{"content":" there"}}]}
{"choices":[{"finish_reason":"stop"}]}
```

**Python streaming:**
```python
import requests

response = requests.post(
    "http://localhost:8080/responses",
    json={"model": "gpt-4", "messages": [...], "stream": True},
    stream=True
)

for line in response.iter_lines():
    if line:
        chunk = json.loads(line)
        if chunk["choices"][0]["delta"].get("content"):
            print(chunk["choices"][0]["delta"]["content"], end="")
```

## Vision (Image Understanding)

```bash
curl -X POST http://localhost:8080/responses \
  -H "Content-Type: application/json" \
  -d '{
    "model": "router:vision",
    "messages": [
      {
        "role": "user",
        "content": [
          {"type": "text", "text": "What is in this image?"},
          {
            "type": "image_url",
            "image_url": {"url": "https://example.com/image.jpg"}
          }
        ]
      }
    ],
    "max_tokens": 100
  }'
```

## Error Handling

Errors are returned as JSON with status codes:

```json
400 Bad Request
{
  "error": {
    "message": "max_tokens is required",
    "type": "invalid_request_error"
  }
}

404 Not Found
{
  "error": {
    "message": "model not found",
    "type": "not_found_error"
  }
}

503 Service Unavailable
{
  "error": {
    "message": "all providers unavailable",
    "type": "service_unavailable"
  }
}
```

**Python error handling:**
```python
response = requests.post("http://localhost:8080/responses", json={...})

if response.status_code == 400:
    error = response.json()["error"]
    print(f"Validation error: {error['message']}")
elif response.status_code == 503:
    print("Gateway overloaded, retry later")
else:
    result = response.json()
    print(result["choices"][0]["message"]["content"])
```

## Batch Requests

Send multiple requests efficiently:

```bash
# Request 1
curl -X POST http://localhost:8080/responses -d '{"model":"gpt-4","messages":[...]}'

# Request 2
curl -X POST http://localhost:8080/responses -d '{"model":"claude-3","messages":[...]}'
```

## Health Check

```bash
curl http://localhost:8080/health

# Response: 200 OK
# {"status": "healthy"}
```

## Get Available Models

```bash
# List all models
curl http://localhost:8080/v1/models

# Filter by capability
curl http://localhost:8080/v1/models?capability=chat

# Filter by provider
curl http://localhost:8080/v1/models?provider=openai

# Include metadata
curl http://localhost:8080/v1/models?include_metadata=true
```

## Response Format

Standard response structure:
```json
{
  "id": "chatcmpl-...",
  "object": "text_completion",
  "created": 1234567890,
  "model": "gpt-4",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello!"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 5,
    "total_tokens": 15
  }
}
```
