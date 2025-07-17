package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS returns a gin.HandlerFunc for handling CORS headers with environment-based security
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Get environment and configure allowed origins
		env := os.Getenv("SERVER_ENVIRONMENT")
		if env == "" {
			env = os.Getenv("ENV")
		}
		if env == "" {
			env = os.Getenv("ENVIRONMENT")
		}
		
		var allowedOrigins []string
		if env == "development" || env == "dev" {
			// Development: Allow localhost and common development origins
			allowedOrigins = []string{
				"http://localhost:3000",
				"http://localhost:8080",
				"http://localhost:5173",
				"http://127.0.0.1:3000",
				"http://127.0.0.1:8080",
				"http://127.0.0.1:5173",
			}
		} else {
			// Production: Restrict to specific domains
			allowedOrigins = []string{
				"https://ccproxy.orchestre.dev",
				"https://docs.ccproxy.orchestre.dev",
			}
			
			// Allow custom origins from environment variable
			if customOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); customOrigins != "" {
				allowedOrigins = append(allowedOrigins, strings.Split(customOrigins, ",")...)
			}
		}
		
		// Check if origin is allowed
		isAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				isAllowed = true
				break
			}
		}
		
		// Set CORS headers
		if isAllowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		} else if env == "development" || env == "dev" {
			// In development, still allow requests but log the origin
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, anthropic-version, x-api-key")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
