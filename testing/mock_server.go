package testing

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// MockServer provides a mock HTTP server for testing
type MockServer struct {
	server    *httptest.Server
	mu        sync.RWMutex
	routes    map[string]*MockRoute
	requests  []RecordedRequest
}

// MockRoute represents a mock route configuration
type MockRoute struct {
	Method      string
	Path        string
	Response    interface{}
	StatusCode  int
	Headers     map[string]string
}

// RecordedRequest represents a recorded request
type RecordedRequest struct {
	Method    string
	Path      string
	Headers   http.Header
	Body      []byte
	Timestamp time.Time
}

// NewMockServer creates a new mock server
func NewMockServer() *MockServer {
	ms := &MockServer{
		routes:   make(map[string]*MockRoute),
		requests: make([]RecordedRequest, 0),
	}
	
	// Create test server with router
	ms.server = httptest.NewServer(http.HandlerFunc(ms.handleRequest))
	
	return ms
}

// GetURL returns the mock server URL
func (ms *MockServer) GetURL() string {
	return ms.server.URL
}

// Close shuts down the mock server
func (ms *MockServer) Close() {
	ms.server.Close()
}

// AddRoute adds a mock route
func (ms *MockServer) AddRoute(method, path string, response interface{}, statusCode int) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	key := method + ":" + path
	ms.routes[key] = &MockRoute{
		Method:     method,
		Path:       path,
		Response:   response,
		StatusCode: statusCode,
		Headers:    make(map[string]string),
	}
}

// GetRequests returns all recorded requests
func (ms *MockServer) GetRequests() []RecordedRequest {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	return append([]RecordedRequest{}, ms.requests...)
}

// handleRequest handles incoming requests
func (ms *MockServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Record request
	ms.mu.Lock()
	ms.requests = append(ms.requests, RecordedRequest{
		Method:    r.Method,
		Path:      r.URL.Path,
		Headers:   r.Header.Clone(),
		Timestamp: time.Now(),
	})
	ms.mu.Unlock()
	
	// Find matching route
	key := r.Method + ":" + r.URL.Path
	
	ms.mu.RLock()
	route, ok := ms.routes[key]
	ms.mu.RUnlock()
	
	if !ok {
		http.NotFound(w, r)
		return
	}
	
	// Set headers
	for k, v := range route.Headers {
		w.Header().Set(k, v)
	}
	
	// Write response
	w.WriteHeader(route.StatusCode)
	
	if route.Response != nil {
		switch v := route.Response.(type) {
		case string:
			w.Write([]byte(v))
		case []byte:
			w.Write(v)
		default:
			json.NewEncoder(w).Encode(v)
		}
	}
}