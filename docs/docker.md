# Docker Support for CCProxy

CCProxy provides comprehensive Docker support for easy deployment and development.

## Quick Start

### Using Docker

```bash
# Pull the image
docker pull ghcr.io/orchestre-dev/ccproxy:latest

# Run with minimal configuration
docker run -d \
  --name ccproxy \
  -p 3456:3456 \
  -e CCPROXY_API_KEY=your-api-key \
  -e CCPROXY_PROVIDERS_0_NAME=anthropic \
  -e CCPROXY_PROVIDERS_0_API_KEY=your-anthropic-key \
  ghcr.io/orchestre-dev/ccproxy:latest
```

### Using Docker Compose

```bash
# Copy example environment file
cp .env.example .env

# Edit .env with your API keys
vim .env

# Start services
docker-compose up -d

# View logs
docker-compose logs -f
```

## Available Images

- `ccproxy:latest` - Latest stable release
- `ccproxy:v1.0.0` - Specific version
- `ccproxy:develop` - Development branch (unstable)

### Supported Architectures

- `linux/amd64` - Standard x86-64
- `linux/arm64` - ARM64 (Apple Silicon, AWS Graviton)

## Configuration

### Environment Variables

CCProxy can be configured entirely through environment variables:

```bash
# Basic configuration
CCPROXY_HOST=0.0.0.0
CCPROXY_PORT=3456
CCPROXY_API_KEY=your-api-key
CCPROXY_LOG=true
CCPROXY_LOG_FILE=/logs/ccproxy.log

# Provider configuration
CCPROXY_PROVIDERS_0_NAME=anthropic
CCPROXY_PROVIDERS_0_API_BASE_URL=https://api.anthropic.com
CCPROXY_PROVIDERS_0_API_KEY=your-anthropic-key
CCPROXY_PROVIDERS_0_ENABLED=true

# Performance settings
CCPROXY_PERFORMANCE_METRICS_ENABLED=true
CCPROXY_PERFORMANCE_RATE_LIMIT_ENABLED=true
CCPROXY_PERFORMANCE_RATE_LIMIT_REQUESTS_PER_MIN=1000
```

### Using Configuration File

Mount a configuration file:

```bash
docker run -d \
  --name ccproxy \
  -p 3456:3456 \
  -v $(pwd)/config.json:/home/ccproxy/.ccproxy/config.json:ro \
  ghcr.io/orchestre-dev/ccproxy:latest
```

## Docker Compose

### Basic Setup

```yaml
version: '3.8'

services:
  ccproxy:
    image: ghcr.io/orchestre-dev/ccproxy:latest
    container_name: ccproxy
    ports:
      - "3456:3456"
    environment:
      - CCPROXY_API_KEY=${CCPROXY_API_KEY}
      - CCPROXY_PROVIDERS_0_API_KEY=${ANTHROPIC_API_KEY}
    volumes:
      - ./config.json:/home/ccproxy/.ccproxy/config.json:ro
      - ccproxy-logs:/home/ccproxy/.ccproxy/logs
    restart: unless-stopped

volumes:
  ccproxy-logs:
```

### With Monitoring

```bash
# Start with monitoring stack
docker-compose --profile monitoring up -d

# Access services
# - CCProxy: http://localhost:3456
# - Prometheus: http://localhost:9090
# - Grafana: http://localhost:3000 (admin/admin)
```

## Development

### Hot Reload Development

```bash
# Start development environment with hot reload
docker-compose -f docker-compose.dev.yml up

# Code changes will automatically trigger rebuilds
```

### Building Images

```bash
# Build for current platform
make docker-build

# Build multi-architecture image
make docker-build-multiarch

# Push to registry
make docker-push
```

## Health Checks

The Docker image includes health checks:

```bash
# Check container health
docker inspect ccproxy --format='{{.State.Health.Status}}'

# View health check logs
docker inspect ccproxy --format='{{range .State.Health.Log}}{{.Output}}{{end}}'
```

## Volumes and Persistence

- `/home/ccproxy/.ccproxy` - Configuration and runtime data
- `/home/ccproxy/.ccproxy/logs` - Log files (if file logging enabled)

## Security

- Runs as non-root user (UID 1000)
- Minimal Alpine Linux base image
- No unnecessary packages or tools
- Read-only root filesystem compatible

### Running with Read-Only Root

```bash
docker run -d \
  --name ccproxy \
  --read-only \
  --tmpfs /tmp \
  -v ccproxy-data:/home/ccproxy/.ccproxy \
  -p 3456:3456 \
  ghcr.io/orchestre-dev/ccproxy:latest
```

## Resource Limits

### Docker Run

```bash
docker run -d \
  --name ccproxy \
  --memory="1g" \
  --memory-swap="1g" \
  --cpus="1.0" \
  -p 3456:3456 \
  ghcr.io/orchestre-dev/ccproxy:latest
```

### Docker Compose

```yaml
services:
  ccproxy:
    # ... other configuration ...
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '0.5'
          memory: 512M
```

## Troubleshooting

### View Logs

```bash
# Docker logs
docker logs ccproxy

# Follow logs
docker logs -f ccproxy

# Last 100 lines
docker logs --tail 100 ccproxy
```

### Debug Mode

```bash
docker run -it --rm \
  --entrypoint sh \
  ghcr.io/orchestre-dev/ccproxy:latest
```

### Common Issues

1. **Port already in use**
   ```bash
   # Change host port
   docker run -p 8080:3456 ...
   ```

2. **Permission denied**
   ```bash
   # Ensure volumes have correct permissions
   sudo chown -R 1000:1000 ./ccproxy-data
   ```

3. **Cannot connect to provider**
   ```bash
   # Check DNS and network
   docker exec ccproxy nslookup api.anthropic.com
   ```

## Integration

### Kubernetes

Kubernetes deployment documentation coming soon.

### CI/CD

The repository includes GitHub Actions workflows for:
- Building multi-architecture images
- Running tests in containers
- Publishing to GitHub Container Registry
- Automated security scanning