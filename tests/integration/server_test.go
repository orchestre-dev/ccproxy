package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/orchestre-dev/ccproxy/internal/config"
	"github.com/orchestre-dev/ccproxy/internal/server"
	testfw "github.com/orchestre-dev/ccproxy/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServerLifecycle tests the complete server lifecycle
func TestServerLifecycle(t *testing.T) {
	// Create a mock provider server
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
	}))
	defer mockProvider.Close()
	
	// Get a free port for testing
	port, err := testfw.GetFreePort()
	require.NoError(t, err, "Failed to get free port")
	
	// Create test configuration
	cfg := &config.Config{
		Host:      "127.0.0.1",
		Port:      port,
		Log:       false,
		Providers: []config.Provider{
			{
				Name:       "mock-provider",
				APIBaseURL: mockProvider.URL,
				APIKey:     "test-key",
				Enabled:    true,
				Models:     []string{"mock-model"},
			},
		},
		Routes:    map[string]config.Route{
			"default": {
				Provider: "mock-provider",
				Model:    "mock-model",
			},
		},
	}

	// Create and start server
	srv, err := server.New(cfg)
	require.NoError(t, err, "Failed to create server")

	// Start server in background
	serverErr := make(chan error, 1)
	serverStarted := make(chan bool, 1)
	go func() {
		// Signal that we're about to start
		serverStarted <- true
		if err := srv.Run(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
			t.Logf("Server error: %v", err)
		}
	}()
	
	// Wait for goroutine to start
	<-serverStarted

	// Check for immediate server errors
	select {
	case err := <-serverErr:
		t.Fatalf("Server failed to start: %v", err)
	case <-time.After(100 * time.Millisecond):
		// No immediate error, continue
	}
	
	// Wait for server to be ready
	retries := 50 // 5 seconds total
	serverReady := false
	for i := 0; i < retries; i++ {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				serverReady = true
				break
			}
		}
		// Check for server errors during wait
		select {
		case err := <-serverErr:
			t.Fatalf("Server error during startup: %v", err)
		default:
			// Continue waiting
		}
		time.Sleep(100 * time.Millisecond)
	}
	
	if !serverReady {
		t.Fatalf("Server did not become ready within timeout on port %d", port)
	}

	// Test health endpoint
	t.Run("Health Check", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var health map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&health)
		require.NoError(t, err)

		assert.Equal(t, "ok", health["status"])
		assert.NotEmpty(t, health["timestamp"])
	})

	// Test status endpoint
	t.Run("Status Check", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/status", port))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var status map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&status)
		require.NoError(t, err)

		assert.NotEmpty(t, status["status"])
		assert.NotEmpty(t, status["timestamp"])
		assert.NotEmpty(t, status["proxy"])
	})

	// Shutdown server
	err = srv.Shutdown()
	assert.NoError(t, err, "Failed to shutdown server")

	// Check if server stopped
	select {
	case err := <-serverErr:
		assert.NoError(t, err, "Server error during shutdown")
	case <-time.After(2 * time.Second):
		// Server stopped gracefully
	}
}

// TestServerWithProviders tests server with configured providers
func TestServerWithProviders(t *testing.T) {
	// Skip if no API key is provided
	apiKey := os.Getenv("TEST_ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping provider test: TEST_ANTHROPIC_API_KEY not set")
	}

	// Get a free port for testing
	port, err := testfw.GetFreePort()
	require.NoError(t, err, "Failed to get free port")
	
	// Create configuration with provider
	cfg := &config.Config{
		Host: "127.0.0.1",
		Port: port,
		Log:  false,
		Providers: []config.Provider{
			{
				Name:       "anthropic",
				APIBaseURL: "https://api.anthropic.com",
				APIKey:     apiKey,
				Models:     []string{"claude-3-sonnet-20240229"},
				Enabled:    true,
			},
		},
		Routes: map[string]config.Route{
			"default": {
				Provider: "anthropic",
				Model:    "claude-3-sonnet-20240229",
			},
		},
	}

	// Create and start server
	srv, err := server.New(cfg)
	require.NoError(t, err)

	// Start server
	go func() {
		_ = srv.Run()
	}()
	defer srv.Shutdown()

	// Wait for server to be ready
	retries := 50 // 5 seconds total
	for i := 0; i < retries; i++ {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Test provider list endpoint
	t.Run("List Providers", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/providers", port))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		providers, ok := result["providers"].([]interface{})
		require.True(t, ok)
		assert.Len(t, providers, 1)

		provider := providers[0].(map[string]interface{})
		assert.Equal(t, "anthropic", provider["name"])
		assert.Equal(t, true, provider["enabled"])
	})
}

// TestMessageEndpoint tests the main messages endpoint
func TestMessageEndpoint(t *testing.T) {
	// Get a free port for testing
	port, err := testfw.GetFreePort()
	require.NoError(t, err, "Failed to get free port")
	
	// Create mock server configuration
	cfg := &config.Config{
		Host:   "127.0.0.1",
		Port:   port,
		APIKey: "test-key",
		Providers: []config.Provider{
			{
				Name:       "mock",
				APIBaseURL: "http://localhost:9999", // Non-existent
				APIKey:     "mock-key",
				Models:     []string{"mock-model"},
				Enabled:    true,
			},
		},
		Routes: map[string]config.Route{
			"default": {
				Provider: "mock",
				Model:    "mock-model",
			},
		},
	}

	// Create and start server
	srv, err := server.New(cfg)
	require.NoError(t, err)

	go func() {
		_ = srv.Run()
	}()
	defer srv.Shutdown()

	time.Sleep(100 * time.Millisecond)

	// Test authentication
	t.Run("Authentication Required", func(t *testing.T) {
		req := map[string]interface{}{
			"messages": []map[string]string{
				{"role": "user", "content": "Hello"},
			},
		}

		body, _ := json.Marshal(req)
		resp, err := http.Post(fmt.Sprintf("http://127.0.0.1:%d/v1/messages", port), "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Valid Authentication", func(t *testing.T) {
		req := map[string]interface{}{
			"messages": []map[string]string{
				{"role": "user", "content": "Hello"},
			},
			"model": "mock-model",
		}

		body, _ := json.Marshal(req)
		httpReq, _ := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:%d/v1/messages", port), bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("X-API-Key", "test-key")

		resp, err := http.DefaultClient.Do(httpReq)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Will fail because mock provider is not reachable, but should not be 401
		assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

// TestCORSHeaders tests CORS header handling
func TestCORSHeaders(t *testing.T) {
	// Create a mock provider server
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockProvider.Close()
	
	// Get a free port for testing
	port, err := testfw.GetFreePort()
	require.NoError(t, err, "Failed to get free port")
	
	cfg := &config.Config{
		Host: "127.0.0.1",
		Port: port,
		Providers: []config.Provider{
			{
				Name:       "mock",
				APIBaseURL: mockProvider.URL,
				APIKey:     "test",
				Enabled:    true,
				Models:     []string{"test"},
			},
		},
		Routes: map[string]config.Route{
			"default": {Provider: "mock", Model: "test"},
		},
	}

	srv, err := server.New(cfg)
	require.NoError(t, err)

	go func() {
		_ = srv.Run()
	}()
	defer srv.Shutdown()

	// Wait for server to be ready
	retries := 50 // 5 seconds total
	for i := 0; i < retries; i++ {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Test OPTIONS request
	t.Run("OPTIONS Request", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", fmt.Sprintf("http://127.0.0.1:%d/v1/messages", port), nil)
		req.Header.Set("Origin", "http://example.com")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
		assert.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), "POST")
		assert.Contains(t, resp.Header.Get("Access-Control-Allow-Headers"), "Content-Type")
	})

	// Test regular request with CORS
	t.Run("GET Request with CORS", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://127.0.0.1:%d/health", port), nil)
		req.Header.Set("Origin", "http://example.com")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	})
}

// TestServerShutdown tests graceful shutdown
func TestServerShutdown(t *testing.T) {
	// Create a mock provider server
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockProvider.Close()
	
	// Get a free port for testing
	port, err := testfw.GetFreePort()
	require.NoError(t, err, "Failed to get free port")
	
	cfg := &config.Config{
		Host: "127.0.0.1",
		Port: port,
		Providers: []config.Provider{
			{
				Name:       "mock",
				APIBaseURL: mockProvider.URL,
				APIKey:     "test",
				Enabled:    true,
				Models:     []string{"test"},
			},
		},
		Routes: map[string]config.Route{
			"default": {Provider: "mock", Model: "test"},
		},
	}

	srv, err := server.New(cfg)
	require.NoError(t, err)

	// Start server
	serverDone := make(chan bool, 1)
	go func() {
		_ = srv.Run()
		close(serverDone)
	}()

	// Wait for server to be ready
	retries := 50 // 5 seconds total
	for i := 0; i < retries; i++ {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Verify server is running
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Shutdown server
	err = srv.Shutdown()
	assert.NoError(t, err)

	// Wait for server to stop
	select {
	case <-serverDone:
		// Server stopped
	case <-time.After(2 * time.Second):
		// Acceptable - server stopped but didn't signal properly
	}

	// Verify server is not running
	_, err = http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
	assert.Error(t, err, "Server should not be reachable after shutdown")
}

// TestStreamingResponse tests SSE streaming functionality
func TestStreamingResponse(t *testing.T) {
	// This is a placeholder for streaming tests
	// In a real scenario, we would need a mock streaming provider
	t.Skip("Streaming tests require mock streaming provider")
}

// Helper function to read response body
func readBody(t *testing.T, resp *http.Response) string {
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(body)
}