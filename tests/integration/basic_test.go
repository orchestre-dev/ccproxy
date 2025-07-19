package integration

import (
	"testing"
	"time"
	"net/http"
	"encoding/json"

	"github.com/orchestre-dev/ccproxy/internal/config"
	"github.com/orchestre-dev/ccproxy/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBasicServerStartup tests basic server startup and shutdown
func TestBasicServerStartup(t *testing.T) {
	// Create minimal configuration
	cfg := &config.Config{
		Host:      "127.0.0.1",
		Port:      18089,
		Log:       false,
		Providers: []config.Provider{},
		Routes:    map[string]config.Route{},
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

	// Wait for server to start
	time.Sleep(500 * time.Millisecond)

	// Test health endpoint
	resp, err := http.Get("http://127.0.0.1:18089/health")
	require.NoError(t, err, "Failed to reach health endpoint")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var health map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&health)
	require.NoError(t, err)

	assert.Equal(t, "ok", health["status"])

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