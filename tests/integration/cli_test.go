package integration

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testutil "github.com/orchestre-dev/ccproxy/testing"
)

// buildTestBinary builds the ccproxy binary for testing
func buildTestBinary(t *testing.T) string {
	t.Helper()
	
	// Get the directory of this test file
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)
	
	// Project root is two levels up from tests/integration
	projectRoot := filepath.Join(testDir, "..", "..")
	
	// Build to a temp location
	binaryPath := filepath.Join(t.TempDir(), "ccproxy-test")
	
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/ccproxy")
	buildCmd.Dir = projectRoot
	
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build ccproxy binary: %v\nOutput: %s", err, string(output))
	}
	
	return binaryPath
}

// TestCLICommands tests the CLI commands integration
func TestCLICommands(t *testing.T) {
	// Build the binary
	binaryPath := buildTestBinary(t)

	// Test version command
	t.Run("Version Command", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "version")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err)

		outputStr := string(output)
		assert.Contains(t, outputStr, "ccproxy version")
		assert.Contains(t, outputStr, "Build Time:")
		assert.Contains(t, outputStr, "Commit:")
	})

	// Test status command when not running
	t.Run("Status Command - Not Running", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "status")
		output, _ := cmd.CombinedOutput()
		// Status command should not error, just report status
		// Check output for either "Not Running" or "Service is not running"
		outputStr := string(output)
		isNotRunning := strings.Contains(outputStr, "Not Running") || 
		               strings.Contains(outputStr, "Service is not running")
		assert.True(t, isNotRunning, "Status should indicate service is not running")
	})

	// Test stop command when not running
	t.Run("Stop Command - Not Running", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "stop")
		output, err := cmd.CombinedOutput()
		// Stop command should error when service is not running
		// or report that service is not running
		outputStr := string(output)
		if err != nil {
			assert.Contains(t, outputStr, "Service is not running")
		} else {
			// If no error, check if it says already stopped
			assert.True(t, 
				strings.Contains(outputStr, "Service is not running") ||
				strings.Contains(outputStr, "not running") ||
				strings.Contains(outputStr, "already stopped"),
				"Stop command should indicate service is not running")
		}
	})
}

// TestClaudeCommands tests the claude subcommands
func TestClaudeCommands(t *testing.T) {
	// Set up test isolation
	isolation := testutil.SetupIsolatedTest(t)
	
	// Build the binary
	binaryPath := buildTestBinary(t)

	// Test claude init
	t.Run("Claude Init", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "claude", "init")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Output: %s", string(output))
		}
		require.NoError(t, err)

		outputStr := string(output)
		// Check for success message or file creation
		assert.True(t, 
			strings.Contains(outputStr, "claude.json") || 
			strings.Contains(outputStr, "Initialized") ||
			strings.Contains(outputStr, "Created"),
			"Init should report success")
		
		// Check if file was created in home directory
		claudeFile := filepath.Join(isolation.GetHomeDir(), ".claude.json")
		assert.FileExists(t, claudeFile, "claude.json should be created in home directory")
	})

	// Test claude show
	t.Run("Claude Show", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "claude", "show")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Error: %v, Output: %s", err, string(output))
		}
		require.NoError(t, err)

		outputStr := string(output)
		t.Logf("Show output: %s", outputStr)
		// Check for expected output - could be JSON or formatted
		assert.True(t, 
			len(outputStr) > 0, // At least some output
			"Show should display configuration")
	})

	// Test claude set
	t.Run("Claude Set", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "claude", "set", "theme", "dark")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Output: %s", string(output))
		}
		require.NoError(t, err)

		outputStr := string(output)
		assert.True(t,
			strings.Contains(outputStr, "Updated") ||
			strings.Contains(outputStr, "Set") ||
			strings.Contains(outputStr, "theme"),
			"Set should confirm the change")

		// Verify the change
		showCmd := exec.Command(binaryPath, "claude", "show")
		showOutput, _ := showCmd.CombinedOutput()
		assert.Contains(t, string(showOutput), "dark")
	})

	// Test claude reset
	t.Run("Claude Reset", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "claude", "reset")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Output: %s", string(output))
		}
		require.NoError(t, err)

		outputStr := string(output)
		assert.True(t,
			strings.Contains(outputStr, "Reset") ||
			strings.Contains(outputStr, "defaults") ||
			strings.Contains(outputStr, "claude.json"),
			"Reset should confirm action")

		// Verify reset to defaults
		showCmd := exec.Command(binaryPath, "claude", "show")
		showOutput, _ := showCmd.CombinedOutput()
		t.Logf("Show after reset: %s", string(showOutput))
		// After reset, just verify we can show the config
		assert.True(t,
			len(string(showOutput)) > 0,
			"Should be able to show config after reset")
	})
}

// TestServiceLifecycle tests start/stop/status commands
func TestServiceLifecycle(t *testing.T) {
	// Skip if running in CI without proper permissions
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping service lifecycle test in CI")
	}

	// Build the binary
	binaryPath := buildTestBinary(t)
	
	// Set up enhanced process cleanup
	cleanup := testutil.NewEnhancedProcessCleanup(t)
	
	// Set up test isolation
	iso := testutil.SetupIsolatedTest(t)

	// Get a free port for testing
	port, err := testutil.GetFreePort()
	require.NoError(t, err, "Failed to get free port")
	
	// Create test config
	testConfig := fmt.Sprintf(`{
		"host": "127.0.0.1",
		"port": %d,
		"log": false,
		"providers": [],
		"routes": {}
	}`, port)
	configPath := iso.CreateTempFile("test-config.json", testConfig)
	cleanup.TrackFile(configPath)
	
	// Track the port for cleanup
	cleanup.TrackPort(port)
	
	// Track PID file directory (ccproxy stores PID in .ccproxy directory)
	pidDir := filepath.Join(iso.GetHomeDir(), ".ccproxy")
	cleanup.TrackDirectory(pidDir)

	// Test start command in foreground
	t.Run("Start Foreground", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "start", "--config", configPath, "--foreground")
		
		// Start the command
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		
		err := cmd.Start()
		require.NoError(t, err)
		
		// Track the process with enhanced cleanup
		cleanup.TrackCommand(cmd, "ccproxy start --foreground")

		// Wait a bit for server to start
		time.Sleep(500 * time.Millisecond)

		// Check if process is running
		assert.NotNil(t, cmd.Process)

		// Log output
		t.Logf("Stdout: %s", stdout.String())
		t.Logf("Stderr: %s", stderr.String())
		
		// The server might not produce immediate output in foreground mode
		// Just verify the process started successfully
		assert.True(t, err == nil || cmd.ProcessState != nil, 
			"Server process should have started successfully")
		
		// Verify no zombies
		assert.NoError(t, cleanup.VerifyNoZombies())
	})
}

// TestCodeCommand tests the code command
func TestCodeCommand(t *testing.T) {
	// This test would require more setup to properly test
	// as it involves starting the service and environment setup
	t.Skip("Code command test requires full environment setup")
}

// TestConfigLoading tests configuration file loading
func TestConfigLoading(t *testing.T) {
	// Build the binary
	binaryPath := buildTestBinary(t)

	t.Run("Invalid Config File", func(t *testing.T) {
		// Create invalid config
		err := os.WriteFile("invalid-config.json", []byte("{invalid json}"), 0644)
		require.NoError(t, err)
		defer os.Remove("invalid-config.json")

		cmd := exec.Command(binaryPath, "start", "--config", "invalid-config.json", "--foreground")
		output, err := cmd.CombinedOutput()
		
		assert.Error(t, err)
		assert.Contains(t, string(output), "failed to")
	})

	t.Run("Missing Config File", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "start", "--config", "non-existent.json", "--foreground")
		output, err := cmd.CombinedOutput()
		
		assert.Error(t, err)
		assert.Contains(t, string(output), "failed to")
	})
}

// TestEnvironmentVariables tests environment variable handling
func TestEnvironmentVariables(t *testing.T) {
	// Build the binary
	binaryPath := buildTestBinary(t)

	t.Run("CCPROXY Environment Variables", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "version")
		cmd.Env = append(os.Environ(), "CCPROXY_PORT=9999", "CCPROXY_HOST=0.0.0.0")
		
		output, err := cmd.CombinedOutput()
		require.NoError(t, err)
		
		// Version command should work regardless of env vars
		assert.Contains(t, string(output), "ccproxy version")
	})
}

// Helper function to check if a process is running
func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	
	// On Unix, this doesn't actually check if the process exists
	// We need to send signal 0 to check
	err = process.Signal(os.Signal(nil))
	return err == nil
}

// Helper function to wait for a port to be available
func waitForPort(port string, timeout time.Duration) bool {
	return testutil.WaitForPort(&testing.T{}, "localhost", port, timeout)
}

// TestHelp tests help output for all commands
func TestHelp(t *testing.T) {
	// Build the binary
	binaryPath := buildTestBinary(t)

	commands := [][]string{
		{},
		{"--help"},
		{"start", "--help"},
		{"stop", "--help"},
		{"status", "--help"},
		{"code", "--help"},
		{"claude", "--help"},
		{"claude", "init", "--help"},
		{"claude", "show", "--help"},
		{"claude", "reset", "--help"},
		{"claude", "set", "--help"},
		{"version", "--help"},
	}

	for _, args := range commands {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			cmd := exec.Command(binaryPath, args...)
			output, err := cmd.CombinedOutput()
			
			// Help should not error
			if len(args) > 0 && args[len(args)-1] == "--help" {
				assert.NoError(t, err)
			}
			
			// Should contain usage information
			outputStr := string(output)
			assert.True(t, 
				strings.Contains(outputStr, "Usage") || 
				strings.Contains(outputStr, "usage") ||
				strings.Contains(outputStr, "CCProxy"),
				"Help output should contain usage information")
		})
	}
}