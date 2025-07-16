package provider

import (
	"fmt"

	"ccproxy/internal/config"
	"ccproxy/internal/provider/groq"
	"ccproxy/internal/provider/openrouter"
	"ccproxy/pkg/logger"
)

// Factory creates providers based on configuration
type Factory struct {
	config *config.Config
	logger *logger.Logger
}

// NewFactory creates a new provider factory
func NewFactory(config *config.Config, logger *logger.Logger) *Factory {
	return &Factory{
		config: config,
		logger: logger,
	}
}

// CreateProvider creates a provider based on the configuration
func (f *Factory) CreateProvider() (Provider, error) {
	if !IsValidProviderType(f.config.Provider) {
		return nil, fmt.Errorf("unsupported provider type: %s", f.config.Provider)
	}

	var provider Provider
	var err error

	switch ProviderType(f.config.Provider) {
	case ProviderTypeGroq:
		provider, err = groq.NewProvider(&f.config.Providers.Groq, f.logger)
	case ProviderTypeOpenRouter:
		provider, err = openrouter.NewProvider(&f.config.Providers.OpenRouter, f.logger)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", f.config.Provider)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create %s provider: %w", f.config.Provider, err)
	}

	// Validate the provider configuration
	if err := provider.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid %s provider configuration: %w", f.config.Provider, err)
	}

	f.logger.Infof("Successfully initialized %s provider with model %s", 
		provider.GetName(), provider.GetModel())

	return provider, nil
}

// GetAvailableProviders returns a list of available provider types
func GetAvailableProviders() []string {
	return []string{
		string(ProviderTypeGroq),
		string(ProviderTypeOpenRouter),
	}
}