package providers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/orchestre-dev/ccproxy/internal/config"
)

func TestService_Initialize(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Providers: []config.Provider{
			{
				Name:       "provider1",
				APIBaseURL: "https://api1.test.com",
				APIKey:     "key1",
				Models:     []string{"model1", "model2"},
				Enabled:    true,
			},
			{
				Name:       "provider2",
				APIBaseURL: "https://api2.test.com",
				APIKey:     "key2",
				Models:     []string{"model3"},
				Enabled:    false,
			},
		},
	}
	
	configService := config.NewService()
	configService.SetConfig(cfg)
	
	service := NewService(configService)
	
	// Initialize service
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}
	
	// Check providers are loaded
	providers := service.GetAllProviders()
	if len(providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(providers))
	}
	
	// Check provider details
	p1, err := service.GetProvider("provider1")
	if err != nil {
		t.Errorf("Failed to get provider1: %v", err)
	}
	if p1.Name != "provider1" {
		t.Errorf("Expected provider name 'provider1', got '%s'", p1.Name)
	}
	
	// Check health status initialized
	health, err := service.GetProviderHealth("provider1")
	if err != nil {
		t.Errorf("Failed to get health: %v", err)
	}
	if !health.Healthy {
		t.Error("Expected provider to be initially healthy")
	}
	
	// Check stats initialized
	stats, err := service.GetProviderStats("provider1")
	if err != nil {
		t.Errorf("Failed to get stats: %v", err)
	}
	if stats.TotalRequests != 0 {
		t.Error("Expected initial requests to be 0")
	}
	
	// Cleanup
	service.Stop()
}

func TestService_GetHealthyProviders(t *testing.T) {
	cfg := &config.Config{
		Providers: []config.Provider{
			{
				Name:    "healthy1",
				Enabled: true,
				Models:  []string{"model1"},
			},
			{
				Name:    "disabled",
				Enabled: false,
				Models:  []string{"model2"},
			},
			{
				Name:    "healthy2",
				Enabled: true,
				Models:  []string{"model3"},
			},
		},
	}
	
	configService := config.NewService()
	configService.SetConfig(cfg)
	
	service := NewService(configService)
	service.Initialize()
	
	// Get healthy providers
	healthy := service.GetHealthyProviders()
	if len(healthy) != 2 {
		t.Errorf("Expected 2 healthy providers, got %d", len(healthy))
	}
	
	// Check names
	names := make(map[string]bool)
	for _, p := range healthy {
		names[p.Name] = true
	}
	
	if !names["healthy1"] || !names["healthy2"] {
		t.Error("Expected healthy1 and healthy2 in results")
	}
	if names["disabled"] {
		t.Error("Disabled provider should not be in healthy list")
	}
	
	service.Stop()
}

func TestService_SelectProvider(t *testing.T) {
	cfg := &config.Config{
		Providers: []config.Provider{
			{
				Name:    "provider1",
				Enabled: true,
				Models:  []string{"model1", "model2"},
			},
			{
				Name:    "provider2",
				Enabled: true,
				Models:  []string{"model2", "model3"},
			},
			{
				Name:    "provider3",
				Enabled: false,
				Models:  []string{"model1"},
			},
		},
	}
	
	configService := config.NewService()
	configService.SetConfig(cfg)
	
	service := NewService(configService)
	service.Initialize()
	
	tests := []struct {
		name        string
		criteria    SelectionCriteria
		wantError   bool
		wantProvider string
	}{
		{
			name: "select provider for model1",
			criteria: SelectionCriteria{
				Model: "model1",
			},
			wantProvider: "provider1",
		},
		{
			name: "select provider for model3",
			criteria: SelectionCriteria{
				Model: "model3",
			},
			wantProvider: "provider2",
		},
		{
			name: "no provider for unknown model",
			criteria: SelectionCriteria{
				Model: "unknown",
			},
			wantError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := service.SelectProvider(tt.criteria)
			
			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if provider.Name != tt.wantProvider {
				t.Errorf("Expected provider %s, got %s", tt.wantProvider, provider.Name)
			}
		})
	}
	
	service.Stop()
}

func TestService_RecordRequest(t *testing.T) {
	cfg := &config.Config{
		Providers: []config.Provider{
			{
				Name:    "provider1",
				Enabled: true,
			},
		},
	}
	
	configService := config.NewService()
	configService.SetConfig(cfg)
	
	service := NewService(configService)
	service.Initialize()
	
	// Record successful request
	service.RecordRequest("provider1", true, 100*time.Millisecond)
	
	// Record failed request
	service.RecordRequest("provider1", false, 200*time.Millisecond)
	
	// Check stats
	stats, err := service.GetProviderStats("provider1")
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}
	
	if stats.TotalRequests != 2 {
		t.Errorf("Expected 2 total requests, got %d", stats.TotalRequests)
	}
	if stats.SuccessfulRequests != 1 {
		t.Errorf("Expected 1 successful request, got %d", stats.SuccessfulRequests)
	}
	if stats.FailedRequests != 1 {
		t.Errorf("Expected 1 failed request, got %d", stats.FailedRequests)
	}
	if stats.AverageLatency != 150*time.Millisecond {
		t.Errorf("Expected average latency 150ms, got %v", stats.AverageLatency)
	}
	
	service.Stop()
}

func TestService_HealthCheck(t *testing.T) {
	// Create test HTTP servers
	healthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer healthyServer.Close()
	
	unhealthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer unhealthyServer.Close()
	
	cfg := &config.Config{
		Providers: []config.Provider{
			{
				Name:       "healthy",
				APIBaseURL: healthyServer.URL,
				APIKey:     "test-key",
				Enabled:    true,
			},
			{
				Name:       "unhealthy",
				APIBaseURL: unhealthyServer.URL,
				APIKey:     "test-key",
				Enabled:    true,
			},
		},
	}
	
	configService := config.NewService()
	configService.SetConfig(cfg)
	
	service := NewService(configService)
	service.Initialize()
	
	// Perform health check
	service.checkAllProviders()
	
	// Check healthy provider
	health, err := service.GetProviderHealth("healthy")
	if err != nil {
		t.Fatalf("Failed to get health: %v", err)
	}
	if !health.Healthy {
		t.Errorf("Expected healthy provider to be healthy, error: %s", health.ErrorMessage)
	}
	if health.ConsecutiveFails != 0 {
		t.Errorf("Expected 0 consecutive fails, got %d", health.ConsecutiveFails)
	}
	
	// Check unhealthy provider (first failure)
	health, err = service.GetProviderHealth("unhealthy")
	if err != nil {
		t.Fatalf("Failed to get health: %v", err)
	}
	if health.ConsecutiveFails != 1 {
		t.Errorf("Expected 1 consecutive fail, got %d", health.ConsecutiveFails)
	}
	// Should still be considered healthy after 1 failure
	if !health.Healthy {
		t.Error("Expected provider to still be healthy after 1 failure")
	}
	
	// Simulate 2 more failures
	service.checkAllProviders()
	service.checkAllProviders()
	
	health, err = service.GetProviderHealth("unhealthy")
	if err != nil {
		t.Fatalf("Failed to get health: %v", err)
	}
	if health.Healthy {
		t.Error("Expected provider to be unhealthy after 3 failures")
	}
	if health.ConsecutiveFails != 3 {
		t.Errorf("Expected 3 consecutive fails, got %d", health.ConsecutiveFails)
	}
	
	service.Stop()
}

func TestService_RefreshProvider(t *testing.T) {
	cfg := &config.Config{
		Providers: []config.Provider{
			{
				Name:       "provider1",
				APIBaseURL: "https://api1.test.com",
				Enabled:    true,
			},
		},
	}
	
	configService := config.NewService()
	configService.SetConfig(cfg)
	
	service := NewService(configService)
	service.Initialize()
	
	// Update provider in config
	cfg.Providers[0].APIBaseURL = "https://api1-new.test.com"
	cfg.Providers[0].Enabled = false
	
	// Refresh provider
	err := service.RefreshProvider("provider1")
	if err != nil {
		t.Fatalf("Failed to refresh provider: %v", err)
	}
	
	// Check updated values
	provider, err := service.GetProvider("provider1")
	if err != nil {
		t.Fatalf("Failed to get provider: %v", err)
	}
	if provider.APIBaseURL != "https://api1-new.test.com" {
		t.Errorf("Expected updated URL, got %s", provider.APIBaseURL)
	}
	if provider.Enabled != false {
		t.Error("Expected provider to be disabled")
	}
	
	// Remove provider from config
	cfg.Providers = []config.Provider{}
	
	// Refresh should remove the provider
	err = service.RefreshProvider("provider1")
	if err != nil {
		t.Fatalf("Failed to refresh provider: %v", err)
	}
	
	// Provider should no longer exist
	_, err = service.GetProvider("provider1")
	if err == nil {
		t.Error("Expected error getting removed provider")
	}
	
	service.Stop()
}

func TestService_HealthCheckInterval(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping health check interval test in short mode")
	}
	
	checkCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		checkCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	cfg := &config.Config{
		Providers: []config.Provider{
			{
				Name:       "provider1",
				APIBaseURL: server.URL,
				Enabled:    true,
			},
		},
	}
	
	configService := config.NewService()
	configService.SetConfig(cfg)
	
	service := NewService(configService)
	service.Initialize()
	
	// Start health checks with 100ms interval
	service.StartHealthChecks(100 * time.Millisecond)
	
	// Wait for multiple checks
	time.Sleep(350 * time.Millisecond)
	
	// Should have performed initial check + 3 interval checks
	if checkCount < 3 || checkCount > 5 {
		t.Errorf("Expected 3-5 health checks, got %d", checkCount)
	}
	
	service.Stop()
}