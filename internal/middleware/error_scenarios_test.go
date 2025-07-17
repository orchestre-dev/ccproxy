// Package middleware provides comprehensive error scenario tests for middleware components
package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"ccproxy/internal/config"
	"ccproxy/pkg/logger"
)

func TestCORSErrorScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		origin         string
		environment    string
		expectedStatus int
		shouldAllow    bool
	}{
		{
			name:           "Valid Production Origin",
			origin:         "https://ccproxy.orchestre.dev",
			environment:    "production",
			expectedStatus: http.StatusOK,
			shouldAllow:    true,
		},
		{
			name:           "Invalid Production Origin",
			origin:         "https://malicious-site.com",
			environment:    "production",
			expectedStatus: http.StatusOK,
			shouldAllow:    false,
		},
		{
			name:           "Development Localhost",
			origin:         "http://localhost:3000",
			environment:    "development",
			expectedStatus: http.StatusOK,
			shouldAllow:    true,
		},
		{
			name:           "Development Any Origin",
			origin:         "https://any-site.com",
			environment:    "development",
			expectedStatus: http.StatusOK,
			shouldAllow:    true, // Dev allows any origin
		},
		{
			name:           "No Origin Header",
			origin:         "",
			environment:    "production",
			expectedStatus: http.StatusOK,
			shouldAllow:    false,
		},
		{
			name:           "Malformed Origin",
			origin:         "not-a-valid-url",
			environment:    "production",
			expectedStatus: http.StatusOK,
			shouldAllow:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment
			t.Setenv("SERVER_ENVIRONMENT", tt.environment)

			// Create test router
			router := gin.New()
			router.Use(CORS())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			// Create request
			req, err := http.NewRequestWithContext(context.Background(), "GET", "/test", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			// Record response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Check status
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d but got %d", tt.expectedStatus, w.Code)
			}

			// Check CORS headers
			allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
			if tt.shouldAllow {
				if allowOrigin != tt.origin {
					t.Errorf("Expected Access-Control-Allow-Origin '%s' but got '%s'", tt.origin, allowOrigin)
				}
			} else {
				// In production, disallowed origins should not have the header set
				if tt.environment == "production" && allowOrigin != "" {
					t.Errorf("Should not set Access-Control-Allow-Origin header for disallowed origin '%s' in production, but got '%s'", tt.origin, allowOrigin)
				}
			}
		})
	}
}

// Removed rate limiting test as it's not needed for a local proxy

func TestLoggingErrorScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create logger for testing
	mockLogger := logger.New(config.LoggingConfig{
		Level:  "info",
		Format: "json",
	})

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		expectLogs     bool
		expectedStatus int
	}{
		{
			name: "Normal Request",
			setupRequest: func() *http.Request {
				req, err := http.NewRequestWithContext(context.Background(), "GET", "/test", nil)
				if err != nil {
					panic(err)
				}
				return req
			},
			expectLogs:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name: "Request with Large Headers",
			setupRequest: func() *http.Request {
				req, err := http.NewRequestWithContext(context.Background(), "GET", "/test", nil)
				if err != nil {
					panic(err)
				}
				// Add very large header
				req.Header.Set("X-Large-Header", strings.Repeat("x", 8192))
				return req
			},
			expectLogs:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name: "Request with Special Characters",
			setupRequest: func() *http.Request {
				req, err := http.NewRequestWithContext(context.Background(), "GET", "/test?param=特殊文字", nil)
				if err != nil {
					panic(err)
				}
				req.Header.Set("User-Agent", "TestAgent/1.0 (特殊文字)")
				return req
			},
			expectLogs:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid HTTP Method",
			setupRequest: func() *http.Request {
				req, err := http.NewRequestWithContext(context.Background(), "INVALID", "/test", nil)
				if err != nil {
					panic(err)
				}
				return req
			},
			expectLogs:     true,
			expectedStatus: http.StatusNotFound, // Gin returns 404 for unregistered routes/methods
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test router with logging middleware
			router := gin.New()
			router.Use(Logger(mockLogger))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := tt.setupRequest()
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Verify response
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d but got %d", tt.expectedStatus, w.Code)
			}

			// In a real test, you'd capture log output and verify it
			// For now, we just verify the middleware doesn't crash
		})
	}
}

func TestRecoveryErrorScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLogger := logger.New(config.LoggingConfig{
		Level:  "error",
		Format: "json",
	})

	tests := []struct {
		name        string
		handler     gin.HandlerFunc
		expectPanic bool
	}{
		{
			name: "Normal Handler",
			handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			},
			expectPanic: false,
		},
		{
			name: "Nil Pointer Panic",
			handler: func(c *gin.Context) {
				var nilPointer *string
				_ = *nilPointer // This will panic
			},
			expectPanic: true,
		},
		{
			name: "Array Index Panic",
			handler: func(c *gin.Context) {
				arr := []int{1, 2, 3}
				_ = arr[10] // This will panic
			},
			expectPanic: true,
		},
		{
			name: "Type Assertion Panic",
			handler: func(c *gin.Context) {
				var i interface{} = "string"
				// This will panic - intentionally not checking the ok value
				v := i.(int)
				_ = v
			},
			expectPanic: true,
		},
		{
			name: "Manual Panic",
			handler: func(c *gin.Context) {
				panic("manual panic for testing")
			},
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test router with recovery middleware
			router := gin.New()
			router.Use(Recovery(mockLogger))
			router.GET("/test", tt.handler)

			req, err := http.NewRequestWithContext(context.Background(), "GET", "/test", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			if tt.expectPanic {
				// Recovery middleware should catch panic and return 500
				if w.Code != http.StatusInternalServerError {
					t.Errorf("Expected status 500 after panic but got %d", w.Code)
				}
			} else {
				// Normal request should succeed
				if w.Code != http.StatusOK {
					t.Errorf("Expected status 200 but got %d", w.Code)
				}
			}
		})
	}
}

func TestMiddlewareChainErrorScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLogger := logger.New(config.LoggingConfig{
		Level:  "info",
		Format: "json",
	})

	t.Run("Full Middleware Chain with Errors", func(t *testing.T) {
		// Create router with full middleware chain
		router := gin.New()
		router.Use(Recovery(mockLogger))
		router.Use(Logger(mockLogger))
		router.Use(CORS())

		// Add handler that sometimes panics
		router.POST("/test", func(c *gin.Context) {
			// Check for panic trigger
			if c.GetHeader("X-Trigger-Panic") == "true" {
				panic("Intentional test panic")
			}
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Test scenarios through full middleware chain
		scenarios := []struct {
			name    string
			headers map[string]string
			origin  string
		}{
			{
				name:    "Normal Request",
				headers: map[string]string{},
				origin:  "http://localhost:3000",
			},
			{
				name: "Request with Panic",
				headers: map[string]string{
					"X-Trigger-Panic": "true",
				},
				origin: "http://localhost:3000",
			},
			{
				name:    "CORS Preflight",
				headers: map[string]string{},
				origin:  "https://unknown-origin.com",
			},
		}

		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				// Set development environment for CORS
				t.Setenv("SERVER_ENVIRONMENT", "development")

				req, err := http.NewRequestWithContext(context.Background(), "POST", "/test", nil)
				if err != nil {
					t.Fatalf("Failed to create request: %v", err)
				}
				if scenario.origin != "" {
					req.Header.Set("Origin", scenario.origin)
				}
				for key, value := range scenario.headers {
					req.Header.Set(key, value)
				}

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				// Verify the middleware chain handled the request
				// (specific status depends on scenario)
				if w.Code == 0 {
					t.Error("Expected some HTTP status code")
				}

				t.Logf("Scenario '%s' resulted in status %d", scenario.name, w.Code)
			})
		}
	})
}

func TestEdgeCaseErrorScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Removed rate limiting tests as it's not needed for a local proxy

	t.Run("Concurrent CORS Requests", func(t *testing.T) {
		t.Setenv("SERVER_ENVIRONMENT", "production")

		router := gin.New()
		router.Use(CORS())
		router.GET("/test", func(c *gin.Context) {
			time.Sleep(10 * time.Millisecond) // Simulate processing
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Launch concurrent requests
		const numRequests = 20
		results := make(chan int, numRequests)

		for i := 0; i < numRequests; i++ {
			go func() {
				req, err := http.NewRequestWithContext(context.Background(), "GET", "/test", nil)
				if err != nil {
					panic(err)
				}
				req.Header.Set("Origin", "https://ccproxy.orchestre.dev")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				results <- w.Code
			}()
		}

		// Collect results
		successCount := 0
		for i := 0; i < numRequests; i++ {
			if code := <-results; code == http.StatusOK {
				successCount++
			}
		}

		if successCount != numRequests {
			t.Errorf("Expected all %d requests to succeed, but only %d did",
				numRequests, successCount)
		}
	})
}
