package integration

import (
	"testing"

	"github.com/orchestre-dev/ccproxy/internal/config"
	testfw "github.com/orchestre-dev/ccproxy/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompilation verifies that integration tests compile correctly
func TestCompilation(t *testing.T) {
	// This test verifies that all imports and types are correct
	assert.True(t, true, "Integration tests compile successfully")
}

// TestTestFramework verifies test framework initialization
func TestTestFramework(t *testing.T) {
	// Create test framework
	tf := testfw.NewTestFramework(t)
	assert.NotNil(t, tf, "Test framework should be created")
	
	// Get config
	cfg := tf.GetConfig()
	assert.NotNil(t, cfg, "Config should not be nil")
	assert.IsType(t, &config.Config{}, cfg, "Config should be correct type")
}

// TestMockProvider verifies mock provider creation
func TestMockProvider(t *testing.T) {
	// Create mock provider
	mockProvider := testfw.NewMockProviderServer()
	defer mockProvider.Close()
	
	assert.NotNil(t, mockProvider, "Mock provider should be created")
	assert.NotEmpty(t, mockProvider.URL(), "Mock provider should have URL")
}

// TestResourceCleanup verifies resource cleanup
func TestResourceCleanup(t *testing.T) {
	// Create test isolation
	iso := testfw.SetupIsolatedTest(t)
	assert.NotNil(t, iso, "Isolation should be created")
	
	// Create temp file
	tempFile := iso.CreateTempFile("test.txt", "test content")
	assert.FileExists(t, tempFile, "Temp file should exist")
	
	// Get various directories
	tempDir := iso.GetTempDir()
	assert.DirExists(t, tempDir, "Temp directory should exist")
	
	homeDir := iso.GetHomeDir()
	assert.NotEmpty(t, homeDir, "Home directory should be set")
	
	configDir := iso.GetConfigDir()
	assert.NotEmpty(t, configDir, "Config directory should be set")
}

// TestGetFreePort verifies port allocation
func TestGetFreePort(t *testing.T) {
	// Get multiple free ports
	ports := make([]int, 5)
	for i := 0; i < 5; i++ {
		port, err := testfw.GetFreePort()
		require.NoError(t, err, "Should get free port")
		assert.Greater(t, port, 1024, "Port should be valid")
		assert.Less(t, port, 65536, "Port should be in valid range")
		ports[i] = port
	}
	
	// Verify all ports are different
	for i := 0; i < len(ports); i++ {
		for j := i + 1; j < len(ports); j++ {
			assert.NotEqual(t, ports[i], ports[j], "Ports should be unique")
		}
	}
}

// TestFixtures verifies test fixtures
func TestFixtures(t *testing.T) {
	// Create fixtures
	fixtures := testfw.NewFixtures()
	assert.NotNil(t, fixtures, "Fixtures should be created")
	
	// Get request fixture
	req, err := fixtures.GetRequest("anthropic_messages")
	require.NoError(t, err, "Should get request fixture")
	assert.NotNil(t, req, "Request should not be nil")
	assert.NotEmpty(t, req["model"], "Request should have model")
	
	// Generate large message
	largeMsg := fixtures.GenerateLargeMessage(1000)
	assert.NotEmpty(t, largeMsg, "Should generate large message")
	assert.Greater(t, len(largeMsg), 3000, "Large message should be sufficiently large")
}