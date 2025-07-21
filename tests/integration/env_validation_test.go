package integration

import (
	"os"
	"testing"

	"github.com/orchestre-dev/ccproxy/internal/utils"
	testfw "github.com/orchestre-dev/ccproxy/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnvironmentValidation demonstrates environment variable validation in tests
func TestEnvironmentValidation(t *testing.T) {
	t.Run("valid_environment", func(t *testing.T) {
		// Set up safe test environment
		env := testfw.NewSafeTestEnvironment(t)
		
		// Set valid environment variables
		require.NoError(t, env.Set("CCPROXY_SPAWN_DEPTH", "5"))
		require.NoError(t, env.Set("CCPROXY_PORT", "8080"))
		require.NoError(t, env.Set("CCPROXY_TEST_MODE", "1"))
		
		// Validate environment
		testfw.ValidateTestEnvironment(t)
		
		// Verify values were set
		assert.Equal(t, "5", os.Getenv("CCPROXY_SPAWN_DEPTH"))
		assert.Equal(t, "8080", os.Getenv("CCPROXY_PORT"))
		assert.Equal(t, "1", os.Getenv("CCPROXY_TEST_MODE"))
	})
	
	t.Run("invalid_spawn_depth", func(t *testing.T) {
		env := testfw.NewSafeTestEnvironment(t)
		
		// Try to set invalid spawn depth
		err := env.Set("CCPROXY_SPAWN_DEPTH", "15")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds maximum allowed depth")
	})
	
	t.Run("invalid_port", func(t *testing.T) {
		env := testfw.NewSafeTestEnvironment(t)
		
		// Try to set invalid port
		err := env.Set("CCPROXY_PORT", "70000")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be between 1 and 65535")
	})
	
	t.Run("validation_report", func(t *testing.T) {
		// Skip if verbose not enabled
		if !testing.Verbose() {
			t.Skip("Skipping validation report test (run with -v to see report)")
		}
		
		env := testfw.NewSafeTestEnvironment(t)
		
		// Set some test values
		env.Set("CCPROXY_SPAWN_DEPTH", "2")
		env.Set("CCPROXY_HOST", "0.0.0.0")
		env.Set("CCPROXY_LOG", "true")
		
		// Generate validation report
		testfw.ValidateTestEnvironmentWithReport(t)
	})
}

// TestSpawnDepthProtection verifies spawn depth protection works
func TestSpawnDepthProtection(t *testing.T) {
	testCases := []struct {
		name      string
		depth     string
		shouldErr bool
		errMsg    string
	}{
		{"valid_zero", "0", false, ""},
		{"valid_five", "5", false, ""},
		{"valid_ten", "10", false, ""},
		{"invalid_negative", "-1", true, "must be non-negative"},
		{"invalid_overflow", "2147483648", true, "value out of range"},
		{"invalid_text", "abc", true, "invalid integer format"},
		{"invalid_eleven", "11", true, "exceeds maximum allowed depth"},
		{"whitespace", " 5 ", false, ""}, // Should be trimmed
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := utils.ValidateSpawnDepth(tc.depth)
			
			if tc.shouldErr {
				assert.Error(t, err)
				if tc.errMsg != "" {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestEnvironmentVariableDocumentation demonstrates getting env var documentation
func TestEnvironmentVariableDocumentation(t *testing.T) {
	doc := utils.GetEnvironmentVariableDocumentation()
	
	// Verify documentation contains key environment variables
	assert.Contains(t, doc, "CCPROXY_SPAWN_DEPTH")
	assert.Contains(t, doc, "CCPROXY_PORT")
	assert.Contains(t, doc, "CCPROXY_TEST_MODE")
	assert.Contains(t, doc, "CCPROXY_HOST")
	assert.Contains(t, doc, "CCPROXY_LOG")
	
	// Verify descriptions are included
	assert.Contains(t, doc, "Tracks the depth of process spawning")
	assert.Contains(t, doc, "Port number to bind the server to")
	assert.Contains(t, doc, "test mode which disables background spawning")
}

// TestValidationReportDetails demonstrates detailed validation reporting
func TestValidationReportDetails(t *testing.T) {
	// Save original env
	origSpawnDepth := os.Getenv("CCPROXY_SPAWN_DEPTH")
	origPort := os.Getenv("CCPROXY_PORT")
	
	// Clean up
	defer func() {
		if origSpawnDepth != "" {
			os.Setenv("CCPROXY_SPAWN_DEPTH", origSpawnDepth)
		} else {
			os.Unsetenv("CCPROXY_SPAWN_DEPTH")
		}
		if origPort != "" {
			os.Setenv("CCPROXY_PORT", origPort)
		} else {
			os.Unsetenv("CCPROXY_PORT")
		}
	}()
	
	t.Run("all_valid", func(t *testing.T) {
		os.Setenv("CCPROXY_SPAWN_DEPTH", "5")
		os.Setenv("CCPROXY_PORT", "8080")
		
		report := utils.ValidateEnvironmentVariablesWithReport()
		assert.True(t, report.Valid)
		assert.Empty(t, report.Errors)
		
		// Check specific details
		spawnDepthDetail := report.Details["CCPROXY_SPAWN_DEPTH"]
		assert.Equal(t, "5", spawnDepthDetail.Value)
		assert.True(t, spawnDepthDetail.Valid)
		assert.Empty(t, spawnDepthDetail.Error)
		
		portDetail := report.Details["CCPROXY_PORT"]
		assert.Equal(t, "8080", portDetail.Value)
		assert.True(t, portDetail.Valid)
		assert.Empty(t, portDetail.Error)
	})
	
	t.Run("with_errors", func(t *testing.T) {
		os.Setenv("CCPROXY_SPAWN_DEPTH", "20")
		os.Setenv("CCPROXY_PORT", "99999")
		
		report := utils.ValidateEnvironmentVariablesWithReport()
		assert.False(t, report.Valid)
		assert.Len(t, report.Errors, 2)
		
		// Check error details
		spawnDepthDetail := report.Details["CCPROXY_SPAWN_DEPTH"]
		assert.Equal(t, "20", spawnDepthDetail.Value)
		assert.False(t, spawnDepthDetail.Valid)
		assert.Contains(t, spawnDepthDetail.Error, "exceeds maximum allowed depth")
		
		portDetail := report.Details["CCPROXY_PORT"]
		assert.Equal(t, "99999", portDetail.Value)
		assert.False(t, portDetail.Valid)
		assert.Contains(t, portDetail.Error, "must be between 1 and 65535")
	})
}