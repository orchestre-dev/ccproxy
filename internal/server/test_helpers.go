package server

import (
	"testing"
	"time"

	"github.com/orchestre-dev/ccproxy/internal/config"
)

// createTestServerWithProvider creates a server with a test provider configured
func createTestServerWithProvider(t *testing.T, cfg *config.Config) *Server {
	// Add a test provider to the config
	if cfg.Providers == nil {
		cfg.Providers = []config.Provider{}
	}
	
	// Add a default test provider
	testProvider := config.Provider{
		Name:       "test-provider",
		APIBaseURL: "http://localhost:9999",
		APIKey:     "test-api-key",
		Models:     []string{"claude-3-opus-20240229", "claude-3-sonnet-20240229"},
		Enabled:    true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	cfg.Providers = append(cfg.Providers, testProvider)
	
	// Also set up default route
	if cfg.Routes == nil {
		cfg.Routes = make(map[string]config.Route)
	}
	cfg.Routes["default"] = config.Route{
		Provider: "test-provider",
		Model:    "claude-3-opus-20240229",
	}
	
	// Create server
	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	
	return srv
}

// createMinimalTestServer creates a server with minimal config for basic tests
func createMinimalTestServer(t *testing.T) *Server {
	cfg := &config.Config{
		APIKey: "test-key",
		Port:   3456,
		Host:   "127.0.0.1",
	}
	
	return createTestServerWithProvider(t, cfg)
}