package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/orchestre-dev/ccproxy/internal/config"
	testfw "github.com/orchestre-dev/ccproxy/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParallelServerInstances demonstrates running multiple server instances
// in parallel without interference
func TestParallelServerInstances(t *testing.T) {
	// Run multiple server tests in parallel
	testCases := []struct {
		name   string
		apiKey string
	}{
		{"server1", "key1"},
		{"server2", "key2"},
		{"server3", "key3"},
		{"server4", "key4"},
		{"server5", "key5"},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Enable parallel execution
			t.Parallel()
			
			// Set up isolated test environment
			isolation := testfw.SetupIsolatedTest(t)
			
			// Create test framework
			tf := testfw.NewTestFramework(t)
			
			// Configure and start server with unique settings
			cfg := &config.Config{
				Host:   "localhost",
				Port:   0, // Let the system assign a free port
				APIKey: tc.apiKey,
				Log:    false, // Reduce log noise in parallel tests
			}
			
			srv := tf.StartServerWithConfig(cfg)
			require.NotNil(t, srv)
			
			// Test that each server has its own configuration
			serverURL := fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port)
			
			// Make a test request
			resp, err := testfw.NewHTTPClient(serverURL).Request("GET", "/health", nil)
			require.NoError(t, err)
			defer resp.Body.Close()
			
			assert.Equal(t, 200, resp.StatusCode)
			
			// Write test data to isolated home
			testFile := filepath.Join(isolation.GetHomeDir(), fmt.Sprintf("%s.txt", tc.name))
			err = os.WriteFile(testFile, []byte(tc.apiKey), 0644)
			require.NoError(t, err)
			
			// Verify file exists in isolated environment
			data, err := os.ReadFile(testFile)
			require.NoError(t, err)
			assert.Equal(t, tc.apiKey, string(data))
		})
	}
}

// TestConfigurationIsolation demonstrates isolated configuration handling
func TestConfigurationIsolation(t *testing.T) {
	// Define different configurations to test
	configs := []struct {
		name     string
		provider string
		model    string
	}{
		{"config1", "anthropic", "claude-3"},
		{"config2", "openai", "gpt-4"},
		{"config3", "google", "gemini-pro"},
	}

	// Track config files to ensure isolation
	configFiles := sync.Map{}

	for _, cfg := range configs {
		cfg := cfg // Capture range variable
		t.Run(cfg.name, func(t *testing.T) {
			t.Parallel()
			
			// Set up isolated environment
			testfw.RunParallel(t, func(t *testing.T) {
				isolation := testfw.SetupIsolatedTest(t)
				
				// Create config file in isolated config directory
				configDir := isolation.GetConfigDir()
				configPath := filepath.Join(configDir, "config.json")
				
				// Write configuration
				configData := fmt.Sprintf(`{
					"provider": "%s",
					"model": "%s",
					"api_key": "test-key-%s"
				}`, cfg.provider, cfg.model, cfg.name)
				
				err := os.WriteFile(configPath, []byte(configData), 0644)
				require.NoError(t, err)
				
				// Store config path to verify isolation
				configFiles.Store(cfg.name, configPath)
				
				// Verify each test has its own config file
				configFiles.Range(func(key, value interface{}) bool {
					name := key.(string)
					path := value.(string)
					
					if name != cfg.name {
						// Other config files should not exist in our environment
						_, err := os.Stat(path)
						assert.True(t, os.IsNotExist(err), 
							"Config file from test %s should not exist in test %s", 
							name, cfg.name)
					}
					return true
				})
				
				// Read and verify our config
				data, err := os.ReadFile(configPath)
				require.NoError(t, err)
				assert.Contains(t, string(data), cfg.provider)
				assert.Contains(t, string(data), cfg.model)
			})
		})
	}
}

// TestEnvironmentVariableIsolation demonstrates environment variable isolation
func TestEnvironmentVariableIsolation(t *testing.T) {
	// Set some global env vars that tests might modify
	os.Setenv("CCPROXY_TEST_GLOBAL", "original")
	defer os.Unsetenv("CCPROXY_TEST_GLOBAL")

	testCases := []struct {
		name  string
		value string
	}{
		{"test1", "value1"},
		{"test2", "value2"},
		{"test3", "value3"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			
			isolation := testfw.SetupIsolatedTest(t)
			
			// Each test modifies the same env var
			isolation.SetEnv("CCPROXY_TEST_GLOBAL", tc.value)
			isolation.SetEnv("CCPROXY_TEST_LOCAL", tc.name)
			
			// Verify our values
			assert.Equal(t, tc.value, os.Getenv("CCPROXY_TEST_GLOBAL"))
			assert.Equal(t, tc.name, os.Getenv("CCPROXY_TEST_LOCAL"))
			
			// Create a file using env var path
			testPath := filepath.Join(isolation.GetTempDir(), os.Getenv("CCPROXY_TEST_LOCAL"))
			err := os.WriteFile(testPath, []byte(tc.value), 0644)
			require.NoError(t, err)
		})
	}

	// After all tests, verify original value is restored
	assert.Equal(t, "original", os.Getenv("CCPROXY_TEST_GLOBAL"))
	assert.Empty(t, os.Getenv("CCPROXY_TEST_LOCAL"))
}

// TestCleanupVerification verifies that cleanup happens properly
func TestCleanupVerification(t *testing.T) {
	var isolatedHome string
	var isolatedConfig string
	
	t.Run("create_resources", func(t *testing.T) {
		isolation := testfw.SetupIsolatedTest(t)
		
		isolatedHome = isolation.GetHomeDir()
		isolatedConfig = isolation.GetConfigDir()
		
		// Create some files
		testFile := filepath.Join(isolatedHome, "test.txt")
		err := os.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)
		
		// Verify they exist
		_, err = os.Stat(testFile)
		require.NoError(t, err)
		_, err = os.Stat(isolatedConfig)
		require.NoError(t, err)
	})
	
	t.Run("verify_cleanup", func(t *testing.T) {
		// The isolated directories should have been cleaned up
		_, err := os.Stat(isolatedHome)
		assert.True(t, os.IsNotExist(err), "Isolated home should be cleaned up")
		
		_, err = os.Stat(isolatedConfig)
		assert.True(t, os.IsNotExist(err), "Isolated config should be cleaned up")
	})
}

// BenchmarkIsolationOverhead measures the overhead of test isolation
func BenchmarkIsolationOverhead(b *testing.B) {
	b.Run("with_isolation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			func() {
				t := &testing.T{}
				isolation := testfw.SetupIsolatedTest(t)
				_ = isolation.GetHomeDir()
				// Cleanup happens automatically
			}()
		}
	})
	
	b.Run("without_isolation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = os.TempDir()
		}
	})
}