package transformer

import (
	"sync"
	"testing"
)

func TestGetRegistry(t *testing.T) {
	// Reset the registry for testing
	globalRegistry = nil
	registryOnce = sync.Once{}

	// First call should initialize the registry
	registry1 := GetRegistry()
	if registry1 == nil {
		t.Fatal("Expected non-nil registry")
	}

	// Subsequent calls should return the same instance
	registry2 := GetRegistry()
	if registry1 != registry2 {
		t.Error("Expected same registry instance")
	}

	// Verify built-in transformers are registered
	transformers := []string{
		"anthropic",
		"deepseek",
		"gemini",
		"openrouter",
		"tooluse",  // Note: transformer name is "tooluse", not "tool-use"
		"tool",
		"openai",
	}

	for _, name := range transformers {
		trans, err := registry1.Get(name)
		if err != nil {
			t.Errorf("Expected transformer %s to be registered, got error: %v", name, err)
		}
		if trans == nil {
			t.Errorf("Expected non-nil transformer for %s", name)
		}
	}
}

func TestRegisterBuiltinTransformers(t *testing.T) {
	// Create a fresh service
	service := NewService()

	// Register built-in transformers
	err := RegisterBuiltinTransformers(service)
	if err != nil {
		t.Fatalf("Failed to register built-in transformers: %v", err)
	}

	// Verify all transformers are registered
	tests := []struct {
		name     string
		expected string
	}{
		{"anthropic", "anthropic"},
		{"deepseek", "deepseek"},
		{"gemini", "gemini"},
		{"openrouter", "openrouter"},
		{"tooluse", "tooluse"},
		{"tool", "tool"},
		{"openai", "openai"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trans, err := service.Get(tt.name)
			if err != nil {
				t.Errorf("Failed to get transformer %s: %v", tt.name, err)
			}
			if trans == nil {
				t.Errorf("Expected non-nil transformer for %s", tt.name)
			}
			if trans != nil && trans.GetName() != tt.expected {
				t.Errorf("Expected transformer name %s, got %s", tt.expected, trans.GetName())
			}
		})
	}
}

func TestRegisterBuiltinTransformers_DuplicateRegistration(t *testing.T) {
	// Create a service and register transformers twice
	service := NewService()

	// First registration should succeed
	err := RegisterBuiltinTransformers(service)
	if err != nil {
		t.Fatalf("First registration failed: %v", err)
	}

	// Second registration should fail due to duplicates
	err = RegisterBuiltinTransformers(service)
	if err == nil {
		t.Error("Expected error for duplicate registration")
	}
}