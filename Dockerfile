# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o ccproxy ./cmd/ccproxy

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 -S ccproxy && \
    adduser -u 1000 -S ccproxy -G ccproxy

# Create necessary directories
RUN mkdir -p /home/ccproxy/.ccproxy && \
    chown -R ccproxy:ccproxy /home/ccproxy

# Copy binary from builder
COPY --from=builder /build/ccproxy /usr/local/bin/ccproxy

# Set user
USER ccproxy

# Set working directory
WORKDIR /home/ccproxy

# Expose default port
EXPOSE 3456

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ccproxy status || exit 1

# Default command
ENTRYPOINT ["ccproxy"]
CMD ["start", "--foreground"]