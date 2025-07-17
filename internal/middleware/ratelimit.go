// Package middleware provides HTTP middleware for the CCProxy server
package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter represents a rate limiter for a specific client
type RateLimiter struct {
	clients   map[string]*ClientLimiter
	lastReset time.Time
	window    time.Duration
	mu        sync.RWMutex
	requests  int
}

// ClientLimiter tracks requests for a specific client
type ClientLimiter struct {
	resetTime time.Time
	mu        sync.RWMutex
	count     int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requests int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests:  requests,
		window:    window,
		clients:   make(map[string]*ClientLimiter),
		lastReset: time.Now(),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if a request from the given client is allowed
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	client, exists := rl.clients[clientID]

	if !exists {
		client = &ClientLimiter{
			count:     1,
			resetTime: now.Add(rl.window),
		}
		rl.clients[clientID] = client
		return true
	}

	client.mu.Lock()
	defer client.mu.Unlock()

	// Reset counter if window has passed
	if now.After(client.resetTime) {
		client.count = 1
		client.resetTime = now.Add(rl.window)
		return true
	}

	// Check if limit exceeded
	if client.count >= rl.requests {
		return false
	}

	client.count++
	return true
}

// cleanup removes expired clients
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()

		for clientID, client := range rl.clients {
			client.mu.RLock()
			if now.After(client.resetTime) {
				delete(rl.clients, clientID)
			}
			client.mu.RUnlock()
		}

		rl.mu.Unlock()
	}
}

// GetClientID extracts client identifier from request
func GetClientID(c *gin.Context) string {
	// Try to get API key from Authorization header
	if auth := c.GetHeader("Authorization"); auth != "" {
		// Use hash of API key for privacy
		return fmt.Sprintf("api_%x", auth)
	}

	// Fall back to IP address
	return c.ClientIP()
}

// RateLimit returns a gin middleware for rate limiting
func RateLimit(requests int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(requests, window)

	return func(c *gin.Context) {
		clientID := GetClientID(c)

		if !limiter.Allow(clientID) {
			// Add rate limit headers
			c.Header("X-RateLimit-Limit", strconv.Itoa(requests))
			c.Header("X-RateLimit-Window", window.String())
			c.Header("Retry-After", strconv.Itoa(int(window.Seconds())))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     fmt.Sprintf("Too many requests. Limit: %d per %s", requests, window),
				"retry_after": window.String(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
