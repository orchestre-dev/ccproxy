package router

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/orchestre-dev/ccproxy/internal/config"
)

func TestRouterMiddleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedNext   bool
		checkContext   bool
		expectedModel  string
	}{
		{
			name:         "non-POST request",
			method:       http.MethodGet,
			path:         "/v1/messages",
			expectedNext: true,
		},
		{
			name:         "non-messages path",
			method:       http.MethodPost,
			path:         "/v1/other",
			expectedNext: true,
		},
		{
			name:         "invalid JSON body",
			method:       http.MethodPost,
			path:         "/v1/messages",
			body:         "invalid json",
			expectedNext: true,
		},
		{
			name:   "missing model field",
			method: http.MethodPost,
			path:   "/v1/messages",
			body: map[string]interface{}{
				"messages": []interface{}{},
			},
			expectedNext: true,
		},
		{
			name:   "valid request with thinking",
			method: http.MethodPost,
			path:   "/v1/messages",
			body: map[string]interface{}{
				"model":    "claude-3-opus",
				"thinking": true,
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
			expectedNext:  true,
			checkContext:  true,
			expectedModel: "openrouter,anthropic/claude-3-opus",
		},
		{
			name:   "request with system and tools",
			method: http.MethodPost,
			path:   "/v1/messages",
			body: map[string]interface{}{
				"model":  "claude-3-sonnet",
				"system": "You are a helpful assistant",
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "What's the weather?",
					},
				},
				"tools": []interface{}{
					map[string]interface{}{
						"name":         "get_weather",
						"description":  "Get weather info",
						"input_schema": map[string]interface{}{"type": "object"},
					},
				},
			},
			expectedNext:  true,
			checkContext:  true,
			expectedModel: "openai,gpt-4",
		},
		{
			name:   "request with invalid message format",
			method: http.MethodPost,
			path:   "/v1/messages",
			body: map[string]interface{}{
				"model": "gpt-4",
				"messages": []interface{}{
					"invalid message",
				},
			},
			expectedNext:  true,
			checkContext:  true,
			expectedModel: "openai,gpt-4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test config
			cfg := &config.Config{
				Providers: []config.Provider{
					{
						Name: "openrouter",
						APIBaseURL: "https://openrouter.ai/api/v1",
						Models: []string{
							"anthropic/claude-3-opus",
							"anthropic/claude-3.5-sonnet",
						},
					},
					{
						Name: "openai",
						APIBaseURL: "https://api.openai.com/v1",
						Models: []string{
							"gpt-4",
							"gpt-3.5-turbo",
						},
					},
				},
				Routes: map[string]config.Route{
					"think": {
						Provider: "openrouter",
						Model:    "anthropic/claude-3-opus",
					},
					"default": {
						Provider: "openai",
						Model:    "gpt-4",
					},
				},
			}

			// Create router
			r := gin.New()
			
			// Track if next was called
			nextCalled := false
			
			// Add middleware
			r.Use(RouterMiddleware(cfg))
			
			// Add test handler
			r.POST("/v1/messages", func(c *gin.Context) {
				nextCalled = true
				
				if tt.checkContext {
					// Check routing decision
					decision, exists := c.Get("routing_decision")
					if !exists {
						t.Error("routing_decision should be set")
					}
					
					tokenCount, exists := c.Get("token_count")
					if !exists {
						t.Error("token_count should be set")
					}
					if tc, ok := tokenCount.(int); ok && tc < 0 {
						t.Errorf("token_count should be >= 0, got %d", tc)
					}
					
					// Check modified body
					var body map[string]interface{}
					bodyBytes, _ := io.ReadAll(c.Request.Body)
					json.Unmarshal(bodyBytes, &body)
					
					if tt.expectedModel != "" && body["model"] != tt.expectedModel {
						t.Errorf("expected model %s, got %s", tt.expectedModel, body["model"])
					}
					
					// Log decision for debugging
					if d, ok := decision.(RouteDecision); ok {
						t.Logf("Routing decision: provider=%s, model=%s, reason=%s",
							d.Provider, d.Model, d.Reason)
					}
				}
			})
			
			// Also add handlers for other paths
			r.GET("/v1/messages", func(c *gin.Context) {
				nextCalled = true
			})
			r.POST("/v1/other", func(c *gin.Context) {
				nextCalled = true
			})

			// Create request
			var bodyReader io.Reader
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					bodyReader = bytes.NewBufferString(str)
				} else {
					bodyBytes, _ := json.Marshal(tt.body)
					bodyReader = bytes.NewBuffer(bodyBytes)
				}
			}
			
			req := httptest.NewRequest(tt.method, tt.path, bodyReader)
			if tt.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			
			// Create response recorder
			w := httptest.NewRecorder()
			
			// Process request
			r.ServeHTTP(w, req)
			
			// Check if next was called
			if tt.expectedNext != nextCalled {
				t.Errorf("expected next() to be called: %v, but was: %v", tt.expectedNext, nextCalled)
			}
		})
	}
}

func TestBodyReader(t *testing.T) {
	data := []byte("test data")
	reader := &bodyReader{
		data: data,
		pos:  0,
	}
	
	// Test reading in chunks
	buf1 := make([]byte, 4)
	n1, err1 := reader.Read(buf1)
	if err1 != nil {
		t.Errorf("unexpected error: %v", err1)
	}
	if n1 != 4 {
		t.Errorf("expected to read 4 bytes, got %d", n1)
	}
	if string(buf1) != "test" {
		t.Errorf("expected 'test', got '%s'", string(buf1))
	}
	
	// Read remaining
	buf2 := make([]byte, 10)
	n2, err2 := reader.Read(buf2)
	if err2 != nil {
		t.Errorf("unexpected error: %v", err2)
	}
	if n2 != 5 {
		t.Errorf("expected to read 5 bytes, got %d", n2)
	}
	if string(buf2[:n2]) != " data" {
		t.Errorf("expected ' data', got '%s'", string(buf2[:n2]))
	}
	
	// Read at end
	buf3 := make([]byte, 10)
	n3, err3 := reader.Read(buf3)
	if err3 != io.EOF {
		t.Errorf("expected io.EOF at end, got: %v", err3)
	}
	if n3 != 0 {
		t.Errorf("expected to read 0 bytes at end, got %d", n3)
	}
	
	// Test Close
	if err := reader.Close(); err != nil {
		t.Errorf("unexpected error on Close: %v", err)
	}
	
	// Test empty reader
	emptyReader := &bodyReader{
		data: []byte{},
		pos:  0,
	}
	buf4 := make([]byte, 10)
	n4, err4 := emptyReader.Read(buf4)
	if err4 != io.EOF {
		t.Errorf("expected io.EOF from empty reader, got: %v", err4)
	}
	if n4 != 0 {
		t.Errorf("expected to read 0 bytes from empty reader, got %d", n4)
	}
}

func TestGetStringValue(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]interface{}
		key      string
		expected string
	}{
		{
			name: "string value exists",
			m: map[string]interface{}{
				"key": "value",
			},
			key:      "key",
			expected: "value",
		},
		{
			name: "key doesn't exist",
			m: map[string]interface{}{
				"other": "value",
			},
			key:      "key",
			expected: "",
		},
		{
			name: "non-string value",
			m: map[string]interface{}{
				"key": 123,
			},
			key:      "key",
			expected: "",
		},
		{
			name:     "nil map",
			m:        nil,
			key:      "key",
			expected: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringValue(tt.m, tt.key)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// Test error cases and edge cases
func TestRouterMiddleware_ErrorCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	cfg := &config.Config{
		Providers: []config.Provider{
			{
				Name: "test",
				Models: []string{"test-model"},
			},
		},
	}
	
	// Test with malformed tools
	r := gin.New()
	r.Use(RouterMiddleware(cfg))
	
	called := false
	r.POST("/v1/messages", func(c *gin.Context) {
		called = true
	})
	
	body := map[string]interface{}{
		"model": "test-model",
		"tools": []interface{}{
			"invalid tool", // Non-map tool
			map[string]interface{}{
				"name": 123, // Non-string name
			},
		},
	}
	
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	if !called {
		t.Error("Handler should be called even with malformed tools")
	}
}