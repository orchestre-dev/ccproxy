package security

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/orchestre-dev/ccproxy/internal/errors"
	"github.com/orchestre-dev/ccproxy/internal/utils"
)

// SecurityMiddleware provides security middleware for HTTP handlers
func SecurityMiddleware(manager *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Validate request
		if err := manager.ValidateRequest(c.Request); err != nil {
			handleSecurityError(c, err)
			return
		}

		// Add security headers
		addSecurityHeaders(c, manager.config)

		// Process request
		c.Next()

		// Log request timing
		duration := time.Since(start)
		if duration > 5*time.Second {
			utils.GetLogger().Warnf("Slow request detected: %s %s took %v", c.Request.Method, c.Request.URL.Path, duration)
		}
	}
}

// AuthMiddleware provides authentication middleware
func AuthMiddleware(manager *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !manager.config.RequireAuth {
			c.Next()
			return
		}

		// Check for API key in header
		apiKey := c.GetHeader(manager.config.APIKeyHeader)
		if apiKey == "" {
			// Check Authorization header
			auth := c.GetHeader("Authorization")
			if auth == "" {
				handleAuthError(c, "missing authentication")
				return
			}

			// Extract token from Bearer auth
			if strings.HasPrefix(auth, "Bearer ") {
				apiKey = strings.TrimPrefix(auth, "Bearer ")
			} else {
				handleAuthError(c, "invalid authorization format")
				return
			}
		}

		// Validate API key
		if err := manager.ValidateAPIKey(apiKey); err != nil {
			handleAuthError(c, err.Error())
			return
		}

		// Store API key hash in context for logging
		c.Set("api_key_hash", manager.hashAPIKey(apiKey))

		c.Next()
	}
}

// RateLimitMiddleware provides rate limiting middleware
func RateLimitMiddleware(limiter *IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil {
			c.Next()
			return
		}

		ip := getClientIP(c)
		if !limiter.Allow(ip) {
			info := limiter.GetLimit(ip)
			
			// Set rate limit headers
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", info.Limit))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", info.Remaining))
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", info.Reset.Unix()))
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
				"retry_after": info.Reset.Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CORSMiddleware provides CORS security middleware
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		
		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Max-Age", "86400")
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// CSRFMiddleware provides CSRF protection
func CSRFMiddleware(tokenHeader string) gin.HandlerFunc {
	if tokenHeader == "" {
		tokenHeader = "X-CSRF-Token"
	}

	return func(c *gin.Context) {
		// Skip CSRF check for safe methods
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Get CSRF token from header
		token := c.GetHeader(tokenHeader)
		if token == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "missing CSRF token",
			})
			c.Abort()
			return
		}

		// Validate token (simplified - in production, use proper CSRF token validation)
		if len(token) < 32 {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "invalid CSRF token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		
		c.Next()
	}
}

// SanitizationMiddleware sanitizes request and response data
func SanitizationMiddleware(sanitizer *DataSanitizer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Store original writer
		originalWriter := c.Writer
		
		// Create response capture writer
		captureWriter := &responseCapture{
			ResponseWriter: c.Writer,
			body:          []byte{},
		}
		c.Writer = captureWriter

		// Process request
		c.Next()

		// Restore original writer
		c.Writer = originalWriter

		// Sanitize response if needed
		if len(captureWriter.body) > 0 && sanitizer != nil {
			// For logging purposes, sanitize the response
			sanitizedBody := sanitizer.SanitizeString(string(captureWriter.body))
			if sanitizedBody != string(captureWriter.body) {
				utils.GetLogger().Debug("Response contained sensitive data that was sanitized for logging")
			}
		}

		// Write original response
		c.Writer.Write(captureWriter.body)
	}
}

// Helper functions

func addSecurityHeaders(c *gin.Context, config *SecurityConfig) {
	// Security headers
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Frame-Options", "DENY")
	c.Header("X-XSS-Protection", "1; mode=block")
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
	
	// HSTS header for HTTPS
	if config.EnableTLS {
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	}

	// CSP header based on security level
	switch config.Level {
	case SecurityLevelStrict, SecurityLevelParanoid:
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';")
	case SecurityLevelBasic:
		c.Header("Content-Security-Policy", "default-src 'self' https:; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline';")
	}
}

func handleSecurityError(c *gin.Context, err error) {
	var statusCode int
	var message string

	if ccErr, ok := err.(*errors.CCProxyError); ok {
		statusCode = ccErr.StatusCode
		message = ccErr.Message
	} else {
		statusCode = http.StatusBadRequest
		message = "security validation failed"
	}

	c.JSON(statusCode, gin.H{
		"error": message,
		"request_id": c.GetString("request_id"),
	})
	c.Abort()
}

func handleAuthError(c *gin.Context, message string) {
	c.Header("WWW-Authenticate", `Bearer realm="ccproxy"`)
	c.JSON(http.StatusUnauthorized, gin.H{
		"error": message,
		"request_id": c.GetString("request_id"),
	})
	c.Abort()
}

func getClientIP(c *gin.Context) string {
	// Try X-Forwarded-For first
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Try X-Real-IP
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return c.ClientIP()
}

// responseCapture captures response for sanitization
type responseCapture struct {
	gin.ResponseWriter
	body []byte
}

func (w *responseCapture) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return len(b), nil
}