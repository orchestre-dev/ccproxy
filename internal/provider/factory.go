package provider

import (
	"fmt"

	"ccproxy/internal/config"
	"ccproxy/internal/provider/gemini"
	"ccproxy/internal/provider/groq"
	"ccproxy/internal/provider/mistral"
	"ccproxy/internal/provider/ollama"
	"ccproxy/internal/provider/openai"
	"ccproxy/internal/provider/openrouter"
	"ccproxy/internal/provider/xai"
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

	switch Type(f.config.Provider) {
	case ProviderTypeGroq:
		provider, err = groq.NewProvider(&f.config.Providers.Groq, f.logger)
	case ProviderTypeOpenRouter:
		provider, err = openrouter.NewProvider(&f.config.Providers.OpenRouter, f.logger)
	case ProviderTypeOpenAI:
		provider, err = openai.NewProvider(&f.config.Providers.OpenAI, f.logger)
	case ProviderTypeXAI:
		provider, err = xai.NewProvider(&f.config.Providers.XAI, f.logger)
	case ProviderTypeGemini:
		provider, err = gemini.NewProvider(&f.config.Providers.Gemini, f.logger)
	case ProviderTypeMistral:
		provider, err = mistral.NewProvider(&f.config.Providers.Mistral, f.logger)
	case ProviderTypeOllama:
		provider, err = ollama.NewProvider(&f.config.Providers.Ollama, f.logger)
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
		string(ProviderTypeOpenAI),
		string(ProviderTypeXAI),
		string(ProviderTypeGemini),
		string(ProviderTypeMistral),
		string(ProviderTypeOllama),
	}
}
