package performance

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter implements rate limiting for requests
type RateLimiter struct {
	config    RateLimitConfig
	limiters  map[string]*rate.Limiter
	lastClean time.Time
	hits      int64
	mu        sync.RWMutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		config:    config,
		limiters:  make(map[string]*rate.Limiter),
		lastClean: time.Now(),
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	limiter, exists := rl.limiters[key]
	if !exists {
		// Create new limiter for this key
		ratePerSecond := rate.Limit(float64(rl.config.RequestsPerMin) / 60.0)
		limiter = rate.NewLimiter(ratePerSecond, rl.config.BurstSize)
		rl.limiters[key] = limiter
	}
	rl.mu.Unlock()

	allowed := limiter.Allow()
	if !allowed {
		atomic.AddInt64(&rl.hits, 1)
	}

	return allowed
}

// AllowN checks if n requests are allowed
func (rl *RateLimiter) AllowN(key string, n int) bool {
	rl.mu.Lock()
	limiter, exists := rl.limiters[key]
	if !exists {
		ratePerSecond := rate.Limit(float64(rl.config.RequestsPerMin) / 60.0)
		limiter = rate.NewLimiter(ratePerSecond, rl.config.BurstSize)
		rl.limiters[key] = limiter
	}
	rl.mu.Unlock()

	allowed := limiter.AllowN(time.Now(), n)
	if !allowed {
		atomic.AddInt64(&rl.hits, int64(n))
	}

	return allowed
}

// Wait blocks until a request is allowed
func (rl *RateLimiter) Wait(key string) error {
	rl.mu.Lock()
	limiter, exists := rl.limiters[key]
	if !exists {
		ratePerSecond := rate.Limit(float64(rl.config.RequestsPerMin) / 60.0)
		limiter = rate.NewLimiter(ratePerSecond, rl.config.BurstSize)
		rl.limiters[key] = limiter
	}
	rl.mu.Unlock()

	return limiter.Wait(context.Background())
}

// GetHits returns the number of rate limit hits
func (rl *RateLimiter) GetHits() int64 {
	return atomic.LoadInt64(&rl.hits)
}

// ResetHits resets the hit counter
func (rl *RateLimiter) ResetHits() {
	atomic.StoreInt64(&rl.hits, 0)
}

// GetLimiterCount returns the number of active limiters
func (rl *RateLimiter) GetLimiterCount() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return len(rl.limiters)
}

// cleanupLoop periodically cleans up unused limiters
func (rl *RateLimiter) cleanupLoop() {
	// Read initial cleanup interval with lock
	rl.mu.RLock()
	cleanupInterval := rl.config.CleanupInterval
	rl.mu.RUnlock()

	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()

		// Update ticker interval if config changed
		rl.mu.RLock()
		newInterval := rl.config.CleanupInterval
		rl.mu.RUnlock()

		if newInterval != cleanupInterval {
			ticker.Stop()
			ticker = time.NewTicker(newInterval)
			cleanupInterval = newInterval
		}
	}
}

// cleanup removes inactive limiters
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	if now.Sub(rl.lastClean) < rl.config.CleanupInterval {
		return
	}

	// For simplicity, we'll clear all limiters that haven't been used recently
	// In a production system, you'd track last usage time per limiter
	oldCount := len(rl.limiters)

	// Keep limiters with available tokens (recently used)
	newLimiters := make(map[string]*rate.Limiter)
	for key, limiter := range rl.limiters {
		// If limiter has less than max tokens, it was used recently
		if limiter.Tokens() < float64(rl.config.BurstSize) {
			newLimiters[key] = limiter
		}
	}

	rl.limiters = newLimiters
	rl.lastClean = now

	// Log cleanup results if needed
	_ = oldCount - len(newLimiters) // removed count (currently unused)
}

// UpdateConfig updates the rate limiter configuration
func (rl *RateLimiter) UpdateConfig(config RateLimitConfig) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.config = config

	// Clear existing limiters to use new config
	rl.limiters = make(map[string]*rate.Limiter)
}

// GetConfig returns the current configuration
func (rl *RateLimiter) GetConfig() RateLimitConfig {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.config
}
