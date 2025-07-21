package testing

import (
	"fmt"
	"os"
	"testing"

	"github.com/orchestre-dev/ccproxy/internal/utils"
)

// ValidateTestEnvironment validates all CCProxy environment variables used in tests
// and returns an error if any are invalid. This should be called at the beginning
// of test suites to catch configuration errors early.
func ValidateTestEnvironment(t *testing.T) {
	t.Helper()

	// Validate all environment variables
	if err := utils.ValidateEnvironmentVariables(); err != nil {
		t.Fatalf("Environment validation failed: %v", err)
	}
}

// ValidateTestEnvironmentWithReport validates environment variables and logs
// a detailed report for debugging
func ValidateTestEnvironmentWithReport(t *testing.T) {
	t.Helper()

	report := utils.ValidateEnvironmentVariablesWithReport()
	
	if !report.Valid {
		t.Logf("Environment Validation Report:")
		t.Logf("================================")
		for _, err := range report.Errors {
			t.Logf("ERROR: %s", err)
		}
		t.Logf("================================")
		t.Fatalf("Environment validation failed with %d errors", len(report.Errors))
	}

	// Log successful validation with any warnings
	if testing.Verbose() {
		t.Logf("Environment validation passed")
		for name, detail := range report.Details {
			if detail.Value != "" {
				t.Logf("  %s = %q", name, detail.Value)
			}
		}
	}
}

// SetTestEnvironmentSafely sets environment variables with validation
func SetTestEnvironmentSafely(t *testing.T, env map[string]string) {
	t.Helper()

	// First, validate each value
	for name, value := range env {
		// Find the environment variable definition
		var found bool
		for _, envVar := range utils.EnvironmentVariables {
			if envVar.Name == name {
				found = true
				if envVar.ValidateFunc != nil {
					if err := envVar.ValidateFunc(value); err != nil {
						t.Fatalf("Invalid value for %s: %v", name, err)
					}
				}
				break
			}
		}
		
		if !found {
			t.Logf("WARNING: Unknown environment variable %s", name)
		}
	}

	// Set the environment variables
	for name, value := range env {
		os.Setenv(name, value)
	}
}

// AssertSpawnDepthValid checks that CCPROXY_SPAWN_DEPTH is within valid bounds
func AssertSpawnDepthValid(t *testing.T) {
	t.Helper()

	if depthStr := os.Getenv("CCPROXY_SPAWN_DEPTH"); depthStr != "" {
		if err := utils.ValidateSpawnDepth(depthStr); err != nil {
			t.Fatalf("Invalid CCPROXY_SPAWN_DEPTH: %v", err)
		}
	}
}

// EnsureTestMode ensures CCPROXY_TEST_MODE is set to prevent background spawning
func EnsureTestMode(t *testing.T) {
	t.Helper()

	if os.Getenv("CCPROXY_TEST_MODE") != "1" {
		t.Logf("Setting CCPROXY_TEST_MODE=1 to prevent background spawning")
		os.Setenv("CCPROXY_TEST_MODE", "1")
	}
}

// SafeTestEnvironment sets up a safe test environment with proper validation
type SafeTestEnvironment struct {
	t           *testing.T
	originalEnv map[string]string
}

// NewSafeTestEnvironment creates a new safe test environment
func NewSafeTestEnvironment(t *testing.T) *SafeTestEnvironment {
	env := &SafeTestEnvironment{
		t:           t,
		originalEnv: make(map[string]string),
	}

	// Ensure test mode is enabled
	EnsureTestMode(t)

	// Validate current environment
	ValidateTestEnvironment(t)

	// Set up cleanup
	t.Cleanup(func() {
		env.Restore()
	})

	return env
}

// Set sets an environment variable with validation
func (e *SafeTestEnvironment) Set(name, value string) error {
	// Save original value
	if _, saved := e.originalEnv[name]; !saved {
		if orig, exists := os.LookupEnv(name); exists {
			e.originalEnv[name] = orig
		} else {
			e.originalEnv[name] = ""
		}
	}

	// Validate before setting
	for _, envVar := range utils.EnvironmentVariables {
		if envVar.Name == name {
			if envVar.ValidateFunc != nil {
				if err := envVar.ValidateFunc(value); err != nil {
					return fmt.Errorf("validation failed for %s: %w", name, err)
				}
			}
			break
		}
	}

	os.Setenv(name, value)
	return nil
}

// Restore restores original environment variables
func (e *SafeTestEnvironment) Restore() {
	for name, value := range e.originalEnv {
		if value == "" {
			os.Unsetenv(name)
		} else {
			os.Setenv(name, value)
		}
	}
}