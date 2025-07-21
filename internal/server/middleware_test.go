package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/orchestre-dev/ccproxy/internal/config"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuthMiddleware(t *testing.T) {
	t.Run("NoAPIKeyLocalhostAccess", func(t *testing.T) {
		middleware := authMiddleware("", true)
		
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		// Test localhost access
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for localhost access, got %d", w.Code)
		}
	})

	t.Run("NoAPIKeyNonLocalhostBlocked", func(t *testing.T) {
		middleware := authMiddleware("", true)
		
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		// Test non-localhost access
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403 for non-localhost access, got %d", w.Code)
		}
	})

	t.Run("ValidBearerToken", func(t *testing.T) {
		apiKey := "test-api-key"
		middleware := authMiddleware(apiKey, true)
		
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+apiKey)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for valid bearer token, got %d", w.Code)
		}
	})

	t.Run("ValidXAPIKey", func(t *testing.T) {
		apiKey := "test-api-key"
		middleware := authMiddleware(apiKey, true)
		
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("x-api-key", apiKey)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for valid x-api-key, got %d", w.Code)
		}
	})

	t.Run("InvalidAPIKey", func(t *testing.T) {
		apiKey := "test-api-key"
		middleware := authMiddleware(apiKey, true)
		
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer wrong-key")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 for invalid API key, got %d", w.Code)
		}
	})

	t.Run("SkipAuthForHealthEndpoints", func(t *testing.T) {
		apiKey := "test-api-key"
		middleware := authMiddleware(apiKey, true)
		
		router := gin.New()
		router.Use(middleware)
		router.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "healthy"})
		})

		// Test access without auth to health endpoint
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for health endpoint without auth, got %d", w.Code)
		}
	})

	t.Run("CaseInsensitiveBearerToken", func(t *testing.T) {
		apiKey := "test-api-key"
		middleware := authMiddleware(apiKey, true)
		
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "BEARER "+apiKey) // uppercase
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for case-insensitive bearer token, got %d", w.Code)
		}
	})
}

func TestIsLocalhost(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		expected   bool
	}{
		{"IPv4 localhost", "127.0.0.1:12345", true},
		{"IPv6 localhost", "[::1]:12345", true},
		{"Private network", "192.168.1.100:12345", false},
		{"Private network 10.x", "10.0.0.1:12345", false},
		{"Private network 172.x", "172.16.0.1:12345", false},
		{"Public IP", "8.8.8.8:12345", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = test.remoteAddr
			c.Request = req

			result := isLocalhost(c)
			if result != test.expected {
				t.Errorf("Expected %t for %s, got %t", test.expected, test.remoteAddr, result)
			}
		})
	}
}

func TestCORSMiddleware(t *testing.T) {
	middleware := corsMiddleware()
	
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	t.Run("NormalRequest", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Check CORS headers
		if w.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Error("Expected Access-Control-Allow-Origin header")
		}

		if w.Header().Get("Access-Control-Allow-Methods") == "" {
			t.Error("Expected Access-Control-Allow-Methods header")
		}

		if w.Header().Get("Access-Control-Allow-Headers") == "" {
			t.Error("Expected Access-Control-Allow-Headers header")
		}
	})

	t.Run("PreflightRequest", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status 204 for OPTIONS request, got %d", w.Code)
		}

		// Check CORS headers
		if w.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Error("Expected Access-Control-Allow-Origin header for preflight")
		}
	})
}

func TestRequestSizeLimitMiddleware(t *testing.T) {
	t.Run("RequestWithinLimit", func(t *testing.T) {
		maxSize := int64(100) // 100 bytes
		middleware := requestSizeLimitMiddleware(maxSize)
		
		router := gin.New()
		router.Use(middleware)
		router.POST("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		smallBody := `{"message": "small"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/test", strings.NewReader(smallBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for small request, got %d", w.Code)
		}
	})

	t.Run("RequestExceedsLimit", func(t *testing.T) {
		maxSize := int64(10) // 10 bytes
		middleware := requestSizeLimitMiddleware(maxSize)
		
		router := gin.New()
		router.Use(middleware)
		router.POST("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		largeBody := `{"message": "this is a very long message that exceeds the limit"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/test", strings.NewReader(largeBody))
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = int64(len(largeBody))
		router.ServeHTTP(w, req)

		if w.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("Expected status 413 for large request, got %d", w.Code)
		}
	})

	t.Run("DisabledSizeLimit", func(t *testing.T) {
		maxSize := int64(0) // Disabled
		middleware := requestSizeLimitMiddleware(maxSize)
		
		router := gin.New()
		router.Use(middleware)
		router.POST("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		largeBody := strings.Repeat("x", 1000000) // 1MB
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/test", strings.NewReader(largeBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 when size limit disabled, got %d", w.Code)
		}
	})

	t.Run("NoContentLength", func(t *testing.T) {
		maxSize := int64(100)
		middleware := requestSizeLimitMiddleware(maxSize)
		
		router := gin.New()
		router.Use(middleware)
		router.POST("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		body := `{"message": "test"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		// Don't set Content-Length
		router.ServeHTTP(w, req)

		// Should still work and be limited by MaxBytesReader
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 when no content-length, got %d", w.Code)
		}
	})
}

func TestLoggingMiddleware(t *testing.T) {
	middleware := loggingMiddleware()
	
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test?param=value", nil)
	req.Header.Set("User-Agent", "test-agent")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test that middleware doesn't crash (logging is hard to test without capturing output)
}

func TestPerformanceMiddleware(t *testing.T) {
	// Create a mock server with performance monitoring
	cfg := &config.Config{
		Host: "127.0.0.1",
		Port: 3456,
		Performance: config.PerformanceConfig{
			MetricsEnabled: true,
		},
		Routes: map[string]config.Route{
			"default": {
				Provider: "openai",
				Model:    "gpt-4",
			},
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	middleware := server.performanceMiddleware()
	
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.Set("provider", "openai")
		c.Set("model", "gpt-4")
		c.JSON(200, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check that requests served counter increased
	if server.requestsServed <= 0 {
		t.Error("Expected requests served counter to increase")
	}
}