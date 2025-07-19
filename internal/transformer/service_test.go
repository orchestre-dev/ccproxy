package transformer

import (
	"context"
	"net/http"
	"testing"

	"github.com/musistudio/ccproxy/internal/config"
)

func TestService_Register(t *testing.T) {
	service := NewService()
	
	// Register transformer
	transformer := NewMockTransformer("test-transformer")
	err := service.Register(transformer)
	if err != nil {
		t.Errorf("Failed to register transformer: %v", err)
	}
	
	// Try to register duplicate
	err = service.Register(transformer)
	if err == nil {
		t.Error("Expected error registering duplicate transformer")
	}
	
	// Get registered transformer
	retrieved, err := service.Get("test-transformer")
	if err != nil {
		t.Errorf("Failed to get transformer: %v", err)
	}
	if retrieved != transformer {
		t.Error("Retrieved transformer doesn't match registered one")
	}
}

func TestService_GetByEndpoint(t *testing.T) {
	service := NewService()
	
	// Register transformers with different endpoints
	t1 := NewBaseTransformer("t1", "/v1/messages")
	t2 := NewBaseTransformer("t2", "/v1/messages")
	t3 := NewBaseTransformer("t3", "/v1/chat/completions")
	t4 := NewBaseTransformer("t4", "")
	
	service.Register(t1)
	service.Register(t2)
	service.Register(t3)
	service.Register(t4)
	
	// Get transformers for /v1/messages
	transformers := service.GetByEndpoint("/v1/messages")
	if len(transformers) != 2 {
		t.Errorf("Expected 2 transformers for /v1/messages, got %d", len(transformers))
	}
	
	// Get transformers for /v1/chat/completions
	transformers = service.GetByEndpoint("/v1/chat/completions")
	if len(transformers) != 1 {
		t.Errorf("Expected 1 transformer for /v1/chat/completions, got %d", len(transformers))
	}
	
	// Get transformers for non-existent endpoint
	transformers = service.GetByEndpoint("/v1/unknown")
	if len(transformers) != 0 {
		t.Errorf("Expected 0 transformers for /v1/unknown, got %d", len(transformers))
	}
}

func TestService_CreateChain(t *testing.T) {
	service := NewService()
	
	// Register transformers
	t1 := NewMockTransformer("transformer1")
	t2 := NewMockTransformer("transformer2")
	service.Register(t1)
	service.Register(t2)
	
	// Create chain from configs
	configs := []config.TransformerConfig{
		{Name: "transformer1"},
		{Name: "transformer2"},
	}
	
	chain, err := service.CreateChain(configs)
	if err != nil {
		t.Fatalf("Failed to create chain: %v", err)
	}
	
	// Test chain works
	ctx := context.Background()
	result, err := chain.TransformRequestIn(ctx, "test", "provider")
	if err != nil {
		t.Fatalf("Chain transformation failed: %v", err)
	}
	
	if result != "test-in-in" {
		t.Errorf("Expected 'test-in-in', got '%v'", result)
	}
	
	// Test with non-existent transformer
	configs = []config.TransformerConfig{
		{Name: "non-existent"},
	}
	
	_, err = service.CreateChain(configs)
	if err == nil {
		t.Error("Expected error creating chain with non-existent transformer")
	}
}

func TestService_CreateChainFromNames(t *testing.T) {
	service := NewService()
	
	// Register transformers
	t1 := NewMockTransformer("t1")
	t2 := NewMockTransformer("t2")
	service.Register(t1)
	service.Register(t2)
	
	// Create chain from names
	chain, err := service.CreateChainFromNames([]string{"t1", "t2"})
	if err != nil {
		t.Fatalf("Failed to create chain: %v", err)
	}
	
	// Test chain works
	ctx := context.Background()
	result, err := chain.TransformRequestIn(ctx, "test", "provider")
	if err != nil {
		t.Fatalf("Chain transformation failed: %v", err)
	}
	
	if result != "test-in-in" {
		t.Errorf("Expected 'test-in-in', got '%v'", result)
	}
}

func TestService_GetOrCreateChain(t *testing.T) {
	service := NewService()
	
	// Register transformer
	t1 := NewMockTransformer("t1")
	service.Register(t1)
	
	// Create provider with transformer
	provider := &config.Provider{
		Name: "test-provider",
		Transformers: []config.TransformerConfig{
			{Name: "t1"},
		},
	}
	
	// Get chain (should create)
	chain1, err := service.GetOrCreateChain(provider)
	if err != nil {
		t.Fatalf("Failed to get/create chain: %v", err)
	}
	
	// Get chain again (should return cached)
	chain2, err := service.GetOrCreateChain(provider)
	if err != nil {
		t.Fatalf("Failed to get cached chain: %v", err)
	}
	
	// Should be the same instance
	if chain1 != chain2 {
		t.Error("Expected to get cached chain instance")
	}
}

func TestParseTransformerConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []config.TransformerConfig
		wantErr  bool
	}{
		{
			name:  "simple string",
			input: "transformer1",
			expected: []config.TransformerConfig{
				{Name: "transformer1"},
			},
		},
		{
			name:  "array of strings",
			input: []interface{}{"t1", "t2"},
			expected: []config.TransformerConfig{
				{Name: "t1"},
				{Name: "t2"},
			},
		},
		{
			name: "array with options",
			input: []interface{}{
				"t1",
				[]interface{}{"t2", map[string]interface{}{"option": "value"}},
			},
			expected: []config.TransformerConfig{
				{Name: "t1"},
				{Name: "t2", Config: map[string]interface{}{"option": "value"}},
			},
		},
		{
			name: "object format",
			input: map[string]interface{}{
				"use": []interface{}{"t1", "t2"},
			},
			expected: []config.TransformerConfig{
				{Name: "t1"},
				{Name: "t2"},
			},
		},
		{
			name:    "invalid format",
			input:   123,
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTransformerConfig(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d configs, got %d", len(tt.expected), len(result))
				return
			}
			
			for i, cfg := range result {
				if cfg.Name != tt.expected[i].Name {
					t.Errorf("Config %d: expected name '%s', got '%s'", 
						i, tt.expected[i].Name, cfg.Name)
				}
				
				// Check config if expected
				if tt.expected[i].Config != nil {
					if cfg.Config == nil {
						t.Errorf("Config %d: expected config, got nil", i)
					}
					// Deep comparison of config would go here
				}
			}
		})
	}
}

func TestService_GetChainForProvider(t *testing.T) {
	service := NewService()
	
	// Test getting chain that doesn't exist
	_, err := service.GetChainForProvider("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent provider")
	}
	
	// Create and cache a chain
	chain := NewTransformerChain()
	chain.Add(NewMockTransformer("test"))
	
	// Manually add it to the cache
	service.mu.Lock()
	service.chains["provider:test-provider"] = chain
	service.mu.Unlock()
	
	// Now get it
	retrievedChain, err := service.GetChainForProvider("test-provider")
	if err != nil {
		t.Errorf("Failed to get cached chain: %v", err)
	}
	
	if retrievedChain != chain {
		t.Error("Retrieved chain doesn't match cached one")
	}
}

func TestService_ApplyRequestTransformation(t *testing.T) {
	service := NewService()
	
	// Register a transformer
	trans := NewMockTransformer("test-trans")
	service.Register(trans)
	
	// Create provider with transformer
	provider := &config.Provider{
		Name: "test-provider",
		Transformers: []config.TransformerConfig{
			{Name: "test-trans"},
		},
	}
	
	// Apply transformation
	ctx := context.Background()
	result, err := service.ApplyRequestTransformation(ctx, provider, "input")
	if err != nil {
		t.Fatalf("Failed to apply transformation: %v", err)
	}
	
	// Mock transformer adds "-in" suffix
	if result != "input-in" {
		t.Errorf("Expected 'input-in', got '%v'", result)
	}
	
	// Test with error case - non-existent transformer
	errorProvider := &config.Provider{
		Name: "error-provider",
		Transformers: []config.TransformerConfig{
			{Name: "non-existent"},
		},
	}
	
	_, err = service.ApplyRequestTransformation(ctx, errorProvider, "input")
	if err == nil {
		t.Error("Expected error for non-existent transformer")
	}
}

func TestService_ApplyResponseTransformation(t *testing.T) {
	service := NewService()
	
	// Register a transformer
	trans := NewMockTransformer("test-trans")
	service.Register(trans)
	
	// Create provider with transformer
	provider := &config.Provider{
		Name: "test-provider",
		Transformers: []config.TransformerConfig{
			{Name: "test-trans"},
		},
	}
	
	// Test with HTTP response
	ctx := context.Background()
	httpResp := &http.Response{
		StatusCode: 200,
		Body:       nil,
	}
	
	resp := &Response{Response: httpResp}
	result, err := service.ApplyResponseTransformation(ctx, provider, resp)
	if err != nil {
		t.Fatalf("Failed to apply transformation: %v", err)
	}
	
	// Should return a Response type
	if _, ok := result.(*Response); !ok {
		t.Error("Expected Response type")
	}
	
	// Test with non-HTTP response
	result, err = service.ApplyResponseTransformation(ctx, provider, "non-http")
	if err != nil {
		t.Fatalf("Failed to apply transformation: %v", err)
	}
	
	// Should return as-is
	if result != "non-http" {
		t.Errorf("Expected 'non-http', got '%v'", result)
	}
	
	// Test error case
	errorProvider := &config.Provider{
		Name: "error-provider",
		Transformers: []config.TransformerConfig{
			{Name: "non-existent"},
		},
	}
	
	_, err = service.ApplyResponseTransformation(ctx, errorProvider, resp)
	if err == nil {
		t.Error("Expected error for non-existent transformer")
	}
}