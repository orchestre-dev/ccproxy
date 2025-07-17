// Package constants defines shared constants for CCProxy
package constants

import "time"

// HTTP and server constants
const (
	// Timeouts
	DefaultRequestTimeout  = 30 * time.Second
	DefaultReadTimeout     = 30 * time.Second
	DefaultWriteTimeout    = 30 * time.Second
	DefaultShutdownTimeout = 5 * time.Second

	// Health check constants
	HealthCheckTimeout = 10 * time.Second
	HealthCheckPath    = "/health"
	StatusPath         = "/status"

	// Content types
	ContentTypeJSON = "application/json"

	// Common headers
	HeaderContentType   = "Content-Type"
	HeaderAuthorization = "Authorization"
	HeaderUserAgent     = "User-Agent"
)

// Provider constants
const (
	// Provider types
	ProviderGroq       = "groq"
	ProviderOpenRouter = "openrouter"
	ProviderOpenAI     = "openai"
	ProviderXAI        = "xai"
	ProviderGemini     = "gemini"
	ProviderMistral    = "mistral"
	ProviderOllama     = "ollama"

	// Default model limits
	DefaultMaxTokens = 4096

	// Health check request content
	HealthCheckMessage = "health check"
	HealthCheckTokens  = 1
)

// API constants
const (
	// OpenAI API paths
	OpenAIChatCompletionsPath = "/chat/completions"

	// Common API endpoints
	ChatCompletionsEndpoint = "/chat/completions"

	// Request context keys
	RequestIDKey = "request_id"

	// Default values
	DefaultRequestID    = "unknown"
	DefaultFinishReason = "unknown"
)
