package performance

import (
	"sync"
	"sync/atomic"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	// StateClosed means requests are allowed
	StateClosed CircuitState = iota
	// StateOpen means requests are blocked
	StateOpen
	// StateHalfOpen means limited requests are allowed for testing
	StateHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name             string
	config           CircuitBreakerConfig
	state            CircuitState
	failures         int64
	successes        int64
	consecutiveFails int64
	lastStateChange  time.Time
	halfOpenRequests int64
	mu               sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		name:            name,
		config:          config,
		state:           StateClosed,
		lastStateChange: time.Now(),
	}
}

// Allow checks if a request should be allowed
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	
	case StateOpen:
		// Check if we should transition to half-open
		if time.Since(cb.lastStateChange) > cb.config.OpenDuration {
			cb.transitionTo(StateHalfOpen)
			return cb.allowHalfOpen()
		}
		return false
	
	case StateHalfOpen:
		return cb.allowHalfOpen()
	
	default:
		return false
	}
}

// Record records the result of a request
func (cb *CircuitBreaker) Record(success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if success {
		cb.recordSuccess()
	} else {
		cb.recordFailure()
	}
}

// GetState returns the current state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	total := cb.failures + cb.successes
	errorRate := float64(0)
	if total > 0 {
		errorRate = float64(cb.failures) / float64(total)
	}

	return map[string]interface{}{
		"name":              cb.name,
		"state":             cb.stateString(),
		"failures":          cb.failures,
		"successes":         cb.successes,
		"consecutive_fails": cb.consecutiveFails,
		"error_rate":        errorRate,
		"last_state_change": cb.lastStateChange,
	}
}

// Reset resets the circuit breaker
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.consecutiveFails = 0
	cb.halfOpenRequests = 0
	cb.lastStateChange = time.Now()
}

// recordSuccess records a successful request
func (cb *CircuitBreaker) recordSuccess() {
	atomic.AddInt64(&cb.successes, 1)
	atomic.StoreInt64(&cb.consecutiveFails, 0)

	switch cb.state {
	case StateHalfOpen:
		// If we've had enough successful requests in half-open state, close the circuit
		if atomic.LoadInt64(&cb.halfOpenRequests) >= int64(cb.config.HalfOpenMaxRequests) {
			cb.transitionTo(StateClosed)
		}
	}
}

// recordFailure records a failed request
func (cb *CircuitBreaker) recordFailure() {
	atomic.AddInt64(&cb.failures, 1)
	consecutive := atomic.AddInt64(&cb.consecutiveFails, 1)

	total := cb.failures + cb.successes
	errorRate := float64(cb.failures) / float64(total)

	switch cb.state {
	case StateClosed:
		// Check if we should open the circuit
		if consecutive >= int64(cb.config.ConsecutiveFailures) || 
		   (total > 10 && errorRate >= cb.config.ErrorThreshold) {
			cb.transitionTo(StateOpen)
		}
	
	case StateHalfOpen:
		// Any failure in half-open state reopens the circuit
		cb.transitionTo(StateOpen)
	}
}

// allowHalfOpen checks if a request should be allowed in half-open state
func (cb *CircuitBreaker) allowHalfOpen() bool {
	current := atomic.AddInt64(&cb.halfOpenRequests, 1)
	return current <= int64(cb.config.HalfOpenMaxRequests)
}

// transitionTo transitions to a new state
func (cb *CircuitBreaker) transitionTo(newState CircuitState) {
	if cb.state != newState {
		cb.state = newState
		cb.lastStateChange = time.Now()
		
		if newState == StateHalfOpen {
			atomic.StoreInt64(&cb.halfOpenRequests, 0)
		}
		
		// Log state change
		// utils.GetLogger().Infof("Circuit breaker %s: %s -> %s", cb.name, cb.stateString(), newStateString(newState))
	}
}

// stateString returns a string representation of the current state
func (cb *CircuitBreaker) stateString() string {
	return stateToString(cb.state)
}

// stateToString converts a circuit state to string
func stateToString(state CircuitState) string {
	switch state {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// IsOpen returns true if the circuit is open
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state == StateOpen
}

// IsClosed returns true if the circuit is closed
func (cb *CircuitBreaker) IsClosed() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state == StateClosed
}

// IsHalfOpen returns true if the circuit is half-open
func (cb *CircuitBreaker) IsHalfOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state == StateHalfOpen
}