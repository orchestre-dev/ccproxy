package performance

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreaker(t *testing.T) {
	t.Run("NewCircuitBreaker", func(t *testing.T) {
		config := CircuitBreakerConfig{
			ErrorThreshold:      0.5,
			ConsecutiveFailures: 3,
			OpenDuration:        30 * time.Second,
			HalfOpenMaxRequests: 2,
		}
		cb := NewCircuitBreaker("test", config)
		require.NotNil(t, cb)
		assert.Equal(t, "test", cb.name)
		assert.Equal(t, StateClosed, cb.GetState())
		assert.True(t, cb.IsClosed())
	})

	t.Run("Consecutive Failures Opens Circuit", func(t *testing.T) {
		config := CircuitBreakerConfig{
			ConsecutiveFailures: 3,
			OpenDuration:        100 * time.Millisecond,
		}
		cb := NewCircuitBreaker("test", config)

		// Should allow initially
		assert.True(t, cb.Allow())

		// Record consecutive failures
		cb.Record(false)
		assert.Equal(t, StateClosed, cb.GetState())

		cb.Record(false)
		assert.Equal(t, StateClosed, cb.GetState())

		cb.Record(false)
		assert.Equal(t, StateOpen, cb.GetState())
		assert.True(t, cb.IsOpen())

		// Should not allow when open
		assert.False(t, cb.Allow())
	})

	t.Run("Error Threshold Opens Circuit", func(t *testing.T) {
		config := CircuitBreakerConfig{
			ErrorThreshold:      0.5,
			ConsecutiveFailures: 100, // High so it doesn't trigger
			OpenDuration:        100 * time.Millisecond,
		}
		cb := NewCircuitBreaker("test", config)

		// Record enough requests to trigger error threshold
		for i := 0; i < 10; i++ {
			cb.Record(true) // Success
		}
		for i := 0; i < 11; i++ {
			cb.Record(false) // Failure
		}

		// Error rate is now > 50%
		assert.Equal(t, StateOpen, cb.GetState())
	})

	t.Run("Transition to Half-Open", func(t *testing.T) {
		config := CircuitBreakerConfig{
			ConsecutiveFailures: 2,
			OpenDuration:        50 * time.Millisecond,
			HalfOpenMaxRequests: 2,
		}
		cb := NewCircuitBreaker("test", config)

		// Open the circuit
		cb.Record(false)
		cb.Record(false)
		assert.Equal(t, StateOpen, cb.GetState())

		// Wait for open duration
		time.Sleep(60 * time.Millisecond)

		// Should transition to half-open on next check
		assert.True(t, cb.Allow())
		assert.Equal(t, StateHalfOpen, cb.GetState())
		assert.True(t, cb.IsHalfOpen())

		// Should allow limited requests in half-open
		assert.True(t, cb.Allow())
		assert.False(t, cb.Allow()) // Exceeds half-open max
	})

	t.Run("Half-Open to Closed", func(t *testing.T) {
		config := CircuitBreakerConfig{
			ConsecutiveFailures: 2,
			OpenDuration:        50 * time.Millisecond,
			HalfOpenMaxRequests: 2,
		}
		cb := NewCircuitBreaker("test", config)

		// Open the circuit
		cb.Record(false)
		cb.Record(false)

		// Wait and transition to half-open
		time.Sleep(60 * time.Millisecond)
		cb.Allow()
		assert.Equal(t, StateHalfOpen, cb.GetState())

		// Successful requests in half-open should close circuit
		cb.Record(true)
		cb.Allow()
		cb.Record(true)
		assert.Equal(t, StateClosed, cb.GetState())
	})

	t.Run("Half-Open to Open", func(t *testing.T) {
		config := CircuitBreakerConfig{
			ConsecutiveFailures: 2,
			OpenDuration:        50 * time.Millisecond,
			HalfOpenMaxRequests: 2,
		}
		cb := NewCircuitBreaker("test", config)

		// Open the circuit
		cb.Record(false)
		cb.Record(false)

		// Wait and transition to half-open
		time.Sleep(60 * time.Millisecond)
		cb.Allow()
		assert.Equal(t, StateHalfOpen, cb.GetState())

		// Failure in half-open should reopen circuit
		cb.Record(false)
		assert.Equal(t, StateOpen, cb.GetState())
	})

	t.Run("GetStats", func(t *testing.T) {
		config := CircuitBreakerConfig{
			ConsecutiveFailures: 3,
		}
		cb := NewCircuitBreaker("test", config)

		// Record some requests
		cb.Record(true)
		cb.Record(true)
		cb.Record(false)

		stats := cb.GetStats()
		assert.Equal(t, "test", stats["name"])
		assert.Equal(t, "closed", stats["state"])
		assert.Equal(t, int64(1), stats["failures"])
		assert.Equal(t, int64(2), stats["successes"])
		assert.Equal(t, int64(1), stats["consecutive_fails"])
		assert.InDelta(t, 0.333, stats["error_rate"], 0.01)
	})

	t.Run("Reset", func(t *testing.T) {
		config := CircuitBreakerConfig{
			ConsecutiveFailures: 2,
		}
		cb := NewCircuitBreaker("test", config)

		// Open the circuit
		cb.Record(false)
		cb.Record(false)
		assert.Equal(t, StateOpen, cb.GetState())

		// Reset
		cb.Reset()
		assert.Equal(t, StateClosed, cb.GetState())
		assert.True(t, cb.Allow())

		stats := cb.GetStats()
		assert.Equal(t, int64(0), stats["failures"])
		assert.Equal(t, int64(0), stats["successes"])
		assert.Equal(t, int64(0), stats["consecutive_fails"])
	})

	t.Run("Concurrent Access", func(t *testing.T) {
		config := CircuitBreakerConfig{
			ConsecutiveFailures: 50,
			ErrorThreshold:      0.5,
			OpenDuration:        1 * time.Second,
		}
		cb := NewCircuitBreaker("test", config)

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					if cb.Allow() {
						// Simulate some successes and failures
						success := (id+j)%3 != 0
						cb.Record(success)
					}
					time.Sleep(1 * time.Millisecond)
				}
			}(i)
		}

		wg.Wait()

		// Should have recorded requests
		stats := cb.GetStats()
		total := stats["failures"].(int64) + stats["successes"].(int64)
		assert.Greater(t, total, int64(0))
	})

	t.Run("State String", func(t *testing.T) {
		assert.Equal(t, "closed", stateToString(StateClosed))
		assert.Equal(t, "open", stateToString(StateOpen))
		assert.Equal(t, "half-open", stateToString(StateHalfOpen))
		assert.Equal(t, "unknown", stateToString(CircuitState(99)))
	})
}