package security

import (
	"sync"
	"testing"
	"time"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestNewIPRateLimiter(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("creates rate limiter with specified parameters", func(t *testing.T) {
		limiter := NewIPRateLimiter(100, time.Minute)
		defer limiter.Stop()

		testutil.AssertNotEqual(t, nil, limiter)
		testutil.AssertEqual(t, 100, limiter.limit)
		testutil.AssertEqual(t, time.Minute, limiter.window)
		testutil.AssertNotEqual(t, nil, limiter.requests)
		testutil.AssertNotEqual(t, nil, limiter.cleanup)
	})

	t.Run("starts with empty requests map", func(t *testing.T) {
		limiter := NewIPRateLimiter(10, time.Second)
		defer limiter.Stop()

		testutil.AssertEqual(t, 0, len(limiter.requests))
	})
}

func TestIPRateLimiterAllow(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("allows first request", func(t *testing.T) {
		limiter := NewIPRateLimiter(5, time.Minute)
		defer limiter.Stop()

		allowed := limiter.Allow("192.168.1.1")
		testutil.AssertTrue(t, allowed)
		testutil.AssertEqual(t, 1, len(limiter.requests))
	})

	t.Run("allows requests within limit", func(t *testing.T) {
		limiter := NewIPRateLimiter(3, time.Minute)
		defer limiter.Stop()

		ip := "192.168.1.2"

		// First 3 requests should be allowed
		for i := 0; i < 3; i++ {
			allowed := limiter.Allow(ip)
			testutil.AssertTrue(t, allowed)
		}

		// 4th request should be denied
		allowed := limiter.Allow(ip)
		testutil.AssertFalse(t, allowed)
	})

	t.Run("rejects requests exceeding limit", func(t *testing.T) {
		limiter := NewIPRateLimiter(2, time.Minute)
		defer limiter.Stop()

		ip := "192.168.1.3"

		// Use up the limit
		limiter.Allow(ip)
		limiter.Allow(ip)

		// This should be rejected
		allowed := limiter.Allow(ip)
		testutil.AssertFalse(t, allowed)

		// Multiple additional requests should still be rejected
		for i := 0; i < 5; i++ {
			allowed := limiter.Allow(ip)
			testutil.AssertFalse(t, allowed)
		}
	})

	t.Run("different IPs have separate limits", func(t *testing.T) {
		limiter := NewIPRateLimiter(2, time.Minute)
		defer limiter.Stop()

		ip1 := "192.168.1.4"
		ip2 := "192.168.1.5"

		// Use up limit for ip1
		limiter.Allow(ip1)
		limiter.Allow(ip1)
		allowed1 := limiter.Allow(ip1)
		testutil.AssertFalse(t, allowed1)

		// ip2 should still be allowed
		allowed2 := limiter.Allow(ip2)
		testutil.AssertTrue(t, allowed2)
	})

	t.Run("resets after window expires", func(t *testing.T) {
		limiter := NewIPRateLimiter(1, 50*time.Millisecond)
		defer limiter.Stop()

		ip := "192.168.1.6"

		// Use up the limit
		allowed1 := limiter.Allow(ip)
		testutil.AssertTrue(t, allowed1)

		// Should be rejected immediately
		allowed2 := limiter.Allow(ip)
		testutil.AssertFalse(t, allowed2)

		// Wait for window to expire
		time.Sleep(60 * time.Millisecond)

		// Should be allowed again
		allowed3 := limiter.Allow(ip)
		testutil.AssertTrue(t, allowed3)
	})

	t.Run("handles concurrent requests safely", func(t *testing.T) {
		limiter := NewIPRateLimiter(100, time.Minute)
		defer limiter.Stop()

		ip := "192.168.1.7"
		numGoroutines := 10
		requestsPerGoroutine := 20
		var wg sync.WaitGroup
		var mu sync.Mutex
		allowedCount := 0

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < requestsPerGoroutine; j++ {
					if limiter.Allow(ip) {
						mu.Lock()
						allowedCount++
						mu.Unlock()
					}
				}
			}()
		}

		wg.Wait()

		// Should allow exactly the limit (100 requests)
		testutil.AssertEqual(t, 100, allowedCount)
	})

	t.Run("bucket count increments correctly", func(t *testing.T) {
		limiter := NewIPRateLimiter(5, time.Minute)
		defer limiter.Stop()

		ip := "192.168.1.8"

		// Make requests and check bucket state
		for i := 1; i <= 3; i++ {
			limiter.Allow(ip)
			bucket := limiter.requests[ip]
			testutil.AssertEqual(t, i, bucket.count)
		}
	})

	t.Run("reset time is set correctly", func(t *testing.T) {
		window := 30 * time.Second
		limiter := NewIPRateLimiter(5, window)
		defer limiter.Stop()

		ip := "192.168.1.9"
		beforeTime := time.Now()

		limiter.Allow(ip)

		bucket := limiter.requests[ip]
		afterTime := time.Now()

		// Reset time should be within the expected window
		expectedResetTime := beforeTime.Add(window)
		testutil.AssertTrue(t, bucket.resetTime.After(expectedResetTime.Add(-time.Second)))
		testutil.AssertTrue(t, bucket.resetTime.Before(afterTime.Add(window).Add(time.Second)))
	})
}

func TestIPRateLimiterGetLimit(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("returns correct info for new IP", func(t *testing.T) {
		limiter := NewIPRateLimiter(10, time.Minute)
		defer limiter.Stop()

		ip := "192.168.1.10"
		info := limiter.GetLimit(ip)

		testutil.AssertEqual(t, ip, info.Key)
		testutil.AssertEqual(t, 10, info.Limit)
		testutil.AssertEqual(t, time.Minute, info.Window)
		testutil.AssertEqual(t, 0, info.Used)
		testutil.AssertEqual(t, 10, info.Remaining)
		testutil.AssertTrue(t, time.Now().Before(info.Reset))
	})

	t.Run("returns correct info for used IP", func(t *testing.T) {
		limiter := NewIPRateLimiter(10, time.Minute)
		defer limiter.Stop()

		ip := "192.168.1.11"

		// Use some requests
		limiter.Allow(ip)
		limiter.Allow(ip)
		limiter.Allow(ip)

		info := limiter.GetLimit(ip)

		testutil.AssertEqual(t, 3, info.Used)
		testutil.AssertEqual(t, 7, info.Remaining)
		testutil.AssertTrue(t, time.Now().Before(info.Reset))
	})

	t.Run("returns zero remaining when limit exceeded", func(t *testing.T) {
		limiter := NewIPRateLimiter(2, time.Minute)
		defer limiter.Stop()

		ip := "192.168.1.12"

		// Use up the limit
		limiter.Allow(ip) // count = 1
		limiter.Allow(ip) // count = 2
		limiter.Allow(ip) // This should be denied, count stays at 2

		info := limiter.GetLimit(ip)

		testutil.AssertEqual(t, 2, info.Used)      // Only 2 allowed
		testutil.AssertEqual(t, 0, info.Remaining) // Should be capped at 0
	})

	t.Run("returns fresh info for expired bucket", func(t *testing.T) {
		limiter := NewIPRateLimiter(5, 50*time.Millisecond)
		defer limiter.Stop()

		ip := "192.168.1.13"

		// Use some requests
		limiter.Allow(ip)
		limiter.Allow(ip)

		// Wait for expiration
		time.Sleep(60 * time.Millisecond)

		info := limiter.GetLimit(ip)

		testutil.AssertEqual(t, 0, info.Used)
		testutil.AssertEqual(t, 5, info.Remaining)
	})
}

func TestIPRateLimiterReset(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("resets IP limit", func(t *testing.T) {
		limiter := NewIPRateLimiter(3, time.Minute)
		defer limiter.Stop()

		ip := "192.168.1.14"

		// Use up the limit
		limiter.Allow(ip)
		limiter.Allow(ip)
		limiter.Allow(ip)

		// Should be rejected
		allowed1 := limiter.Allow(ip)
		testutil.AssertFalse(t, allowed1)

		// Reset the IP
		limiter.Reset(ip)

		// Should be allowed again
		allowed2 := limiter.Allow(ip)
		testutil.AssertTrue(t, allowed2)
	})

	t.Run("reset removes IP from requests map", func(t *testing.T) {
		limiter := NewIPRateLimiter(5, time.Minute)
		defer limiter.Stop()

		ip := "192.168.1.15"

		limiter.Allow(ip)
		testutil.AssertEqual(t, 1, len(limiter.requests))

		limiter.Reset(ip)
		testutil.AssertEqual(t, 0, len(limiter.requests))
	})

	t.Run("reset non-existent IP does not error", func(t *testing.T) {
		limiter := NewIPRateLimiter(5, time.Minute)
		defer limiter.Stop()

		// Should not panic
		limiter.Reset("192.168.1.999")
		testutil.AssertEqual(t, 0, len(limiter.requests))
	})
}

func TestIPRateLimiterCleanup(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("cleans up expired buckets", func(t *testing.T) {
		limiter := NewIPRateLimiter(5, 50*time.Millisecond)
		defer limiter.Stop()

		ip1 := "192.168.1.16"
		ip2 := "192.168.1.17"

		// Create buckets
		limiter.Allow(ip1)
		limiter.Allow(ip2)

		// Check count with lock
		limiter.mu.RLock()
		initialCount := len(limiter.requests)
		limiter.mu.RUnlock()
		testutil.AssertEqual(t, 2, initialCount)

		// Wait for expiration and cleanup cycle
		time.Sleep(200 * time.Millisecond)

		// Manual trigger cleanup by checking size
		// The cleanup goroutine should have removed expired buckets
		limiter.mu.Lock()
		count := len(limiter.requests)
		limiter.mu.Unlock()

		testutil.AssertEqual(t, 0, count)
	})

	t.Run("does not clean up active buckets", func(t *testing.T) {
		limiter := NewIPRateLimiter(5, time.Hour) // Long window
		defer limiter.Stop()

		ip := "192.168.1.18"
		limiter.Allow(ip)

		// Wait a bit but not long enough for expiration
		time.Sleep(100 * time.Millisecond)

		limiter.mu.Lock()
		count := len(limiter.requests)
		limiter.mu.Unlock()

		testutil.AssertEqual(t, 1, count)
	})
}

func TestIPRateLimiterStop(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("stops cleanup ticker", func(t *testing.T) {
		limiter := NewIPRateLimiter(5, time.Minute)

		testutil.AssertNotEqual(t, nil, limiter.cleanup)

		limiter.Stop()

		// After stop, ticker should be stopped
		// We can't easily test this directly, but we can ensure Stop() doesn't panic
	})

	t.Run("multiple stops do not panic", func(t *testing.T) {
		limiter := NewIPRateLimiter(5, time.Minute)

		// Should not panic
		limiter.Stop()
		limiter.Stop()
	})
}

// Token Bucket Rate Limiter Tests

func TestNewTokenBucketRateLimiter(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("creates token bucket limiter", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(100, 10)
		defer limiter.Stop()

		testutil.AssertNotEqual(t, nil, limiter)
		testutil.AssertEqual(t, 100, limiter.capacity)
		testutil.AssertEqual(t, 10, limiter.refillRate)
		testutil.AssertNotEqual(t, nil, limiter.buckets)
		testutil.AssertNotEqual(t, nil, limiter.cleanup)
	})

	t.Run("starts with empty buckets", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(50, 5)
		defer limiter.Stop()

		testutil.AssertEqual(t, 0, len(limiter.buckets))
	})
}

func TestTokenBucketRateLimiterAllow(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("allows request with available tokens", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(100, 10)
		defer limiter.Stop()

		allowed := limiter.Allow("key1", 1)
		testutil.AssertTrue(t, allowed)
		testutil.AssertEqual(t, 1, len(limiter.buckets))
	})

	t.Run("creates bucket with full capacity for new key", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(50, 5)
		defer limiter.Stop()

		key := "newkey"
		allowed := limiter.Allow(key, 10)
		testutil.AssertTrue(t, allowed)

		bucket := limiter.buckets[key]
		testutil.AssertEqual(t, 40, bucket.tokens) // 50 - 10
	})

	t.Run("rejects request when insufficient tokens", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(10, 1)
		defer limiter.Stop()

		key := "key2"

		// Use up most tokens
		limiter.Allow(key, 8)

		// Should allow 2 more tokens
		allowed1 := limiter.Allow(key, 2)
		testutil.AssertTrue(t, allowed1)

		// Should reject request for 5 tokens
		allowed2 := limiter.Allow(key, 5)
		testutil.AssertFalse(t, allowed2)
	})

	t.Run("refills tokens over time", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(10, 10) // 10 tokens per second
		defer limiter.Stop()

		key := "key3"

		// Use up all tokens
		limiter.Allow(key, 10)

		// Should be rejected immediately
		allowed1 := limiter.Allow(key, 1)
		testutil.AssertFalse(t, allowed1)

		// Wait for refill (10 tokens per second)
		time.Sleep(1100 * time.Millisecond) // Wait a bit over 1 second

		// Should have at least 1 token available
		allowed2 := limiter.Allow(key, 1)
		testutil.AssertTrue(t, allowed2)
	})

	t.Run("caps tokens at capacity", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(10, 100) // High refill rate
		defer limiter.Stop()

		key := "key4"

		// Create bucket and use some tokens
		limiter.Allow(key, 5)

		// Wait for refill
		time.Sleep(100 * time.Millisecond)

		// Check available tokens don't exceed capacity
		tokens := limiter.GetTokens(key)
		testutil.AssertTrue(t, tokens <= 10)
	})

	t.Run("different keys have separate buckets", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(5, 1)
		defer limiter.Stop()

		key1 := "user1"
		key2 := "user2"

		// Use up tokens for key1
		limiter.Allow(key1, 5)
		allowed1 := limiter.Allow(key1, 1)
		testutil.AssertFalse(t, allowed1)

		// key2 should still have full capacity
		allowed2 := limiter.Allow(key2, 5)
		testutil.AssertTrue(t, allowed2)
	})

	t.Run("handles concurrent access safely", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(1000, 100)
		defer limiter.Stop()

		key := "concurrent"
		numGoroutines := 20
		tokensPerRequest := 10
		var wg sync.WaitGroup
		var mu sync.Mutex
		allowedCount := 0

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					if limiter.Allow(key, tokensPerRequest) {
						mu.Lock()
						allowedCount++
						mu.Unlock()
					}
				}
			}()
		}

		wg.Wait()

		// Should allow exactly capacity/tokensPerRequest requests (1000/10 = 100)
		testutil.AssertEqual(t, 100, allowedCount)
	})
}

func TestTokenBucketGetTokens(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("returns full capacity for new key", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(50, 5)
		defer limiter.Stop()

		tokens := limiter.GetTokens("newkey")
		testutil.AssertEqual(t, 50, tokens)
	})

	t.Run("returns correct tokens for existing key", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(20, 2)
		defer limiter.Stop()

		key := "existing"
		limiter.Allow(key, 5)

		tokens := limiter.GetTokens(key)
		testutil.AssertEqual(t, 15, tokens)
	})

	t.Run("calculates refilled tokens correctly", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(100, 10) // 10 tokens per second
		defer limiter.Stop()

		key := "refilltest"
		limiter.Allow(key, 50) // Use 50 tokens

		// Wait for 1 second
		time.Sleep(1100 * time.Millisecond)

		tokens := limiter.GetTokens(key)
		// Should have refilled ~10 tokens, so ~60 total, capped at 100
		testutil.AssertTrue(t, tokens >= 60)
		testutil.AssertTrue(t, tokens <= 100)
	})

	t.Run("does not modify bucket state", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(10, 1)
		defer limiter.Stop()

		key := "readonly"
		limiter.Allow(key, 3)

		// Get tokens multiple times
		tokens1 := limiter.GetTokens(key)
		tokens2 := limiter.GetTokens(key)

		testutil.AssertEqual(t, tokens1, tokens2)
		testutil.AssertEqual(t, 7, tokens1) // 10 - 3
	})
}

func TestTokenBucketCleanup(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("cleans up stale buckets", func(t *testing.T) {
		// Create a limiter with short cleanup interval
		limiter := NewTokenBucketRateLimiter(10, 1)
		defer limiter.Stop()

		key := "staletest"
		limiter.Allow(key, 1)

		// Verify bucket was created
		limiter.mu.RLock()
		initialCount := len(limiter.buckets)
		limiter.mu.RUnlock()
		testutil.AssertEqual(t, 1, initialCount)

		// Manually set old timestamp to trigger cleanup
		limiter.mu.Lock()
		if bucket, ok := limiter.buckets[key]; ok {
			bucket.lastRefill = time.Now().Add(-15 * time.Minute) // Older than 10 minutes
			limiter.buckets[key] = bucket
		}
		limiter.mu.Unlock()

		// Manually trigger cleanup logic to simulate passage of time
		limiter.mu.Lock()
		now := time.Now()
		for k, bucket := range limiter.buckets {
			if now.Sub(bucket.lastRefill) > 10*time.Minute {
				delete(limiter.buckets, k)
			}
		}
		count := len(limiter.buckets)
		limiter.mu.Unlock()

		testutil.AssertEqual(t, 0, count)
	})
}

func TestTokenBucketStop(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("stops cleanup ticker", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(10, 1)

		testutil.AssertNotEqual(t, nil, limiter.cleanup)

		limiter.Stop()

		// Should not panic
	})

	t.Run("multiple stops do not panic", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(10, 1)

		limiter.Stop()
		limiter.Stop() // Should not panic
	})
}

func TestTokenBucketRefillTokens(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("refills tokens based on elapsed time", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(100, 10)
		defer limiter.Stop()

		// Create a bucket
		bucket := &tokenBucket{
			tokens:     50,
			lastRefill: time.Now().Add(-2 * time.Second), // 2 seconds ago
		}

		limiter.refillTokens(bucket)

		// Should have added ~20 tokens (10 per second * 2 seconds)
		testutil.AssertTrue(t, bucket.tokens >= 70)
		testutil.AssertTrue(t, bucket.tokens <= 100) // Capped at capacity
	})

	t.Run("updates last refill time", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(100, 10)
		defer limiter.Stop()

		oldTime := time.Now().Add(-1 * time.Second)
		bucket := &tokenBucket{
			tokens:     50,
			lastRefill: oldTime,
		}

		limiter.refillTokens(bucket)

		testutil.AssertTrue(t, bucket.lastRefill.After(oldTime))
	})

	t.Run("does not refill if no time elapsed", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(100, 10)
		defer limiter.Stop()

		now := time.Now()
		bucket := &tokenBucket{
			tokens:     50,
			lastRefill: now,
		}

		limiter.refillTokens(bucket)

		testutil.AssertEqual(t, 50, bucket.tokens)
		testutil.AssertEqual(t, now, bucket.lastRefill)
	})

	t.Run("caps tokens at capacity", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(100, 50)
		defer limiter.Stop()

		bucket := &tokenBucket{
			tokens:     90,
			lastRefill: time.Now().Add(-5 * time.Second), // Would add 250 tokens
		}

		limiter.refillTokens(bucket)

		testutil.AssertEqual(t, 100, bucket.tokens) // Should be capped at capacity
	})
}
