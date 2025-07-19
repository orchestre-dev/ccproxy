package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	ccproxyPort = 8080
	mockPort    = 9090
	testAPIKey  = "test-e2e-key"
)

// TestMain sets up and tears down the test environment
func TestMain(m *testing.M) {
	// Clean up before starting
	cleanup()

	// Run tests
	code := m.Run()

	// Clean up after tests
	cleanup()

	os.Exit(code)
}

// cleanup ensures no processes are left running
func cleanup() {
	// Kill any existing ccproxy processes
	exec.Command("pkill", "-f", "ccproxy").Run()
	
	// Remove PID file
	homeDir, _ := os.UserHomeDir()
	pidFile := filepath.Join(homeDir, ".ccproxy", ".ccproxy.pid")
	os.Remove(pidFile)
	
	// Wait for ports to be released
	time.Sleep(100 * time.Millisecond)
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
	
	configPath := filepath.Join(t.TempDir(), "test-config.json")
	data, err := json.MarshalIndent(config, "", "  ")
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
	
	// Return cleanup function
	return func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait()
			// Log output on cleanup for debugging
			if t.Failed() {
				t.Logf("ccproxy stdout:\n%s", stdout.String())
				t.Logf("ccproxy stderr:\n%s", stderr.String())
			}
		}
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