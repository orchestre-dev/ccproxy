# Messages Endpoint

The `/v1/messages` endpoint is the main proxy endpoint that converts Anthropic Messages API requests to the configured provider's format.

## Endpoint

```
POST /v1/messages
```

## Authentication

CCProxy supports API key authentication when configured:

- **Bearer Token**: `Authorization: Bearer your-api-key`
- **API Key Header**: `x-api-key: your-api-key`

If no API key is configured in `config.json`, the service is only accessible from localhost.

## Request Format

The request format follows the Anthropic Messages API specification:

### Headers

```http
Content-Type: application/json
Accept: application/json
```

### Request Body

```json
{
  "model": "claude-3-sonnet",
  "messages": [
    {
      "role": "user",
      "content": "Hello, how are you?"
    }
  ],
  "max_tokens": 100,
  "temperature": 0.7,
  "top_p": 0.9,
  "tools": [],
  "system": "You are a helpful assistant."
}
```

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `model` | string | Yes | Model name (mapped to provider model) |
| `messages` | array | Yes | Array of message objects |
| `max_tokens` | integer | No | Maximum tokens to generate |
| `temperature` | number | No | Sampling temperature (0-1) |
| `top_p` | number | No | Nucleus sampling parameter |
| `tools` | array | No | Available tools/functions |
| `system` | string | No | System message |
| `stop` | array | No | Stop sequences |

### Message Object

```json
{
  "role": "user|assistant|system",
  "content": "string or array of content blocks"
}
```

### Content Blocks

#### Text Content
```json
{
  "type": "text",
  "text": "Hello, how are you?"
}
```

#### Tool Use
```json
{
  "type": "tool_use",
  "id": "call_123",
  "name": "function_name",
  "input": {
    "parameter": "value"
  }
}
```

#### Tool Result
```json
{
  "type": "tool_result",
  "tool_use_id": "call_123",
  "content": "Result of the function call"
}
```

### Tool Definition

```json
{
  "name": "get_weather",
  "description": "Get current weather for a location",
  "input_schema": {
    "type": "object",
    "properties": {
      "location": {
        "type": "string",
        "description": "City name"
      }
    },
    "required": ["location"]
  }
}
```

## Response Format

Responses follow the Anthropic Messages API format:

### Success Response

```json
{
  "id": "msg_123abc",
  "type": "message",
  "role": "assistant",
  "model": "anthropic/claude-3-sonnet-20240229",
  "content": [
    {
      "type": "text",
      "text": "Hello! I'm doing well, thank you for asking."
    }
  ],
  "stop_reason": "end_turn",
  "stop_sequence": null,
  "usage": {
    "input_tokens": 12,
    "output_tokens": 20
  }
}
```

### Tool Use Response

```json
{
  "id": "msg_456def",
  "type": "message",
  "role": "assistant", 
  "model": "anthropic/claude-3-sonnet-20240229",
  "content": [
    {
      "type": "tool_use",
      "id": "call_123",
      "name": "get_weather",
      "input": {
        "location": "San Francisco"
      }
    }
  ],
  "stop_reason": "tool_use",
  "stop_sequence": null,
  "usage": {
    "input_tokens": 25,
    "output_tokens": 15
  }
}
```

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique message identifier |
| `type` | string | Always "message" |
| `role` | string | Always "assistant" |
| `model` | string | Provider/model that generated the response |
| `content` | array | Array of content blocks |
| `stop_reason` | string | Why generation stopped |
| `stop_sequence` | string | Stop sequence that triggered end |
| `usage` | object | Token usage information |

### Stop Reasons

| Reason | Description |
|--------|-------------|
| `end_turn` | Natural end of response |
| `max_tokens` | Hit token limit |
| `stop_sequence` | Triggered stop sequence |
| `tool_use` | Model wants to use a tool |

## Example Requests

### Basic Text Request

```bash
curl -X POST http://localhost:3456/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [
      {
        "role": "user",
        "content": "What is the capital of France?"
      }
    ],
    "max_tokens": 50
  }'
```

### Multi-turn Conversation

```bash
curl -X POST http://localhost:3456/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [
      {
        "role": "user",
        "content": "Hello!"
      },
      {
        "role": "assistant", 
        "content": [{"type": "text", "text": "Hello! How can I help you?"}]
      },
      {
        "role": "user",
        "content": "What is 2+2?"
      }
    ],
    "max_tokens": 100
  }'
```

### Request with System Message

```bash
curl -X POST http://localhost:3456/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet",
    "system": "You are a helpful math tutor.",
    "messages": [
      {
        "role": "user",
        "content": "Explain calculus basics"
      }
    ],
    "max_tokens": 500
  }'
```

### Request with Tools

```bash
curl -X POST http://localhost:3456/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [
      {
        "role": "user",
        "content": "What is the weather like in Tokyo?"
      }
    ],
    "tools": [
      {
        "name": "get_weather",
        "description": "Get current weather for a location",
        "input_schema": {
          "type": "object",
          "properties": {
            "location": {"type": "string"},
            "unit": {"type": "string", "enum": ["celsius", "fahrenheit"]}
          },
          "required": ["location"]
        }
      }
    ],
    "max_tokens": 200
  }'
```

### Tool Result Follow-up

```bash
curl -X POST http://localhost:3456/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [
      {
        "role": "user",
        "content": "What is the weather like in Tokyo?"
      },
      {
        "role": "assistant",
        "content": [
          {
            "type": "tool_use",
            "id": "call_123",
            "name": "get_weather",
            "input": {"location": "Tokyo"}
          }
        ]
      },
      {
        "role": "user",
        "content": [
          {
            "type": "tool_result",
            "tool_use_id": "call_123",
            "content": "Temperature: 22°C, Sunny, Light breeze"
          }
        ]
      }
    ],
    "tools": [
      {
        "name": "get_weather",
        "description": "Get current weather for a location",
        "input_schema": {
          "type": "object",
          "properties": {
            "location": {"type": "string"}
          },
          "required": ["location"]
        }
      }
    ],
    "max_tokens": 150
  }'
```

## Error Responses

### 400 Bad Request

```json
{
  "error": {
    "type": "invalid_request_error",
    "message": "Missing required field: messages"
  }
}
```

### 401 Unauthorized

```json
{
  "error": {
    "type": "authentication_error", 
    "message": "Provider API key is invalid"
  }
}
```

### 429 Too Many Requests

```json
{
  "error": {
    "type": "rate_limit_error",
    "message": "Rate limit exceeded. Please retry after 60 seconds."
  }
}
```

### 500 Internal Server Error

```json
{
  "error": {
    "type": "internal_server_error",
    "message": "An unexpected error occurred"
  }
}
```

### 502 Bad Gateway

```json
{
  "error": {
    "type": "api_error",
    "message": "Provider returned an error: Model not available"
  }
}
```

## Provider-Specific Behavior

### Model Mapping

CCProxy automatically maps Anthropic model names to provider-specific models. The mapping is configured in your `config.json` file.

Example mappings:

| Provider | Anthropic Model | Mapped Model |
|----------|----------------|--------------|
| Anthropic | claude-3-sonnet | claude-3-sonnet-20240229 |
| OpenRouter | claude-3-sonnet | anthropic/claude-3.5-sonnet |
| OpenAI | claude-3-sonnet | gpt-4 |
| Gemini | claude-3-sonnet | gemini-1.5-flash |
| DeepSeek | claude-3-sonnet | deepseek-coder |

### Feature Support

Provider capabilities vary:

| Feature | Anthropic | OpenRouter | OpenAI | Gemini | DeepSeek |
|---------|-----------|------------|--------|--------|-----------|
| Text | ✅ | ✅ | ✅ | ✅ | ✅ |
| Tools | ✅ | ✅ | ✅ | ✅ | ✅ |
| Vision | ✅ | ✅* | ✅ | ✅ | ❌ |
| Streaming | ✅ | ✅ | ✅ | ✅ | ✅ |

*Depends on specific model

## Rate Limits

Rate limits are enforced by the underlying providers and vary by plan:

| Provider | Default Limits | Notes |
|----------|----------------|-------|
| Anthropic | Per API key | Check your Anthropic dashboard |
| OpenRouter | Model-specific | Varies by model |
| OpenAI | Tier-based | Based on usage tier |
| Gemini | 15 RPM (free) | Higher limits for paid plans |
| DeepSeek | API key based | Check provider documentation |

## Best Practices

### 1. Error Handling

Always implement proper error handling:

```javascript
try {
  const response = await fetch('/v1/messages', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request)
  });
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error.message);
  }
  
  const data = await response.json();
  return data;
} catch (error) {
  console.error('API Error:', error.message);
}
```

### 2. Token Management

Monitor token usage to optimize costs:

```javascript
const response = await callAPI(request);
console.log(`Used ${response.usage.input_tokens} input tokens`);
console.log(`Used ${response.usage.output_tokens} output tokens`);
```

### 3. Tool Use Patterns

Implement proper tool use flows:

```javascript
function handleToolUse(response) {
  for (const content of response.content) {
    if (content.type === 'tool_use') {
      const result = executeFunction(content.name, content.input);
      // Send result back in next message
      return {
        role: 'user',
        content: [{
          type: 'tool_result',
          tool_use_id: content.id,
          content: result
        }]
      };
    }
  }
}
```

### 4. Retry Logic

Implement exponential backoff for rate limits:

```javascript
async function retryRequest(request, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await callAPI(request);
    } catch (error) {
      if (error.status === 429 && i < maxRetries - 1) {
        await new Promise(resolve => 
          setTimeout(resolve, Math.pow(2, i) * 1000)
        );
        continue;
      }
      throw error;
    }
  }
}
```

## Next Steps

- Learn about [health endpoints](/api/health) for monitoring
- Explore [Claude Code integration](/api/claude) for seamless usage
- Check out [provider-specific features](/providers/) for optimization