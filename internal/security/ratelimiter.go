package security

import (
	"sync"
	"time"
)

// IPRateLimiter implements a simple IP-based rate limiter
type IPRateLimiter struct {
	requests map[string]*rateLimitBucket
	limit    int
	window   time.Duration
	mu       sync.RWMutex
	cleanup  *time.Ticker
	done     chan struct{}
	wg       sync.WaitGroup
	stopOnce sync.Once
}

// rateLimitBucket tracks requests for a single IP
type rateLimitBucket struct {
	count     int
	resetTime time.Time
}

// NewIPRateLimiter creates a new IP rate limiter
func NewIPRateLimiter(limit int, window time.Duration) *IPRateLimiter {
	rl := &IPRateLimiter{
		requests: make(map[string]*rateLimitBucket),
		limit:    limit,
		window:   window,
		cleanup:  time.NewTicker(window),
		done:     make(chan struct{}),
	}

	// Start cleanup goroutine
	rl.wg.Add(1)
	go rl.cleanupExpired()

	return rl
}

// Allow checks if a request from the given IP is allowed
func (rl *IPRateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, exists := rl.requests[ip]

	if !exists || now.After(bucket.resetTime) {
		// Create new bucket or reset expired one
		rl.requests[ip] = &rateLimitBucket{
			count:     1,
			resetTime: now.Add(rl.window),
		}
		return true
	}

	// Check if limit exceeded
	if bucket.count >= rl.limit {
		return false
	}

	// Increment counter
	bucket.count++
	return true
}

// GetLimit returns the current limit for an IP
func (rl *IPRateLimiter) GetLimit(ip string) RateLimitInfo {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	info := RateLimitInfo{
		Key:    ip,
		Limit:  rl.limit,
		Window: rl.window,
	}

	bucket, exists := rl.requests[ip]
	if exists && time.Now().Before(bucket.resetTime) {
		info.Used = bucket.count
		info.Reset = bucket.resetTime
		info.Remaining = rl.limit - bucket.count
		if info.Remaining < 0 {
			info.Remaining = 0
		}
	} else {
		info.Used = 0
		info.Reset = time.Now().Add(rl.window)
		info.Remaining = rl.limit
	}

	return info
}

// Reset resets the rate limit for an IP
func (rl *IPRateLimiter) Reset(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.requests, ip)
}

// cleanupExpired removes expired buckets
func (rl *IPRateLimiter) cleanupExpired() {
	defer rl.wg.Done()
	for {
		select {
		case <-rl.cleanup.C:
			rl.mu.Lock()
			now := time.Now()
			for ip, bucket := range rl.requests {
				if now.After(bucket.resetTime) {
					delete(rl.requests, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.done:
			return
		}
	}
}

// Stop stops the rate limiter
func (rl *IPRateLimiter) Stop() {
	rl.stopOnce.Do(func() {
		rl.cleanup.Stop()
		close(rl.done)
		rl.wg.Wait()
	})
}

// TokenBucketRateLimiter implements a token bucket rate limiter
type TokenBucketRateLimiter struct {
	buckets    map[string]*tokenBucket
	capacity   int
	refillRate int
	mu         sync.RWMutex
	cleanup    *time.Ticker
	done       chan struct{}
	wg         sync.WaitGroup
	stopOnce   sync.Once
}

// tokenBucket represents a token bucket for rate limiting
type tokenBucket struct {
	tokens     int
	lastRefill time.Time
}

// NewTokenBucketRateLimiter creates a new token bucket rate limiter
func NewTokenBucketRateLimiter(capacity, refillRate int) *TokenBucketRateLimiter {
	rl := &TokenBucketRateLimiter{
		buckets:    make(map[string]*tokenBucket),
		capacity:   capacity,
		refillRate: refillRate,
		cleanup:    time.NewTicker(5 * time.Minute),
		done:       make(chan struct{}),
	}

	rl.wg.Add(1)
	go rl.cleanupStale()
	return rl
}

// Allow checks if a request is allowed and consumes a token
func (rl *TokenBucketRateLimiter) Allow(key string, tokens int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[key]
	if !exists {
		bucket = &tokenBucket{
			tokens:     rl.capacity,
			lastRefill: time.Now(),
		}
		rl.buckets[key] = bucket
	}

	// Refill tokens based on time elapsed
	rl.refillTokens(bucket)

	// Check if enough tokens available
	if bucket.tokens >= tokens {
		bucket.tokens -= tokens
		return true
	}

	return false
}

// refillTokens refills tokens based on elapsed time
func (rl *TokenBucketRateLimiter) refillTokens(bucket *tokenBucket) {
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill)

	// Calculate tokens to add
	tokensToAdd := int(elapsed.Seconds()) * rl.refillRate
	if tokensToAdd > 0 {
		bucket.tokens += tokensToAdd
		if bucket.tokens > rl.capacity {
			bucket.tokens = rl.capacity
		}
		bucket.lastRefill = now
	}
}

// GetTokens returns the current token count for a key
func (rl *TokenBucketRateLimiter) GetTokens(key string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	bucket, exists := rl.buckets[key]
	if !exists {
		return rl.capacity
	}

	// Calculate current tokens without modifying
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill)
	tokensToAdd := int(elapsed.Seconds()) * rl.refillRate

	tokens := bucket.tokens + tokensToAdd
	if tokens > rl.capacity {
		tokens = rl.capacity
	}

	return tokens
}

// cleanupStale removes buckets that haven't been used recently
func (rl *TokenBucketRateLimiter) cleanupStale() {
	defer rl.wg.Done()
	for {
		select {
		case <-rl.cleanup.C:
			rl.mu.Lock()
			now := time.Now()
			for key, bucket := range rl.buckets {
				if now.Sub(bucket.lastRefill) > 10*time.Minute {
					delete(rl.buckets, key)
				}
			}
			rl.mu.Unlock()
		case <-rl.done:
			return
		}
	}
}

// Stop stops the rate limiter
func (rl *TokenBucketRateLimiter) Stop() {
	rl.stopOnce.Do(func() {
		rl.cleanup.Stop()
		close(rl.done)
		rl.wg.Wait()
	})
}
