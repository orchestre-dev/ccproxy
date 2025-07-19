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
		RegisterBuiltinTransformers(globalRegistry)
	})
	return globalRegistry
}

// RegisterBuiltinTransformers registers all built-in transformers
func RegisterBuiltinTransformers(service *Service) {
	// TODO: Register built-in transformers as they are implemented
	// service.Register(NewAnthropicTransformer())
	// service.Register(NewDeepseekTransformer())
	// service.Register(NewGeminiTransformer())
	// service.Register(NewOpenRouterTransformer())
	// service.Register(NewToolUseTransformer())
	// service.Register(NewMaxTokenTransformer())
	// service.Register(NewOpenAITransformer())
}