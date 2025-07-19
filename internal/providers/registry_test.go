package providers

import (
	"testing"

	"github.com/musistudio/ccproxy/internal/config"
)

// MockProvider for testing
type MockProvider struct {
	BaseProvider
	healthy bool
}

func (m *MockProvider) IsHealthy() bool {
	return m.healthy
}

func NewMockProvider(name string, models []string, healthy bool) *MockProvider {
	return &MockProvider{
		BaseProvider: BaseProvider{
			Config: &config.Provider{
				Name:    name,
				Models:  models,
				Enabled: true,
			},
		},
		healthy: healthy,
	}
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	registry := &Registry{
		providers: make(map[string]Provider),
	}
	
	// Create mock providers
	p1 := NewMockProvider("provider1", []string{"model1", "model2"}, true)
	p2 := NewMockProvider("provider2", []string{"model3"}, true)
	
	// Register providers
	err := registry.Register(p1)
	if err != nil {
		t.Errorf("Failed to register provider1: %v", err)
	}
	
	err = registry.Register(p2)
	if err != nil {
		t.Errorf("Failed to register provider2: %v", err)
	}
	
	// Try to register duplicate
	err = registry.Register(p1)
	if err == nil {
		t.Error("Expected error registering duplicate provider")
	}
	
	// Get provider
	provider, err := registry.Get("provider1")
	if err != nil {
		t.Errorf("Failed to get provider1: %v", err)
	}
	if provider.GetName() != "provider1" {
		t.Errorf("Expected provider1, got %s", provider.GetName())
	}
	
	// Get non-existent provider
	_, err = registry.Get("nonexistent")
	if err == nil {
		t.Error("Expected error getting non-existent provider")
	}
}

func TestRegistry_GetAll(t *testing.T) {
	registry := &Registry{
		providers: make(map[string]Provider),
	}
	
	// Register multiple providers
	p1 := NewMockProvider("provider1", []string{"model1"}, true)
	p2 := NewMockProvider("provider2", []string{"model2"}, true)
	p3 := NewMockProvider("provider3", []string{"model3"}, false)
	
	registry.Register(p1)
	registry.Register(p2)
	registry.Register(p3)
	
	// Get all providers
	all := registry.GetAll()
	if len(all) != 3 {
		t.Errorf("Expected 3 providers, got %d", len(all))
	}
	
	// Check all providers are present
	names := make(map[string]bool)
	for _, p := range all {
		names[p.GetName()] = true
	}
	
	if !names["provider1"] || !names["provider2"] || !names["provider3"] {
		t.Error("Not all providers returned")
	}
}

func TestRegistry_GetByModel(t *testing.T) {
	registry := &Registry{
		providers: make(map[string]Provider),
	}
	
	// Register providers with different models
	p1 := NewMockProvider("provider1", []string{"model1", "model2"}, true)
	p2 := NewMockProvider("provider2", []string{"model2", "model3"}, true)
	p3 := NewMockProvider("provider3", []string{"model3"}, true)
	
	registry.Register(p1)
	registry.Register(p2)
	registry.Register(p3)
	
	// Get providers for model1
	providers := registry.GetByModel("model1")
	if len(providers) != 1 {
		t.Errorf("Expected 1 provider for model1, got %d", len(providers))
	}
	if providers[0].GetName() != "provider1" {
		t.Errorf("Expected provider1 for model1, got %s", providers[0].GetName())
	}
	
	// Get providers for model2
	providers = registry.GetByModel("model2")
	if len(providers) != 2 {
		t.Errorf("Expected 2 providers for model2, got %d", len(providers))
	}
	
	// Get providers for model3
	providers = registry.GetByModel("model3")
	if len(providers) != 2 {
		t.Errorf("Expected 2 providers for model3, got %d", len(providers))
	}
	
	// Get providers for non-existent model
	providers = registry.GetByModel("model4")
	if len(providers) != 0 {
		t.Errorf("Expected 0 providers for model4, got %d", len(providers))
	}
}

func TestRegistry_Unregister(t *testing.T) {
	registry := &Registry{
		providers: make(map[string]Provider),
	}
	
	// Register and unregister
	p1 := NewMockProvider("provider1", []string{"model1"}, true)
	registry.Register(p1)
	
	// Verify it exists
	_, err := registry.Get("provider1")
	if err != nil {
		t.Error("Provider should exist before unregister")
	}
	
	// Unregister
	registry.Unregister("provider1")
	
	// Verify it's gone
	_, err = registry.Get("provider1")
	if err == nil {
		t.Error("Provider should not exist after unregister")
	}
	
	// Unregister non-existent (should not panic)
	registry.Unregister("nonexistent")
}

func TestRegistry_Clear(t *testing.T) {
	registry := &Registry{
		providers: make(map[string]Provider),
	}
	
	// Register multiple providers
	p1 := NewMockProvider("provider1", []string{"model1"}, true)
	p2 := NewMockProvider("provider2", []string{"model2"}, true)
	
	registry.Register(p1)
	registry.Register(p2)
	
	// Verify they exist
	all := registry.GetAll()
	if len(all) != 2 {
		t.Error("Providers should exist before clear")
	}
	
	// Clear
	registry.Clear()
	
	// Verify empty
	all = registry.GetAll()
	if len(all) != 0 {
		t.Errorf("Expected 0 providers after clear, got %d", len(all))
	}
}

func TestGetRegistry(t *testing.T) {
	// Get global registry
	r1 := GetRegistry()
	r2 := GetRegistry()
	
	// Should be the same instance
	if r1 != r2 {
		t.Error("GetRegistry should return the same instance")
	}
	
	// Should be initialized
	if r1.providers == nil {
		t.Error("Registry providers map should be initialized")
	}
}

func TestBaseProvider(t *testing.T) {
	cfg := &config.Provider{
		Name:    "test-provider",
		Models:  []string{"model1", "model2", "model3"},
		Enabled: true,
	}
	
	provider := &BaseProvider{Config: cfg}
	
	// Test GetName
	if provider.GetName() != "test-provider" {
		t.Errorf("Expected name 'test-provider', got '%s'", provider.GetName())
	}
	
	// Test IsHealthy
	if !provider.IsHealthy() {
		t.Error("Expected provider to be healthy when enabled")
	}
	
	cfg.Enabled = false
	if provider.IsHealthy() {
		t.Error("Expected provider to be unhealthy when disabled")
	}
	
	// Test GetModels
	models := provider.GetModels()
	if len(models) != 3 {
		t.Errorf("Expected 3 models, got %d", len(models))
	}
	
	// Test SupportsModel
	if !provider.SupportsModel("model1") {
		t.Error("Expected provider to support model1")
	}
	if !provider.SupportsModel("model2") {
		t.Error("Expected provider to support model2")
	}
	if provider.SupportsModel("model4") {
		t.Error("Expected provider to not support model4")
	}
}