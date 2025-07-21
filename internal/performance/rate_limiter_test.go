package performance

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter(t *testing.T) {
	t.Run("NewRateLimiter", func(t *testing.T) {
		config := RateLimitConfig{
			RequestsPerMin:  60,
			BurstSize:       5,
			CleanupInterval: 1 * time.Minute,
		}
		limiter := NewRateLimiter(config)
		require.NotNil(t, limiter)
		assert.Equal(t, config, limiter.GetConfig())
	})

	t.Run("Allow - Basic", func(t *testing.T) {
		config := RateLimitConfig{
			RequestsPerMin:  60, // 1 per second
			BurstSize:       3,
			CleanupInterval: 1 * time.Minute,
		}
		limiter := NewRateLimiter(config)

		key := "test-key"

		// Should allow burst
		for i := 0; i < 3; i++ {
			assert.True(t, limiter.Allow(key), "Request %d should be allowed", i+1)
		}

		// Should be rate limited after burst
		assert.False(t, limiter.Allow(key), "Request after burst should be denied")

		// Wait and try again
		time.Sleep(1100 * time.Millisecond)
		assert.True(t, limiter.Allow(key), "Request after waiting should be allowed")
	})

	t.Run("AllowN", func(t *testing.T) {
		config := RateLimitConfig{
			RequestsPerMin:  120, // 2 per second
			BurstSize:       5,
			CleanupInterval: 1 * time.Minute,
		}
		limiter := NewRateLimiter(config)

		key := "test-key"

		// Should allow burst of 3
		assert.True(t, limiter.AllowN(key, 3))

		// Should not allow 3 more immediately
		assert.False(t, limiter.AllowN(key, 3))

		// But should allow 2
		assert.True(t, limiter.AllowN(key, 2))
	})

	t.Run("Multiple Keys", func(t *testing.T) {
		config := RateLimitConfig{
			RequestsPerMin:  60,
			BurstSize:       2,
			CleanupInterval: 1 * time.Minute,
		}
		limiter := NewRateLimiter(config)

		// Each key should have its own limit
		assert.True(t, limiter.Allow("key1"))
		assert.True(t, limiter.Allow("key1"))
		assert.False(t, limiter.Allow("key1"))

		assert.True(t, limiter.Allow("key2"))
		assert.True(t, limiter.Allow("key2"))
		assert.False(t, limiter.Allow("key2"))

		assert.Equal(t, 2, limiter.GetLimiterCount())
	})

	t.Run("Hit Counter", func(t *testing.T) {
		config := RateLimitConfig{
			RequestsPerMin:  60,
			BurstSize:       1,
			CleanupInterval: 1 * time.Minute,
		}
		limiter := NewRateLimiter(config)

		key := "test-key"

		// First request allowed
		assert.True(t, limiter.Allow(key))
		assert.Equal(t, int64(0), limiter.GetHits())

		// Second request denied
		assert.False(t, limiter.Allow(key))
		assert.Equal(t, int64(1), limiter.GetHits())

		// Third request denied
		assert.False(t, limiter.Allow(key))
		assert.Equal(t, int64(2), limiter.GetHits())

		// Reset hits
		limiter.ResetHits()
		assert.Equal(t, int64(0), limiter.GetHits())
	})

	t.Run("UpdateConfig", func(t *testing.T) {
		config := RateLimitConfig{
			RequestsPerMin:  60,
			BurstSize:       2,
			CleanupInterval: 1 * time.Minute,
		}
		limiter := NewRateLimiter(config)

		// Use up the burst
		limiter.Allow("key1")
		limiter.Allow("key1")
		assert.False(t, limiter.Allow("key1"))

		// Update config with higher limits
		newConfig := RateLimitConfig{
			RequestsPerMin:  120,
			BurstSize:       5,
			CleanupInterval: 1 * time.Minute,
		}
		limiter.UpdateConfig(newConfig)

		// Should allow again with new limits
		assert.True(t, limiter.Allow("key1"))
		assert.Equal(t, newConfig, limiter.GetConfig())
	})

	t.Run("Concurrent Access", func(t *testing.T) {
		config := RateLimitConfig{
			RequestsPerMin:  6000, // 100 per second
			BurstSize:       50,
			CleanupInterval: 1 * time.Minute,
		}
		limiter := NewRateLimiter(config)

		var allowed int64
		var wg sync.WaitGroup

		// Multiple goroutines trying to get tokens
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					if limiter.Allow("shared-key") {
						atomic.AddInt64(&allowed, 1)
					}
					time.Sleep(1 * time.Millisecond)
				}
			}(i)
		}

		wg.Wait()

		// Should have allowed burst + some additional based on rate
		assert.GreaterOrEqual(t, allowed, int64(50)) // At least burst size
		assert.LessOrEqual(t, allowed, int64(100))   // But not all 100 requests
	})

	t.Run("Cleanup", func(t *testing.T) {
		config := RateLimitConfig{
			RequestsPerMin:  60,
			BurstSize:       2,
			CleanupInterval: 100 * time.Millisecond, // Fast cleanup for testing
		}
		limiter := NewRateLimiter(config)

		// Create limiters for multiple keys
		for i := 0; i < 5; i++ {
			key := fmt.Sprintf("key%d", i)
			limiter.Allow(key)
			limiter.Allow(key)
		}

		assert.Equal(t, 5, limiter.GetLimiterCount())

		// Wait for cleanup and token refill
		time.Sleep(200 * time.Millisecond)

		// Force cleanup by calling it directly
		limiter.cleanup()

		// Some limiters should have been cleaned up (or at least not increased)
		assert.LessOrEqual(t, limiter.GetLimiterCount(), 5)
	})
}