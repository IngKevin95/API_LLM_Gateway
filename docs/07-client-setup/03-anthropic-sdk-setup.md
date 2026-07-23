# Anthropic SDK Setup with API LLM Gateway

Use the official Anthropic SDK with the gateway as a proxy.

## Installation

```bash
pip install anthropic
```

## Configuration

```python
from anthropic import Anthropic

client = Anthropic(
    api_key="your-api-key",
    base_url="http://localhost:8080/v1"  # Point to gateway
)

message = client.messages.create(
    model="claude-3-opus-20240229",
    max_tokens=1024,  # REQUIRED for Anthropic
    messages=[{"role": "user", "content": "Hello"}]
)
print(message.content[0].text)
```

## Key Differences

### max_tokens is Required
Anthropic **requires** `max_tokens` in every request (unlike OpenAI where it's optional):

```python
# ✅ Correct
message = client.messages.create(
    model="claude-3-opus-20240229",
    max_tokens=1024,  # Required
    messages=[{"role": "user", "content": "Hello"}]
)

# ❌ Wrong - will return 400 error
message = client.messages.create(
    model="claude-3-opus-20240229",
    messages=[{"role": "user", "content": "Hello"}]
)
```

### Temperature Range
Anthropic temperature: [0, 1] (vs. OpenAI: [0, 2])
Gateway auto-clamps values:

```python
# Temperature 1.5 will be clamped to 1.0 for Anthropic
message = client.messages.create(
    model="claude-3-opus-20240229",
    max_tokens=1024,
    temperature=1.5,  # Clamped to 1.0
    messages=[{"role": "user", "content": "Hello"}]
)
```

## Parameter Support

**Supported:**
- `temperature` (0-1) — Auto-clamped
- `top_p` (0-1) — Nucleus sampling
- `top_k` — Top-K sampling (Anthropic-specific)
- `max_tokens` — **Required**
- `stop_sequences` — Stop conditions
- `tools` — Function definitions
- `tool_choice` — "auto", "required", "none"

**Not Supported (silently filtered):**
- `seed` — Anthropic doesn't support deterministic seeding
- `response_format` — Use raw response + client-side parsing
- `presence_penalty`, `frequency_penalty` — Different penalty model

## Extended Thinking (Claude 3.5)

```python
message = client.messages.create(
    model="claude-3-5-sonnet-20241022",
    max_tokens=16000,
    thinking={"type": "enabled", "budget_tokens": 10000},
    messages=[{"role": "user", "content": "Solve this complex problem..."}]
)

# Access thinking and response
for block in message.content:
    if block.type == "thinking":
        print(f"Thinking: {block.thinking}")
    elif block.type == "text":
        print(f"Response: {block.text}")
```

## Tool Use (Function Calling)

```python
message = client.messages.create(
    model="claude-3-opus-20240229",
    max_tokens=1024,
    tools=[
        {
            "name": "weather",
            "description": "Get weather for a location",
            "input_schema": {
                "type": "object",
                "properties": {
                    "location": {"type": "string"}
                },
                "required": ["location"]
            }
        }
    ],
    messages=[{"role": "user", "content": "What's the weather in Paris?"}]
)

# Process tool calls
for block in message.content:
    if block.type == "tool_use":
        print(f"Called tool: {block.name}")
        print(f"Input: {block.input}")
```

## Streaming

```python
with client.messages.stream(
    model="claude-3-opus-20240229",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Hello"}]
) as stream:
    for text in stream.text_stream:
        print(text, end="", flush=True)
```

## Vision (Image Understanding)

```python
message = client.messages.create(
    model="claude-3-opus-20240229",
    max_tokens=1024,
    messages=[
        {
            "role": "user",
            "content": [
                {"type": "text", "text": "What's in this image?"},
                {
                    "type": "image",
                    "source": {
                        "type": "url",
                        "url": "https://example.com/image.jpg"
                    }
                }
            ]
        }
    ]
)
```

## Error Handling

```python
from anthropic import APIError, APIStatusError

try:
    message = client.messages.create(
        model="claude-3-opus-20240229",
        max_tokens=1024,
        messages=[{"role": "user", "content": "Hello"}]
    )
except APIStatusError as e:
    if "max_tokens is required" in str(e):
        print("Remember: Anthropic requires max_tokens")
    elif e.status_code == 404:
        print("Model not found - check /v1/models")
    else:
        print(f"API error: {e}")
```

## Capability-Based Routing

```python
# Let gateway pick best Claude model
message = client.messages.create(
    model="router:chat",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Hello"}]
)
```

## Conversation History

```python
from anthropic import Anthropic

client = Anthropic(
    api_key="your-api-key",
    base_url="http://localhost:8080/v1"
)

conversation = []

# First message
response1 = client.messages.create(
    model="claude-3-opus-20240229",
    max_tokens=1024,
    messages=[{"role": "user", "content": "What is 2+2?"}]
)
conversation.append({"role": "user", "content": "What is 2+2?"})
conversation.append({"role": "assistant", "content": response1.content[0].text})

# Follow-up with context
response2 = client.messages.create(
    model="claude-3-opus-20240229",
    max_tokens=1024,
    messages=conversation + [{"role": "user", "content": "What about 3+3?"}]
)
```

## Environment Variables

```bash
export ANTHROPIC_API_KEY="your-api-key"
export ANTHROPIC_BASE_URL="http://localhost:8080/v1"
```

Note: Base URL support varies by SDK version. Verify your version supports it.
