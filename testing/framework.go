// Package testing provides testing utilities for ccproxy
package testing

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

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
	t       *testing.T
	config  TestConfig
	cleanup []func()
	mu      sync.Mutex
}

// NewTestContext creates a new test context
func NewTestContext(t *testing.T) *TestContext {
	tempDir, err := os.MkdirTemp("", "ccproxy-test-*")
	require.NoError(t, err)
	
	tc := &TestContext{
		t: t,
		config: TestConfig{
			TempDir: tempDir,
			Timeout: 30 * time.Second,
		},
		cleanup: make([]func(), 0),
	}
	
	// Register cleanup
	tc.AddCleanup(func() {
		os.RemoveAll(tempDir)
	})
	
	t.Cleanup(tc.Cleanup)
	
	return tc
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