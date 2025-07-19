# Anthropic Messages API vs OpenAI API Format Comparison

## 1. Message Structure Differences

### Anthropic Messages Format
```json
{
  "messages": [
    {
      "role": "user",
      "content": "Hello, Claude"
    },
    {
      "role": "assistant",
      "content": "Hi, I'm Claude. How can I help?"
    }
  ]
}
```

### OpenAI Messages Format
```json
{
  "messages": [
    {
      "role": "user",
      "content": "Hello"
    },
    {
      "role": "assistant",
      "content": "Hello! How can I help you today?"
    }
  ]
}
```

**Key Differences:**
- Both use similar message structure with `role` and `content`
- Anthropic supports content as string or array of content blocks
- OpenAI also supports content as string or structured format
- Both support `user` and `assistant` roles

## 2. System Messages Handling

### Anthropic
- System message is passed as a separate parameter at the API level:
```json
{
  "system": "You are a helpful assistant",
  "messages": [...]
}
```

### OpenAI
- System message is included in the messages array:
```json
{
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful assistant"
    },
    {
      "role": "user",
      "content": "Hello"
    }
  ]
}
```

## 3. Tool/Function Calling Format

### Anthropic Tool Format
```json
{
  "tools": [
    {
      "name": "get_weather",
      "description": "Get the current weather in a given location",
      "input_schema": {
        "type": "object",
        "properties": {
          "location": {
            "type": "string",
            "description": "The city and state, e.g. San Francisco, CA"
          }
        },
        "required": ["location"]
      }
    }
  ],
  "tool_choice": {"type": "any"}
}
```

**Tool Use Response:**
```json
{
  "type": "tool_use",
  "id": "toolu_01T1x1fJ34qAmk2tNTrN7Up6",
  "name": "get_weather",
  "input": {
    "location": "San Francisco, CA"
  }
}
```

**Tool Result Format:**
```json
{
  "type": "tool_result",
  "tool_use_id": "toolu_01T1x1fJ34qAmk2tNTrN7Up6",
  "content": "72°F, sunny"
}
```

### OpenAI Tool Format
```json
{
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "get_weather",
        "description": "Get the current weather in a given location",
        "parameters": {
          "type": "object",
          "properties": {
            "location": {
              "type": "string",
              "description": "The city and state, e.g. San Francisco, CA"
            }
          },
          "required": ["location"]
        },
        "strict": true  // For structured outputs
      }
    }
  ],
  "tool_choice": "auto"
}
```

**Tool Call Response:**
```json
{
  "tool_calls": [
    {
      "id": "call_abc123",
      "type": "function",
      "function": {
        "name": "get_weather",
        "arguments": "{\"location\": \"San Francisco, CA\"}"
      }
    }
  ]
}
```

## 4. Streaming Response Format

### Anthropic Streaming Events
```
event: message_start
data: {"type":"message_start","message":{"id":"msg_01...","type":"message","role":"assistant","content":[]}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":15}}

event: message_stop
data: {"type":"message_stop"}
```

**Tool Use Streaming:**
```
event: content_block_start
data: {"type":"content_block_start","index":1,"content_block":{"type":"tool_use","id":"toolu_01...","name":"get_weather","input":{}}}

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"{\"location\":"}}

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":" \"San Francisco"}}
```

### OpenAI Streaming Format
- Uses Server-Sent Events (SSE) similar to Anthropic
- Streams chunks with delta objects
- Tool calls are announced upfront with arguments streamed thereafter
- Usage information provided in the last chunk

**Key Streaming Differences:**
- Anthropic uses more granular event types (message_start, content_block_start, etc.)
- OpenAI uses a simpler chunk-based approach
- Anthropic supports fine-grained tool streaming with `input_json_delta`
- Both support incremental text streaming

## 5. Special Fields

### Anthropic Special Fields

#### Thinking Blocks (Claude 4 models)
```json
{
  "thinking": {
    "type": "enabled",
    "budget_tokens": 16000
  }
}
```

**Thinking Streaming Events:**
```
event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"thinking","thinking":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"Let me solve this step by step..."}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"signature_delta","signature":"EqQBCgIYAhIM..."}}
```

#### Cache Control
```json
{
  "cache_control": {
    "type": "ephemeral",
    "ttl": "5m"
  }
}
```

#### Beta Features
- Fine-grained tool streaming: `anthropic-beta: fine-grained-tool-streaming-2025-05-14`
- Token-efficient tools: `anthropic-beta: token-efficient-tools-2025-02-19`

### OpenAI Special Fields

#### Structured Outputs
```json
{
  "response_format": {
    "type": "json_schema",
    "json_schema": {
      "name": "test_response",
      "strict": true,
      "schema": {...}
    }
  }
}
```

#### JSON Mode
```json
{
  "response_format": {
    "type": "json_object"
  }
}
```

**Note:** Messages must contain the word "json" when using JSON mode.

## Summary of Key Differences

1. **System Messages**: Anthropic uses a separate parameter, OpenAI includes in messages array
2. **Tool Definitions**: Anthropic uses `input_schema`, OpenAI uses `parameters` with optional `strict` mode
3. **Tool Responses**: Anthropic uses `tool_use` blocks, OpenAI uses `tool_calls` array
4. **Streaming**: Anthropic has more granular event types, OpenAI uses simpler chunk approach
5. **Special Features**: 
   - Anthropic: Thinking blocks, fine-grained tool streaming, cache control
   - OpenAI: Structured outputs with 100% schema adherence, JSON mode
6. **Tool Arguments**: Anthropic uses parsed JSON objects, OpenAI uses stringified JSON in `arguments`
7. **Event Types**: Anthropic has dedicated events for different phases (start, delta, stop), OpenAI uses unified chunk format