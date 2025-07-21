// Package testing provides testing utilities for ccproxy
package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/orchestre-dev/ccproxy/internal/config"
	"github.com/orchestre-dev/ccproxy/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig provides test configuration
type TestConfig struct {
	TempDir string
	BaseURL string
	Timeout time.Duration
}

// TestContext provides a test execution context
type TestContext struct {
	t         *testing.T
	config    TestConfig
	cleanup   []func()
	mu        sync.Mutex
	isolation *TestIsolation
}

// NewTestContext creates a new test context
func NewTestContext(t *testing.T) *TestContext {
	// Use isolated test setup
	isolation := SetupIsolatedTest(t)
	
	tc := &TestContext{
		t:         t,
		isolation: isolation,
		config: TestConfig{
			TempDir: isolation.GetTempDir(),
			Timeout: 30 * time.Second,
		},
		cleanup: make([]func(), 0),
	}
	
	t.Cleanup(tc.Cleanup)
	
	return tc
}

// GetIsolation returns the test isolation environment
func (tc *TestContext) GetIsolation() *TestIsolation {
	return tc.isolation
}

// AddCleanup adds a cleanup function
func (tc *TestContext) AddCleanup(fn func()) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.cleanup = append(tc.cleanup, fn)
}

// Cleanup runs all cleanup functions
func (tc *TestContext) Cleanup() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	
	for i := len(tc.cleanup) - 1; i >= 0; i-- {
		tc.cleanup[i]()
	}
}

// CreateTempFile creates a temporary file
func (tc *TestContext) CreateTempFile(name, content string) string {
	path := filepath.Join(tc.config.TempDir, name)
	dir := filepath.Dir(path)
	
	err := os.MkdirAll(dir, 0755)
	require.NoError(tc.t, err)
	
	err = os.WriteFile(path, []byte(content), 0644)
	require.NoError(tc.t, err)
	
	return path
}

// HTTPClient creates an HTTP client for testing
type HTTPClient struct {
	client  *http.Client
	baseURL string
	headers map[string]string
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
		headers: make(map[string]string),
	}
}

// SetHeader sets a default header
func (c *HTTPClient) SetHeader(key, value string) {
	c.headers[key] = value
}

// Request makes an HTTP request
func (c *HTTPClient) Request(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}
	
	url := c.baseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	
	// Set default headers
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	return c.client.Do(req)
}

// SimpleTestServer provides a simple test server
type SimpleTestServer struct {
	server *httptest.Server
	mux    *http.ServeMux
}

// NewSimpleTestServer creates a new test server
func NewSimpleTestServer() *SimpleTestServer {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	
	return &SimpleTestServer{
		server: server,
		mux:    mux,
	}
}

// URL returns the server URL
func (s *SimpleTestServer) URL() string {
	return s.server.URL
}

// Close shuts down the server
func (s *SimpleTestServer) Close() {
	s.server.Close()
}

// HandleFunc registers a handler function
func (s *SimpleTestServer) HandleFunc(pattern string, handler http.HandlerFunc) {
	s.mux.HandleFunc(pattern, handler)
}

// HandleJSON registers a JSON response handler
func (s *SimpleTestServer) HandleJSON(pattern string, response interface{}, statusCode int) {
	s.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(response)
	})
}

// AssertHelpers provides assertion helpers
type AssertHelpers struct {
	t *testing.T
}

// NewAssertHelpers creates assertion helpers
func NewAssertHelpers(t *testing.T) *AssertHelpers {
	return &AssertHelpers{t: t}
}

// AssertJSONEqual asserts JSON equality
func (a *AssertHelpers) AssertJSONEqual(expected, actual string) {
	var expectedObj, actualObj interface{}
	
	err := json.Unmarshal([]byte(expected), &expectedObj)
	require.NoError(a.t, err)
	
	err = json.Unmarshal([]byte(actual), &actualObj)
	require.NoError(a.t, err)
	
	assert.Equal(a.t, expectedObj, actualObj)
}

// AssertEventually asserts a condition becomes true
func (a *AssertHelpers) AssertEventually(condition func() bool, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	
	a.t.Fatal("Condition never became true")
}

// DataGenerator generates test data
type DataGenerator struct{}

// NewDataGenerator creates a data generator
func NewDataGenerator() *DataGenerator {
	return &DataGenerator{}
}

// GenerateString generates a string of specified length
func (g *DataGenerator) GenerateString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}

// GenerateJSON generates a JSON object
func (g *DataGenerator) GenerateJSON(fields map[string]interface{}) string {
	data, _ := json.Marshal(fields)
	return string(data)
}

// TestFramework provides a comprehensive testing framework
type TestFramework struct {
	t         *testing.T
	context   *TestContext
	config    *config.Config
	servers   []*server.Server
	mu        sync.Mutex
	isolation *TestIsolation
}

// NewTestFramework creates a new test framework with isolation
func NewTestFramework(t *testing.T) *TestFramework {
	// Setup isolation first
	isolation := SetupIsolatedTest(t)
	
	tf := &TestFramework{
		t:         t,
		context:   NewTestContext(t),
		isolation: isolation,
		config: &config.Config{
			Host:   "localhost",
			Port:   0, // Will be assigned a free port when starting server
			APIKey: "test-key",
			Log:    true,
		},
		servers: make([]*server.Server, 0),
	}
	
	// Register cleanup
	t.Cleanup(tf.Cleanup)
	
	return tf
}

// GetIsolation returns the test isolation environment
func (tf *TestFramework) GetIsolation() *TestIsolation {
	return tf.isolation
}

// GetConfig returns the test configuration
func (tf *TestFramework) GetConfig() *config.Config {
	return tf.config
}

// AddProvider adds a provider configuration
func (tf *TestFramework) AddProvider(name, baseURL string) {
	tf.mu.Lock()
	defer tf.mu.Unlock()
	
	provider := config.Provider{
		Name:       name,
		APIBaseURL: baseURL,
		APIKey:     "test-key",
		Enabled:    true,
		Models:     []string{"*"}, // Accept all models for testing
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	tf.config.Providers = append(tf.config.Providers, provider)
	
	// Also add routes for this provider
	if tf.config.Routes == nil {
		tf.config.Routes = make(map[string]config.Route)
	}
	
	// Add a route for the provider name
	tf.config.Routes[name] = config.Route{
		Provider: name,
		Model:    "*",
	}
	
	// If this is the first provider, make it the default
	if _, hasDefault := tf.config.Routes["default"]; !hasDefault {
		tf.config.Routes["default"] = config.Route{
			Provider: name,
			Model:    "*",
		}
	}
}

// StartServer starts a test server with optional configuration
func (tf *TestFramework) StartServer(cfg ...*config.Config) (*server.Server, error) {
	var configToUse *config.Config
	if len(cfg) > 0 && cfg[0] != nil {
		configToUse = cfg[0]
	} else {
		configToUse = tf.config
	}
	
	// Get a free port if needed
	if configToUse.Port == 0 {
		port, err := GetFreePort()
		if err != nil {
			return nil, fmt.Errorf("failed to get free port: %w", err)
		}
		configToUse.Port = port
	}
	
	return tf.StartServerWithError(configToUse)
}

// StartServerWithError starts a test server with the provided configuration and returns error
func (tf *TestFramework) StartServerWithError(cfg *config.Config) (*server.Server, error) {
	// If port is not set or is a common test port, get a free one
	if cfg.Port == 0 || cfg.Port == 8080 || cfg.Port == 9090 {
		port, err := GetFreePort()
		if err != nil {
			return nil, fmt.Errorf("failed to get free port: %w", err)
		}
		cfg.Port = port
	}
	
	// Create server
	srv, err := server.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}
	
	// Track server for cleanup
	tf.mu.Lock()
	tf.servers = append(tf.servers, srv)
	tf.mu.Unlock()
	
	// Create a done channel to signal server shutdown
	serverReady := make(chan error, 1)
	
	// Start in background with proper error handling
	go func() {
		if err := srv.Run(); err != nil && err != http.ErrServerClosed {
			tf.t.Logf("Server error: %v", err)
			serverReady <- err
		}
	}()
	
	// Wait for server to be ready with timeout
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case err := <-serverReady:
			return nil, fmt.Errorf("server failed to start: %w", err)
		case <-timeout:
			return nil, fmt.Errorf("server failed to start within timeout")
		case <-ticker.C:
			resp, err := http.Get(fmt.Sprintf("http://%s:%d/health", cfg.Host, cfg.Port))
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return srv, nil
				}
			}
		}
	}
}

// StartServerWithConfig starts a test server with the provided configuration (helper method)
func (tf *TestFramework) StartServerWithConfig(cfg *config.Config) *server.Server {
	srv, err := tf.StartServerWithError(cfg)
	if err != nil {
		tf.t.Fatalf("Failed to start server: %v", err)
	}
	return srv
}

// Cleanup cleans up all resources
func (tf *TestFramework) Cleanup() {
	tf.mu.Lock()
	defer tf.mu.Unlock()
	
	// Shutdown all servers
	for _, srv := range tf.servers {
		if srv != nil {
			if err := srv.Shutdown(); err != nil {
				tf.t.Logf("Error shutting down server: %v", err)
			}
		}
	}
	
	// Clean up context
	if tf.context != nil {
		tf.context.Cleanup()
	}
}