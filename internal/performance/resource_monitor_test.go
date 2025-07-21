package performance

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceMonitor(t *testing.T) {
	t.Run("NewResourceMonitor", func(t *testing.T) {
		limits := ResourceLimits{
			MaxMemoryMB:      1024,
			MaxGoroutines:    5000,
			MaxCPUPercent:    80.0,
			MaxRequestBodyMB: 10,
		}
		monitor := NewResourceMonitor(limits)
		require.NotNil(t, monitor)
		assert.Equal(t, limits, monitor.GetLimits())
	})

	t.Run("GetMetrics", func(t *testing.T) {
		limits := ResourceLimits{
			MaxMemoryMB:   1024,
			MaxGoroutines: 5000,
		}
		monitor := NewResourceMonitor(limits)

		metrics := monitor.GetMetrics()
		assert.Greater(t, metrics.MemoryUsage, uint64(0))
		assert.Equal(t, uint64(1024*1024*1024), metrics.MemoryLimit)
		assert.Greater(t, metrics.GoroutineCount, 0)
		assert.Equal(t, 5000, metrics.GoroutineLimit)
		assert.GreaterOrEqual(t, metrics.CPUUsagePercent, float64(0))
		assert.LessOrEqual(t, metrics.CPUUsagePercent, float64(100))
	})

	t.Run("CheckLimits - Within Limits", func(t *testing.T) {
		limits := ResourceLimits{
			MaxMemoryMB:   10000, // Very high limit
			MaxGoroutines: 50000, // Very high limit
			MaxCPUPercent: 100.0,
		}
		monitor := NewResourceMonitor(limits)

		err := monitor.CheckLimits()
		assert.NoError(t, err)
	})

	t.Run("CheckLimits - Goroutine Limit", func(t *testing.T) {
		currentGoroutines := runtime.NumGoroutine()
		limits := ResourceLimits{
			MaxGoroutines: currentGoroutines - 1, // Set limit below current
		}
		monitor := NewResourceMonitor(limits)

		err := monitor.CheckLimits()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "goroutine limit exceeded")
	})

	t.Run("CheckRequestSize", func(t *testing.T) {
		limits := ResourceLimits{
			MaxRequestBodyMB: 10,
		}
		monitor := NewResourceMonitor(limits)

		// Within limit
		err := monitor.CheckRequestSize(5 * 1024 * 1024) // 5MB
		assert.NoError(t, err)

		// Exceeds limit
		err = monitor.CheckRequestSize(15 * 1024 * 1024) // 15MB
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request body too large")

		// No limit set
		monitor2 := NewResourceMonitor(ResourceLimits{})
		err = monitor2.CheckRequestSize(100 * 1024 * 1024) // 100MB
		assert.NoError(t, err)
	})

	t.Run("CheckResponseSize", func(t *testing.T) {
		limits := ResourceLimits{
			MaxResponseBodyMB: 50,
		}
		monitor := NewResourceMonitor(limits)

		// Within limit
		err := monitor.CheckResponseSize(30 * 1024 * 1024) // 30MB
		assert.NoError(t, err)

		// Exceeds limit
		err = monitor.CheckResponseSize(60 * 1024 * 1024) // 60MB
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "response body too large")
	})

	t.Run("GetMemoryStats", func(t *testing.T) {
		monitor := NewResourceMonitor(ResourceLimits{})
		stats := monitor.GetMemoryStats()
		assert.Greater(t, stats.Alloc, uint64(0))
		assert.Greater(t, stats.TotalAlloc, uint64(0))
	})

	t.Run("SetLimits", func(t *testing.T) {
		monitor := NewResourceMonitor(ResourceLimits{
			MaxMemoryMB: 512,
		})

		newLimits := ResourceLimits{
			MaxMemoryMB:   1024,
			MaxGoroutines: 10000,
		}
		monitor.SetLimits(newLimits)

		assert.Equal(t, newLimits, monitor.GetLimits())
	})

	t.Run("CPU Usage Approximation", func(t *testing.T) {
		monitor := NewResourceMonitor(ResourceLimits{})

		// Force CPU usage update
		monitor.updateCPUUsage()
		metrics := monitor.GetMetrics()

		// CPU usage should be a reasonable percentage
		assert.GreaterOrEqual(t, metrics.CPUUsagePercent, float64(0))
		assert.LessOrEqual(t, metrics.CPUUsagePercent, float64(100))
	})
}