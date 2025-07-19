package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/orchestre-dev/ccproxy/internal/config"
)

func TestAuthMiddleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name           string
		apiKey         string
		requestPath    string
		authHeader     string
		xApiKeyHeader  string
		clientIP       string
		expectedStatus int
		expectedBody   string
	}{
		// Health endpoints should bypass auth
		{
			name:           "root endpoint bypasses auth",
			apiKey:         "test-key",
			requestPath:    "/",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "health endpoint bypasses auth",
			apiKey:         "test-key",
			requestPath:    "/health",
			expectedStatus: http.StatusOK,
		},
		
		// No API key configured - localhost access
		{
			name:           "no api key - localhost allowed",
			apiKey:         "",
			requestPath:    "/v1/messages",
			clientIP:       "127.0.0.1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no api key - ::1 localhost allowed",
			apiKey:         "",
			requestPath:    "/v1/messages",
			clientIP:       "::1",
			expectedStatus: http.StatusOK,
		},
		
		// API key configured - valid auth
		{
			name:           "valid bearer token",
			apiKey:         "test-key",
			requestPath:    "/v1/messages",
			authHeader:     "Bearer test-key",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid x-api-key header",
			apiKey:         "test-key",
			requestPath:    "/v1/messages",
			xApiKeyHeader:  "test-key",
			expectedStatus: http.StatusOK,
		},
		
		// API key configured - invalid auth
		{
			name:           "invalid bearer token",
			apiKey:         "test-key",
			requestPath:    "/v1/messages",
			authHeader:     "Bearer wrong-key",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid API key",
		},
		{
			name:           "invalid x-api-key header",
			apiKey:         "test-key",
			requestPath:    "/v1/messages",
			xApiKeyHeader:  "wrong-key",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid API key",
		},
		{
			name:           "missing auth",
			apiKey:         "test-key",
			requestPath:    "/v1/messages",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid API key",
		},
		{
			name:           "malformed bearer token",
			apiKey:         "test-key",
			requestPath:    "/v1/messages",
			authHeader:     "Bearer",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid API key",
		},
		{
			name:           "wrong auth type",
			apiKey:         "test-key",
			requestPath:    "/v1/messages",
			authHeader:     "Basic test-key",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid API key",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router with middleware
			router := gin.New()
			router.Use(authMiddleware(tt.apiKey, false))
			
			// Add test routes
			router.GET("/", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})
			router.GET("/health", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})
			router.POST("/v1/messages", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})
			
			// Create request
			req := httptest.NewRequest("GET", tt.requestPath, nil)
			if tt.requestPath == "/v1/messages" {
				req.Method = "POST"
			}
			
			// Set headers
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			if tt.xApiKeyHeader != "" {
				req.Header.Set("x-api-key", tt.xApiKeyHeader)
			}
			
			// Set client IP if specified
			if tt.clientIP != "" {
				req.Header.Set("X-Real-IP", tt.clientIP)
			}
			
			// Perform request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			// Check status
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			
			// Check body if expected
			if tt.expectedBody != "" && !contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("Expected body to contain '%s', got '%s'", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestIsLocalhost(t *testing.T) {
	tests := []struct {
		name     string
		clientIP string
		expected bool
	}{
		{"127.0.0.1", "127.0.0.1", true},
		{"::1", "::1", true},
		{"localhost resolves to 127.0.0.1", "127.0.0.1", true}, // Gin resolves localhost to IP
		{"external IP", "8.8.8.8", false},
		{"local network 192.168", "192.168.1.100", false},
		{"local network 10.x", "10.0.0.1", false},
		{"local network 172.x", "172.16.0.1", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)
			c.Request.Header.Set("X-Real-IP", tt.clientIP)
			
			result := isLocalhost(c)
			if result != tt.expected {
				t.Errorf("Expected isLocalhost(%s) = %v, got %v", tt.clientIP, tt.expected, result)
			}
		})
	}
}

func TestSecurityConstraint(t *testing.T) {
	tests := []struct {
		name         string
		config       *config.Config
		expectedHost string
	}{
		{
			name: "no api key with external host forces localhost",
			config: &config.Config{
				APIKey: "",
				Host:   "0.0.0.0",
				Port:   3456,
			},
			expectedHost: "127.0.0.1",
		},
		{
			name: "no api key with localhost keeps localhost",
			config: &config.Config{
				APIKey: "",
				Host:   "127.0.0.1",
				Port:   3456,
			},
			expectedHost: "127.0.0.1",
		},
		{
			name: "api key set allows external host",
			config: &config.Config{
				APIKey: "test-key",
				Host:   "0.0.0.0",
				Port:   3456,
			},
			expectedHost: "0.0.0.0",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create server (which applies security constraint)
			server, err := New(tt.config)
			if err != nil {
				t.Fatalf("Failed to create server: %v", err)
			}
			
			// Check that host was set correctly
			if server.config.Host != tt.expectedHost {
				t.Errorf("Expected host %s, got %s", tt.expectedHost, server.config.Host)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}