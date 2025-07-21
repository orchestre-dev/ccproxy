package performance

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Middleware creates performance monitoring middleware for Gin
func Middleware(monitor *Monitor) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip performance monitoring for health/status endpoints
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/status" {
			c.Next()
			return
		}

		// Check resource limits before processing
		if err := monitor.CheckResourceLimits(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": gin.H{
					"message": "Service temporarily unavailable due to resource constraints",
					"type":    "resource_exhausted",
					"code":    "RESOURCE_LIMIT_EXCEEDED",
				},
			})
			c.Abort()
			return
		}

		// Extract provider from request
		provider := extractProvider(c)

		// Check circuit breaker if provider is known
		if provider != "" && !monitor.CheckCircuitBreaker(provider) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": gin.H{
					"message": "Provider temporarily unavailable",
					"type":    "provider_unavailable",
					"code":    "CIRCUIT_BREAKER_OPEN",
				},
			})
			c.Abort()
			return
		}

		// Check rate limiting
		rateLimitKey := getRateLimitKey(c, monitor.config.RateLimit)
		if !monitor.CheckRateLimit(rateLimitKey) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"message": "Rate limit exceeded",
					"type":    "rate_limit_error",
					"code":    "RATE_LIMIT_EXCEEDED",
				},
			})
			c.Abort()
			return
		}

		// Capture request size
		requestSize := int64(0)
		if c.Request.Body != nil {
			// Read body to calculate size
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				requestSize = int64(len(bodyBytes))
				// Restore body for downstream handlers
				c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

				// Check request size limit
				if err := monitor.resourceMonitor.CheckRequestSize(requestSize); err != nil {
					c.JSON(http.StatusRequestEntityTooLarge, gin.H{
						"error": gin.H{
							"message": err.Error(),
							"type":    "request_too_large",
							"code":    "REQUEST_SIZE_LIMIT_EXCEEDED",
						},
					})
					c.Abort()
					return
				}
			}
		}

		// Start timing
		startTime := time.Now()

		// Capture response for metrics
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		// Process request
		c.Next()

		// Calculate metrics
		latency := time.Since(startTime)

		// Extract model from request context
		model := c.GetString("model")
		if model == "" {
			model = extractModel(c)
		}

		// Record request metrics
		metrics := RequestMetrics{
			Provider:     provider,
			Model:        model,
			StartTime:    startTime,
			EndTime:      time.Now(),
			Latency:      latency,
			Success:      writer.status < 400,
			StatusCode:   writer.status,
			RequestSize:  requestSize,
			ResponseSize: int64(writer.body.Len()),
		}

		// Extract token counts if available
		if tokensIn, exists := c.Get("tokens_in"); exists {
			if tokens, ok := tokensIn.(int); ok {
				metrics.TokensIn = tokens
			}
		}
		if tokensOut, exists := c.Get("tokens_out"); exists {
			if tokens, ok := tokensOut.(int); ok {
				metrics.TokensOut = tokens
			}
		}

		// Record the metrics
		monitor.RecordRequest(metrics)

		// Update circuit breaker
		if provider != "" {
			monitor.RecordProviderError(provider, metrics.Success)
		}
	}
}

// extractProvider extracts the provider name from the request context or headers
func extractProvider(c *gin.Context) string {
	// First check if provider was set in context by routing
	if provider, exists := c.Get("provider"); exists {
		if p, ok := provider.(string); ok {
			return p
		}
	}

	// Try to extract from model parameter (format: provider,model)
	model := extractModel(c)
	if model != "" {
		if idx := bytes.IndexByte([]byte(model), ','); idx > 0 {
			return model[:idx]
		}
	}

	return ""
}

// extractModel extracts the model from request
func extractModel(c *gin.Context) string {
	// Check context first
	if model, exists := c.Get("model"); exists {
		if m, ok := model.(string); ok {
			return m
		}
	}

	// Try to get from request body
	var body map[string]interface{}
	if c.Request.Body != nil {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err == nil {
			// Restore body
			c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			// Try to parse
			if err := json.Unmarshal(bodyBytes, &body); err == nil {
				if model, ok := body["model"].(string); ok {
					return model
				}
			}
		}
	}

	return ""
}

// getRateLimitKey determines the rate limit key based on configuration
func getRateLimitKey(c *gin.Context, config RateLimitConfig) string {
	if config.PerAPIKey {
		// Extract API key from Authorization header
		auth := c.GetHeader("Authorization")
		if auth != "" {
			// Simple hash of API key for privacy
			return fmt.Sprintf("key:%x", sha256.Sum256([]byte(auth)))
		}
	}

	if config.PerProvider {
		provider := extractProvider(c)
		if provider != "" {
			return fmt.Sprintf("provider:%s", provider)
		}
	}

	// Default to IP-based rate limiting
	return fmt.Sprintf("ip:%s", c.ClientIP())
}

// responseWriter captures response data for metrics
type responseWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// GetPerformanceHandler returns a handler for performance metrics endpoint
func GetPerformanceHandler(monitor *Monitor) gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics := monitor.GetMetrics()
		c.JSON(http.StatusOK, metrics)
	}
}
