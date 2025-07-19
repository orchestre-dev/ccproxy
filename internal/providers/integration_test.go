package providers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/musistudio/ccproxy/internal/config"
	"github.com/musistudio/ccproxy/internal/server"
)

func TestProviderServiceIntegration(t *testing.T) {
	// Create mock provider server
	providerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate healthy provider
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer providerServer.Close()
	
	// Create configuration with test provider
	cfg := &config.Config{
		Host:   "127.0.0.1",
		Port:   0,
		APIKey: "test-key",
		Providers: []config.Provider{
			{
				Name:       "test-provider",
				APIBaseURL: providerServer.URL,
				APIKey:     "provider-key",
				Models:     []string{"test-model"},
				Enabled:    true,
			},
		},
		Routes: map[string]config.Route{
			"default": {
				Provider: "test-provider",
				Model:    "test-model",
			},
		},
	}
	
	// Create server with provider service
	srv, err := server.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Shutdown()
	
	// Create test server
	ts := httptest.NewServer(srv.GetRouter())
	defer ts.Close()
	
	// Wait for initial health check
	time.Sleep(100 * time.Millisecond)
	
	// Test health endpoint includes provider information
	req, err := http.NewRequest("GET", ts.URL+"/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	
	// Parse response
	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	// Check provider information is included
	providers, ok := health["providers"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected providers field in health response")
	}
	
	total, ok := providers["total"].(float64)
	if !ok || total != 1 {
		t.Errorf("Expected 1 total provider, got %v", providers["total"])
	}
	
	healthy, ok := providers["healthy"].(float64)
	if !ok || healthy != 1 {
		t.Errorf("Expected 1 healthy provider, got %v", providers["healthy"])
	}
	
	// Check provider details
	details, ok := providers["details"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected provider details in health response")
	}
	
	testProvider, ok := details["test-provider"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected test-provider in details")
	}
	
	if !testProvider["healthy"].(bool) {
		t.Error("Expected test-provider to be healthy")
	}
	
	if !testProvider["enabled"].(bool) {
		t.Error("Expected test-provider to be enabled")
	}
}