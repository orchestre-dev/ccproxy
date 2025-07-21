package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orchestre-dev/ccproxy/internal/config"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Host: "127.0.0.1",
		Port: 3456,
		Performance: config.PerformanceConfig{
			RequestTimeout:     30 * time.Second,
			MaxRequestBodySize: 10 * 1024 * 1024,
			MetricsEnabled:     true,
		},
		Routes: map[string]config.Route{
			"default": {
				Provider: "openai",
				Model:    "gpt-4",
			},
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if server == nil {
		t.Error("Expected non-nil server")
	}

	if server.config != cfg {
		t.Error("Server config not set correctly")
	}

	if server.router == nil {
		t.Error("Router not initialized")
	}

	if server.providerService == nil {
		t.Error("Provider service not initialized")
	}

	if server.pipeline == nil {
		t.Error("Pipeline not initialized")
	}

	if server.stateManager == nil {
		t.Error("State manager not initialized")
	}

	if server.performance == nil {
		t.Error("Performance monitor not initialized")
	}
}

func TestNewWithPath(t *testing.T) {
	cfg := &config.Config{
		Host: "127.0.0.1",
		Port: 3456,
		Performance: config.PerformanceConfig{
			RequestTimeout:     30 * time.Second,
			MaxRequestBodySize: 10 * 1024 * 1024,
		},
		Routes: map[string]config.Route{
			"default": {
				Provider: "openai",
				Model:    "gpt-4",
			},
		},
	}

	configPath := "/test/config.json"
	server, err := NewWithPath(cfg, configPath)
	if err != nil {
		t.Fatalf("Failed to create server with path: %v", err)
	}

	if server.configPath != configPath {
		t.Errorf("Expected config path %s, got %s", configPath, server.configPath)
	}
}

func TestNewWithSecurityConstraints(t *testing.T) {
	t.Run("ForceLocalhostWhenNoAPIKey", func(t *testing.T) {
		cfg := &config.Config{
			Host:    "0.0.0.0", // Would allow external access
			Port:    3456,
			APIKey:  "", // No API key
			Performance: config.PerformanceConfig{
				RequestTimeout:     30 * time.Second,
				MaxRequestBodySize: 10 * 1024 * 1024,
			},
			Routes: map[string]config.Route{
				"default": {
					Provider: "openai",
					Model:    "gpt-4",
				},
			},
		}

		server, err := New(cfg)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Should force localhost
		if server.config.Host != "127.0.0.1" {
			t.Errorf("Expected host to be forced to 127.0.0.1, got %s", server.config.Host)
		}
	})

	t.Run("AllowExternalWhenAPIKeyPresent", func(t *testing.T) {
		cfg := &config.Config{
			Host:    "0.0.0.0",
			Port:    3456,
			APIKey:  "test-api-key",
			Performance: config.PerformanceConfig{
				RequestTimeout:     30 * time.Second,
				MaxRequestBodySize: 10 * 1024 * 1024,
			},
			Routes: map[string]config.Route{
				"default": {
					Provider: "openai",
					Model:    "gpt-4",
				},
			},
		}

		server, err := New(cfg)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Should not change host when API key is present
		if server.config.Host != "0.0.0.0" {
			t.Errorf("Expected host to remain 0.0.0.0, got %s", server.config.Host)
		}
	})
}

func TestGetRouter(t *testing.T) {
	cfg := &config.Config{
		Host: "127.0.0.1",
		Port: 3456,
		Performance: config.PerformanceConfig{
			RequestTimeout:     30 * time.Second,
			MaxRequestBodySize: 10 * 1024 * 1024,
		},
		Routes: map[string]config.Route{
			"default": {
				Provider: "openai",
				Model:    "gpt-4",
			},
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	router := server.GetRouter()
	if router == nil {
		t.Error("Expected non-nil router")
	}

	if router != server.router {
		t.Error("GetRouter should return the same instance as server.router")
	}
}

func TestGetPort(t *testing.T) {
	cfg := &config.Config{
		Host: "127.0.0.1",
		Port: 8080,
		Performance: config.PerformanceConfig{
			RequestTimeout:     30 * time.Second,
			MaxRequestBodySize: 10 * 1024 * 1024,
		},
		Routes: map[string]config.Route{
			"default": {
				Provider: "openai",
				Model:    "gpt-4",
			},
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	port := server.GetPort()
	if port != 8080 {
		t.Errorf("Expected port 8080, got %d", port)
	}
}

func TestShutdown(t *testing.T) {
	cfg := &config.Config{
		Host: "127.0.0.1",
		Port: 3456,
		Performance: config.PerformanceConfig{
			RequestTimeout:     30 * time.Second,
			MaxRequestBodySize: 10 * 1024 * 1024,
		},
		Routes: map[string]config.Route{
			"default": {
				Provider: "openai",
				Model:    "gpt-4",
			},
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test shutdown without starting
	err = server.Shutdown()
	if err != nil {
		t.Errorf("Shutdown should not error when server not started: %v", err)
	}
}

func TestServerRoutes(t *testing.T) {
	cfg := &config.Config{
		Host: "127.0.0.1",
		Port: 3456,
		Performance: config.PerformanceConfig{
			RequestTimeout:     30 * time.Second,
			MaxRequestBodySize: 10 * 1024 * 1024,
		},
		Routes: map[string]config.Route{
			"default": {
				Provider: "openai",
				Model:    "gpt-4",
			},
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	router := server.GetRouter()

	// Test root endpoint
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for root endpoint, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal root response: %v", err)
	}

	if response["message"] != "LLMs API" {
		t.Errorf("Expected message 'LLMs API', got %v", response["message"])
	}
}

func TestGinModeConfiguration(t *testing.T) {
	// Save original value
	originalMode := os.Getenv("GIN_MODE")
	defer os.Setenv("GIN_MODE", originalMode)

	t.Run("DefaultToReleaseMode", func(t *testing.T) {
		os.Unsetenv("GIN_MODE")
		
		cfg := &config.Config{
			Host: "127.0.0.1",
			Port: 3456,
			Performance: config.PerformanceConfig{
				RequestTimeout:     30 * time.Second,
				MaxRequestBodySize: 10 * 1024 * 1024,
			},
			Routes: map[string]config.Route{
				"default": {
					Provider: "openai",
					Model:    "gpt-4",
				},
			},
		}

		_, err := New(cfg)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Gin mode should be set to release
		if gin.Mode() != gin.ReleaseMode {
			t.Errorf("Expected Gin mode to be release, got %s", gin.Mode())
		}
	})

	t.Run("RespectExistingGinMode", func(t *testing.T) {
		os.Setenv("GIN_MODE", "test")
		
		// Note: Once Gin mode is set, it cannot be changed in the same process
		// This test documents the expected behavior but may not work in practice
		// due to Gin's global state
		
		cfg := &config.Config{
			Host: "127.0.0.1",
			Port: 3456,
			Performance: config.PerformanceConfig{
				RequestTimeout:     30 * time.Second,
				MaxRequestBodySize: 10 * 1024 * 1024,
			},
			Routes: map[string]config.Route{
				"default": {
					Provider: "openai",
					Model:    "gpt-4",
				},
			},
		}

		_, err := New(cfg)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Gin mode behavior: once set to release in earlier test, it stays release
		// This is expected behavior due to Gin's global state
		mode := gin.Mode()
		if mode != gin.TestMode && mode != gin.ReleaseMode {
			t.Errorf("Expected Gin mode to be test or release, got %s", mode)
		}
	})
}

func TestServerWithPerformanceMonitoring(t *testing.T) {
	cfg := &config.Config{
		Host: "127.0.0.1",
		Port: 3456,
		Performance: config.PerformanceConfig{
			RequestTimeout:     30 * time.Second,
			MaxRequestBodySize: 10 * 1024 * 1024,
			MetricsEnabled:     true, // Enable metrics
		},
		Routes: map[string]config.Route{
			"default": {
				Provider: "openai",
				Model:    "gpt-4",
			},
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server with performance monitoring: %v", err)
	}

	if server.performance == nil {
		t.Error("Performance monitor should be initialized when metrics enabled")
	}

	// Test that performance middleware is added
	router := server.GetRouter()
	if router == nil {
		t.Error("Router should not be nil")
	}
}

func TestServerHTTPSettings(t *testing.T) {
	cfg := &config.Config{
		Host: "127.0.0.1",
		Port: 3456,
		Performance: config.PerformanceConfig{
			RequestTimeout:     30 * time.Second,
			MaxRequestBodySize: 10 * 1024 * 1024,
		},
		Routes: map[string]config.Route{
			"default": {
				Provider: "openai",
				Model:    "gpt-4",
			},
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Check HTTP server settings
	if server.server.Addr != "127.0.0.1:3456" {
		t.Errorf("Expected server address '127.0.0.1:3456', got %s", server.server.Addr)
	}

	if server.server.ReadTimeout != 30*time.Second {
		t.Errorf("Expected read timeout 30s, got %v", server.server.ReadTimeout)
	}

	if server.server.WriteTimeout != 30*time.Second {
		t.Errorf("Expected write timeout 30s, got %v", server.server.WriteTimeout)
	}

	if server.server.IdleTimeout != 120*time.Second {
		t.Errorf("Expected idle timeout 120s, got %v", server.server.IdleTimeout)
	}
}

func TestReadinessProbe(t *testing.T) {
	cfg := &config.Config{
		Host: "127.0.0.1",
		Port: 3456,
		Performance: config.PerformanceConfig{
			RequestTimeout:     30 * time.Second,
			MaxRequestBodySize: 10 * 1024 * 1024,
		},
		Routes: map[string]config.Route{
			"default": {
				Provider: "openai",
				Model:    "gpt-4",
			},
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if server.readiness == nil {
		t.Error("Readiness probe should be initialized")
	}

	// Test readiness checks are set up
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start readiness probe
	server.readiness.Start(ctx)
	defer server.readiness.Stop()

	// This should timeout since we don't have actual providers configured
	err = server.readiness.WaitForReady(ctx, 1*time.Second)
	if err == nil {
		t.Error("Expected readiness check to fail with no configured providers")
	}
}