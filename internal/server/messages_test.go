package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/musistudio/ccproxy/internal/config"
)

func TestMessagesEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		request        interface{}
		apiKey         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "valid message request",
			request: MessageRequest{
				Model: "claude-3-opus-20240229",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				MaxTokens: 100,
			},
			apiKey:         "test-key",
			expectedStatus: http.StatusNotImplemented,
			expectedBody:   "Message processing pipeline not yet implemented",
		},
		{
			name: "missing model",
			request: map[string]interface{}{
				"messages": []map[string]string{
					{"role": "user", "content": "Hello"},
				},
			},
			apiKey:         "test-key",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing messages",
			request: map[string]interface{}{
				"model": "claude-3-opus-20240229",
			},
			apiKey:         "test-key",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty messages array",
			request: MessageRequest{
				Model:    "claude-3-opus-20240229",
				Messages: []Message{},
			},
			apiKey:         "test-key",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "message without role",
			request: map[string]interface{}{
				"model": "claude-3-opus-20240229",
				"messages": []map[string]string{
					{"content": "Hello"},
				},
			},
			apiKey:         "test-key",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "message without content",
			request: map[string]interface{}{
				"model": "claude-3-opus-20240229",
				"messages": []map[string]string{
					{"role": "user"},
				},
			},
			apiKey:         "test-key",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "with streaming enabled",
			request: MessageRequest{
				Model: "claude-3-opus-20240229",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				Stream: true,
			},
			apiKey:         "test-key",
			expectedStatus: http.StatusNotImplemented,
		},
		{
			name: "with system prompt",
			request: MessageRequest{
				Model: "claude-3-opus-20240229",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				System: "You are a helpful assistant",
			},
			apiKey:         "test-key",
			expectedStatus: http.StatusNotImplemented,
		},
	}

	// Create test server
	cfg := &config.Config{
		APIKey: "test-key",
		Port:   3456,
		Host:   "127.0.0.1",
	}

	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if tt.apiKey != "" {
				req.Header.Set("x-api-key", tt.apiKey)
			}

			// Perform request
			w := httptest.NewRecorder()
			srv.router.ServeHTTP(w, req)

			// Check status
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			// Check body
			if tt.expectedBody != "" && !contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("Expected body to contain '%s', got '%s'", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestMessagesAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test with no API key configured (localhost only)
	cfg := &config.Config{
		APIKey: "",
		Port:   3456,
		Host:   "127.0.0.1",
	}

	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	validRequest := MessageRequest{
		Model: "claude-3-opus-20240229",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	tests := []struct {
		name           string
		clientIP       string
		expectedStatus int
	}{
		{
			name:           "localhost allowed without API key",
			clientIP:       "127.0.0.1",
			expectedStatus: http.StatusNotImplemented, // Because endpoint is not implemented yet
		},
		{
			name:           "::1 allowed without API key",
			clientIP:       "::1",
			expectedStatus: http.StatusNotImplemented,
		},
		{
			name:           "external IP blocked without API key",
			clientIP:       "1.2.3.4",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(validRequest)
			req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Real-IP", tt.clientIP)

			w := httptest.NewRecorder()
			srv.router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}