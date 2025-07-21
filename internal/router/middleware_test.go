package router

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/orchestre-dev/ccproxy/internal/config"
)

func TestRouterMiddleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Routes: map[string]config.Route{
			"default": {
				Provider: "openai",
				Model:    "gpt-4",
			},
			"longContext": {
				Provider: "anthropic",
				Model:    "claude-3-opus",
			},
			"think": {
				Provider: "anthropic",
				Model:    "claude-3-sonnet",
			},
		},
	}

	t.Run("ValidRequest", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello",
				},
			},
		}

		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		middleware := RouterMiddleware(cfg)

		// Set up a handler to capture the modified request
		var capturedBody map[string]interface{}
		middleware(c)

		// Read the modified body
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		json.Unmarshal(bodyBytes, &capturedBody)

		// Check that model was updated (should use default routing)
		if capturedBody["model"] != "openai,gpt-4" {
			t.Errorf("Expected model to be routed to 'openai,gpt-4', got %v", capturedBody["model"])
		}

		// Check that routing decision was stored
		decision, exists := c.Get("routing_decision")
		if !exists {
			t.Error("Expected routing decision to be stored in context")
		}

		routingDecision := decision.(RouteDecision)
		if routingDecision.Provider != "openai" {
			t.Errorf("Expected provider 'openai', got %s", routingDecision.Provider)
		}

		// Check that token count was stored
		tokenCount, exists := c.Get("token_count")
		if !exists {
			t.Error("Expected token count to be stored in context")
		}

		if tokenCount.(int) <= 0 {
			t.Error("Expected positive token count")
		}
	})

	t.Run("RequestWithThinking", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"model": "gpt-4",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Think step by step",
				},
			},
			"thinking": true,
		}

		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		middleware := RouterMiddleware(cfg)
		middleware(c)

		// Read the modified body
		var capturedBody map[string]interface{}
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		json.Unmarshal(bodyBytes, &capturedBody)

		// Should route to thinking model
		if capturedBody["model"] != "anthropic,claude-3-sonnet" {
			t.Errorf("Expected thinking model, got %v", capturedBody["model"])
		}

		decision, _ := c.Get("routing_decision")
		routingDecision := decision.(RouteDecision)
		if routingDecision.Reason != "thinking parameter enabled" {
			t.Errorf("Expected thinking routing reason, got %s", routingDecision.Reason)
		}
	})

	t.Run("RequestWithSystemAndTools", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"model": "gpt-4",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "What's the weather?",
				},
			},
			"system": "You are a helpful assistant.",
			"tools": []interface{}{
				map[string]interface{}{
					"name":        "get_weather",
					"description": "Get weather information",
					"input_schema": map[string]interface{}{
						"type": "object",
					},
				},
			},
		}

		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		middleware := RouterMiddleware(cfg)
		middleware(c)

		// Should process complex request with system and tools
		tokenCount, _ := c.Get("token_count")
		if tokenCount.(int) <= 0 {
			t.Error("Expected positive token count for complex request")
		}
	})

	t.Run("NonPostRequest", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/v1/messages", nil)

		middleware := RouterMiddleware(cfg)

		// Just call the middleware directly
		middleware(c)

		// Should not set routing decision for non-POST
		_, exists := c.Get("routing_decision")
		if exists {
			t.Error("Should not set routing decision for non-POST request")
		}
	})

	t.Run("WrongPath", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, engine := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/completions", nil)

		middleware := RouterMiddleware(cfg)

		called := false
		engine.Use(middleware)
		engine.POST("/v1/completions", func(c *gin.Context) {
			called = true
		})

		// Process the request
		engine.ServeHTTP(w, c.Request)

		if !called {
			t.Error("Expected next handler to be called")
		}

		// Should not process wrong path
		_, exists := c.Get("routing_decision")
		if exists {
			t.Error("Should not process requests to wrong path")
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, engine := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/messages", strings.NewReader("invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")

		middleware := RouterMiddleware(cfg)

		called := false
		engine.Use(middleware)
		engine.POST("/v1/messages", func(c *gin.Context) {
			called = true
		})

		// Process the request
		engine.ServeHTTP(w, c.Request)

		if !called {
			t.Error("Expected next handler to be called even with invalid JSON")
		}

		// Should not set routing decision for invalid JSON
		_, exists := c.Get("routing_decision")
		if exists {
			t.Error("Should not set routing decision for invalid JSON")
		}
	})

	t.Run("MissingModel", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello",
				},
			},
			// No model field
		}

		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, engine := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		middleware := RouterMiddleware(cfg)

		called := false
		engine.Use(middleware)
		engine.POST("/v1/messages", func(c *gin.Context) {
			called = true
		})

		// Process the request
		engine.ServeHTTP(w, c.Request)

		if !called {
			t.Error("Expected next handler to be called")
		}

		// Should not set routing decision for missing model
		_, exists := c.Get("routing_decision")
		if exists {
			t.Error("Should not set routing decision for missing model")
		}
	})
}

func TestBodyReader(t *testing.T) {
	t.Run("BasicRead", func(t *testing.T) {
		data := []byte("hello world")
		reader := &bodyReader{data: data, pos: 0}

		buf := make([]byte, 5)
		n, err := reader.Read(buf)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if n != 5 {
			t.Errorf("Expected to read 5 bytes, got %d", n)
		}

		if string(buf) != "hello" {
			t.Errorf("Expected 'hello', got %s", string(buf))
		}

		if reader.pos != 5 {
			t.Errorf("Expected position 5, got %d", reader.pos)
		}
	})

	t.Run("ReadToEnd", func(t *testing.T) {
		data := []byte("test")
		reader := &bodyReader{data: data, pos: 0}

		buf := make([]byte, 10) // Buffer larger than data
		n, err := reader.Read(buf)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if n != 4 {
			t.Errorf("Expected to read 4 bytes, got %d", n)
		}

		if string(buf[:n]) != "test" {
			t.Errorf("Expected 'test', got %s", string(buf[:n]))
		}
	})

	t.Run("ReadPastEnd", func(t *testing.T) {
		data := []byte("test")
		reader := &bodyReader{data: data, pos: 4} // Already at end

		buf := make([]byte, 5)
		n, err := reader.Read(buf)

		if err != io.EOF {
			t.Errorf("Expected EOF, got %v", err)
		}

		if n != 0 {
			t.Errorf("Expected 0 bytes read, got %d", n)
		}
	})

	t.Run("MultipleReads", func(t *testing.T) {
		data := []byte("hello world")
		reader := &bodyReader{data: data, pos: 0}

		// First read
		buf1 := make([]byte, 5)
		n1, err1 := reader.Read(buf1)
		if err1 != nil || n1 != 5 || string(buf1) != "hello" {
			t.Error("First read failed")
		}

		// Second read
		buf2 := make([]byte, 6)
		n2, err2 := reader.Read(buf2)
		if err2 != nil || n2 != 6 || string(buf2) != " world" {
			t.Error("Second read failed")
		}

		// Third read should return EOF
		buf3 := make([]byte, 1)
		_, err3 := reader.Read(buf3)
		if err3 != io.EOF {
			t.Error("Expected EOF on third read")
		}
	})

	t.Run("Close", func(t *testing.T) {
		reader := &bodyReader{data: []byte("test"), pos: 0}

		err := reader.Close()
		if err != nil {
			t.Errorf("Unexpected error on close: %v", err)
		}

		// Should be able to close multiple times
		err = reader.Close()
		if err != nil {
			t.Errorf("Unexpected error on second close: %v", err)
		}
	})
}

func TestGetStringValue(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]interface{}
		key      string
		expected string
	}{
		{
			name:     "ValidString",
			m:        map[string]interface{}{"role": "user"},
			key:      "role",
			expected: "user",
		},
		{
			name:     "NonExistentKey",
			m:        map[string]interface{}{"role": "user"},
			key:      "content",
			expected: "",
		},
		{
			name:     "NonStringValue",
			m:        map[string]interface{}{"count": 42},
			key:      "count",
			expected: "",
		},
		{
			name:     "NilValue",
			m:        map[string]interface{}{"content": nil},
			key:      "content",
			expected: "",
		},
		{
			name:     "EmptyString",
			m:        map[string]interface{}{"content": ""},
			key:      "content",
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := getStringValue(test.m, test.key)
			if result != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, result)
			}
		})
	}
}
