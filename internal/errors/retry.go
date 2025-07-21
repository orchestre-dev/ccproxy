package errors

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/orchestre-dev/ccproxy/internal/utils"
)

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxAttempts     int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	Multiplier      float64
	Jitter          bool
	RetryableErrors []ErrorType
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
		RetryableErrors: []ErrorType{
			ErrorTypeInternal,
			ErrorTypeBadGateway,
			ErrorTypeServiceUnavailable,
			ErrorTypeGatewayTimeout,
			ErrorTypeTooManyRequests,
			ErrorTypeRateLimitError,
		},
	}
}

// RetryFunc is a function that can be retried
type RetryFunc func(ctx context.Context) error

// RetryWithConfig executes a function with retry logic
func RetryWithConfig(ctx context.Context, config *RetryConfig, fn RetryFunc) error {
	logger := utils.GetLogger()
	
	var lastErr error
	delay := config.InitialDelay
	
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Check context
		if err := ctx.Err(); err != nil {
			return Wrap(err, ErrorTypeInternal, "Context cancelled")
		}
		
		// Execute function
		err := fn(ctx)
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// Check if error is retryable
		if !isRetryableWithConfig(err, config) {
			return err
		}
		
		// Check if we've exhausted attempts
		if attempt >= config.MaxAttempts {
			break
		}
		
		// Calculate delay
		actualDelay := calculateDelay(delay, config, err)
		
		logger.WithField("attempt", attempt).
			WithField("delay_ms", actualDelay.Milliseconds()).
			WithField("error", err.Error()).
			Debug("Retrying after error")
		
		// Wait with context
		select {
		case <-ctx.Done():
			return Wrap(ctx.Err(), ErrorTypeInternal, "Context cancelled during retry")
		case <-time.After(actualDelay):
			// Continue to next attempt
		}
		
		// Update delay for next attempt
		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}
	
	// Wrap the last error to indicate retry exhaustion
	return Wrap(lastErr, ErrorTypeInternal, "Max retry attempts exceeded")
}

// Retry executes a function with default retry configuration
func Retry(ctx context.Context, fn RetryFunc) error {
	return RetryWithConfig(ctx, DefaultRetryConfig(), fn)
}

// RetryWithBackoff executes a function with exponential backoff
func RetryWithBackoff(ctx context.Context, attempts int, delay time.Duration, fn RetryFunc) error {
	config := &RetryConfig{
		MaxAttempts:  attempts,
		InitialDelay: delay,
		MaxDelay:     delay * time.Duration(math.Pow(2, float64(attempts-1))),
		Multiplier:   2.0,
		Jitter:       true,
	}
	return RetryWithConfig(ctx, config, fn)
}

// isRetryableWithConfig checks if an error is retryable based on config
func isRetryableWithConfig(err error, config *RetryConfig) bool {
	// First check if it's a CCProxyError
	var ccErr *CCProxyError
	if e, ok := err.(*CCProxyError); ok {
		ccErr = e
		
		// Check if error type is in retryable list
		if config.RetryableErrors != nil {
			for _, retryableType := range config.RetryableErrors {
				if ccErr.Type == retryableType {
					return true
				}
			}
			// If we have a specific list and the error isn't in it, don't retry
			return false
		}
		
		// Otherwise use the error's retryable flag
		return ccErr.Retryable
	}
	
	// For non-CCProxyError, use generic check
	return IsRetryableError(err)
}

// calculateDelay calculates the actual delay based on config and error
func calculateDelay(baseDelay time.Duration, config *RetryConfig, err error) time.Duration {
	delay := baseDelay
	
	// Check if error specifies retry after
	if retryAfter := GetRetryAfter(err); retryAfter != nil {
		// Use the larger of retry after or calculated delay
		if *retryAfter > delay {
			delay = *retryAfter
		}
	}
	
	// Apply jitter if configured
	if config.Jitter {
		// Add ±25% jitter
		jitter := rand.Float64()*0.5 - 0.25
		delay = time.Duration(float64(delay) * (1 + jitter))
	}
	
	// Ensure delay doesn't exceed max
	if delay > config.MaxDelay {
		delay = config.MaxDelay
	}
	
	return delay
}

// CircuitBreaker provides circuit breaker functionality
type CircuitBreaker struct {
	name            string
	maxFailures     int
	resetTimeout    time.Duration
	halfOpenTimeout time.Duration
	
	failures      int
	lastFailTime  time.Time
	state         CircuitState
	stateChanged  time.Time
}

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:            name,
		maxFailures:     maxFailures,
		resetTimeout:    resetTimeout,
		halfOpenTimeout: resetTimeout / 2,
		state:           CircuitClosed,
		stateChanged:    time.Now(),
	}
}

// Execute runs a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn RetryFunc) error {
	// Check circuit state
	state := cb.getState()
	
	switch state {
	case CircuitOpen:
		return New(ErrorTypeServiceUnavailable, "Circuit breaker is open").
			WithDetails(map[string]interface{}{
				"circuit": cb.name,
				"state":   "open",
			})
		
	case CircuitHalfOpen:
		// Allow one request through
		err := fn(ctx)
		if err == nil {
			cb.onSuccess()
			return nil
		}
		cb.onFailure()
		return err
		
	case CircuitClosed:
		err := fn(ctx)
		if err == nil {
			cb.onSuccess()
			return nil
		}
		
		// Only count retryable errors as failures
		if IsRetryableError(err) {
			cb.onFailure()
		}
		return err
		
	default:
		return New(ErrorTypeInternal, "Unknown circuit breaker state")
	}
}

// getState returns the current circuit state
func (cb *CircuitBreaker) getState() CircuitState {
	now := time.Now()
	
	switch cb.state {
	case CircuitOpen:
		if now.Sub(cb.stateChanged) > cb.resetTimeout {
			cb.state = CircuitHalfOpen
			cb.stateChanged = now
			utils.GetLogger().Infof("Circuit breaker %s: transitioning to half-open", cb.name)
		}
		
	case CircuitHalfOpen:
		if now.Sub(cb.stateChanged) > cb.halfOpenTimeout {
			// Timeout in half-open state, reset to closed
			cb.state = CircuitClosed
			cb.failures = 0
			cb.stateChanged = now
			utils.GetLogger().Infof("Circuit breaker %s: resetting to closed", cb.name)
		}
	}
	
	return cb.state
}

// onSuccess handles successful execution
func (cb *CircuitBreaker) onSuccess() {
	cb.failures = 0
	if cb.state != CircuitClosed {
		cb.state = CircuitClosed
		cb.stateChanged = time.Now()
		utils.GetLogger().Infof("Circuit breaker %s: closed after success", cb.name)
	}
}

// onFailure handles failed execution
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()
	
	if cb.failures >= cb.maxFailures {
		cb.state = CircuitOpen
		cb.stateChanged = time.Now()
		utils.GetLogger().Warnf("Circuit breaker %s: opened after %d failures", cb.name, cb.failures)
	}
}

// GetState returns the current state name
func (cb *CircuitBreaker) GetState() string {
	switch cb.getState() {
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	case CircuitClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"name":         cb.name,
		"state":        cb.GetState(),
		"failures":     cb.failures,
		"last_failure": cb.lastFailTime,
		"state_since":  cb.stateChanged,
	}
}