package providers

import (
	"fmt"
	"sync"

	"github.com/orchestre-dev/ccproxy/internal/config"
)

// Registry maintains a global registry of provider clients
type Registry struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

// Provider interface that all provider implementations must satisfy
type Provider interface {
	// GetName returns the provider name
	GetName() string

	// IsHealthy checks if the provider is operational
	IsHealthy() bool

	// SupportsModel checks if the provider supports a specific model
	SupportsModel(model string) bool

	// GetModels returns all supported models
	GetModels() []string
}

// BaseProvider provides common functionality for all providers
type BaseProvider struct {
	Config *config.Provider
}

func (p *BaseProvider) GetName() string {
	return p.Config.Name
}

func (p *BaseProvider) IsHealthy() bool {
	return p.Config.Enabled
}

func (p *BaseProvider) SupportsModel(model string) bool {
	for _, m := range p.Config.Models {
		if m == model {
			return true
		}
	}
	return false
}

func (p *BaseProvider) GetModels() []string {
	return p.Config.Models
}

var (
	globalRegistry *Registry
	registryOnce   sync.Once
)

// GetRegistry returns the global provider registry
func GetRegistry() *Registry {
	registryOnce.Do(func() {
		globalRegistry = &Registry{
			providers: make(map[string]Provider),
		}
	})
	return globalRegistry
}

// Register adds a provider to the registry
func (r *Registry) Register(provider Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := provider.GetName()
	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider already registered: %s", name)
	}

	r.providers[name] = provider
	return nil
}

// Unregister removes a provider from the registry
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.providers, name)
}

// Get retrieves a provider by name
func (r *Registry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", name)
	}

	return provider, nil
}

// GetAll returns all registered providers
func (r *Registry) GetAll() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]Provider, 0, len(r.providers))
	for _, p := range r.providers {
		providers = append(providers, p)
	}

	return providers
}

// GetByModel returns all providers that support a specific model
func (r *Registry) GetByModel(model string) []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var providers []Provider
	for _, p := range r.providers {
		if p.SupportsModel(model) {
			providers = append(providers, p)
		}
	}

	return providers
}

// Clear removes all providers from the registry
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.providers = make(map[string]Provider)
}
