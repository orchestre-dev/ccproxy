# Environment Variables

CCProxy is configured entirely through environment variables. This guide covers all available configuration options.

## Core Configuration

### Provider Selection

```bash
# Select the AI provider to use
PROVIDER=groq  # Options: groq, openai, gemini, mistral, ollama, xai, openrouter
```

### Server Configuration

```bash
# Server settings
SERVER_HOST=0.0.0.0           # Host to bind to
SERVER_PORT=3456              # Port to listen on
SERVER_ENVIRONMENT=production # Environment: development, production
```

## Provider-Specific Configuration

### Groq Configuration

```bash
GROQ_API_KEY=your_api_key_here
GROQ_BASE_URL=https://api.groq.com/openai/v1
GROQ_MODEL=moonshotai/kimi-k2-instruct
GROQ_MAX_TOKENS=16384
```

### OpenAI Configuration

```bash
OPENAI_API_KEY=your_api_key_here
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_MODEL=gpt-4o
OPENAI_MAX_TOKENS=4096
```

### Gemini Configuration

```bash
GEMINI_API_KEY=your_api_key_here
GEMINI_BASE_URL=https://generativelanguage.googleapis.com
GEMINI_MODEL=gemini-2.0-flash
GEMINI_MAX_TOKENS=32768
```

### Mistral Configuration

```bash
MISTRAL_API_KEY=your_api_key_here
MISTRAL_BASE_URL=https://api.mistral.ai/v1
MISTRAL_MODEL=mistral-large-latest
MISTRAL_MAX_TOKENS=32768
```

### XAI Configuration

```bash
XAI_API_KEY=your_api_key_here
XAI_BASE_URL=https://api.x.ai/v1
XAI_MODEL=grok-beta
XAI_MAX_TOKENS=128000
```

### Ollama Configuration

```bash
OLLAMA_API_KEY=ollama  # Default value
OLLAMA_BASE_URL=http://localhost:11434
OLLAMA_MODEL=llama3.2
OLLAMA_MAX_TOKENS=4096
```

### OpenRouter Configuration

```bash
OPENROUTER_API_KEY=your_api_key_here
OPENROUTER_BASE_URL=https://openrouter.ai/api/v1
OPENROUTER_MODEL=openai/gpt-4o
OPENROUTER_MAX_TOKENS=4096
OPENROUTER_SITE_URL=https://yoursite.com
OPENROUTER_SITE_NAME=Your Site
```

## Logging Configuration

```bash
LOG_LEVEL=info    # Options: debug, info, warn, error
LOG_FORMAT=json   # Options: json, text
```

## Security Configuration

```bash
# CORS settings
CORS_ALLOWED_ORIGINS=https://yoursite.com,https://anotherdomain.com
```

## Example Configuration Files

### Development (.env.development)

```bash
PROVIDER=groq
GROQ_API_KEY=your_development_key
SERVER_ENVIRONMENT=development
LOG_LEVEL=debug
```

### Production (.env.production)

```bash
PROVIDER=groq
GROQ_API_KEY=your_production_key
SERVER_ENVIRONMENT=production
LOG_LEVEL=info
LOG_FORMAT=json
```

## Configuration Validation

CCProxy validates all configuration at startup. If required environment variables are missing or invalid, the application will exit with an error message.

## Best Practices

1. **Use separate API keys** for development and production
2. **Store secrets securely** - never commit API keys to version control
3. **Use environment-specific configurations** for different deployment stages
4. **Monitor API usage** to avoid hitting rate limits
5. **Set appropriate timeouts** based on your use case