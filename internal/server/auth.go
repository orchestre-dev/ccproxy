package server

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// authMiddleware creates authentication middleware
func authMiddleware(apiKey string, enforceLocalhost bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for health endpoints
		path := c.Request.URL.Path
		if path == "/" || path == "/health" {
			c.Next()
			return
		}
		
		// If no API key is configured
		if apiKey == "" {
			// Enforce localhost-only access
			if enforceLocalhost && !isLocalhost(c) {
				Forbidden(c, "API access is restricted to localhost when no API key is configured")
				c.Abort()
				return
			}
			// Allow access from localhost
			c.Next()
			return
		}
		
		// Check Authorization header (Bearer token)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				if parts[1] == apiKey {
					c.Next()
					return
				}
			}
		}
		
		// Check x-api-key header
		if c.GetHeader("x-api-key") == apiKey {
			c.Next()
			return
		}
		
		// Authentication failed
		Unauthorized(c, "Invalid API key")
		c.Abort()
	}
}

// isLocalhost checks if the request is from localhost
func isLocalhost(c *gin.Context) bool {
	// Get client IP
	clientIP := c.ClientIP()
	
	// Check for localhost addresses
	localhostAddrs := []string{
		"127.0.0.1",
		"::1",
	}
	
	for _, addr := range localhostAddrs {
		if clientIP == addr {
			return true
		}
	}
	
	// Check if request is from local network (optional, more permissive)
	// This includes 192.168.x.x, 10.x.x.x, 172.16-31.x.x
	if strings.HasPrefix(clientIP, "192.168.") ||
		strings.HasPrefix(clientIP, "10.") ||
		strings.HasPrefix(clientIP, "172.") {
		// For now, we'll be strict and not allow local network
		return false
	}
	
	return false
}