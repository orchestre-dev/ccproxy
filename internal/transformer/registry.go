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
	
	// Register DeepSeek transformer
	if err := service.Register(NewDeepSeekTransformer()); err != nil {
		return err
	}
	
	// Register Gemini transformer
	if err := service.Register(NewGeminiTransformer()); err != nil {
		return err
	}
	
	// Register OpenRouter transformer
	if err := service.Register(NewOpenRouterTransformer()); err != nil {
		return err
	}
	
	// Register ToolUse transformer
	if err := service.Register(NewToolUseTransformer()); err != nil {
		return err
	}
	
	// Register Tool transformer (for general tool handling)
	if err := service.Register(NewToolTransformer()); err != nil {
		return err
	}
	
	// Register OpenAI transformer
	if err := service.Register(NewOpenAITransformer()); err != nil {
		return err
	}
	
	// TODO: Register MaxToken transformer when implemented
	// service.Register(NewMaxTokenTransformer())
	
	return nil
}