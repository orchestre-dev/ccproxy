package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/musistudio/ccproxy/internal/config"
	"github.com/musistudio/ccproxy/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServerLifecycle tests the complete server lifecycle
func TestServerLifecycle(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Host:      "127.0.0.1",
		Port:      13456, // Use fixed port for testing
		Log:       false,
		Providers: []config.Provider{},
		Routes:    map[string]config.Route{},
	}

	// Create and start server
	srv, err := server.New(cfg)
	require.NoError(t, err, "Failed to create server")

	// Start server in background
	serverErr := make(chan error, 1)
	go func() {
		if err := srv.Run(); err != nil {
			serverErr <- err
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test health endpoint
	t.Run("Health Check", func(t *testing.T) {
		resp, err := http.Get("http://127.0.0.1:13456/health")
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
		resp, err := http.Get("http://127.0.0.1:13456/status")
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

	// Create configuration with provider
	cfg := &config.Config{
		Host: "127.0.0.1",
		Port: 13456,
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

	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	// Test provider list endpoint
	t.Run("List Providers", func(t *testing.T) {
		resp, err := http.Get("http://127.0.0.1:13456/providers")
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
	// Create mock server configuration
	cfg := &config.Config{
		Host:   "127.0.0.1",
		Port:   13457, // Use different fixed port to avoid conflicts
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
		resp, err := http.Post("http://127.0.0.1:13457/v1/messages", "application/json", bytes.NewReader(body))
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
		httpReq, _ := http.NewRequest("POST", "http://127.0.0.1:13457/v1/messages", bytes.NewReader(body))
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
	cfg := &config.Config{
		Host: "127.0.0.1",
		Port: 13458, // Use different fixed port
	}

	srv, err := server.New(cfg)
	require.NoError(t, err)

	go func() {
		_ = srv.Run()
	}()
	defer srv.Shutdown()

	time.Sleep(100 * time.Millisecond)

	// Test OPTIONS request
	t.Run("OPTIONS Request", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", "http://127.0.0.1:13458/v1/messages", nil)
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
		req, _ := http.NewRequest("GET", "http://127.0.0.1:13458/health", nil)
		req.Header.Set("Origin", "http://example.com")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	})
}

// TestServerShutdown tests graceful shutdown
func TestServerShutdown(t *testing.T) {
	cfg := &config.Config{
		Host: "127.0.0.1",
		Port: 13459, // Use different fixed port
	}

	srv, err := server.New(cfg)
	require.NoError(t, err)

	// Start server
	serverDone := make(chan bool, 1)
	go func() {
		_ = srv.Run()
		close(serverDone)
	}()

	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	// Verify server is running
	resp, err := http.Get("http://127.0.0.1:13459/health")
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
	_, err = http.Get("http://127.0.0.1:13459/health")
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