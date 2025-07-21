package performance

import (
	"time"
)

// Metrics represents performance metrics for the proxy
type Metrics struct {
	// Request metrics
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	AverageLatency     time.Duration `json:"average_latency"`
	P50Latency         time.Duration `json:"p50_latency"`
	P95Latency         time.Duration `json:"p95_latency"`
	P99Latency         time.Duration `json:"p99_latency"`

	// Provider metrics
	ProviderMetrics map[string]*ProviderMetrics `json:"provider_metrics"`

	// Resource metrics
	MemoryUsage     uint64  `json:"memory_usage_bytes"`
	GoroutineCount  int     `json:"goroutine_count"`
	CPUUsagePercent float64 `json:"cpu_usage_percent"`

	// Rate limiting
	RateLimitHits int64 `json:"rate_limit_hits"`

	// Time window
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// ProviderMetrics represents metrics for a specific provider
type ProviderMetrics struct {
	Name               string        `json:"name"`
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	AverageLatency     time.Duration `json:"average_latency"`
	TokensProcessed    int64         `json:"tokens_processed"`
	ErrorRate          float64       `json:"error_rate"`
	HealthStatus       string        `json:"health_status"`
}

// ResourceLimits defines resource limits for the proxy
type ResourceLimits struct {
	MaxMemoryMB       uint64        `json:"max_memory_mb"`
	MaxGoroutines     int           `json:"max_goroutines"`
	MaxCPUPercent     float64       `json:"max_cpu_percent"`
	RequestTimeout    time.Duration `json:"request_timeout"`
	MaxRequestBodyMB  int           `json:"max_request_body_mb"`
	MaxResponseBodyMB int           `json:"max_response_body_mb"`
}

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	Enabled         bool          `json:"enabled"`
	RequestsPerMin  int           `json:"requests_per_min"`
	BurstSize       int           `json:"burst_size"`
	PerProvider     bool          `json:"per_provider"`
	PerAPIKey       bool          `json:"per_api_key"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

// CircuitBreakerConfig defines circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled             bool          `json:"enabled"`
	ErrorThreshold      float64       `json:"error_threshold"`
	ConsecutiveFailures int           `json:"consecutive_failures"`
	OpenDuration        time.Duration `json:"open_duration"`
	HalfOpenMaxRequests int           `json:"half_open_max_requests"`
}

// PerformanceConfig combines all performance-related configurations
type PerformanceConfig struct {
	ResourceLimits  ResourceLimits       `json:"resource_limits"`
	RateLimit       RateLimitConfig      `json:"rate_limit"`
	CircuitBreaker  CircuitBreakerConfig `json:"circuit_breaker"`
	MetricsEnabled  bool                 `json:"metrics_enabled"`
	MetricsInterval time.Duration        `json:"metrics_interval"`
	ProfilerEnabled bool                 `json:"profiler_enabled"`
}

// DefaultPerformanceConfig returns default performance configuration
func DefaultPerformanceConfig() *PerformanceConfig {
	return &PerformanceConfig{
		ResourceLimits: ResourceLimits{
			MaxMemoryMB:       2048, // 2GB
			MaxGoroutines:     10000,
			MaxCPUPercent:     80.0,
			RequestTimeout:    5 * time.Minute,
			MaxRequestBodyMB:  10,
			MaxResponseBodyMB: 100,
		},
		RateLimit: RateLimitConfig{
			Enabled:         false,
			RequestsPerMin:  1000,
			BurstSize:       100,
			PerProvider:     true,
			PerAPIKey:       false,
			CleanupInterval: 5 * time.Minute,
		},
		CircuitBreaker: CircuitBreakerConfig{
			Enabled:             false,
			ErrorThreshold:      0.5, // 50% error rate
			ConsecutiveFailures: 5,
			OpenDuration:        30 * time.Second,
			HalfOpenMaxRequests: 3,
		},
		MetricsEnabled:  true,
		MetricsInterval: 1 * time.Minute,
		ProfilerEnabled: false,
	}
}

// RequestMetrics represents metrics for a single request
type RequestMetrics struct {
	Provider     string
	Model        string
	StartTime    time.Time
	EndTime      time.Time
	Latency      time.Duration
	TokensIn     int
	TokensOut    int
	Success      bool
	Error        error
	StatusCode   int
	RequestSize  int64
	ResponseSize int64
}
