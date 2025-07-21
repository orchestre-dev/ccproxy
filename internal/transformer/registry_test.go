package transformer

import (
	"testing"
)

func TestGetRegistry(t *testing.T) {
	t.Run("SingletonBehavior", func(t *testing.T) {
		// Get registry multiple times
		reg1 := GetRegistry()
		reg2 := GetRegistry()

		if reg1 == nil {
			t.Error("Expected non-nil registry")
		}

		if reg1 != reg2 {
			t.Error("Expected same registry instance (singleton pattern)")
		}
	})

	t.Run("BuiltinTransformersRegistered", func(t *testing.T) {
		reg := GetRegistry()

		// Check that built-in transformers are registered
		expectedTransformers := []string{
			"anthropic",
			"deepseek",
			"gemini",
			"openrouter",
			"tooluse",
			"tool",
			"openai",
			"maxtoken",
			"parameters",
		}

		for _, name := range expectedTransformers {
			transformer, err := reg.Get(name)
			if err != nil {
				t.Errorf("Expected transformer %q to be registered, got error: %v", name, err)
				continue
			}

			if transformer.GetName() != name {
				t.Errorf("Expected transformer name %q, got %q", name, transformer.GetName())
			}
		}
	})

	t.Run("RegistryServiceMethods", func(t *testing.T) {
		reg := GetRegistry()

		// Test that the registry has service methods available
		if reg.transformers == nil {
			t.Error("Expected transformers map to be initialized")
		}

		if reg.chains == nil {
			t.Error("Expected chains map to be initialized")
		}

		if reg.maxCacheSize != 100 {
			t.Errorf("Expected maxCacheSize 100, got %d", reg.maxCacheSize)
		}
	})
}

func TestRegisterBuiltinTransformers(t *testing.T) {
	t.Run("RegisterToFreshService", func(t *testing.T) {
		// Create a fresh service
		service := NewService()

		// Initially should be empty
		if len(service.transformers) != 0 {
			t.Errorf("Expected empty service, got %d transformers", len(service.transformers))
		}

		// Register built-in transformers
		err := RegisterBuiltinTransformers(service)
		if err != nil {
			t.Fatalf("Unexpected error registering built-in transformers: %v", err)
		}

		// Should now have transformers
		if len(service.transformers) == 0 {
			t.Error("Expected transformers to be registered")
		}
	})

	t.Run("TransformerTypes", func(t *testing.T) {
		service := NewService()
		err := RegisterBuiltinTransformers(service)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Test specific transformer types
		tests := []struct {
			name     string
			endpoint string
		}{
			{"anthropic", "/v1/messages"},
			{"openai", "/v1/chat/completions"},
			{"gemini", "/v1beta/models/:modelAndAction"},
			{"deepseek", "/v1/chat/completions"},
			{"openrouter", "/api/v1/chat/completions"},
			{"maxtoken", ""},
			{"parameters", ""},
			{"tool", ""},
			{"tooluse", ""},
		}

		for _, test := range tests {
			transformer, err := service.Get(test.name)
			if err != nil {
				t.Errorf("Failed to get transformer %q: %v", test.name, err)
				continue
			}

			if transformer.GetName() != test.name {
				t.Errorf("Transformer %q: expected name %q, got %q", test.name, test.name, transformer.GetName())
			}

			if transformer.GetEndpoint() != test.endpoint {
				t.Errorf("Transformer %q: expected endpoint %q, got %q", test.name, test.endpoint, transformer.GetEndpoint())
			}
		}
	})

	t.Run("DuplicateRegistration", func(t *testing.T) {
		service := NewService()

		// Register once
		err := RegisterBuiltinTransformers(service)
		if err != nil {
			t.Fatalf("Unexpected error on first registration: %v", err)
		}

		// Register again - should fail
		err = RegisterBuiltinTransformers(service)
		if err == nil {
			t.Error("Expected error on duplicate registration")
		}
	})
}

func TestRegistryIntegration(t *testing.T) {
	t.Run("CreateChainWithBuiltins", func(t *testing.T) {
		reg := GetRegistry()

		// Create a chain with built-in transformers
		transformerNames := []string{"anthropic", "maxtoken"}
		chain, err := reg.CreateChainFromNames(transformerNames)
		if err != nil {
			t.Fatalf("Failed to create chain: %v", err)
		}

		if chain == nil {
			t.Error("Expected non-nil chain")
		}

		if len(chain.transformers) != 2 {
			t.Errorf("Expected 2 transformers in chain, got %d", len(chain.transformers))
		}
	})

	t.Run("GetChainForProvider", func(t *testing.T) {
		reg := GetRegistry()

		// Get chain for anthropic provider
		chain := reg.GetChainForProvider("anthropic")
		if chain == nil {
			t.Error("Expected non-nil chain for anthropic provider")
		}

		// Should have at least the anthropic transformer
		if len(chain.transformers) == 0 {
			t.Error("Expected at least one transformer in provider chain")
		}

		// First transformer should be the provider-specific one
		if len(chain.transformers) > 0 && chain.transformers[0].GetName() != "anthropic" {
			t.Errorf("Expected first transformer to be 'anthropic', got %q", chain.transformers[0].GetName())
		}
	})

	t.Run("GetByEndpoint", func(t *testing.T) {
		reg := GetRegistry()

		// Get transformers for /v1/messages endpoint
		transformers := reg.GetByEndpoint("/v1/messages")
		if len(transformers) == 0 {
			t.Error("Expected at least one transformer for /v1/messages endpoint")
		}

		// Should include anthropic transformer
		found := false
		for _, t := range transformers {
			if t.GetName() == "anthropic" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected to find anthropic transformer for /v1/messages endpoint")
		}
	})
}

func TestRegistryPanicRecovery(t *testing.T) {
	// This test verifies that if registration fails, the function panics as expected
	// We can't easily test the panic recovery without causing actual panics,
	// but we can verify the normal flow works
	t.Run("NormalFlow", func(t *testing.T) {
		// This should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic: %v", r)
			}
		}()

		reg := GetRegistry()
		if reg == nil {
			t.Error("Expected non-nil registry")
		}
	})
}
