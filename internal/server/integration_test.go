package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/musistudio/ccproxy/internal/config"
)

func TestServerIntegration(t *testing.T) {
	// Create test config with API key
	cfg := &config.Config{
		Host:   "127.0.0.1",
		Port:   0, // Use random port
		APIKey: "test-integration-key",
		Log:    false,
		Providers: []config.Provider{
			{
				Name:       "test-provider",
				APIBaseURL: "https://api.test.com",
				APIKey:     "provider-key",
				Models:     []string{"test-model"},
				Enabled:    true,
			},
		},
	}
	
	// Create server
	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	
	// Start test server
	ts := httptest.NewServer(srv.router)
	defer ts.Close()
	
	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)
	
	tests := []struct {
		name           string
		path           string
		headers        map[string]string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "health endpoint no auth",
			path:           "/",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"LLMs API","version":"1.0.0"}`,
		},
		{
			name:           "providers without auth",
			path:           "/providers",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid API key",
		},
		{
			name: "providers with valid auth",
			path: "/providers",
			headers: map[string]string{
				"x-api-key": "test-integration-key",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "providers with invalid auth",
			path: "/providers",
			headers: map[string]string{
				"x-api-key": "wrong-key",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid API key",
		},
		{
			name: "providers with bearer token",
			path: "/providers",
			headers: map[string]string{
				"Authorization": "Bearer test-integration-key",
			},
			expectedStatus: http.StatusOK,
		},
	}
	
	client := &http.Client{Timeout: 5 * time.Second}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", ts.URL+tt.path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			
			// Add headers
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()
			
			// Check status
			if resp.StatusCode != tt.expectedStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, resp.StatusCode, body)
			}
			
			// Check body if specified
			if tt.expectedBody != "" {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read body: %v", err)
				}
				
				if !contains(string(body), tt.expectedBody) {
					t.Errorf("Expected body to contain '%s', got '%s'", tt.expectedBody, string(body))
				}
			}
		})
	}
}