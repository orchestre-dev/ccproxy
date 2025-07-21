package state

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ReadinessCheck represents a readiness check function
type ReadinessCheck func(ctx context.Context) error

// ReadinessProbe manages readiness checks for components
type ReadinessProbe struct {
	mu       sync.RWMutex
	checks   map[string]ReadinessCheck
	results  map[string]*CheckResult
	manager  *Manager
	interval time.Duration
	timeout  time.Duration
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// CheckResult represents the result of a readiness check
type CheckResult struct {
	Success     bool
	Error       error
	LastChecked time.Time
	Duration    time.Duration
}

// NewReadinessProbe creates a new readiness probe
func NewReadinessProbe(manager *Manager, interval, timeout time.Duration) *ReadinessProbe {
	return &ReadinessProbe{
		checks:   make(map[string]ReadinessCheck),
		results:  make(map[string]*CheckResult),
		manager:  manager,
		interval: interval,
		timeout:  timeout,
	}
}

// RegisterCheck registers a readiness check for a component
func (rp *ReadinessProbe) RegisterCheck(name string, check ReadinessCheck) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	rp.checks[name] = check
	rp.results[name] = &CheckResult{
		Success:     false,
		LastChecked: time.Time{},
	}

	// Register component with state manager
	if rp.manager != nil {
		rp.manager.RegisterComponent(name)
	}
}

// Start begins running readiness checks
func (rp *ReadinessProbe) Start(ctx context.Context) {
	rp.mu.Lock()
	if rp.cancel != nil {
		rp.mu.Unlock()
		return // Already running
	}

	ctx, cancel := context.WithCancel(ctx)
	rp.cancel = cancel
	rp.mu.Unlock()

	// Initial check
	rp.runChecks(ctx)

	// Start periodic checks
	rp.wg.Add(1)
	go func() {
		defer rp.wg.Done()
		
		ticker := time.NewTicker(rp.interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				rp.runChecks(ctx)
			}
		}
	}()
}

// Stop stops running readiness checks
func (rp *ReadinessProbe) Stop() {
	rp.mu.Lock()
	if rp.cancel != nil {
		rp.cancel()
		rp.cancel = nil
	}
	rp.mu.Unlock()

	rp.wg.Wait()
}

// runChecks executes all registered checks
func (rp *ReadinessProbe) runChecks(ctx context.Context) {
	rp.mu.RLock()
	checks := make(map[string]ReadinessCheck)
	for name, check := range rp.checks {
		checks[name] = check
	}
	rp.mu.RUnlock()

	var wg sync.WaitGroup
	for name, check := range checks {
		wg.Add(1)
		go func(n string, c ReadinessCheck) {
			defer wg.Done()
			rp.runCheck(ctx, n, c)
		}(name, check)
	}
	wg.Wait()
}

// runCheck executes a single readiness check
func (rp *ReadinessProbe) runCheck(ctx context.Context, name string, check ReadinessCheck) {
	start := time.Now()

	// Create timeout context
	checkCtx, cancel := context.WithTimeout(ctx, rp.timeout)
	defer cancel()

	// Run the check
	err := check(checkCtx)
	duration := time.Since(start)

	// Update result
	rp.mu.Lock()
	result := &CheckResult{
		Success:     err == nil,
		Error:       err,
		LastChecked: time.Now(),
		Duration:    duration,
	}
	rp.results[name] = result
	rp.mu.Unlock()

	// Update component state
	if rp.manager != nil {
		if err == nil {
			rp.manager.SetComponentState(name, StateReady, nil)
		} else {
			rp.manager.SetComponentState(name, StateError, err)
		}
	}
}

// GetResult returns the result of a specific check
func (rp *ReadinessProbe) GetResult(name string) (*CheckResult, bool) {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	result, exists := rp.results[name]
	if !exists {
		return nil, false
	}

	// Return a copy
	copy := *result
	return &copy, true
}

// GetAllResults returns all check results
func (rp *ReadinessProbe) GetAllResults() map[string]CheckResult {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	results := make(map[string]CheckResult)
	for name, result := range rp.results {
		results[name] = *result
	}
	return results
}

// IsReady returns true if all checks are passing
func (rp *ReadinessProbe) IsReady() bool {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	if len(rp.results) == 0 {
		return false
	}

	for _, result := range rp.results {
		if !result.Success {
			return false
		}
	}

	return true
}

// WaitForReady waits for all checks to be ready
func (rp *ReadinessProbe) WaitForReady(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for readiness")
		case <-ticker.C:
			if rp.IsReady() {
				return nil
			}
		}
	}
}

// RunOnce executes all checks once and returns results
func (rp *ReadinessProbe) RunOnce(ctx context.Context) map[string]CheckResult {
	rp.runChecks(ctx)
	return rp.GetAllResults()
}