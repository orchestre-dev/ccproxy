package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// MockProviderServer provides a mock server for testing provider interactions
type MockProviderServer struct {
	server          *httptest.Server
	mux             *http.ServeMux
	requests        []MockRequest
	mu              sync.Mutex
	responseDelay   time.Duration
	defaultResponse interface{}
	errorRate       float64
	errorCount      int
	totalRequests   int
}

// MockRequest stores information about a received request
type MockRequest struct {
	Method      string
	Path        string
	Headers     http.Header
	Body        []byte
	ReceivedAt  time.Time
}

// NewMockProviderServer creates a new mock provider server
func NewMockProviderServer(providerName ...string) *MockProviderServer {
	// Determine provider type
	provider := "anthropic" // default
	if len(providerName) > 0 && providerName[0] != "" {
		provider = providerName[0]
	}
	
	// Create base server
	mps := &MockProviderServer{
		mux:             http.NewServeMux(),
		requests:        make([]MockRequest, 0),
		responseDelay:   0,
		errorRate:       0,
	}
	
	// Set provider-specific default response
	switch provider {
	case "anthropic":
		mps.defaultResponse = map[string]interface{}{
			"id": "msg_test",
			"type": "message",
			"role": "assistant",
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": "This is a mock Anthropic response",
				},
			},
			"model": "claude-3-sonnet-20240229",
			"usage": map[string]interface{}{
				"input_tokens":  10,
				"output_tokens": 20,
			},
		}
	case "openai":
		mps.defaultResponse = map[string]interface{}{
			"id": "chatcmpl-test",
			"object": "chat.completion",
			"created": time.Now().Unix(),
			"model": "gpt-4",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "This is a mock OpenAI response",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     10,
				"completion_tokens": 20,
				"total_tokens":      30,
			},
		}
	default:
		// Generic response
		mps.defaultResponse = map[string]interface{}{
			"id": "test-response-id",
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "This is a mock response",
					},
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     10,
				"completion_tokens": 20,
				"total_tokens":      30,
			},
		}
	}
	
	// Set up default handlers
	mps.mux.HandleFunc("/", mps.handleRequest)
	
	// Create test server
	mps.server = httptest.NewServer(mps.mux)
	
	return mps
}

// URL returns the server URL
func (m *MockProviderServer) URL() string {
	return m.server.URL
}

// Close shuts down the server
func (m *MockProviderServer) Close() {
	m.server.Close()
}

// SetResponseDelay sets the delay before responding
func (m *MockProviderServer) SetResponseDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responseDelay = delay
}

// SetDefaultResponse sets the default response
func (m *MockProviderServer) SetDefaultResponse(response interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultResponse = response
}

// SetErrorRate sets the error rate (0.0 to 1.0)
func (m *MockProviderServer) SetErrorRate(rate float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorRate = rate
}

// GetRequests returns all received requests
func (m *MockProviderServer) GetRequests() []MockRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]MockRequest{}, m.requests...)
}

// GetLastRequest returns the last received request
func (m *MockProviderServer) GetLastRequest() *MockRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.requests) == 0 {
		return nil
	}
	return &m.requests[len(m.requests)-1]
}

// ClearRequests clears all stored requests
func (m *MockProviderServer) ClearRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests = make([]MockRequest, 0)
}

// handleRequest handles incoming requests
func (m *MockProviderServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	
	// Record request
	body := make([]byte, 0)
	if r.Body != nil {
		body, _ = io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(body))
	}
	
	m.requests = append(m.requests, MockRequest{
		Method:     r.Method,
		Path:       r.URL.Path,
		Headers:    r.Header.Clone(),
		Body:       body,
		ReceivedAt: time.Now(),
	})
	
	m.totalRequests++
	
	// Check if we should return an error
	if m.errorRate > 0 && float64(m.errorCount)/float64(m.totalRequests) < m.errorRate {
		m.errorCount++
		m.mu.Unlock()
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Mock server error",
				"type":    "internal_error",
			},
		})
		return
	}
	
	delay := m.responseDelay
	response := m.defaultResponse
	m.mu.Unlock()
	
	// Apply delay if configured
	if delay > 0 {
		time.Sleep(delay)
	}
	
	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleFunc allows custom handlers for specific paths
func (m *MockProviderServer) HandleFunc(pattern string, handler http.HandlerFunc) {
	m.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		// Record request first
		m.mu.Lock()
		body := make([]byte, 0)
		if r.Body != nil {
			body, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(body))
		}
		
		m.requests = append(m.requests, MockRequest{
			Method:     r.Method,
			Path:       r.URL.Path,
			Headers:    r.Header.Clone(),
			Body:       body,
			ReceivedAt: time.Now(),
		})
		m.mu.Unlock()
		
		// Call custom handler
		handler(w, r)
	})
}

// HandleStreamingResponse sets up a streaming response handler
func (m *MockProviderServer) HandleStreamingResponse(pattern string, events []string) {
	m.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		// Record request
		m.mu.Lock()
		body := make([]byte, 0)
		if r.Body != nil {
			body, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(body))
		}
		
		m.requests = append(m.requests, MockRequest{
			Method:     r.Method,
			Path:       r.URL.Path,
			Headers:    r.Header.Clone(),
			Body:       body,
			ReceivedAt: time.Now(),
		})
		m.mu.Unlock()
		
		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}
		
		// Send events
		for _, event := range events {
			fmt.Fprintf(w, "data: %s\n\n", event)
			flusher.Flush()
			time.Sleep(10 * time.Millisecond) // Small delay between events
		}
	})
}

// GetRequestCount returns the number of requests received
func (m *MockProviderServer) GetRequestCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.requests)
}

// AssertRequestReceived checks if a request was received
func (m *MockProviderServer) AssertRequestReceived(t *testing.T, method, path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for _, req := range m.requests {
		if req.Method == method && req.Path == path {
			return
		}
	}
	
	t.Errorf("Expected request %s %s not received", method, path)
}

// GetRequestsForPath returns all requests for a specific path
func (m *MockProviderServer) GetRequestsForPath(path string) []MockRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	var result []MockRequest
	for _, req := range m.requests {
		if req.Path == path {
			result = append(result, req)
		}
	}
	return result
}

// AddStreamingRoute adds a custom SSE streaming route handler
func (m *MockProviderServer) AddStreamingRoute(path string, handler func(w http.ResponseWriter, r *http.Request)) {
	m.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		// Record request
		m.mu.Lock()
		body := make([]byte, 0)
		if r.Body != nil {
			body, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(body))
		}
		
		m.requests = append(m.requests, MockRequest{
			Method:     r.Method,
			Path:       r.URL.Path,
			Headers:    r.Header.Clone(),
			Body:       body,
			ReceivedAt: time.Now(),
		})
		m.mu.Unlock()
		
		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		
		// Call the custom handler
		handler(w, r)
	})
}

// AddConditionalRoute adds a route that only returns the specified response when the condition is met
func (m *MockProviderServer) AddConditionalRoute(method, path string, condition func(*http.Request) bool, response interface{}, statusCode int) {
	m.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		// Only handle the specified method
		if r.Method != method {
			http.NotFound(w, r)
			return
		}
		
		// Record request
		m.mu.Lock()
		body := make([]byte, 0)
		if r.Body != nil {
			body, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(body))
		}
		
		m.requests = append(m.requests, MockRequest{
			Method:     r.Method,
			Path:       r.URL.Path,
			Headers:    r.Header.Clone(),
			Body:       body,
			ReceivedAt: time.Now(),
		})
		m.mu.Unlock()
		
		// Check condition
		if condition(r) {
			// Return specified response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			json.NewEncoder(w).Encode(response)
		} else {
			// Return default response
			m.handleRequest(w, r)
		}
	})
}