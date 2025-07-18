# Docker Deployment

This guide covers deploying CCProxy using Docker containers.

## Quick Start with Docker

### Using Docker Compose

```yaml
version: '3.8'

services:
  ccproxy:
    image: ccproxy:latest
    ports:
      - "7187:7187"
    environment:
      - PROVIDER=groq
      - GROQ_API_KEY=your_api_key_here
      - SERVER_ENVIRONMENT=production
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:7187/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

### Running with Docker

```bash
docker run -d \
  --name ccproxy \
  -p 7187:7187 \
  -e PROVIDER=groq \
  -e GROQ_API_KEY=your_api_key_here \
  ccproxy:latest
```

## Building from Source

```bash
# Clone the repository
git clone https://github.com/orchestre-dev/ccproxy.git
cd ccproxy

# Build the Docker image
docker build -t ccproxy:latest -f docker/Dockerfile .

# Run the container
docker run -d -p 7187:7187 -e PROVIDER=groq -e GROQ_API_KEY=your_key ccproxy:latest
```

## Configuration

All configuration is done via environment variables. See the [Configuration Guide](/guide/configuration) for details.

## Health Checks

The Docker image includes built-in health checks:

```bash
# Check container health
docker ps

# View health check logs
docker inspect ccproxy | grep -A 20 "Health"
```

## Production Deployment

For production deployments, consider:

- Using secrets management for API keys
- Setting up proper logging and monitoring
- Configuring resource limits
- Using a reverse proxy (nginx, Traefik)

## Troubleshooting

### Common Issues

1. **Container exits immediately**: Check environment variables and logs
2. **Connection refused**: Ensure ports are properly mapped
3. **Provider errors**: Verify API keys and provider configuration

### Viewing Logs

```bash
# View container logs
docker logs ccproxy

# Follow logs in real-time
docker logs -f ccproxy
```