package transformer

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/orchestre-dev/ccproxy/internal/config"
	"github.com/orchestre-dev/ccproxy/internal/utils"
)

// Service manages transformers and their lifecycle
type Service struct {
	transformers map[string]Transformer
	chains       map[string]*TransformerChain
	mu           sync.RWMutex
}

// NewService creates a new transformer service
func NewService() *Service {
	return &Service{
		transformers: make(map[string]Transformer),
		chains:       make(map[string]*TransformerChain),
	}
}

// Register registers a transformer
func (s *Service) Register(transformer Transformer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	name := transformer.GetName()
	if _, exists := s.transformers[name]; exists {
		return fmt.Errorf("transformer already registered: %s", name)
	}
	
	s.transformers[name] = transformer
	utils.GetLogger().Debugf("Registered transformer: %s", name)
	return nil
}

// Get retrieves a transformer by name
func (s *Service) Get(name string) (Transformer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	transformer, exists := s.transformers[name]
	if !exists {
		return nil, fmt.Errorf("transformer not found: %s", name)
	}
	
	return transformer, nil
}

// GetByEndpoint retrieves transformers that handle a specific endpoint
func (s *Service) GetByEndpoint(endpoint string) []Transformer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var result []Transformer
	for _, t := range s.transformers {
		if t.GetEndpoint() == endpoint {
			result = append(result, t)
		}
	}
	
	return result
}

// CreateChain creates a transformer chain from configuration
func (s *Service) CreateChain(configs []config.TransformerConfig) (*TransformerChain, error) {
	chain := NewTransformerChain()
	
	for _, cfg := range configs {
		transformer, err := s.Get(cfg.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get transformer %s: %w", cfg.Name, err)
		}
		
		// TODO: Apply transformer-specific configuration from cfg.Config
		
		chain.Add(transformer)
	}
	
	return chain, nil
}

// CreateChainFromNames creates a transformer chain from transformer names
func (s *Service) CreateChainFromNames(names []string) (*TransformerChain, error) {
	chain := NewTransformerChain()
	
	for _, name := range names {
		transformer, err := s.Get(name)
		if err != nil {
			return nil, fmt.Errorf("failed to get transformer %s: %w", name, err)
		}
		
		chain.Add(transformer)
	}
	
	return chain, nil
}

// GetChainForProvider gets the transformer chain for a provider by name
func (s *Service) GetChainForProvider(providerName string) (*TransformerChain, error) {
	s.mu.RLock()
	chainKey := fmt.Sprintf("provider:%s", providerName)
	chain, exists := s.chains[chainKey]
	s.mu.RUnlock()
	
	if exists {
		return chain, nil
	}
	
	return nil, fmt.Errorf("no transformer chain for provider: %s", providerName)
}

// GetOrCreateChain gets or creates a transformer chain for a provider
func (s *Service) GetOrCreateChain(provider *config.Provider) (*TransformerChain, error) {
	// First check if chain exists with read lock
	s.mu.RLock()
	chainKey := fmt.Sprintf("provider:%s", provider.Name)
	chain, exists := s.chains[chainKey]
	s.mu.RUnlock()
	
	if exists {
		return chain, nil
	}
	
	// Create new chain (without holding lock)
	chain, err := s.CreateChain(provider.Transformers)
	if err != nil {
		return nil, err
	}
	
	// Cache the chain with write lock
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Double-check in case another goroutine created it
	if existingChain, exists := s.chains[chainKey]; exists {
		return existingChain, nil
	}
	
	s.chains[chainKey] = chain
	return chain, nil
}

// ApplyRequestTransformation applies request transformations for a provider
func (s *Service) ApplyRequestTransformation(ctx context.Context, provider *config.Provider, request interface{}) (interface{}, error) {
	chain, err := s.GetOrCreateChain(provider)
	if err != nil {
		return nil, err
	}
	
	result, err := chain.TransformRequestIn(ctx, request, provider.Name)
	if err != nil {
		utils.LogTransformer(provider.Name, "request", false, err)
		return nil, err
	}
	
	utils.LogTransformer(provider.Name, "request", true, nil)
	return result, nil
}

// ApplyResponseTransformation applies response transformations for a provider
func (s *Service) ApplyResponseTransformation(ctx context.Context, provider *config.Provider, response interface{}) (interface{}, error) {
	chain, err := s.GetOrCreateChain(provider)
	if err != nil {
		return nil, err
	}
	
	// For HTTP responses, we need to handle them differently
	if httpResp, ok := response.(*Response); ok {
		result, err := chain.TransformResponseOut(ctx, httpResp.Response)
		if err != nil {
			utils.LogTransformer(provider.Name, "response", false, err)
			return nil, err
		}
		
		utils.LogTransformer(provider.Name, "response", true, nil)
		return &Response{Response: result}, nil
	}
	
	// For other types, log and return as-is
	utils.LogTransformer(provider.Name, "response", true, nil)
	return response, nil
}

// Response wraps http.Response for type safety
type Response struct {
	Response *http.Response
}

// ParseTransformerConfig parses transformer configuration from various formats
func ParseTransformerConfig(input interface{}) ([]config.TransformerConfig, error) {
	switch v := input.(type) {
	case string:
		// Simple string format: "transformer1"
		return []config.TransformerConfig{{Name: v}}, nil
		
	case []interface{}:
		// Array format: ["transformer1", ["transformer2", {options}]]
		var configs []config.TransformerConfig
		for _, item := range v {
			switch t := item.(type) {
			case string:
				configs = append(configs, config.TransformerConfig{Name: t})
			case []interface{}:
				if len(t) >= 1 {
					if name, ok := t[0].(string); ok {
						cfg := config.TransformerConfig{Name: name}
						if len(t) >= 2 {
							if options, ok := t[1].(map[string]interface{}); ok {
								cfg.Config = options
							}
						}
						configs = append(configs, cfg)
					}
				}
			}
		}
		return configs, nil
		
	case map[string]interface{}:
		// Object format: {use: ["transformer1"], model: {use: ["transformer2"]}}
		var configs []config.TransformerConfig
		
		// Global transformers
		if use, ok := v["use"].([]interface{}); ok {
			for _, t := range use {
				if name, ok := t.(string); ok {
					configs = append(configs, config.TransformerConfig{Name: name})
				}
			}
		}
		
		// Model-specific transformers
		// TODO: Handle model-specific transformer configuration
		
		return configs, nil
		
	default:
		return nil, fmt.Errorf("invalid transformer configuration format: %T", input)
	}
}