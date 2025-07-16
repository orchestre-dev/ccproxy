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
}

// ProviderType represents the available provider types
type ProviderType string

const (
	ProviderTypeGroq       ProviderType = "groq"
	ProviderTypeOpenRouter ProviderType = "openrouter"
)

// IsValidProviderType checks if the provider type is supported
func IsValidProviderType(providerType string) bool {
	switch ProviderType(providerType) {
	case ProviderTypeGroq, ProviderTypeOpenRouter:
		return true
	default:
		return false
	}
}