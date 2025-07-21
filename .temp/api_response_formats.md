# CCProxy API Response Formats and Behaviors

## 1. Health Endpoints

### Root Endpoint (`/`)

**Method:** GET  
**Response:**
```json
{
  "message": "LLMs API",
  "version": "1.0.0"
}
```

**Behavior:**
- Simple health check endpoint
- No authentication required
- Returns immediately with service status

### Health Endpoint (`/health`)

**Method:** GET  
**Response:**
```json
{
  "status": "ok",
  "timestamp": "2025-01-17T10:30:00.000Z"
}
```

**Behavior:**
- Returns current timestamp in ISO 8601 format
- No authentication required
- Used for service monitoring

## 2. Messages Endpoint (`/v1/messages`)

**Method:** POST  
**Endpoint Handler:** Implemented via the Anthropic transformer

### Request Format:
```json
{
  "model": "claude-3-sonnet",
  "messages": [
    {
      "role": "user",
      "content": "string or array of content blocks"
    }
  ],
  "max_tokens": 100,
  "temperature": 0.7,
  "stream": false,
  "system": "string or array",
  "tools": [],
  "thinking": false
}
```

### Special Behaviors:
1. **Model Routing:** The router middleware (`router.ts`) implements intelligent model selection:
   - If model contains comma, it's used as-is (provider,model format)
   - If token count > 60,000 and `config.Router.longContext` is set, uses long context model
   - If model starts with "claude-3-5-haiku" and `config.Router.background` is set, uses background model
   - If `thinking` is true and `config.Router.think` is set, uses think model
   - Otherwise uses `config.Router.default`

2. **Token Counting:** The router counts tokens for:
   - Message content (text, tool_use, tool_result)
   - System messages
   - Tool definitions (name + description + input_schema)

3. **Request Transformation:** The Anthropic transformer converts:
   - System messages to appropriate format
   - Tool results to proper tool messages
   - Assistant messages with tool calls

4. **Response Handling:**
   - For streaming requests: Returns `text/event-stream` with appropriate headers
   - For non-streaming: Returns JSON response from provider

## 3. Provider Management API

### List Providers
**Method:** GET `/providers`  
**Authentication:** Required (uses authMiddleware)  
**Response:** Array of provider objects

### Create Provider
**Method:** POST `/providers`  
**Authentication:** Required  
**Request Body:**
```json
{
  "id": "string",
  "name": "string",
  "type": "openai|anthropic",
  "baseUrl": "string",
  "apiKey": "string",
  "models": ["string"]
}
```
**Validation:**
- Name must not be empty
- BaseUrl must be valid URL
- API key required
- At least one model required
- Provider name must be unique

**Response:** Created provider object

### Get Provider
**Method:** GET `/providers/:id`  
**Authentication:** Required  
**Response:** Provider object or 404 error

### Update Provider
**Method:** PUT `/providers/:id`  
**Authentication:** Required  
**Request Body:** Partial provider object  
**Response:** Updated provider object or 404 error

### Delete Provider
**Method:** DELETE `/providers/:id`  
**Authentication:** Required  
**Response:**
```json
{
  "message": "Provider deleted successfully"
}
```

### Toggle Provider
**Method:** PATCH `/providers/:id/toggle`  
**Authentication:** Required  
**Request Body:**
```json
{
  "enabled": boolean
}
```
**Response:**
```json
{
  "message": "Provider enabled successfully"
}
```
or
```json
{
  "message": "Provider disabled successfully"
}
```

## 4. Middleware and Special Handling

### Authentication Middleware
- Applied to all provider management endpoints
- Not applied to health endpoints (/, /health)
- Not applied to /v1/messages endpoint

### CORS Middleware
- Allows all origins (*)
- Supports methods: GET, POST, PUT, DELETE, OPTIONS, PATCH
- Handles preflight requests

### Model Provider Middleware
- Extracts provider from model string (format: "provider,model")
- Sets `req.provider` for use in route handlers
- Returns 400 error if model is missing

### Error Handling
- Custom error handler using `createApiError` function
- Returns structured error responses with:
  - Status code
  - Error message
  - Error code (e.g., "provider_not_found")

### Request Flow for /v1/messages:
1. Model provider middleware extracts provider from model
2. Router middleware counts tokens and selects appropriate model
3. Anthropic transformer converts request to provider format
4. Provider-specific transformers apply if configured
5. Request sent to provider with proper authentication
6. Response transformed back through provider transformers
7. Final response transformed by Anthropic transformer
8. Streaming or JSON response returned to client

## 5. Key Differences from Go Implementation

The current Go implementation in `internal/server/server.go`:
- Has placeholder implementations for most endpoints
- Returns simpler responses without all fields
- Doesn't implement the complex request transformation logic
- Missing token counting and model routing logic
- Doesn't handle streaming responses
- Provider management endpoints are not fully implemented

To replicate the JavaScript behavior, the Go implementation needs:
1. Full request/response transformation logic
2. Token counting implementation
3. Model routing based on configuration
4. Streaming response support
5. Complete provider management CRUD operations
6. Proper error handling with structured responses