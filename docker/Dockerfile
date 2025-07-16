# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=docker -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o ccproxy \
    ./cmd/proxy

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 appgroup && \
    adduser -D -u 1001 -G appgroup appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/ccproxy .

# Change ownership to non-root user
RUN chown appuser:appgroup ccproxy

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 7187

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:7187/health || exit 1

# Run the application
CMD ["./ccproxy"]