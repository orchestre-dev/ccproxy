package middleware

import (
	"net/http"

	"ccproxy/pkg/logger"

	"github.com/gin-gonic/gin"
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

				logger.WithRequestID(requestID.(string)).WithField("panic", err).Error("Panic recovered")

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
					"request_id": requestID,
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}