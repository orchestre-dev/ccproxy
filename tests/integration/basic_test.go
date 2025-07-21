package integration

import (
	"testing"
	"time"
	"net/http"
	"net/http/httptest"
	"encoding/json"
	"fmt"

	"github.com/orchestre-dev/ccproxy/internal/config"
	"github.com/orchestre-dev/ccproxy/internal/server"
	testfw "github.com/orchestre-dev/ccproxy/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBasicServerStartup tests basic server startup and shutdown
func TestBasicServerStartup(t *testing.T) {
	// Create a mock provider server
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockProvider.Close()
	
	// Get a free port for testing
	port, err := testfw.GetFreePort()
	require.NoError(t, err, "Failed to get free port")
	
	// Create minimal configuration
	cfg := &config.Config{
		Host:      "127.0.0.1",
		Port:      port,
		Log:       false,
		Providers: []config.Provider{
			{
				Name:       "mock",
				APIBaseURL: mockProvider.URL,
				APIKey:     "test",
				Enabled:    true,
				Models:     []string{"test"},
			},
		},
		Routes:    map[string]config.Route{
			"default": {Provider: "mock", Model: "test"},
		},
	}

	// Create server
	srv, err := server.New(cfg)
	require.NoError(t, err, "Failed to create server")

	// Start server in background
	serverErr := make(chan error, 1)
	go func() {
		if err := srv.Run(); err != nil {
			serverErr <- err
		}
	}()

	// Wait for server to be ready
	retries := 100 // 10 seconds total
	serverReady := false
	var lastError error
	for i := 0; i < retries; i++ {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
		if err == nil {
			if resp.StatusCode == http.StatusOK {
				serverReady = true
				resp.Body.Close()
				break
			}
			resp.Body.Close()
			lastError = fmt.Errorf("health endpoint returned status %d", resp.StatusCode)
		} else {
			lastError = err
		}
		
		// Check if server errored
		select {
		case err := <-serverErr:
			t.Fatalf("Server failed to start: %v", err)
		default:
			// Continue waiting
		}
		
		time.Sleep(100 * time.Millisecond)
	}
	
	if !serverReady && lastError != nil {
		t.Logf("Last error while waiting for server: %v", lastError)
	}

	require.True(t, serverReady, "Server failed to start within timeout")

	// Test health endpoint
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
	require.NoError(t, err, "Failed to reach health endpoint")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var health map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&health)
	require.NoError(t, err)

	// Check for valid status values
	status, ok := health["status"].(string)
	assert.True(t, ok, "Status should be a string")
	assert.Contains(t, []string{"ok", "healthy", "running"}, status, "Status should indicate healthy state")

	// Shutdown server
	err = srv.Shutdown()
	assert.NoError(t, err, "Failed to shutdown server")

	// Wait for server to stop
	select {
	case err := <-serverErr:
		if err != nil {
			t.Logf("Server error: %v", err)
		}
	case <-time.After(2 * time.Second):
		// Server stopped gracefully
	}
}