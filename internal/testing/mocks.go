package testing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/orchestre-dev/ccproxy/internal/config"
)

// MockHTTPClient is a mock implementation of http.Client
type MockHTTPClient struct {
	mu            sync.RWMutex
	responses     map[string]*http.Response
	requests      []*http.Request
	responseFunc  func(*http.Request) (*http.Response, error)
	errorToReturn error
}

// NewMockHTTPClient creates a new mock HTTP client
func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		responses: make(map[string]*http.Response),
		requests:  make([]*http.Request, 0),
	}
}

// Do implements the http.Client interface
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Store the request
	m.requests = append(m.requests, req)
	
	// Return error if set
	if m.errorToReturn != nil {
		return nil, m.errorToReturn
	}
	
	// Use response function if set
	if m.responseFunc != nil {
		return m.responseFunc(req)
	}
	
	// Look for exact URL match
	url := req.URL.String()
	if resp, exists := m.responses[url]; exists {
		return resp, nil
	}
	
	// Default response
	return &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("{}")),
	}, nil
}

// SetResponse sets a response for a specific URL
func (m *MockHTTPClient) SetResponse(url string, response *http.Response) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[url] = response
}

// SetResponseFunc sets a function to generate responses
func (m *MockHTTPClient) SetResponseFunc(fn func(*http.Request) (*http.Response, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responseFunc = fn
}

// SetError sets an error to return
func (m *MockHTTPClient) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorToReturn = err
}

// GetRequests returns all captured requests
func (m *MockHTTPClient) GetRequests() []*http.Request {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*http.Request, len(m.requests))
	copy(result, m.requests)
	return result
}

// GetLastRequest returns the last captured request
func (m *MockHTTPClient) GetLastRequest() *http.Request {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.requests) == 0 {
		return nil
	}
	return m.requests[len(m.requests)-1]
}

// Reset clears all captured requests and responses
func (m *MockHTTPClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses = make(map[string]*http.Response)
	m.requests = make([]*http.Request, 0)
	m.responseFunc = nil
	m.errorToReturn = nil
}

// CreateMockResponse creates a mock HTTP response
func CreateMockResponse(statusCode int, body string, headers map[string]string) *http.Response {
	resp := &http.Response{
		StatusCode: statusCode,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
	
	for key, value := range headers {
		resp.Header.Set(key, value)
	}
	
	return resp
}

// CreateJSONResponse creates a mock HTTP response with JSON body
func CreateJSONResponse(statusCode int, data interface{}) *http.Response {
	jsonData, _ := json.Marshal(data)
	return &http.Response{
		StatusCode: statusCode,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(bytes.NewReader(jsonData)),
	}
}

// MockServer wraps httptest.Server with additional utilities
type MockServer struct {
	*httptest.Server
	mu             sync.RWMutex
	requests       []*http.Request
	responseFunc   func(*http.Request) (int, interface{})
	defaultStatus  int
	defaultBody    interface{}
}

// NewMockServer creates a new mock server
func NewMockServer() *MockServer {
	ms := &MockServer{
		requests:      make([]*http.Request, 0),
		defaultStatus: 200,
		defaultBody:   map[string]string{"status": "ok"},
	}
	
	ms.Server = httptest.NewServer(http.HandlerFunc(ms.handler))
	return ms
}

// NewMockTLSServer creates a new mock HTTPS server
func NewMockTLSServer() *MockServer {
	ms := &MockServer{
		requests:      make([]*http.Request, 0),
		defaultStatus: 200,
		defaultBody:   map[string]string{"status": "ok"},
	}
	
	ms.Server = httptest.NewTLSServer(http.HandlerFunc(ms.handler))
	return ms
}

func (ms *MockServer) handler(w http.ResponseWriter, r *http.Request) {
	ms.mu.Lock()
	ms.requests = append(ms.requests, r)
	ms.mu.Unlock()
	
	status := ms.defaultStatus
	body := ms.defaultBody
	
	if ms.responseFunc != nil {
		status, body = ms.responseFunc(r)
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if err := json.NewEncoder(w).Encode(body); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(`{"error": "encoding error"}`))
	}
}

// SetResponseFunc sets a function to generate responses
func (ms *MockServer) SetResponseFunc(fn func(*http.Request) (int, interface{})) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.responseFunc = fn
}

// SetDefaultResponse sets the default response
func (ms *MockServer) SetDefaultResponse(status int, body interface{}) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.defaultStatus = status
	ms.defaultBody = body
}

// GetRequests returns all captured requests
func (ms *MockServer) GetRequests() []*http.Request {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	result := make([]*http.Request, len(ms.requests))
	copy(result, ms.requests)
	return result
}

// Reset clears all captured requests
func (ms *MockServer) Reset() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.requests = make([]*http.Request, 0)
	ms.responseFunc = nil
}

// MockProvider creates a mock provider configuration
func MockProvider(name string) config.Provider {
	return config.Provider{
		Name:       name,
		APIBaseURL: "https://api." + name + ".com",
		APIKey:     "mock-api-key-" + name,
		Models:     []string{"mock-model-1", "mock-model-2"},
		Enabled:    true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// MockConfig creates a mock configuration
func MockConfig() *config.Config {
	return &config.Config{
		Providers: []config.Provider{
			MockProvider("anthropic"),
			MockProvider("openai"),
		},
		Host:            "127.0.0.1",
		Port:            3456,
		Log:             true,
		ShutdownTimeout: 30 * time.Second,
	}
}

// MockContext creates a mock context with timeout
func MockContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// MockContextWithValue creates a mock context with a value
func MockContextWithValue(key, value interface{}) context.Context {
	return context.WithValue(context.Background(), key, value)
}

// CaptureLog captures log output for testing
type CaptureLog struct {
	mu      sync.RWMutex
	entries []string
}

// NewCaptureLog creates a new log capturer
func NewCaptureLog() *CaptureLog {
	return &CaptureLog{
		entries: make([]string, 0),
	}
}

// Write implements io.Writer
func (c *CaptureLog) Write(p []byte) (n int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = append(c.entries, string(p))
	return len(p), nil
}

// GetEntries returns all captured log entries
func (c *CaptureLog) GetEntries() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]string, len(c.entries))
	copy(result, c.entries)
	return result
}

// GetLastEntry returns the last captured log entry
func (c *CaptureLog) GetLastEntry() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.entries) == 0 {
		return ""
	}
	return c.entries[len(c.entries)-1]
}

// Contains checks if any log entry contains the substring
func (c *CaptureLog) Contains(substr string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, entry := range c.entries {
		if ContainsString(entry, substr) {
			return true
		}
	}
	return false
}

// Reset clears all captured log entries
func (c *CaptureLog) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make([]string, 0)
}

// TestLogger provides a logger for testing
type TestLogger struct {
	t        *testing.T
	captured *CaptureLog
}

// NewTestLogger creates a new test logger
func NewTestLogger(t *testing.T) *TestLogger {
	return &TestLogger{
		t:        t,
		captured: NewCaptureLog(),
	}
}

// Info logs an info message
func (tl *TestLogger) Info(msg string, args ...interface{}) {
	tl.captured.Write([]byte(fmt.Sprintf("INFO: "+msg, args...)))
	tl.t.Logf("INFO: "+msg, args...)
}

// Error logs an error message
func (tl *TestLogger) Error(msg string, args ...interface{}) {
	tl.captured.Write([]byte(fmt.Sprintf("ERROR: "+msg, args...)))
	tl.t.Logf("ERROR: "+msg, args...)
}

// Debug logs a debug message
func (tl *TestLogger) Debug(msg string, args ...interface{}) {
	tl.captured.Write([]byte(fmt.Sprintf("DEBUG: "+msg, args...)))
	tl.t.Logf("DEBUG: "+msg, args...)
}

// GetCaptured returns the captured log output
func (tl *TestLogger) GetCaptured() *CaptureLog {
	return tl.captured
}