// Package provider defines the interface and types for AI provider implementations
package provider

import (
	"context"

	"ccproxy/internal/models"
)

// Provider represents a generic AI provider interface
type Provider interface {
	// CreateChatCompletion sends a chat completion request to the provider
	CreateChatCompletion(ctx context.Context, req *models.ChatCompletionRequest) (*models.ChatCompletionResponse, error)

	// GetName returns the provider name (e.g., "groq", "openrouter")
	GetName() string

	// GetModel returns the configured model for this provider
	GetModel() string

	// GetMaxTokens returns the maximum tokens allowed for this provider
	GetMaxTokens() int

	// ValidateConfig validates the provider configuration
	ValidateConfig() error

	// GetBaseURL returns the API base URL for this provider
	GetBaseURL() string

	// HealthCheck performs a health check on the provider
	HealthCheck(ctx context.Context) error
}

// Type represents the available provider types
type Type string

// Provider type constants
const (
	ProviderTypeGroq       Type = "groq"
	ProviderTypeOpenRouter Type = "openrouter"
	ProviderTypeOpenAI     Type = "openai"
	ProviderTypeXAI        Type = "xai"
	ProviderTypeGemini     Type = "gemini"
	ProviderTypeMistral    Type = "mistral"
	ProviderTypeOllama     Type = "ollama"
)

// IsValidProviderType checks if the provider type is supported
func IsValidProviderType(providerType string) bool {
	switch Type(providerType) {
	case ProviderTypeGroq,
		ProviderTypeOpenRouter,
		ProviderTypeOpenAI,
		ProviderTypeXAI,
		ProviderTypeGemini,
		ProviderTypeMistral,
		ProviderTypeOllama:
		return true
	default:
		return false
	}
}
