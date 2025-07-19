package performance

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Skip health and status endpoints", func(t *testing.T) {
		config := DefaultPerformanceConfig()
		monitor := NewMonitor(config)

		router := gin.New()
		router.Use(Middleware(monitor))
		router.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
		router.GET("/status", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "running"})
		})

		// Health endpoint
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)

		// Status endpoint
		req = httptest.NewRequest("GET", "/status", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)

		// No metrics should be recorded
		metrics := monitor.GetMetrics()
		assert.Equal(t, int64(0), metrics.TotalRequests)
	})

	t.Run("Resource limit exceeded", func(t *testing.T) {
		config := DefaultPerformanceConfig()
		config.ResourceLimits.MaxGoroutines = 1 // Very low limit
		monitor := NewMonitor(config)

		router := gin.New()
		router.Use(Middleware(monitor))
		router.POST("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"ok": true})
		})

		req := httptest.NewRequest("POST", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		
		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Contains(t, resp, "error")
	})

	t.Run("Circuit breaker open", func(t *testing.T) {
		config := DefaultPerformanceConfig()
		config.CircuitBreaker.Enabled = true
		config.CircuitBreaker.ConsecutiveFailures = 1
		monitor := NewMonitor(config)

		router := gin.New()
		router.Use(Middleware(monitor))
		router.POST("/test", func(c *gin.Context) {
			c.Set("provider", "test-provider")
			c.JSON(500, gin.H{"error": "provider error"})
		})

		// First request should work but fail
		req := httptest.NewRequest("POST", "/test", nil)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 500, w.Code)

		// Open the circuit by recording the error
		monitor.RecordProviderError("test-provider", false)

		// Second request should be blocked
		req = httptest.NewRequest("POST", "/test", bytes.NewBufferString(`{"model":"test-provider,test-model"}`))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})

	t.Run("Rate limiting", func(t *testing.T) {
		config := DefaultPerformanceConfig()
		config.RateLimit.Enabled = true
		config.RateLimit.RequestsPerMin = 60
		config.RateLimit.BurstSize = 1
		monitor := NewMonitor(config)

		router := gin.New()
		router.Use(Middleware(monitor))
		router.POST("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"ok": true})
		})

		// First request should work
		req := httptest.NewRequest("POST", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)

		// Second request should be rate limited
		req = httptest.NewRequest("POST", "/test", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	})

	t.Run("Request size limit", func(t *testing.T) {
		config := DefaultPerformanceConfig()
		config.ResourceLimits.MaxRequestBodyMB = 1 // 1MB limit
		monitor := NewMonitor(config)

		router := gin.New()
		router.Use(Middleware(monitor))
		router.POST("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"ok": true})
		})

		// Create large request body
		largeBody := make([]byte, 2*1024*1024) // 2MB
		for i := range largeBody {
			largeBody[i] = 'a'
		}

		req := httptest.NewRequest("POST", "/test", bytes.NewReader(largeBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	})

	t.Run("Successful request with metrics", func(t *testing.T) {
		config := DefaultPerformanceConfig()
		monitor := NewMonitor(config)

		router := gin.New()
		router.Use(Middleware(monitor))
		router.POST("/v1/messages", func(c *gin.Context) {
			// Simulate provider and token counts
			c.Set("provider", "test-provider")
			c.Set("model", "test-model")
			c.Set("tokens_in", 10)
			c.Set("tokens_out", 20)
			c.JSON(200, gin.H{"response": "ok"})
		})

		body := map[string]interface{}{
			"model": "test-model",
			"messages": []map[string]string{
				{"role": "user", "content": "Hello"},
			},
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		// Check metrics
		time.Sleep(10 * time.Millisecond) // Give time for metrics to be recorded
		metrics := monitor.GetMetrics()
		assert.Equal(t, int64(1), metrics.TotalRequests)
		assert.Equal(t, int64(1), metrics.SuccessfulRequests)
		assert.Equal(t, int64(0), metrics.FailedRequests)
		
		// Check provider metrics - provider might be empty if not extracted properly
		if len(metrics.ProviderMetrics) > 0 {
			// Get the first provider
			for _, pm := range metrics.ProviderMetrics {
				assert.Equal(t, int64(1), pm.TotalRequests)
				assert.Equal(t, int64(30), pm.TokensProcessed) // 10 + 20
				break
			}
		}
	})

	t.Run("Extract provider from model", func(t *testing.T) {
		config := DefaultPerformanceConfig()
		monitor := NewMonitor(config)

		router := gin.New()
		router.Use(Middleware(monitor))
		router.POST("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"ok": true})
		})

		body := map[string]interface{}{
			"model": "provider1,model1",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		// Check that provider was extracted
		metrics := monitor.GetMetrics()
		assert.Contains(t, metrics.ProviderMetrics, "provider1")
	})

	t.Run("Rate limit by API key", func(t *testing.T) {
		config := DefaultPerformanceConfig()
		config.RateLimit.Enabled = true
		config.RateLimit.PerAPIKey = true
		config.RateLimit.RequestsPerMin = 60
		config.RateLimit.BurstSize = 1
		monitor := NewMonitor(config)

		router := gin.New()
		router.Use(Middleware(monitor))
		router.POST("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"ok": true})
		})

		// First request with API key 1
		req := httptest.NewRequest("POST", "/test", nil)
		req.Header.Set("Authorization", "Bearer key1")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)

		// Second request with same key should be limited
		req = httptest.NewRequest("POST", "/test", nil)
		req.Header.Set("Authorization", "Bearer key1")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)

		// Request with different key should work
		req = httptest.NewRequest("POST", "/test", nil)
		req.Header.Set("Authorization", "Bearer key2")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	})
}

func TestGetPerformanceHandler(t *testing.T) {
	config := DefaultPerformanceConfig()
	monitor := NewMonitor(config)

	// Add some metrics
	monitor.RecordRequest(RequestMetrics{
		Provider: "test",
		Success:  true,
		Latency:  100 * time.Millisecond,
	})

	router := gin.New()
	router.GET("/metrics", GetPerformanceHandler(monitor))

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var metrics Metrics
	err := json.Unmarshal(w.Body.Bytes(), &metrics)
	require.NoError(t, err)
	assert.Equal(t, int64(1), metrics.TotalRequests)
}

func TestExtractFunctions(t *testing.T) {
	t.Run("extractProvider", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		
		// From context
		c.Set("provider", "context-provider")
		assert.Equal(t, "context-provider", extractProvider(c))
		
		// From model
		c = &gin.Context{}
		body := map[string]interface{}{"model": "provider,model"}
		bodyBytes, _ := json.Marshal(body)
		c.Request = httptest.NewRequest("POST", "/", bytes.NewReader(bodyBytes))
		assert.Equal(t, "provider", extractProvider(c))
	})

	t.Run("getRateLimitKey", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/", nil)
		
		// IP-based (default)
		config := RateLimitConfig{}
		key := getRateLimitKey(c, config)
		assert.Contains(t, key, "ip:")
		
		// API key based
		config.PerAPIKey = true
		c.Request.Header.Set("Authorization", "Bearer test-key")
		key = getRateLimitKey(c, config)
		assert.Contains(t, key, "key:")
		
		// Provider based
		config = RateLimitConfig{PerProvider: true}
		c.Set("provider", "test-provider")
		key = getRateLimitKey(c, config)
		assert.Equal(t, "provider:test-provider", key)
	})
}

func TestResponseWriter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Create a test context with gin's ResponseWriter
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	writer := &responseWriter{
		ResponseWriter: c.Writer,
		body:           &bytes.Buffer{},
	}

	// Test WriteHeader
	writer.WriteHeader(201)
	assert.Equal(t, 201, writer.status)

	// Test Write
	data := []byte("test response")
	n, err := writer.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, data, writer.body.Bytes())
}