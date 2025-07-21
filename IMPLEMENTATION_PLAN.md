# CCProxy Implementation Plan

## Project Overview
Create a new Go-based proxy called `ccproxy` inspired by Claude Code Router. This proxy sits between Claude Code and various LLM providers, offering model routing, request transformation, and authentication. Built as a single compiled binary for easy distribution across all platforms.

## Requirements (EARS Framework)

### Environmental Requirements

**ENV-001**: The system SHALL operate on Linux (amd64, arm64), macOS (amd64, arm64), and Windows (amd64) platforms.

**ENV-002**: The system SHALL run as a background service on port 3456 by default.

**ENV-003**: The system SHALL support operation behind corporate proxies when PROXY_URL is configured.

**ENV-004**: The system SHALL maintain a home directory at ~/.ccproxy for configuration and logs.

**ENV-005**: The system SHALL set ANTHROPIC_AUTH_TOKEN="test" and ANTHROPIC_BASE_URL="http://127.0.0.1:3456" when executing Claude Code.

**ENV-006**: The system SHALL support API_TIMEOUT_MS="600000" environment variable for Claude Code execution.

**ENV-007**: The system SHALL support CLAUDE_PATH environment variable to override claude command path.


### Functional Requirements

**FUN-001**: The system SHALL provide a CLI with commands: start, stop, status, and code.

**FUN-002**: The system SHALL proxy requests from Claude Code to configured LLM providers.

**FUN-003**: The system SHALL route requests to different models based on:
- Token count (>60K tokens → longContext model)
- Model type (claude-3-5-haiku → background model)
- Thinking parameter presence (→ think model)
- Explicit model selection via comma-separated format

**FUN-004**: The system SHALL count tokens using tiktoken-compatible tokenization.

**FUN-005**: The system SHALL authenticate requests using API key validation when APIKEY is configured.

**FUN-006**: The system SHALL transform requests/responses based on provider-specific requirements.

**FUN-007**: The system SHALL support dynamic model switching via the /model command format.

**FUN-008**: The system SHALL maintain process lifecycle with PID file management.

**FUN-009**: The system SHALL track reference counts for multiple concurrent Claude Code instances.

**FUN-010**: The system SHALL automatically start the service when executing the code command if not running.

**FUN-011**: The system SHALL wait up to 10 seconds for service startup with 1 second initial delay.

**FUN-012**: The system SHALL use APIKEY for authentication configuration.

**FUN-013**: The system SHALL display "✅ Service is already running in the background" when start is called on running service.

**FUN-014**: The system SHALL support transformer options via array format: ["transformer_name", {options}].

**FUN-015**: The system SHALL display exact status messages with emojis (📊 CCProxy Status, ✅ Status: Running, etc.).

**FUN-016**: The system SHALL only auto-stop service when reference count reaches 0.

**FUN-017**: The system SHALL check LOG environment variable as string equality to "true", not truthy.

**FUN-018**: The system SHALL spawn claude command with shell:true and inherit stdio.

**FUN-019**: The system SHALL handle missing tool responses by injecting default success responses.

### System Requirements

**SYS-001**: The system SHALL be implemented as a single compiled Go binary named `ccproxy`.

**SYS-002**: The system SHALL use the Gin web framework for HTTP server implementation.

**SYS-003**: The system SHALL maintain configuration in JSON format at ~/.ccproxy/config.json.

**SYS-004**: The system SHALL log to ~/.ccproxy/ccproxy.log when LOG is enabled.

**SYS-005**: The system SHALL save PID to ~/.ccproxy/.ccproxy.pid for process management.

**SYS-006**: The system SHALL maintain reference count in system temp directory.

**SYS-007**: The system SHALL support plugin loading from ~/.ccproxy/plugins directory.

### State-driven Requirements

**STATE-001**: WHEN the service is not running, the start command SHALL initialize the service in background mode.

**STATE-002**: WHEN the service is already running, the start command SHALL report success without creating a new instance.

**STATE-003**: WHEN APIKEY is not set AND HOST is configured, the system SHALL force HOST to 127.0.0.1.

**STATE-004**: WHEN executing the code command AND service is not running, the system SHALL automatically start the service.

**STATE-005**: WHEN the last Claude Code instance exits, the system SHALL automatically stop the service if no other instances are running.

### Event-driven Requirements

**EVENT-001**: WHEN receiving SIGINT or SIGTERM, the system SHALL clean up PID file and exit gracefully.

**EVENT-002**: WHEN a request is received, the system SHALL apply authentication middleware before routing.

**EVENT-003**: WHEN routing a request, the system SHALL calculate token count before selecting the target model.

**EVENT-004**: WHEN forwarding a request, the system SHALL apply provider-specific transformations.

### Unwanted Behavior Requirements

**UNW-001**: The system SHALL NOT expose the service on non-localhost addresses when APIKEY is not configured.

**UNW-002**: The system SHALL NOT allow multiple instances of the service to run simultaneously.

**UNW-003**: The system SHALL NOT forward requests without proper authentication when APIKEY is set.

**UNW-004**: The system SHALL NOT expose internal errors to external clients.

## Key Components Analysis

1. **Core Server**: Currently uses @musistudio/llms (a Fastify-based server) - needs to be replaced with Gin framework
2. **Token Counting**: Uses tiktoken for counting tokens - will use tiktoken-go
3. **Process Management**: PID file handling and reference counting for background service
4. **CLI Interface**: Commands: start, stop, status, code
5. **Configuration**: JSON-based config with providers, routers, and transformers
6. **Middleware**: Authentication and routing logic

## Implementation Plan

### Phase 1: Project Setup & Core Structure
1. Initialize Go module: `go mod init github.com/musistudio/ccproxy`
2. Create directory structure:
   ```
   ccproxy/
   ├── cmd/
   │   └── ccproxy/
   │       └── main.go          # CLI entry point
   ├── internal/
   │   ├── server/
   │   │   ├── server.go        # HTTP server implementation
   │   │   ├── middleware/
   │   │   │   ├── auth.go      # Authentication middleware
   │   │   │   └── router.go    # Routing middleware
   │   │   └── handlers/
   │   │       └── proxy.go     # Proxy handler
   │   ├── config/
   │   │   └── config.go        # Configuration management
   │   ├── process/
   │   │   └── manager.go       # Process management (PID, reference counting)
   │   ├── transformer/
   │   │   ├── transformer.go   # Transformer interface
   │   │   └── builtin/         # Built-in transformers
   │   └── utils/
   │       ├── tokenizer.go     # Token counting utilities
   │       └── logger.go        # Logging utilities
   ├── pkg/
   │   └── types/
   │       └── types.go         # Shared types/interfaces
   ├── go.mod
   ├── go.sum
   └── Makefile                 # Build scripts for multi-platform
   ```

### Phase 2: Core Dependencies & Types
1. Add Go dependencies:
   - `github.com/gin-gonic/gin` - Web framework
   - `github.com/pkoukk/tiktoken-go` - Token counting
   - `github.com/spf13/cobra` - CLI framework
   - `github.com/spf13/viper` - Configuration management
   - `github.com/sirupsen/logrus` - Logging
   - `github.com/google/uuid` - UUID generation

2. Define core types:
   - Provider configuration
   - Router configuration
   - Transformer configuration
   - Request/Response types for Anthropic API

### Phase 3: Server Implementation
1. Implement Gin-based HTTP server with endpoints:
   - `POST /v1/messages` - Main Claude API endpoint (Anthropic transformer)
   - `GET /` - Health check returning version info
   - `GET /health` - Health check returning status
   - `POST /providers` - Create new provider (auth required)
   - `GET /providers` - List all providers (auth required)
   - `GET /providers/:id` - Get specific provider (auth required)
   - `PUT /providers/:id` - Update provider (auth required)
   - `DELETE /providers/:id` - Delete provider (auth required)
   - `PATCH /providers/:id/toggle` - Enable/disable provider (auth required)
   - Dynamic transformer endpoints based on transformer configuration
2. Create proxy handler to forward requests to LLM providers
3. Implement middleware:
   - Pre-handler to extract provider,model from request
   - Authentication (API key validation)
   - Router (model selection based on token count and rules)
   - Request/Response transformation pipeline

### Phase 4: CLI Implementation
1. Use Cobra for CLI commands:
   - `ccproxy start` - Start background service
   - `ccproxy stop` - Stop service
   - `ccproxy status` - Show service status
   - `ccproxy code [args]` - Execute Claude Code with routing

2. Implement process management:
   - PID file handling
   - Reference counting for multiple Claude instances
   - Service lifecycle management

### Phase 5: Configuration & Utils
1. Implement configuration loading/validation
2. Implement token counting with tiktoken-go
3. Create logging system with configurable output
4. Handle ~/.claude.json initialization for Claude Code integration

### Phase 6: Transformer System
1. Define transformer interface
2. Implement built-in transformers:
   - deepseek
   - gemini
   - openrouter
   - groq
   - maxtoken
   - tooluse
3. Create plugin loading system for external transformers

### Phase 7: Build & Distribution
1. Create Makefile for cross-platform builds:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64)
2. Set up GitHub Actions for automated builds
3. Create installation scripts

## Key Challenges & Solutions

1. **@musistudio/llms replacement**: Need to implement the proxy logic ourselves using Gin
2. **Dynamic plugin loading**: Go's plugin system is limited; may need to use subprocess or embedded Lua/JavaScript
3. **Claude Code integration**: Implement Anthropic API endpoints for seamless integration
4. **Streaming responses**: Implement SSE (Server-Sent Events) for streaming responses

## Testing Strategy
1. Unit tests for each component
2. Integration tests for the full proxy flow
3. Compatibility tests with Claude Code
4. Performance benchmarks for optimization
5. Multi-platform testing on all target OS/architectures

## Documentation
1. Create comprehensive README for ccproxy
2. Document installation and setup process
3. API documentation for transformer plugins
4. Usage examples and configuration guide

## Technical Decisions

1. **HTTP Client**: Use Go's standard `net/http` with custom transport for proxy support
2. **JSON Handling**: Use `encoding/json` with struct tags for configuration
3. **Concurrent Request Handling**: Leverage Go's goroutines for better performance
4. **Configuration Format**: Clean JSON-based configuration structure
5. **Binary Name**: `ccproxy` - a clean, descriptive name

## Go Technology Stack
- `gin-gonic/gin`: Web framework for HTTP server
- `pkoukk/tiktoken-go`: Token counting functionality
- `spf13/cobra`: CLI framework
- `spf13/viper`: Configuration management
- `sirupsen/logrus`: Structured logging
- `google/uuid`: UUID generation
- Standard `net/http`: HTTP client with proxy support

## Extracted Business Logic from @musistudio/llms

### Core Server Architecture
The server is built on Fastify and consists of several key services:

1. **ConfigService**: Manages configuration loading and access
2. **ProviderService**: Manages LLM providers (registration, updates, model routing)
3. **TransformerService**: Manages request/response transformers
4. **LLMService**: Orchestrates LLM interactions (minimal implementation found)

### Key Business Logic Components

#### 1. Request Flow
1. Request arrives at `/v1/messages` endpoint (or transformer-specific endpoints)
2. Pre-handler middleware extracts provider and model from comma-separated format
3. Authentication middleware validates API key if configured
4. Router middleware selects appropriate model based on:
   - Token count calculation using tiktoken
   - Model type (haiku → background)
   - Thinking parameter presence
   - Explicit model selection
5. Request transformation pipeline:
   - Apply transformer's `transformRequestOut` 
   - Apply provider's global transformers `transformRequestIn`
   - Apply model-specific transformers
6. Forward request to provider's API endpoint
7. Response transformation pipeline:
   - Apply provider's transformers `transformResponseOut`
   - Apply transformer's `transformResponseIn`
8. Stream or return JSON response

#### 2. Provider Management
- Providers stored in Map with name as key
- Model routes maintained for quick lookup (both `model` and `provider,model` formats)
- Provider configuration includes:
  - name, baseUrl, apiKey, models array
  - transformer configuration (global and per-model)
- Dynamic provider CRUD operations via API endpoints

#### 3. Transformer System
- Built-in transformers: anthropic, gemini, deepseek, tooluse, openrouter
- Transformers can define:
  - Custom endpoints (e.g., `/v1/messages` for Anthropic)
  - Request transformation (in/out)
  - Response transformation (in/out)
- Plugin loading from filesystem paths
- Transformer chaining support

#### 4. Model Routing Algorithm
```javascript
const getUseModel = (req, tokenCount, config) => {
  if (req.body.model.includes(",")) return req.body.model;
  if (tokenCount > 60000 && config.Router.longContext) 
    return config.Router.longContext;
  if (req.body.model?.startsWith("claude-3-5-haiku") && config.Router.background)
    return config.Router.background;
  if (req.body.thinking && config.Router.think)
    return config.Router.think;
  return config.Router.default;
};
```

#### 5. Transformer Implementation Details

**Anthropic Transformer** (`/v1/messages` endpoint):
- Converts between OpenAI and Anthropic message formats
- Handles streaming responses with SSE (Server-Sent Events)
- Manages tool calls/function calling translation
- Supports thinking blocks in responses with signature
- Complex stream processing with content blocks
- Converts system messages to user role messages
- Handles tool_result content type conversion

**Gemini Transformer** (`/v1beta/models/:modelAndAction` endpoint):
- Custom URL construction based on stream mode
- Role mapping: assistant → model, others → user
- Strips unsupported JSON schema fields ($schema, additionalProperties)
- Uses x-goog-api-key header instead of Authorization
- Converts function declarations format
- Handles functionCall in response parts

**DeepSeek Transformer**:
- Limits max_tokens to 8192 (hard limit)
- Converts reasoning_content to thinking blocks
- Manages state for reasoning completion
- Adds signature to thinking blocks
- Increments index for content after reasoning

**OpenRouter Transformer**:
- Handles reasoning content similar to DeepSeek
- Converts reasoning to thinking blocks with signature
- Manages index incrementing for tool calls
- Processes streaming responses only

**Tooluse Transformer**:
- Adds system reminder about tool mode
- Injects ExitTool function for graceful exit
- Sets tool_choice to "required"
- Intercepts ExitTool responses and converts to content
- Handles streaming ExitTool responses

**OpenAI Transformer**:
- Minimal implementation - just defines `/v1/chat/completions` endpoint

**Maxtoken Transformer** (mentioned but not found in code):
- Likely a configuration-based transformer
- Should accept max_tokens as parameter
- Applied via transformer options

#### 6. Request Forwarding
- Uses undici for HTTP requests with proxy support
- Configurable timeout (default 60 minutes)
- Bearer token authentication
- JSON request/response handling
- Stream support for real-time responses

### Critical Implementation Details

1. **Model Resolution**: The system supports both direct model names and `provider,model` format
2. **Transformer Chaining**: Multiple transformers can be applied in sequence
3. **Stream Processing**: Complex conversion between OpenAI and Anthropic streaming formats
4. **Error Handling**: Structured API errors with status codes and error types
5. **Configuration Loading**: Supports both array and object formats for providers
6. **Plugin System**: Dynamic loading of transformer modules from filesystem

### Configuration Service Details

The ConfigService manages multiple configuration sources with the following priority:
1. Initial config (passed to constructor)
2. JSON file (from jsonPath option)
3. .env file (if useEnvFile is true)
4. Environment variables (if useEnvironmentVariables is true)

Key features:
- Supports absolute and relative paths for config files
- Sets LOG_FILE and LOG environment variables from config
- Provides getHttpsProxy() method checking multiple proxy variables
- Config reload capability
- Configuration summary reporting

### Authentication Details

- Uses APIKEY from configuration
- Expects key in x-api-key header
- Only applies to provider management endpoints
- Health check endpoints are excluded from auth

### Error Response Format

```json
{
  "error": {
    "message": "Error message",
    "type": "api_error",
    "code": "error_code"
  }
}
```

### Provider Service Details

- Providers stored in Map by name
- Model routes support both short and full format
- Provider transformer configuration parsing:
  - Array format: ["transformer1", ["transformer2", {options}]]
  - Object format: {use: ["transformer1"], model: {use: ["transformer2"]}}
- Model availability endpoint returns OpenAI-compatible format
- Enabled/disabled state management

### Transformer Application Order

1. Transformer's `transformRequestOut` (for endpoint-specific transformers)
2. Provider's global transformers `transformRequestIn`
3. Provider's model-specific transformers `transformRequestIn`
4. Forward request to provider
5. Provider's global transformers `transformResponseOut`
6. Provider's model-specific transformers `transformResponseOut`  
7. Transformer's `transformResponseIn` (for endpoint-specific transformers)

### Streaming Response Handling

- Content-Type must include "text/event-stream"
- SSE format: `event: <event_type>\ndata: <json_data>\n\n`
- Event types: message_start, content_block_start, content_block_delta, content_block_stop, message_delta, message_stop
- Special handling for thinking blocks with signatures
- Tool call streaming with index tracking
- Proper stream cleanup on errors

### Reference Counting Logic

- Increment on Claude Code start
- Decrement on Claude Code exit
- Service auto-stops when count reaches 0
- Reference count stored in system temp directory
- Prevents premature service termination

### Claude.json Initialization

Creates default config if not exists:
```json
{
  "numStartups": 184,
  "autoUpdaterStatus": "enabled",
  "userID": "<random-64-char-hex>",
  "hasCompletedOnboarding": true,
  "lastOnboardingVersion": "1.0.17",
  "projects": {}
}
```

## Success Criteria
1. All core features work seamlessly with Claude Code
2. Single binary distribution for each platform
3. Improved performance (lower latency, higher throughput)
4. Reduced memory footprint
5. Easier installation process (no Node.js dependency)
6. Clean, maintainable Go codebase
7. Comprehensive test coverage

## Task Breakdown

### Phase 1 Tasks
- [ ] Initialize Go project structure
- [ ] Set up development environment
- [ ] Create basic CLI skeleton with Cobra
- [ ] Implement configuration types and structures

### Phase 2 Tasks
- [ ] Add all Go dependencies
- [ ] Define Anthropic API types (MessageCreateParams, Response types)
- [ ] Define OpenAI API types for transformation
- [ ] Create provider configuration types with transformer support
- [ ] Create router configuration types
- [ ] Implement configuration loading from JSON

### Phase 3 Tasks
- [ ] Set up Gin server with basic endpoints
- [ ] Implement health check endpoints
- [ ] Create services architecture (ConfigService, ProviderService, TransformerService)
- [ ] Implement pre-handler middleware for provider,model extraction
- [ ] Create proxy handler for forwarding requests with timeout and proxy support
- [ ] Implement authentication middleware with header support
- [ ] Implement routing middleware with token counting logic
- [ ] Add request/response transformation pipeline
- [ ] Implement provider CRUD endpoints
- [ ] Handle streaming responses (SSE)

### Phase 4 Tasks
- [ ] Implement start command with daemon mode
- [ ] Implement stop command with PID management
- [ ] Implement status command
- [ ] Implement code command with auto-start
- [ ] Add reference counting logic
- [ ] Handle graceful shutdown

### Phase 5 Tasks
- [ ] Implement configuration validation logic
- [ ] Integrate tiktoken-go for token counting
- [ ] Create structured logging system with file output
- [ ] Implement ~/.claude.json initialization
- [ ] Add proxy support for HTTP client
- [ ] Implement log function that respects LOG environment variable

### Phase 6 Tasks
- [ ] Define transformer interface with endpoint, transform methods
- [ ] Implement anthropic transformer (OpenAI ↔ Anthropic conversion)
- [ ] Implement deepseek transformer
- [ ] Implement gemini transformer
- [ ] Implement openrouter transformer
- [ ] Implement groq transformer (not found in source, may need research)
- [ ] Implement maxtoken transformer (config-based)
- [ ] Implement tooluse transformer
- [ ] Create plugin loading mechanism with filesystem support
- [ ] Implement transformer chaining logic

### Phase 7 Tasks
- [ ] Create Makefile with cross-compilation targets
- [ ] Set up GitHub Actions workflow
- [ ] Create release automation
- [ ] Write installation scripts
- [ ] Create distribution packages

### Testing Tasks
- [ ] Write unit tests for all components
- [ ] Create integration test suite
- [ ] Set up CI/CD pipeline
- [ ] Perform integration testing with Claude Code
- [ ] Conduct performance benchmarking
- [ ] Test on all target platforms

### Documentation Tasks
- [ ] Create README.md
- [ ] Write API documentation
- [ ] Create transformer plugin guide
- [ ] Add troubleshooting guide
- [ ] Installation and setup guide

## Exhaustive Implementation Details

### Logging System Differences

**CCProxy Logging**:
- Only logs when LOG="true" (string comparison)
- Logs to ~/.ccproxy/ccproxy.log
- Creates directory if not exists
- No console output

**@musistudio/llms Logging**:
- ALWAYS calls console.log first
- Then checks LOG="true" for file logging
- Uses LOG_FILE environment variable or defaults to "app.log"
- Both console and file output when enabled

**Decision**: Use clean logging approach (no console output, only file when LOG="true")

### Status Command Output Format

Must display exactly:
```
📊 CCProxy Status
════════════════════════════════════════
✅ Status: Running
🆔 Process ID: [pid]
🌐 Port: 3456
📡 API Endpoint: http://127.0.0.1:3456
📄 PID File: [path]

🚀 Ready to use! Run the following commands:
   ccproxy code    # Start coding with Claude
   ccproxy stop    # Stop the service
```

Or when not running:
```
❌ Status: Not Running

💡 To start the service:
   ccproxy start
```

### Tool Call Handling Complexity

The converter utility implements complex logic:
1. **Tool Response Queue**: Accumulates tool responses by tool_call_id
2. **Pending State Management**: Tracks pending tool calls and text for assistant messages
3. **Auto-injection**: If tool response missing, injects: `{"success": true, "message": "Tool call executed successfully", "tool_call_id": "[id]"}`
4. **Message Consolidation**: Combines multiple assistant blocks into single message
5. **Content Nullification**: Sets content to null when only tool_calls present

### Process Management Edge Cases

1. **PID File Cleanup**: Always cleanup on error, even if process kill fails
2. **Reference Count**: Never negative (Math.max(0, count - 1))
3. **Service Stop Messages**:
   - Success: "claude code router service has been successfully stopped."
   - Already stopped: "Failed to stop the service. It may have already been stopped."
   - No service running: "No service is currently running."

### Transformer Return Value Handling

`transformRequestIn` can return:
- Direct transformed request object
- Object with `{body: request, config: {url?, headers?, ...}}`

Config merging order:
1. Base config
2. Transformer global config
3. Model-specific transformer config
4. Final config passed to HTTP client

### Error Messages and Responses

Standard error messages:
- "Missing model in request body"
- "Provider '[name]' not found"
- "Error from provider: [error]"
- "Invalid API key"
- "Provider with name '[name]' already exists"
- "Model [model] not found. Available models: [list]"
- "Failed to start claude command: [error]"
- "Make sure Claude Code is installed: npm install -g @anthropic-ai/claude-code"

### Critical Timing and Delays

1. **Service Startup**: 1000ms initial delay, then 100ms polling
2. **Service Ready**: Additional 500ms after detection
3. **Request Timeout**: Default 60 minutes (60 * 1000 * 60)
4. **Stream Cleanup**: Always release reader lock in finally block

### Design Considerations

1. **Docker Support**: Include ENABLE_ROUTER environment variable for Docker deployments
2. **Provider Identification**: Use consistent provider.name throughout codebase
3. **Model Format**: Support both "model" and "provider,model" formats for flexibility

### Transformer-Specific Implementation Notes

**Anthropic Transformer Stream Processing**:
- Track content block indices separately
- Handle thinking blocks with signatures
- Increment index for tool calls after text content
- Complex state machine for message events
- Clean up on stream errors with proper reader.releaseLock()

**Gemini Transformer Quirks**:
- Strip $schema and additionalProperties from parameters
- Special format field handling (only keep enum, date-time)
- URL construction: `/:model:streamGenerateContent?alt=sse` or `/:model:generateContent`
- Role mapping must handle edge cases

**DeepSeek/OpenRouter Reasoning**:
- Accumulate reasoning_content separately
- Insert thinking block with signature when transitioning to content
- Track isReasoningComplete state
- Handle index increments for proper ordering

### File Path Handling

1. **Config Path Resolution**:
   - Check if absolute (starts with "/" or contains ":")
   - Otherwise join with process.cwd()
   
2. **Log File Locations**:
   - CCProxy: ~/.ccproxy/ccproxy.log
   - @musistudio/llms: LOG_FILE env or "app.log"
   - Resolution: Use ~/.ccproxy/ccproxy.log as default

### Edge Cases That Must Work

1. **Empty Tool Responses**: Auto-inject success message
2. **Multiple Assistant Blocks**: Consolidate into single message
3. **Null Content**: Set to null when only tool_calls present
4. **Missing Models**: Provide helpful error with available models list
5. **Service Already Running**: Show friendly message, don't error
6. **Malformed JSON**: Pass through original in converters
7. **Stream Interruption**: Proper cleanup with reader.releaseLock()

### Implementation Notes

1. **Dockerfile**: Create proper entrypoint for Go binary
2. **Provider Naming**: Use consistent provider.name approach
3. **ENABLE_ROUTER**: Support as optional feature flag
4. **Maxtoken Transformer**: Implement as configuration-based transformer
5. **Logging**: Implement clean logging without console output when disabled

### Critical Implementation Pitfalls to Avoid

1. **Token Counting**: Use cl100k_base tokenizer for consistent token counting
2. **Stream Event Order**: Message events must follow correct sequence for proper operation
3. **Tool Call IDs**: Must maintain consistent format and uniqueness
4. **Path Resolution**: Implement consistent absolute vs relative path logic
5. **Signal Handling**: Must clean up PID file even on unexpected exit
6. **Model Format**: Both "model" and "provider,model" must work everywhere
7. **Error Status Codes**: Must match exact codes or Claude Code error handling fails

### Implementation Priority Order

Given the complexity, implement in this order:
1. Configuration loading (foundation for everything)
2. Basic HTTP server with health endpoints
3. Provider management (CRUD operations)
4. Authentication middleware
5. Model routing logic
6. Request forwarding (without transformers)
7. Basic transformers (OpenAI, minimal)
8. Complex transformers (Anthropic with streaming)
9. Tool handling and conversion
10. Process management and CLI
11. Reference counting and auto-stop
12. Status command with exact formatting
13. Docker support and edge cases

## Critical Business Logic Summary

To ensure complete functionality, the following business logic MUST be implemented exactly:

1. **Token Counting & Routing**:
   - Use cl100k_base encoding for token counting
   - Route to longContext model when tokens > 60,000
   - Route haiku models to background configuration
   - Route based on thinking parameter presence
   - Support explicit provider,model format

2. **Transformer Pipeline**:
   - Apply transformers in specific order (request out → provider in → forward → provider out → response in)
   - Support transformer chaining with options
   - Handle both streaming and non-streaming responses
   - Handle thinking blocks and reasoning content

3. **Service Management**:
   - PID file in ~/.ccproxy/.ccproxy.pid
   - Reference counting in system temp directory
   - Auto-start service with 10s timeout
   - Clean shutdown on SIGINT/SIGTERM
   - Display status messages correctly

4. **Configuration Hierarchy**:
   - Initial config → JSON file → .env file → environment variables
   - Use APIKEY for authentication configuration
   - Force localhost when no API key configured
   - Log to file when LOG is enabled

5. **Claude Code Integration**:
   - Set correct environment variables (ANTHROPIC_AUTH_TOKEN="test" and ANTHROPIC_BASE_URL="http://127.0.0.1:3456")
   - Initialize ~/.claude.json if missing
   - Pass through command arguments correctly
   - Handle reference counting for multiple instances

6. **Error Handling**:
   - Structured error responses with type, code, message
   - Proper HTTP status codes
   - Stream error handling with cleanup
   - Timeout handling (default 60 minutes)

7. **Provider Management**:
   - Dynamic CRUD operations
   - Model route resolution (short and full format)
   - Enable/disable functionality
   - Transformer configuration parsing

### API Compatibility Requirements

**Request Format Preservation**:
- Accept both `model` and `provider,model` in request body
- Handle `thinking` parameter for model routing
- Support both streaming and non-streaming via `stream` boolean
- Support all tool-related fields

**Response Format Preservation**:
- SSE events must follow exact format: `event: [type]\ndata: [json]\n\n`
- Error responses must use exact structure with type, code, message
- Stream chunks must use correct field names and structure
- Tool calls must use consistent ID formats

**Header Handling**:
- Support both Authorization and x-api-key headers
- Use Content-Type for streams: "text/event-stream"
- Include Cache-Control and Connection headers for streams
- Pass through provider-specific headers

**Timeout Behavior**:
- Default 60 minutes for all requests
- Support custom timeout via API_TIMEOUT_MS
- Combine timeout with abort signals properly
- Clean up on timeout with proper error

This implementation plan provides a comprehensive blueprint for building ccproxy as new software inspired by Claude Code Router, incorporating the best practices and features while creating a clean, efficient Go implementation.




   