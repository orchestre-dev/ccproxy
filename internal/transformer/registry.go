package transformer

import (
	"sync"
)

var (
	globalRegistry *Service
	registryOnce   sync.Once
)

// GetRegistry returns the global transformer registry
func GetRegistry() *Service {
	registryOnce.Do(func() {
		globalRegistry = NewService()
		
		// Register built-in transformers
		if err := RegisterBuiltinTransformers(globalRegistry); err != nil {
			// Log error but continue - registry is still usable
			// Individual transformers can be registered later
			panic(err)
		}
	})
	return globalRegistry
}

// RegisterBuiltinTransformers registers all built-in transformers
func RegisterBuiltinTransformers(service *Service) error {
	// Register Anthropic transformer
	if err := service.Register(NewAnthropicTransformer()); err != nil {
		return err
	}
	
	// TODO: Register other built-in transformers as they are implemented
	// service.Register(NewDeepseekTransformer())
	// service.Register(NewGeminiTransformer())
	// service.Register(NewOpenRouterTransformer())
	// service.Register(NewToolUseTransformer())
	// service.Register(NewMaxTokenTransformer())
	// service.Register(NewOpenAITransformer())
	
	return nil
}