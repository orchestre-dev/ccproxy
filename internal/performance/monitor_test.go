package performance

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonitor(t *testing.T) {
	t.Run("NewMonitor", func(t *testing.T) {
		config := DefaultPerformanceConfig()
		monitor := NewMonitor(config)
		require.NotNil(t, monitor)
		assert.NotNil(t, monitor.config)
		assert.NotNil(t, monitor.metrics)
		assert.NotNil(t, monitor.latencyTracker)
		assert.NotNil(t, monitor.resourceMonitor)
	})

	t.Run("Start and Stop", func(t *testing.T) {
		config := DefaultPerformanceConfig()
		config.MetricsEnabled = true
		monitor := NewMonitor(config)

		monitor.Start()
		time.Sleep(100 * time.Millisecond)
		monitor.Stop()
	})

	t.Run("RecordRequest", func(t *testing.T) {
		monitor := NewMonitor(nil)

		// Record successful request
		monitor.RecordRequest(RequestMetrics{
			Provider:     "test-provider",
			Model:        "test-model",
			StartTime:    time.Now(),
			EndTime:      time.Now().Add(100 * time.Millisecond),
			Latency:      100 * time.Millisecond,
			TokensIn:     10,
			TokensOut:    20,
			Success:      true,
			StatusCode:   200,
			RequestSize:  1024,
			ResponseSize: 2048,
		})

		metrics := monitor.GetMetrics()
		assert.Equal(t, int64(1), metrics.TotalRequests)
		assert.Equal(t, int64(1), metrics.SuccessfulRequests)
		assert.Equal(t, int64(0), metrics.FailedRequests)

		// Record failed request
		monitor.RecordRequest(RequestMetrics{
			Provider:   "test-provider",
			Model:      "test-model",
			StartTime:  time.Now(),
			EndTime:    time.Now().Add(50 * time.Millisecond),
			Latency:    50 * time.Millisecond,
			Success:    false,
			StatusCode: 500,
		})

		metrics = monitor.GetMetrics()
		assert.Equal(t, int64(2), metrics.TotalRequests)
		assert.Equal(t, int64(1), metrics.SuccessfulRequests)
		assert.Equal(t, int64(1), metrics.FailedRequests)
	})

	t.Run("Provider Metrics", func(t *testing.T) {
		monitor := NewMonitor(nil)

		// Record requests for different providers
		monitor.RecordRequest(RequestMetrics{
			Provider:  "provider1",
			Latency:   100 * time.Millisecond,
			TokensIn:  10,
			TokensOut: 20,
			Success:   true,
		})

		monitor.RecordRequest(RequestMetrics{
			Provider:  "provider1",
			Latency:   200 * time.Millisecond,
			TokensIn:  15,
			TokensOut: 25,
			Success:   false,
		})

		monitor.RecordRequest(RequestMetrics{
			Provider: "provider2",
			Latency:  50 * time.Millisecond,
			Success:  true,
		})

		metrics := monitor.GetMetrics()
		require.Contains(t, metrics.ProviderMetrics, "provider1")
		require.Contains(t, metrics.ProviderMetrics, "provider2")

		p1 := metrics.ProviderMetrics["provider1"]
		assert.Equal(t, int64(2), p1.TotalRequests)
		assert.Equal(t, int64(1), p1.SuccessfulRequests)
		assert.Equal(t, int64(1), p1.FailedRequests)
		assert.Equal(t, 150*time.Millisecond, p1.AverageLatency)
		assert.Equal(t, int64(70), p1.TokensProcessed) // 10+20+15+25
		assert.Equal(t, 0.5, p1.ErrorRate)

		p2 := metrics.ProviderMetrics["provider2"]
		assert.Equal(t, int64(1), p2.TotalRequests)
		assert.Equal(t, int64(1), p2.SuccessfulRequests)
		assert.Equal(t, int64(0), p2.FailedRequests)
		assert.Equal(t, 0.0, p2.ErrorRate)
	})

	t.Run("Rate Limiting", func(t *testing.T) {
		config := DefaultPerformanceConfig()
		config.RateLimit.Enabled = true
		config.RateLimit.RequestsPerMin = 60
		config.RateLimit.BurstSize = 5
		monitor := NewMonitor(config)

		// Should allow burst
		for i := 0; i < 5; i++ {
			assert.True(t, monitor.CheckRateLimit("test-key"))
		}

		// Should be rate limited after burst
		allowed := 0
		for i := 0; i < 10; i++ {
			if monitor.CheckRateLimit("test-key") {
				allowed++
			}
		}
		assert.Less(t, allowed, 10)
	})

	t.Run("Circuit Breaker", func(t *testing.T) {
		config := DefaultPerformanceConfig()
		config.CircuitBreaker.Enabled = true
		config.CircuitBreaker.ConsecutiveFailures = 3
		monitor := NewMonitor(config)

		provider := "test-provider"

		// Should allow initially
		assert.True(t, monitor.CheckCircuitBreaker(provider))

		// Record consecutive failures
		for i := 0; i < 3; i++ {
			monitor.RecordProviderError(provider, false)
		}

		// Should be open after consecutive failures
		assert.False(t, monitor.CheckCircuitBreaker(provider))
	})

	t.Run("Reset Metrics", func(t *testing.T) {
		monitor := NewMonitor(nil)

		// Add some data
		monitor.RecordRequest(RequestMetrics{
			Provider: "test",
			Success:  true,
			Latency:  100 * time.Millisecond,
		})

		metrics := monitor.GetMetrics()
		assert.Equal(t, int64(1), metrics.TotalRequests)

		// Reset
		monitor.ResetMetrics()

		metrics = monitor.GetMetrics()
		assert.Equal(t, int64(0), metrics.TotalRequests)
		assert.Equal(t, int64(0), metrics.SuccessfulRequests)
		assert.Equal(t, int64(0), metrics.FailedRequests)
		assert.Empty(t, metrics.ProviderMetrics)
	})

	t.Run("Resource Limits Check", func(t *testing.T) {
		config := DefaultPerformanceConfig()
		monitor := NewMonitor(config)

		// Should not error with default limits
		err := monitor.CheckResourceLimits()
		assert.NoError(t, err)
	})

	t.Run("ForceGC", func(t *testing.T) {
		monitor := NewMonitor(nil)
		// Should not panic
		monitor.ForceGC()
	})
}

func TestMonitorWithContext(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.MetricsEnabled = true
	config.MetricsInterval = 50 * time.Millisecond
	monitor := NewMonitor(config)

	monitor.Start()

	// Let it run for a bit
	time.Sleep(150 * time.Millisecond)

	// Add some metrics
	monitor.RecordRequest(RequestMetrics{
		Provider: "test",
		Success:  true,
		Latency:  100 * time.Millisecond,
	})

	monitor.Stop()

	// Verify it stopped cleanly
	metrics := monitor.GetMetrics()
	assert.NotNil(t, metrics)
}