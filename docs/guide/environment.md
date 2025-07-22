# Environment Variables

CCProxy uses a JSON configuration file (`config.json`) as its primary configuration method. However, it supports environment variables for specific use cases, including sensitive data like API keys and Claude Code integration.

## Environment Variable Support

### Variable Substitution in config.json

CCProxy supports environment variable substitution in configuration files using the `${VAR_NAME}` syntax:

```json
{
  "providers": [
    {
      "name": "anthropic",
      "api_key": "${ANTHROPIC_API_KEY}",
      "enabled": true
    },
    {
      "name": "openai",
      "api_key": "${OPENAI_API_KEY}",
      "enabled": true
    }
  ]
}
```

Set the environment variables:
```bash
export ANTHROPIC_API_KEY="sk-ant-..."
export OPENAI_API_KEY="sk-..."
./ccproxy start
```

### CCProxy-Specific Variables

```bash
# Override default configuration file location
export CCPROXY_CONFIG="/path/to/config.json"

# Override default port (takes precedence over config.json)
export CCPROXY_PORT=8080

# Override default host (takes precedence over config.json)
export CCPROXY_HOST="0.0.0.0"

# Set CCProxy API key for authentication
export CCPROXY_API_KEY="your-secure-api-key"

# Enable logging to file
export LOG=true
```

### Claude Code Integration Variables

When using `./ccproxy code`, the following environment variables are automatically set:

```bash
# Set by ccproxy code command
export ANTHROPIC_BASE_URL=http://127.0.0.1:3456
export ANTHROPIC_AUTH_TOKEN=test
export API_TIMEOUT_MS=600000
```

## Best Practices

### 1. Sensitive Data Management

Store API keys in environment variables rather than hardcoding in config.json:

```bash
# .env file (not committed to git)
ANTHROPIC_API_KEY="sk-ant-..."
OPENAI_API_KEY="sk-..."
GEMINI_API_KEY="AI..."
DEEPSEEK_API_KEY="sk-..."
OPENROUTER_API_KEY="sk-or-v1-..."
CCPROXY_API_KEY="your-secure-key"
```

Load variables before starting:
```bash
source .env
./ccproxy start
```

### 2. Docker Environment

When using Docker, pass environment variables:

```bash
docker run -p 3456:3456 \
  -e ANTHROPIC_API_KEY="$ANTHROPIC_API_KEY" \
  -e OPENAI_API_KEY="$OPENAI_API_KEY" \
  -v $(pwd)/config.json:/home/ccproxy/.ccproxy/config.json \
  ccproxy:latest
```

Or use Docker Compose:
```yaml
version: '3.8'
services:
  ccproxy:
    image: ccproxy:latest
    ports:
      - "3456:3456"
    environment:
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - CCPROXY_API_KEY=${CCPROXY_API_KEY}
    volumes:
      - ./config.json:/home/ccproxy/.ccproxy/config.json
```

### 3. CI/CD Integration

Use environment variables in CI/CD pipelines:

```yaml
# GitHub Actions example
- name: Run CCProxy Tests
  env:
    ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
    OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
  run: |
    ./ccproxy start --config test.config.json
    npm test
```

### 4. Production Deployment

Use a proper secrets management system:

```bash
# AWS Systems Manager Parameter Store
export ANTHROPIC_API_KEY=$(aws ssm get-parameter --name /ccproxy/anthropic-key --query 'Parameter.Value' --output text)
export OPENAI_API_KEY=$(aws ssm get-parameter --name /ccproxy/openai-key --query 'Parameter.Value' --output text)

# HashiCorp Vault
export ANTHROPIC_API_KEY=$(vault kv get -field=api_key secret/ccproxy/anthropic)
export OPENAI_API_KEY=$(vault kv get -field=api_key secret/ccproxy/openai)

# Kubernetes Secrets
kubectl create secret generic ccproxy-secrets \
  --from-literal=anthropic-api-key=$ANTHROPIC_API_KEY \
  --from-literal=openai-api-key=$OPENAI_API_KEY
```

## Environment Variable Reference

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `CCPROXY_CONFIG` | Path to configuration file | `~/.ccproxy/config.json` | `/etc/ccproxy/config.json` |
| `CCPROXY_PORT` | Override port setting | From config.json | `8080` |
| `CCPROXY_HOST` | Override host setting | From config.json | `0.0.0.0` |
| `CCPROXY_API_KEY` | API key for CCProxy authentication | None | `secure-key-123` |
| `LOG` | Enable file logging | `false` | `true` |
| `ANTHROPIC_BASE_URL` | Set by `ccproxy code` | None | `http://127.0.0.1:3456` |
| `ANTHROPIC_AUTH_TOKEN` | Set by `ccproxy code` | None | `test` |
| `API_TIMEOUT_MS` | Set by `ccproxy code` | None | `600000` |

## Viewing Current Environment

Use the `ccproxy env` command to see environment variable documentation:

```bash
./ccproxy env
```

This will display:
- All supported environment variables
- Their current values (if set)
- Usage examples

## Security Considerations

1. **Never commit `.env` files** to version control
2. **Use different API keys** for development and production
3. **Rotate API keys regularly**
4. **Limit API key permissions** to only what's needed
5. **Use secrets management** in production environments

## Troubleshooting

### Variable Not Being Read

```bash
# Check if variable is exported
echo $ANTHROPIC_API_KEY

# Ensure variable is exported, not just set
export ANTHROPIC_API_KEY="sk-ant-..."

# Verify CCProxy sees the variable
./ccproxy start --config test.json
```

### Variable Substitution Not Working

Ensure your config.json uses the correct syntax:
```json
{
  "api_key": "${ANTHROPIC_API_KEY}"  // Correct
  "api_key": "$ANTHROPIC_API_KEY"    // Wrong
  "api_key": "ANTHROPIC_API_KEY"     // Wrong
}
```

## Next Steps

- [Configuration Guide](/guide/configuration) - Full configuration reference
- [Security Guide](/guide/security) - Security best practices
- [Docker Guide](/guide/docker) - Using CCProxy with Docker