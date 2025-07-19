package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIPRateLimiter(t *testing.T) {
	t.Run("basic rate limiting", func(t *testing.T) {
		rl := NewIPRateLimiter(5, 100*time.Millisecond)
		defer rl.Stop()

		ip := "192.168.1.1"

		// First 5 requests should pass
		for i := 0; i < 5; i++ {
			assert.True(t, rl.Allow(ip), "Request %d should be allowed", i+1)
		}

		// 6th request should fail
		assert.False(t, rl.Allow(ip), "6th request should be denied")

		// Wait for window to reset
		time.Sleep(110 * time.Millisecond)

		// Should be allowed again
		assert.True(t, rl.Allow(ip), "Request should be allowed after window reset")
	})

	t.Run("different IPs", func(t *testing.T) {
		rl := NewIPRateLimiter(2, 100*time.Millisecond)
		defer rl.Stop()

		ip1 := "192.168.1.1"
		ip2 := "192.168.1.2"

		// Both IPs should have their own limits
		assert.True(t, rl.Allow(ip1))
		assert.True(t, rl.Allow(ip1))
		assert.False(t, rl.Allow(ip1))

		assert.True(t, rl.Allow(ip2))
		assert.True(t, rl.Allow(ip2))
		assert.False(t, rl.Allow(ip2))
	})

	t.Run("get limit info", func(t *testing.T) {
		rl := NewIPRateLimiter(10, 1*time.Second)
		defer rl.Stop()

		ip := "192.168.1.1"

		// Initial state
		info := rl.GetLimit(ip)
		assert.Equal(t, 10, info.Limit)
		assert.Equal(t, 0, info.Used)
		assert.Equal(t, 10, info.Remaining)

		// After some requests
		rl.Allow(ip)
		rl.Allow(ip)
		rl.Allow(ip)

		info = rl.GetLimit(ip)
		assert.Equal(t, 10, info.Limit)
		assert.Equal(t, 3, info.Used)
		assert.Equal(t, 7, info.Remaining)
	})

	t.Run("reset", func(t *testing.T) {
		rl := NewIPRateLimiter(1, 1*time.Second)
		defer rl.Stop()

		ip := "192.168.1.1"

		// Use up the limit
		assert.True(t, rl.Allow(ip))
		assert.False(t, rl.Allow(ip))

		// Reset
		rl.Reset(ip)

		// Should be allowed again
		assert.True(t, rl.Allow(ip))
	})

	t.Run("cleanup expired", func(t *testing.T) {
		rl := NewIPRateLimiter(1, 50*time.Millisecond)
		defer rl.Stop()

		ip := "192.168.1.1"

		// Create entry
		rl.Allow(ip)
		assert.Len(t, rl.requests, 1)

		// Wait for cleanup
		time.Sleep(100 * time.Millisecond)

		// Entry should be cleaned up
		assert.Empty(t, rl.requests)
	})
}

func TestTokenBucketRateLimiter(t *testing.T) {
	t.Run("basic token consumption", func(t *testing.T) {
		rl := NewTokenBucketRateLimiter(10, 5) // 10 capacity, 5 tokens/sec
		defer rl.Stop()

		key := "user1"

		// Should have full capacity initially
		assert.True(t, rl.Allow(key, 5))
		assert.True(t, rl.Allow(key, 5))
		assert.False(t, rl.Allow(key, 5)) // Exceeds capacity

		// Check remaining tokens
		tokens := rl.GetTokens(key)
		assert.Equal(t, 0, tokens)
	})

	t.Run("token refill", func(t *testing.T) {
		rl := NewTokenBucketRateLimiter(10, 10) // 10 capacity, 10 tokens/sec
		defer rl.Stop()

		key := "user1"

		// Use all tokens
		assert.True(t, rl.Allow(key, 10))
		assert.Equal(t, 0, rl.GetTokens(key))

		// Wait for refill
		time.Sleep(1 * time.Second)

		// Should have refilled
		tokens := rl.GetTokens(key)
		assert.True(t, tokens >= 9) // Allow for slight timing variations
	})

	t.Run("multiple keys", func(t *testing.T) {
		rl := NewTokenBucketRateLimiter(5, 1)
		defer rl.Stop()

		key1 := "user1"
		key2 := "user2"

		// Each key has its own bucket
		assert.True(t, rl.Allow(key1, 3))
		assert.True(t, rl.Allow(key2, 3))

		assert.Equal(t, 2, rl.GetTokens(key1))
		assert.Equal(t, 2, rl.GetTokens(key2))
	})

	t.Run("capacity limit", func(t *testing.T) {
		rl := NewTokenBucketRateLimiter(10, 10) // 10 capacity, 10 tokens/sec
		defer rl.Stop()

		key := "user1"

		// Use all tokens
		assert.True(t, rl.Allow(key, 10))
		assert.Equal(t, 0, rl.GetTokens(key))

		// Wait for refill (need at least 1 second for refill to work)
		time.Sleep(1100 * time.Millisecond)

		// Should be at capacity, not over
		tokens := rl.GetTokens(key)
		assert.Equal(t, 10, tokens)
	})

	t.Run("stale bucket cleanup", func(t *testing.T) {
		rl := NewTokenBucketRateLimiter(10, 1)
		defer rl.Stop()

		// Create entries
		rl.Allow("user1", 1)
		rl.Allow("user2", 1)
		assert.Len(t, rl.buckets, 2)

		// Wait for cleanup (simulated by updating lastRefill time)
		// This is a simplified test - in reality cleanup runs periodically
		rl.mu.Lock()
		for _, bucket := range rl.buckets {
			bucket.lastRefill = time.Now().Add(-15 * time.Minute)
		}
		rl.mu.Unlock()

		// Trigger cleanup
		rl.cleanupStale()

		// Buckets should be cleaned
		assert.Empty(t, rl.buckets)
	})
}

func BenchmarkIPRateLimiter(b *testing.B) {
	rl := NewIPRateLimiter(1000, 1*time.Second)
	defer rl.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		ip := "192.168.1.1"
		for pb.Next() {
			rl.Allow(ip)
		}
	})
}

func BenchmarkTokenBucketRateLimiter(b *testing.B) {
	rl := NewTokenBucketRateLimiter(1000, 100)
	defer rl.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		key := "user1"
		for pb.Next() {
			rl.Allow(key, 1)
		}
	})
}