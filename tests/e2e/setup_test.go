package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	testfw "github.com/orchestre-dev/ccproxy/testing"
	"github.com/stretchr/testify/require"
)

const (
	testAPIKey = "test-e2e-key"
)

var (
	ccproxyPort int
	mockPort    int
)

func init() {
	// Get free ports for testing
	ports, err := testfw.GetFreePorts(2)
	if err != nil {
		panic(fmt.Sprintf("Failed to get free ports: %v", err))
	}
	ccproxyPort = ports[0]
	mockPort = ports[1]
}

// TestMain sets up and tears down the test environment
func TestMain(m *testing.M) {
	// Set test mode environment variable
	os.Setenv("CCPROXY_TEST_MODE", "1")
	
	// Set up signal handling to ensure cleanup on interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	
	// Run cleanup in background on signal
	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal, cleaning up...")
		cleanup()
		os.Exit(1)
	}()
	
	// Clean up before starting
	cleanup()

	// Run tests with panic recovery
	code := 1
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovered from panic in TestMain: %v\n", r)
				cleanup()
			}
		}()
		code = m.Run()
	}()

	// Clean up after tests
	cleanup()

	os.Exit(code)
}

// cleanup ensures no processes are left running
func cleanup() {
	// Add panic recovery to ensure cleanup always completes
	defer func() {
		if r := recover(); r != nil {
			// Log but continue cleanup
			fmt.Printf("Recovered from panic during cleanup: %v\n", r)
		}
	}()
	
	// Try multiple locations for PID files (in case tests use different HOME dirs)
	possiblePIDFiles := []string{
		filepath.Join(os.Getenv("HOME"), ".ccproxy", ".ccproxy.pid"),
	}
	
	// Also check original home directory
	if origHome, err := os.UserHomeDir(); err == nil {
		possiblePIDFiles = append(possiblePIDFiles, filepath.Join(origHome, ".ccproxy", ".ccproxy.pid"))
	}
	
	// Kill processes based on PID files
	for _, pidFile := range possiblePIDFiles {
		if pidData, err := os.ReadFile(pidFile); err == nil {
			pidStr := strings.TrimSpace(string(pidData))
			if pid, err := strconv.Atoi(pidStr); err == nil && pid > 0 {
				// Kill the specific process
				if proc, err := os.FindProcess(pid); err == nil {
					// First try SIGTERM for graceful shutdown
					if err := proc.Signal(syscall.SIGTERM); err == nil {
						// Give it a moment to shut down gracefully
						time.Sleep(100 * time.Millisecond)
					}
					// Then force kill if still running
					if err := proc.Kill(); err != nil && err != os.ErrProcessDone {
						// Process might already be dead, continue cleanup
						fmt.Printf("Failed to kill process %d: %v\n", pid, err)
					}
				}
			}
			// Remove PID file
			os.Remove(pidFile)
		}
	}
	
	// Clean up any stray lock files
	for _, pidFile := range possiblePIDFiles {
		os.Remove(pidFile + ".lock")
		os.Remove(pidFile + ".startup.lock")
	}
	
	// Only use pkill as last resort and be very specific
	// Look for ccproxy processes started with --foreground flag (test processes)
	exec.Command("pkill", "-f", "ccproxy.*--foreground").Run()
	
	// Wait for ports to be released
	time.Sleep(200 * time.Millisecond)
}

// testConfig represents the test configuration
type testConfig struct {
	Host      string            `json:"host"`
	Port      int               `json:"port"`
	APIKey    string            `json:"apikey"`  // Note: no underscore
	Providers []providerConfig  `json:"providers"`
	Routes    map[string]route  `json:"routes"`
	Log       bool              `json:"log"`
}

type providerConfig struct {
	Name         string      `json:"name"`
	APIBaseURL   string      `json:"api_base_url"`
	APIKey       string      `json:"api_key"`
	Models       []string    `json:"models"`
	Enabled      bool        `json:"enabled"`
	Transformers interface{} `json:"transformers,omitempty"`
}

type route struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

// createTestConfig creates a test configuration file
func createTestConfig(t *testing.T) string {
	t.Helper()
	
	// Create isolated log file in temp directory
	logFile := filepath.Join(t.TempDir(), "ccproxy-test.log")
	
	config := testConfig{
		Host:   "127.0.0.1",
		Port:   ccproxyPort,
		APIKey: testAPIKey,
		Providers: []providerConfig{
			{
				Name:         "mock",
				APIBaseURL:   fmt.Sprintf("http://localhost:%d", mockPort),
				APIKey:       "mock-key",
				Models:       []string{"mock-model", "mock-model-2"},
				Enabled:      true,
				// No transformers - mock provider returns responses as-is
			},
		},
		Routes: map[string]route{
			"default": {
				Provider: "mock",
				Model:    "mock-model",
			},
		},
		Log: true,
	}
	
	// Add log file to config map
	configData := map[string]interface{}{
		"host":      config.Host,
		"port":      config.Port,
		"apikey":    config.APIKey,
		"providers": config.Providers,
		"routes":    config.Routes,
		"log":       config.Log,
		"log_file":  logFile,
	}
	
	configPath := filepath.Join(t.TempDir(), "test-config.json")
	data, err := json.MarshalIndent(configData, "", "  ")
	require.NoError(t, err)
	
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)
	
	return configPath
}

// startCCProxy starts the ccproxy server
func startCCProxy(t *testing.T, configPath string) func() {
	t.Helper()
	
	// Build ccproxy if needed
	ccproxyPath := filepath.Join("..", "..", "ccproxy")
	if _, err := os.Stat(ccproxyPath); err != nil {
		cmd := exec.Command("go", "build", "-o", "ccproxy", "./cmd/ccproxy")
		cmd.Dir = filepath.Join("..", "..")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Failed to build ccproxy: %s", string(output))
	}
	
	// Start ccproxy
	cmd := exec.Command(ccproxyPath, "start", "--config", configPath, "--foreground")
	
	// Set environment variables to prevent infinite spawning and enable test mode
	cmd.Env = append(os.Environ(),
		"CCPROXY_FOREGROUND=1",
		"CCPROXY_SPAWN_DEPTH=1",
		"CCPROXY_TEST_MODE=1",
		// Use isolated home directory for PID files
		fmt.Sprintf("HOME=%s", t.TempDir()),
		// Disable any existing proxy settings
		"HTTP_PROXY=",
		"HTTPS_PROXY=",
		"NO_PROXY=*",
	)
	
	// Capture output for debugging
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Start()
	require.NoError(t, err)
	
	// Wait for server to be ready
	ready := false
	for i := 0; i < 50; i++ { // 5 seconds timeout
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", ccproxyPort))
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			ready = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	
	if !ready {
		cmd.Process.Kill()
		t.Fatalf("ccproxy failed to start. Stdout: %s\nStderr: %s", stdout.String(), stderr.String())
	}
	
	// Store process PID for debugging if needed
	// processPID := cmd.Process.Pid
	
	// Register cleanup with t.Cleanup for guaranteed execution
	t.Cleanup(func() {
		// Add panic recovery
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic during ccproxy cleanup: %v", r)
			}
		}()
		
		if cmd.Process != nil {
			// First try SIGTERM for graceful shutdown
			if err := cmd.Process.Signal(syscall.SIGTERM); err == nil {
				// Give it a moment to shut down gracefully
				done := make(chan error, 1)
				go func() {
					done <- cmd.Wait()
				}()
				
				select {
				case <-done:
					// Process exited gracefully
					t.Logf("ccproxy shut down gracefully")
				case <-time.After(2 * time.Second):
					// Force kill if it doesn't exit in time
					t.Logf("ccproxy didn't respond to SIGTERM, force killing")
					if err := cmd.Process.Kill(); err != nil && err != os.ErrProcessDone {
						t.Logf("Failed to kill ccproxy process: %v", err)
					}
					cmd.Wait()
				}
			} else {
				// If SIGTERM fails, force kill
				t.Logf("Failed to send SIGTERM, force killing: %v", err)
				if err := cmd.Process.Kill(); err != nil && err != os.ErrProcessDone {
					t.Logf("Failed to kill ccproxy process: %v", err)
				}
				cmd.Wait()
			}
			
			// Clean up PID files related to this process
			tempHome := filepath.Dir(filepath.Dir(configPath))
			pidFile := filepath.Join(tempHome, ".ccproxy", ".ccproxy.pid")
			os.Remove(pidFile)
			os.Remove(pidFile + ".lock")
			os.Remove(pidFile + ".startup.lock")
			
			// Log output on cleanup for debugging
			if t.Failed() {
				t.Logf("ccproxy stdout:\n%s", stdout.String())
				t.Logf("ccproxy stderr:\n%s", stderr.String())
			}
		}
	})
	
	// Return cleanup function for backward compatibility
	return func() {
		// The actual cleanup is handled by t.Cleanup above
		// This is just for backward compatibility
	}
}

// makeRequest makes a request to ccproxy
func makeRequest(t *testing.T, method, path string, body interface{}, headers map[string]string) (*http.Response, []byte) {
	t.Helper()
	
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(data)
	}
	
	req, err := http.NewRequest(method, fmt.Sprintf("http://localhost:%d%s", ccproxyPort, path), bodyReader)
	require.NoError(t, err)
	
	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, ok := headers["Authorization"]; !ok {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", testAPIKey)
	}
	
	// Apply custom headers
	for k, v := range headers {
		if v == "" {
			// Delete the header if value is empty
			req.Header.Del(k)
		} else {
			req.Header.Set(k, v)
		}
	}
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Logf("HTTP request failed: %v", err)
		t.Logf("Request URL: %s", req.URL)
		t.Logf("Request method: %s", req.Method)
		require.NoError(t, err)
	}
	
	// Log response details
	t.Logf("Response status: %d %s", resp.StatusCode, resp.Status)
	t.Logf("Response Content-Length: %d", resp.ContentLength)
	t.Logf("Response Transfer-Encoding: %v", resp.TransferEncoding)
	t.Logf("Response headers: %v", resp.Header)
	
	// Try to read the body in chunks to see where it fails
	var respBody []byte
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			respBody = append(respBody, buf[:n]...)
			t.Logf("Read %d bytes, total so far: %d", n, len(respBody))
		}
		if err != nil {
			if err != io.EOF {
				t.Logf("Failed to read response body: %v", err)
				t.Logf("Response status: %d", resp.StatusCode)
				t.Logf("Bytes read so far: %d", len(respBody))
				resp.Body.Close()
				require.NoError(t, err)
			}
			break
		}
	}
	resp.Body.Close()
	
	if len(respBody) == 0 {
		t.Logf("Warning: Empty response body")
		t.Logf("Response status: %d", resp.StatusCode)
		t.Logf("Response headers: %v", resp.Header)
	}
	
	return resp, respBody
}

// waitForCondition waits for a condition to be true
func waitForCondition(timeout time.Duration, check func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

// testEnv provides an isolated test environment
type testEnv struct {
	t          *testing.T
	tempDir    string
	configPath string
	ccproxy    func()
	mock       *mockProvider
}

// newTestEnv creates a new isolated test environment
func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	
	// Create completely isolated temp directory
	tempDir := t.TempDir()
	
	// Set isolated HOME for this test
	oldHome := os.Getenv("HOME")
	testHome := filepath.Join(tempDir, "home")
	require.NoError(t, os.MkdirAll(testHome, 0755))
	os.Setenv("HOME", testHome)
	
	// Create .ccproxy directory in isolated home
	ccproxyDir := filepath.Join(testHome, ".ccproxy")
	require.NoError(t, os.MkdirAll(ccproxyDir, 0755))
	
	// Restore original HOME on cleanup
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})
	
	env := &testEnv{
		t:       t,
		tempDir: tempDir,
	}
	
	// Create test config
	env.configPath = env.createConfig()
	
	// Start mock provider
	env.mock = startMockProvider(t)
	
	// Start ccproxy
	env.ccproxy = startCCProxy(t, env.configPath)
	
	return env
}

// createConfig creates an isolated test configuration
func (e *testEnv) createConfig() string {
	logFile := filepath.Join(e.tempDir, "ccproxy.log")
	
	config := map[string]interface{}{
		"host":      "127.0.0.1",
		"port":      ccproxyPort,
		"apikey":    testAPIKey,
		"log":       true,
		"log_file":  logFile,
		"providers": []map[string]interface{}{
			{
				"name":         "mock",
				"api_base_url": fmt.Sprintf("http://localhost:%d", mockPort),
				"api_key":      "mock-key",
				"models":       []string{"mock-model", "mock-model-2"},
				"enabled":      true,
			},
		},
		"routes": map[string]interface{}{
			"default": map[string]string{
				"provider": "mock",
				"model":    "mock-model",
			},
		},
	}
	
	configPath := filepath.Join(e.tempDir, "config.json")
	data, err := json.MarshalIndent(config, "", "  ")
	require.NoError(e.t, err)
	require.NoError(e.t, os.WriteFile(configPath, data, 0644))
	
	return configPath
}

// request makes a request to the test ccproxy instance
func (e *testEnv) request(method, path string, body interface{}, headers map[string]string) (*http.Response, []byte) {
	return makeRequest(e.t, method, path, body, headers)
}