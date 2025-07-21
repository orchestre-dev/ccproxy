package testing

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// LoadTestConfig defines load test configuration
type LoadTestConfig struct {
	Duration        time.Duration
	ConcurrentUsers int
	RampUpTime      time.Duration
	RequestsPerUser int
	ThinkTime       time.Duration
	MaxGoroutines   int // Add limit to prevent resource exhaustion
}

// LoadTestResults contains load test results
type LoadTestResults struct {
	TotalRequests   int64
	SuccessfulReqs  int64
	FailedReqs      int64
	SuccessRequests int64         // Alias for SuccessfulReqs
	FailedRequests  int64         // Alias for FailedReqs
	RequestsPerSec  float64
	ErrorRate       float64
	AverageLatency  time.Duration
	MinLatency      time.Duration
	MaxLatency      time.Duration
	TotalDuration   time.Duration // Total duration of the test
}

// LoadTester provides load testing functionality
type LoadTester struct {
	framework *TestFramework
	config    LoadTestConfig
	results   LoadTestResults
	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewLoadTester creates a new load tester
func NewLoadTester(framework *TestFramework, config LoadTestConfig) *LoadTester {
	// Set sensible defaults to prevent resource exhaustion
	if config.MaxGoroutines == 0 {
		config.MaxGoroutines = 1000 // Reasonable limit
	}
	if config.ConcurrentUsers > config.MaxGoroutines {
		config.ConcurrentUsers = config.MaxGoroutines
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), config.Duration)
	
	lt := &LoadTester{
		framework: framework,
		config:    config,
		ctx:       ctx,
		cancel:    cancel,
	}
	
	// Register cleanup
	framework.t.Cleanup(func() {
		lt.cancel()
	})
	
	return lt
}

// Run executes the load test
func (lt *LoadTester) Run(testFunc func() error) LoadTestResults {
	startTime := time.Now()
	defer func() {
		lt.results.TotalDuration = time.Since(startTime)
		// Sync alias fields
		lt.results.SuccessRequests = lt.results.SuccessfulReqs
		lt.results.FailedRequests = lt.results.FailedReqs
	}()
	
	// Use semaphore to limit concurrent goroutines
	sem := make(chan struct{}, lt.config.ConcurrentUsers)
	
	// Ramp up users gradually
	usersPerInterval := 1
	if lt.config.RampUpTime > 0 {
		intervals := int(lt.config.RampUpTime / (100 * time.Millisecond))
		if intervals > 0 {
			usersPerInterval = lt.config.ConcurrentUsers / intervals
			if usersPerInterval < 1 {
				usersPerInterval = 1
			}
		}
	}
	
	var wg sync.WaitGroup
	done := make(chan struct{})
	
	// Start result collector
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				lt.updateMetrics(startTime)
			case <-done:
				return
			}
		}
	}()
	
	// Launch users with ramp-up
	for i := 0; i < lt.config.ConcurrentUsers; i++ {
		if i > 0 && i%usersPerInterval == 0 && lt.config.RampUpTime > 0 {
			time.Sleep(100 * time.Millisecond)
		}
		
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			lt.runUser(userID, testFunc, sem)
		}(i)
	}
	
	// Wait for all users to complete
	wg.Wait()
	close(done)
	
	// Final metrics update
	lt.updateMetrics(startTime)
	
	return lt.results
}

// runUser simulates a single user
func (lt *LoadTester) runUser(userID int, testFunc func() error, sem chan struct{}) {
	requestCount := 0
	
	for {
		select {
		case <-lt.ctx.Done():
			return
		default:
			// Acquire semaphore slot
			sem <- struct{}{}
			
			// Execute request
			start := time.Now()
			err := testFunc()
			duration := time.Since(start)
			
			// Release semaphore slot
			<-sem
			
			// Record results
			lt.recordResult(err, duration)
			
			requestCount++
			
			// Check if user has completed their requests
			if lt.config.RequestsPerUser > 0 && requestCount >= lt.config.RequestsPerUser {
				return
			}
			
			// Apply think time
			if lt.config.ThinkTime > 0 {
				time.Sleep(lt.config.ThinkTime)
			}
		}
	}
}

// recordResult records a single request result
func (lt *LoadTester) recordResult(err error, duration time.Duration) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	
	atomic.AddInt64(&lt.results.TotalRequests, 1)
	
	if err == nil {
		atomic.AddInt64(&lt.results.SuccessfulReqs, 1)
	} else {
		atomic.AddInt64(&lt.results.FailedReqs, 1)
	}
	
	// Update latency stats
	if lt.results.MinLatency == 0 || duration < lt.results.MinLatency {
		lt.results.MinLatency = duration
	}
	if duration > lt.results.MaxLatency {
		lt.results.MaxLatency = duration
	}
	
	// Update average (simplified - in production use running average)
	totalReqs := atomic.LoadInt64(&lt.results.TotalRequests)
	if totalReqs > 0 {
		currentAvg := lt.results.AverageLatency
		lt.results.AverageLatency = (currentAvg*time.Duration(totalReqs-1) + duration) / time.Duration(totalReqs)
	}
}

// updateMetrics updates test metrics
func (lt *LoadTester) updateMetrics(startTime time.Time) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	
	elapsed := time.Since(startTime).Seconds()
	if elapsed > 0 {
		lt.results.RequestsPerSec = float64(lt.results.TotalRequests) / elapsed
	}
	
	if lt.results.TotalRequests > 0 {
		lt.results.ErrorRate = float64(lt.results.FailedReqs) / float64(lt.results.TotalRequests)
	}
}