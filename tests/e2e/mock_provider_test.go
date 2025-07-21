package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// mockProvider represents a mock LLM provider for testing
type mockProvider struct {
	server     *http.Server
	router     *gin.Engine
	mu         sync.Mutex
	requests   []mockRequest
	responses  map[string]mockResponse
	streamData map[string][]string
}

type mockRequest struct {
	Path    string
	Method  string
	Headers http.Header
	Body    json.RawMessage
}

type mockResponse struct {
	Status  int
	Headers map[string]string
	Body    interface{}
	Stream  bool
}

// startMockProvider starts a mock provider server
func startMockProvider(t *testing.T) *mockProvider {
	gin.SetMode(gin.TestMode)
	
	m := &mockProvider{
		router:     gin.New(),
		requests:   make([]mockRequest, 0),
		responses:  make(map[string]mockResponse),
		streamData: make(map[string][]string),
	}
	
	// Setup routes
	m.router.Any("/*path", m.handleRequest)
	
	// Start server with proper shutdown handling
	m.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", mockPort),
		Handler: m.router,
	}
	
	// Channel to signal server has started
	serverStarted := make(chan error, 1)
	
	go func() {
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Only report error if server hasn't started yet
			select {
			case serverStarted <- err:
				// Server failed to start
			default:
				// Server was already running, this is a normal shutdown
				t.Logf("Mock provider server stopped: %v", err)
			}
		}
	}()
	
	// Register cleanup with t.Cleanup
	t.Cleanup(func() {
		// Add panic recovery
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic during mock provider cleanup: %v", r)
			}
		}()
		
		if m.server != nil {
			// Use context for graceful shutdown
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			
			if err := m.server.Shutdown(ctx); err != nil {
				t.Logf("Error shutting down mock provider: %v", err)
				// Force close if graceful shutdown fails
				m.server.Close()
			}
		}
	})
	
	// Wait for server to be ready with better error handling
	ready := false
	startTimeout := time.After(5 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	
	for !ready {
		select {
		case err := <-serverStarted:
			t.Fatalf("Mock provider failed to start: %v", err)
		case <-startTimeout:
			t.Fatal("Mock provider failed to start within timeout")
		case <-ticker.C:
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", mockPort))
			if err == nil {
				resp.Body.Close()
				ready = true
			}
		}
	}
	
	return m
}

// Stop stops the mock provider gracefully
func (m *mockProvider) Stop() {
	if m.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if err := m.server.Shutdown(ctx); err != nil {
			// Force close if graceful shutdown fails
			m.server.Close()
		}
	}
}

// SetResponse sets a response for a specific path
func (m *mockProvider) SetResponse(path string, response mockResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[path] = response
}

// SetStreamData sets streaming data for a path
func (m *mockProvider) SetStreamData(path string, chunks []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.streamData[path] = chunks
}

// GetRequests returns all captured requests
func (m *mockProvider) GetRequests() []mockRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]mockRequest{}, m.requests...)
}

// ClearRequests clears captured requests
func (m *mockProvider) ClearRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests = m.requests[:0]
}

// handleRequest handles incoming requests
func (m *mockProvider) handleRequest(c *gin.Context) {
	// Capture request
	body, _ := io.ReadAll(c.Request.Body)
	
	m.mu.Lock()
	m.requests = append(m.requests, mockRequest{
		Path:    c.Request.URL.Path,
		Method:  c.Request.Method,
		Headers: c.Request.Header.Clone(),
		Body:    json.RawMessage(body),
	})
	
	// Get response
	resp, ok := m.responses[c.Request.URL.Path]
	m.mu.Unlock()
	
	if !ok {
		// Default response
		c.JSON(200, gin.H{
			"id": "msg_mock",
			"type": "message",
			"content": []gin.H{
				{
					"type": "text",
					"text": "Mock response",
				},
			},
		})
		return
	}
	
	// Set headers
	for k, v := range resp.Headers {
		c.Header(k, v)
	}
	
	// Handle streaming
	if resp.Stream {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		
		c.Stream(func(w io.Writer) bool {
			m.mu.Lock()
			chunks, ok := m.streamData[c.Request.URL.Path]
			m.mu.Unlock()
			
			if !ok {
				return false
			}
			
			for _, chunk := range chunks {
				fmt.Fprintf(w, "data: %s\n\n", chunk)
				c.Writer.Flush()
				time.Sleep(10 * time.Millisecond) // Simulate streaming delay
			}
			
			return false
		})
		return
	}
	
	// Regular response
	c.JSON(resp.Status, resp.Body)
}

// createMockStreamChunks creates mock SSE chunks
func createMockStreamChunks(messages ...string) []string {
	chunks := make([]string, 0, len(messages)+2)
	
	// Start event
	chunks = append(chunks, `{"type":"message_start","message":{"id":"msg_mock","type":"message","role":"assistant","content":[],"model":"mock-model"}}`)
	
	// Content blocks
	for i, msg := range messages {
		chunks = append(chunks, fmt.Sprintf(`{"type":"content_block_start","index":%d,"content_block":{"type":"text","text":""}}`, i))
		chunks = append(chunks, fmt.Sprintf(`{"type":"content_block_delta","index":%d,"delta":{"type":"text_delta","text":"%s"}}`, i, msg))
		chunks = append(chunks, fmt.Sprintf(`{"type":"content_block_stop","index":%d}`, i))
	}
	
	// End event
	chunks = append(chunks, `{"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null}}`)
	chunks = append(chunks, `{"type":"message_stop"}`)
	
	return chunks
}