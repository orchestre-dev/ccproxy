package performance

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/orchestre-dev/ccproxy/internal/errors"
)

// ResourceMonitor monitors system resource usage
type ResourceMonitor struct {
	limits          ResourceLimits
	lastCPUCheck    time.Time
	cpuUsagePercent float64
	mu              sync.RWMutex
}

// ResourceMetrics represents current resource usage
type ResourceMetrics struct {
	MemoryUsage     uint64  `json:"memory_usage"`
	MemoryLimit     uint64  `json:"memory_limit"`
	GoroutineCount  int     `json:"goroutine_count"`
	GoroutineLimit  int     `json:"goroutine_limit"`
	CPUUsagePercent float64 `json:"cpu_usage_percent"`
	CPULimit        float64 `json:"cpu_limit"`
}

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor(limits ResourceLimits) *ResourceMonitor {
	return &ResourceMonitor{
		limits:       limits,
		lastCPUCheck: time.Now(),
	}
}

// CheckLimits checks if any resource limits are exceeded
func (rm *ResourceMonitor) CheckLimits() error {
	metrics := rm.GetMetrics()

	// Check memory limit
	if rm.limits.MaxMemoryMB > 0 {
		maxMemoryBytes := rm.limits.MaxMemoryMB * 1024 * 1024
		if metrics.MemoryUsage > maxMemoryBytes {
			return errors.New(
				errors.ErrorTypeResourceExhausted,
				fmt.Sprintf("memory limit exceeded: %d MB > %d MB",
					metrics.MemoryUsage/(1024*1024),
					rm.limits.MaxMemoryMB),
			).WithCode("MEMORY_LIMIT_EXCEEDED")
		}
	}

	// Check goroutine limit
	if rm.limits.MaxGoroutines > 0 && metrics.GoroutineCount > rm.limits.MaxGoroutines {
		return errors.New(
			errors.ErrorTypeResourceExhausted,
			fmt.Sprintf("goroutine limit exceeded: %d > %d",
				metrics.GoroutineCount,
				rm.limits.MaxGoroutines),
		).WithCode("GOROUTINE_LIMIT_EXCEEDED")
	}

	// Check CPU limit
	if rm.limits.MaxCPUPercent > 0 && metrics.CPUUsagePercent > rm.limits.MaxCPUPercent {
		return errors.New(
			errors.ErrorTypeResourceExhausted,
			fmt.Sprintf("CPU limit exceeded: %.1f%% > %.1f%%",
				metrics.CPUUsagePercent,
				rm.limits.MaxCPUPercent),
		).WithCode("CPU_LIMIT_EXCEEDED")
	}

	return nil
}

// GetMetrics returns current resource metrics
func (rm *ResourceMonitor) GetMetrics() ResourceMetrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	rm.updateCPUUsage()

	rm.mu.RLock()
	cpuUsage := rm.cpuUsagePercent
	rm.mu.RUnlock()

	return ResourceMetrics{
		MemoryUsage:     memStats.Alloc,
		MemoryLimit:     rm.limits.MaxMemoryMB * 1024 * 1024,
		GoroutineCount:  runtime.NumGoroutine(),
		GoroutineLimit:  rm.limits.MaxGoroutines,
		CPUUsagePercent: cpuUsage,
		CPULimit:        rm.limits.MaxCPUPercent,
	}
}

// updateCPUUsage updates CPU usage percentage
// Note: This is a simplified implementation. In production, you might want to use
// more sophisticated CPU monitoring libraries
func (rm *ResourceMonitor) updateCPUUsage() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	now := time.Now()
	if now.Sub(rm.lastCPUCheck) < 1*time.Second {
		return // Don't update too frequently
	}

	// Get current CPU stats
	// This is a placeholder - in a real implementation, you would read
	// from /proc/stat on Linux or use appropriate system calls
	// For now, we'll use a simple approximation based on goroutines
	goroutines := runtime.NumGoroutine()
	numCPU := runtime.NumCPU()

	// Very rough approximation: assume each active goroutine uses some CPU
	// This is not accurate but provides a basic metric
	rm.cpuUsagePercent = float64(goroutines) / float64(numCPU) * 10.0
	if rm.cpuUsagePercent > 100.0 {
		rm.cpuUsagePercent = 100.0
	}

	rm.lastCPUCheck = now
}

// GetMemoryStats returns detailed memory statistics
func (rm *ResourceMonitor) GetMemoryStats() runtime.MemStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return memStats
}

// CheckRequestSize checks if a request body size is within limits
func (rm *ResourceMonitor) CheckRequestSize(size int64) error {
	if rm.limits.MaxRequestBodyMB > 0 {
		maxBytes := int64(rm.limits.MaxRequestBodyMB * 1024 * 1024)
		if size > maxBytes {
			return errors.New(
				errors.ErrorTypeBadRequest,
				fmt.Sprintf("request body too large: %d MB > %d MB",
					size/(1024*1024),
					rm.limits.MaxRequestBodyMB),
			).WithCode("REQUEST_TOO_LARGE")
		}
	}
	return nil
}

// CheckResponseSize checks if a response body size is within limits
func (rm *ResourceMonitor) CheckResponseSize(size int64) error {
	if rm.limits.MaxResponseBodyMB > 0 {
		maxBytes := int64(rm.limits.MaxResponseBodyMB * 1024 * 1024)
		if size > maxBytes {
			return errors.New(
				errors.ErrorTypeInternal,
				fmt.Sprintf("response body too large: %d MB > %d MB",
					size/(1024*1024),
					rm.limits.MaxResponseBodyMB),
			).WithCode("RESPONSE_TOO_LARGE")
		}
	}
	return nil
}

// GetLimits returns the configured resource limits
func (rm *ResourceMonitor) GetLimits() ResourceLimits {
	return rm.limits
}

// SetLimits updates resource limits
func (rm *ResourceMonitor) SetLimits(limits ResourceLimits) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.limits = limits
}
