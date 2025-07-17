// Package middleware provides HTTP middleware for the CCProxy server
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ccproxy/pkg/logger"
)

// Recovery returns a gin.HandlerFunc for panic recovery
func Recovery(logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, exists := c.Get("request_id")
				if !exists {
					requestID = "unknown"
				}

				if requestIDStr, ok := requestID.(string); ok {
					logger.WithRequestID(requestIDStr).WithField("panic", err).Error("Panic recovered")
				} else {
					logger.WithField("panic", err).Error("Panic recovered")
				}

				c.JSON(http.StatusInternalServerError, gin.H{
					"error":      "Internal server error",
					"request_id": requestID,
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}
