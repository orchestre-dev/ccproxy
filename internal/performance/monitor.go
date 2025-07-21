package performance

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/orchestre-dev/ccproxy/internal/utils"
)

// Monitor tracks performance metrics and enforces resource limits
type Monitor struct {
	config          *PerformanceConfig
	metrics         *Metrics
	latencyTracker  *LatencyTracker
	resourceMonitor *ResourceMonitor
	rateLimiter     *RateLimiter
	circuitBreakers map[string]*CircuitBreaker
	
	requestCount    int64
	successCount    int64
	failureCount    int64
	
	startTime       time.Time
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	mu              sync.RWMutex
}

// NewMonitor creates a new performance monitor
func NewMonitor(config *PerformanceConfig) *Monitor {
	if config == nil {
		config = DefaultPerformanceConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	m := &Monitor{
		config:          config,
		metrics:         &Metrics{
			ProviderMetrics: make(map[string]*ProviderMetrics),
			StartTime:       time.Now(),
		},
		latencyTracker:  NewLatencyTracker(),
		resourceMonitor: NewResourceMonitor(config.ResourceLimits),
		circuitBreakers: make(map[string]*CircuitBreaker),
		startTime:       time.Now(),
		ctx:             ctx,
		cancel:          cancel,
	}

	if config.RateLimit.Enabled {
		m.rateLimiter = NewRateLimiter(config.RateLimit)
	}

	return m
}

// Start begins monitoring
func (m *Monitor) Start() {
	if !m.config.MetricsEnabled {
		return
	}

	m.wg.Add(1)
	go m.metricsCollector()

	if m.config.ProfilerEnabled {
		m.wg.Add(1)
		go m.profiler()
	}

	utils.GetLogger().Info("Performance monitoring started")
}

// Stop stops monitoring
func (m *Monitor) Stop() {
	m.cancel()
	m.wg.Wait()
	utils.GetLogger().Info("Performance monitoring stopped")
}

// RecordRequest records metrics for a request
func (m *Monitor) RecordRequest(metrics RequestMetrics) {
	// Update counters
	atomic.AddInt64(&m.requestCount, 1)
	if metrics.Success {
		atomic.AddInt64(&m.successCount, 1)
	} else {
		atomic.AddInt64(&m.failureCount, 1)
	}

	// Record latency
	m.latencyTracker.Record(metrics.Latency)

	// Update provider metrics
	m.updateProviderMetrics(metrics)
}

// CheckRateLimit checks if a request should be rate limited
func (m *Monitor) CheckRateLimit(key string) bool {
	if m.rateLimiter == nil {
		return true // No rate limiting
	}
	return m.rateLimiter.Allow(key)
}

// CheckCircuitBreaker checks if requests to a provider should be allowed
func (m *Monitor) CheckCircuitBreaker(provider string) bool {
	m.mu.RLock()
	cb, exists := m.circuitBreakers[provider]
	m.mu.RUnlock()

	if !exists || !m.config.CircuitBreaker.Enabled {
		return true
	}

	return cb.Allow()
}

// RecordProviderError records an error for circuit breaker tracking
func (m *Monitor) RecordProviderError(provider string, success bool) {
	if !m.config.CircuitBreaker.Enabled {
		return
	}

	m.mu.Lock()
	cb, exists := m.circuitBreakers[provider]
	if !exists {
		cb = NewCircuitBreaker(provider, m.config.CircuitBreaker)
		m.circuitBreakers[provider] = cb
	}
	m.mu.Unlock()

	cb.Record(success)
}

// GetMetrics returns current performance metrics
func (m *Monitor) GetMetrics() *Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Clone metrics to avoid races
	metrics := &Metrics{
		TotalRequests:      atomic.LoadInt64(&m.requestCount),
		SuccessfulRequests: atomic.LoadInt64(&m.successCount),
		FailedRequests:     atomic.LoadInt64(&m.failureCount),
		StartTime:          m.startTime,
		EndTime:            time.Now(),
		ProviderMetrics:    make(map[string]*ProviderMetrics),
	}

	// Get latency percentiles
	percentiles := m.latencyTracker.GetPercentiles()
	metrics.AverageLatency = percentiles.Average
	metrics.P50Latency = percentiles.P50
	metrics.P95Latency = percentiles.P95
	metrics.P99Latency = percentiles.P99

	// Copy provider metrics
	for k, v := range m.metrics.ProviderMetrics {
		metrics.ProviderMetrics[k] = &ProviderMetrics{
			Name:               v.Name,
			TotalRequests:      v.TotalRequests,
			SuccessfulRequests: v.SuccessfulRequests,
			FailedRequests:     v.FailedRequests,
			AverageLatency:     v.AverageLatency,
			TokensProcessed:    v.TokensProcessed,
			ErrorRate:          v.ErrorRate,
			HealthStatus:       v.HealthStatus,
		}
	}

	// Get resource metrics
	if m.resourceMonitor != nil {
		resourceMetrics := m.resourceMonitor.GetMetrics()
		metrics.MemoryUsage = resourceMetrics.MemoryUsage
		metrics.GoroutineCount = resourceMetrics.GoroutineCount
		metrics.CPUUsagePercent = resourceMetrics.CPUUsagePercent
	}

	// Get rate limit hits
	if m.rateLimiter != nil {
		metrics.RateLimitHits = m.rateLimiter.GetHits()
	}

	return metrics
}

// CheckResourceLimits checks if resource limits are exceeded
func (m *Monitor) CheckResourceLimits() error {
	if m.resourceMonitor != nil {
		return m.resourceMonitor.CheckLimits()
	}
	return nil
}

// metricsCollector periodically collects and logs metrics
func (m *Monitor) metricsCollector() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.collectMetrics()
		}
	}
}

// collectMetrics collects current metrics
func (m *Monitor) collectMetrics() {
	metrics := m.GetMetrics()
	
	// Log summary
	utils.GetLogger().Infof(
		"Performance metrics - Requests: %d (success: %d, failed: %d), Avg latency: %v, Memory: %d MB, Goroutines: %d",
		metrics.TotalRequests,
		metrics.SuccessfulRequests,
		metrics.FailedRequests,
		metrics.AverageLatency,
		metrics.MemoryUsage/(1024*1024),
		metrics.GoroutineCount,
	)

	// Check resource limits
	if err := m.CheckResourceLimits(); err != nil {
		utils.GetLogger().Warnf("Resource limit warning: %v", err)
	}
}

// profiler runs CPU and memory profiling if enabled
func (m *Monitor) profiler() {
	defer m.wg.Done()

	// CPU profiling would be implemented here
	// Memory profiling would be implemented here
}

// updateProviderMetrics updates metrics for a specific provider
func (m *Monitor) updateProviderMetrics(req RequestMetrics) {
	m.mu.Lock()
	defer m.mu.Unlock()

	pm, exists := m.metrics.ProviderMetrics[req.Provider]
	if !exists {
		pm = &ProviderMetrics{
			Name: req.Provider,
		}
		m.metrics.ProviderMetrics[req.Provider] = pm
	}

	pm.TotalRequests++
	if req.Success {
		pm.SuccessfulRequests++
	} else {
		pm.FailedRequests++
	}

	// Update average latency (simple moving average)
	if pm.TotalRequests == 1 {
		pm.AverageLatency = req.Latency
	} else {
		pm.AverageLatency = time.Duration(
			(int64(pm.AverageLatency)*(pm.TotalRequests-1) + int64(req.Latency)) / pm.TotalRequests,
		)
	}

	// Update tokens
	pm.TokensProcessed += int64(req.TokensIn + req.TokensOut)

	// Calculate error rate
	if pm.TotalRequests > 0 {
		pm.ErrorRate = float64(pm.FailedRequests) / float64(pm.TotalRequests)
	}
}

// ResetMetrics resets all metrics
func (m *Monitor) ResetMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.StoreInt64(&m.requestCount, 0)
	atomic.StoreInt64(&m.successCount, 0)
	atomic.StoreInt64(&m.failureCount, 0)

	m.metrics = &Metrics{
		ProviderMetrics: make(map[string]*ProviderMetrics),
		StartTime:       time.Now(),
	}

	m.latencyTracker.Reset()
	m.startTime = time.Now()

	utils.GetLogger().Info("Performance metrics reset")
}

// GetResourceMonitor returns the resource monitor
func (m *Monitor) GetResourceMonitor() *ResourceMonitor {
	return m.resourceMonitor
}

// ForceGC forces a garbage collection cycle
func (m *Monitor) ForceGC() {
	runtime.GC()
	utils.GetLogger().Debug("Forced garbage collection")
}