package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ccproxy/pkg/logger"
)

// Logger returns a gin.HandlerFunc for request logging
func Logger(logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()

		// Add request ID to context
		c.Set("request_id", requestID)

		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(start).Milliseconds()

		// Log request
		logger.HTTPLog(
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
			requestID,
		)
	}
}
