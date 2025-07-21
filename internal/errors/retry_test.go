package errors

import (
	"context"
	"testing"
	"time"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	testutil.AssertEqual(t, 3, config.MaxAttempts)
	testutil.AssertEqual(t, 1*time.Second, config.InitialDelay)
	testutil.AssertEqual(t, 30*time.Second, config.MaxDelay)
	testutil.AssertEqual(t, 2.0, config.Multiplier)
	testutil.AssertEqual(t, true, config.Jitter)
	testutil.AssertEqual(t, 6, len(config.RetryableErrors))

	// Check that all expected retryable error types are included
	expectedTypes := []ErrorType{
		ErrorTypeInternal,
		ErrorTypeBadGateway,
		ErrorTypeServiceUnavailable,
		ErrorTypeGatewayTimeout,
		ErrorTypeTooManyRequests,
		ErrorTypeRateLimitError,
	}

	for _, expectedType := range expectedTypes {
		found := false
		for _, retryableType := range config.RetryableErrors {
			if retryableType == expectedType {
				found = true
				break
			}
		}
		testutil.AssertTrue(t, found)
	}
}

func TestRetryWithConfig(t *testing.T) {
	t.Run("SuccessOnFirstAttempt", func(t *testing.T) {
		config := &RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     100 * time.Millisecond,
			Multiplier:   2.0,
			Jitter:       false,
		}

		attempts := 0
		fn := func(ctx context.Context) error {
			attempts++
			return nil
		}

		ctx := context.Background()
		err := RetryWithConfig(ctx, config, fn)

		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 1, attempts)
	})

	t.Run("SuccessOnSecondAttempt", func(t *testing.T) {
		config := &RetryConfig{
			MaxAttempts:     3,
			InitialDelay:    10 * time.Millisecond,
			MaxDelay:        100 * time.Millisecond,
			Multiplier:      2.0,
			Jitter:          false,
			RetryableErrors: []ErrorType{ErrorTypeInternal},
		}

		attempts := 0
		fn := func(ctx context.Context) error {
			attempts++
			if attempts == 1 {
				return New(ErrorTypeInternal, "temporary error")
			}
			return nil
		}

		ctx := context.Background()
		err := RetryWithConfig(ctx, config, fn)

		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 2, attempts)
	})

	t.Run("NonRetryableError", func(t *testing.T) {
		config := &RetryConfig{
			MaxAttempts:     3,
			InitialDelay:    10 * time.Millisecond,
			MaxDelay:        100 * time.Millisecond,
			Multiplier:      2.0,
			Jitter:          false,
			RetryableErrors: []ErrorType{ErrorTypeInternal},
		}

		attempts := 0
		expectedErr := New(ErrorTypeBadRequest, "bad request")
		fn := func(ctx context.Context) error {
			attempts++
			return expectedErr
		}

		ctx := context.Background()
		err := RetryWithConfig(ctx, config, fn)

		testutil.AssertEqual(t, expectedErr, err)
		testutil.AssertEqual(t, 1, attempts)
	})

	t.Run("MaxAttemptsExceeded", func(t *testing.T) {
		config := &RetryConfig{
			MaxAttempts:     2,
			InitialDelay:    10 * time.Millisecond,
			MaxDelay:        100 * time.Millisecond,
			Multiplier:      2.0,
			Jitter:          false,
			RetryableErrors: []ErrorType{ErrorTypeInternal},
		}

		attempts := 0
		originalErr := New(ErrorTypeInternal, "persistent error")
		fn := func(ctx context.Context) error {
			attempts++
			return originalErr
		}

		ctx := context.Background()
		err := RetryWithConfig(ctx, config, fn)

		testutil.AssertEqual(t, false, err == nil)
		testutil.AssertEqual(t, 2, attempts)

		ccErr, ok := err.(*CCProxyError)
		testutil.AssertTrue(t, ok)
		testutil.AssertEqual(t, ErrorTypeInternal, ccErr.Type)
		testutil.AssertEqual(t, "Max retry attempts exceeded", ccErr.Message)
		testutil.AssertEqual(t, originalErr, ccErr.Unwrap())
	})

	t.Run("ContextCancelled", func(t *testing.T) {
		config := &RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     100 * time.Millisecond,
			Multiplier:   2.0,
			Jitter:       false,
		}

		attempts := 0
		fn := func(ctx context.Context) error {
			attempts++
			return New(ErrorTypeInternal, "temporary error")
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := RetryWithConfig(ctx, config, fn)

		testutil.AssertEqual(t, false, err == nil)
		testutil.AssertEqual(t, 0, attempts)

		ccErr, ok := err.(*CCProxyError)
		testutil.AssertTrue(t, ok)
		testutil.AssertEqual(t, ErrorTypeInternal, ccErr.Type)
		testutil.AssertEqual(t, "Context cancelled", ccErr.Message)
	})

	t.Run("ContextCancelledDuringRetry", func(t *testing.T) {
		config := &RetryConfig{
			MaxAttempts:     3,
			InitialDelay:    100 * time.Millisecond,
			MaxDelay:        1 * time.Second,
			Multiplier:      2.0,
			Jitter:          false,
			RetryableErrors: []ErrorType{ErrorTypeInternal},
		}

		attempts := 0
		fn := func(ctx context.Context) error {
			attempts++
			return New(ErrorTypeInternal, "temporary error")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		err := RetryWithConfig(ctx, config, fn)

		testutil.AssertEqual(t, false, err == nil)
		testutil.AssertEqual(t, 1, attempts) // Should have tried once

		ccErr, ok := err.(*CCProxyError)
		testutil.AssertTrue(t, ok)
		testutil.AssertEqual(t, ErrorTypeInternal, ccErr.Type)
		testutil.AssertEqual(t, "Context cancelled during retry", ccErr.Message)
	})

	t.Run("WithRetryAfter", func(t *testing.T) {
		config := &RetryConfig{
			MaxAttempts:     3,
			InitialDelay:    10 * time.Millisecond,
			MaxDelay:        1 * time.Second,
			Multiplier:      2.0,
			Jitter:          false,
			RetryableErrors: []ErrorType{ErrorTypeRateLimitError},
		}

		attempts := 0
		fn := func(ctx context.Context) error {
			attempts++
			if attempts == 1 {
				return New(ErrorTypeRateLimitError, "rate limited").WithRetryAfter(50 * time.Millisecond)
			}
			return nil
		}

		start := time.Now()
		ctx := context.Background()
		err := RetryWithConfig(ctx, config, fn)
		elapsed := time.Since(start)

		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 2, attempts)
		testutil.AssertTrue(t, elapsed >= 50*time.Millisecond) // Should have waited at least 50ms
	})
}

func TestRetry(t *testing.T) {
	attempts := 0
	fn := func(ctx context.Context) error {
		attempts++
		return nil
	}

	ctx := context.Background()
	err := Retry(ctx, fn)

	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, 1, attempts)
}

func TestRetryWithBackoff(t *testing.T) {
	attempts := 0
	fn := func(ctx context.Context) error {
		attempts++
		if attempts == 1 {
			return New(ErrorTypeInternal, "temporary error")
		}
		return nil
	}

	ctx := context.Background()
	err := RetryWithBackoff(ctx, 3, 10*time.Millisecond, fn)

	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, 2, attempts)
}

func TestIsRetryableWithConfig(t *testing.T) {
	t.Run("CCProxyErrorWithRetryableList", func(t *testing.T) {
		config := &RetryConfig{
			RetryableErrors: []ErrorType{ErrorTypeInternal, ErrorTypeBadGateway},
		}

		// Should be retryable
		retryableErr := New(ErrorTypeInternal, "internal error")
		testutil.AssertTrue(t, isRetryableWithConfig(retryableErr, config))

		// Should not be retryable
		nonRetryableErr := New(ErrorTypeBadRequest, "bad request")
		testutil.AssertFalse(t, isRetryableWithConfig(nonRetryableErr, config))
	})

	t.Run("CCProxyErrorWithoutRetryableList", func(t *testing.T) {
		config := &RetryConfig{
			RetryableErrors: nil, // No specific list
		}

		// Should use error's retryable flag
		retryableErr := New(ErrorTypeInternal, "internal error")
		testutil.AssertTrue(t, isRetryableWithConfig(retryableErr, config))

		nonRetryableErr := New(ErrorTypeBadRequest, "bad request")
		testutil.AssertFalse(t, isRetryableWithConfig(nonRetryableErr, config))
	})

	t.Run("RegularError", func(t *testing.T) {
		config := &RetryConfig{
			RetryableErrors: []ErrorType{ErrorTypeInternal},
		}

		// Should use generic check
		retryableErr := &testError{"connection refused"}
		testutil.AssertTrue(t, isRetryableWithConfig(retryableErr, config))

		nonRetryableErr := &testError{"invalid input"}
		testutil.AssertFalse(t, isRetryableWithConfig(nonRetryableErr, config))
	})
}

func TestCalculateDelay(t *testing.T) {
	t.Run("WithoutJitter", func(t *testing.T) {
		config := &RetryConfig{
			MaxDelay: 1 * time.Second,
			Jitter:   false,
		}

		baseDelay := 100 * time.Millisecond
		delay := calculateDelay(baseDelay, config, &testError{"test error"})

		testutil.AssertEqual(t, baseDelay, delay)
	})

	t.Run("WithJitter", func(t *testing.T) {
		config := &RetryConfig{
			MaxDelay: 1 * time.Second,
			Jitter:   true,
		}

		baseDelay := 100 * time.Millisecond
		delay := calculateDelay(baseDelay, config, &testError{"test error"})

		// Should be different from base delay (with very high probability)
		// but within reasonable bounds (75-125ms)
		minDelay := time.Duration(float64(baseDelay) * 0.75)
		maxDelayBound := time.Duration(float64(baseDelay) * 1.25)
		testutil.AssertTrue(t, delay >= minDelay && delay <= maxDelayBound)
	})

	t.Run("WithRetryAfter", func(t *testing.T) {
		config := &RetryConfig{
			MaxDelay: 1 * time.Second,
			Jitter:   false,
		}

		baseDelay := 100 * time.Millisecond
		retryAfter := 300 * time.Millisecond
		err := New(ErrorTypeRateLimitError, "rate limited").WithRetryAfter(retryAfter)

		delay := calculateDelay(baseDelay, config, err)

		testutil.AssertEqual(t, retryAfter, delay) // Should use the larger retry after
	})

	t.Run("RetryAfterSmallerThanBase", func(t *testing.T) {
		config := &RetryConfig{
			MaxDelay: 1 * time.Second,
			Jitter:   false,
		}

		baseDelay := 300 * time.Millisecond
		retryAfter := 100 * time.Millisecond
		err := New(ErrorTypeRateLimitError, "rate limited").WithRetryAfter(retryAfter)

		delay := calculateDelay(baseDelay, config, err)

		testutil.AssertEqual(t, baseDelay, delay) // Should use the larger base delay
	})

	t.Run("ExceedsMaxDelay", func(t *testing.T) {
		config := &RetryConfig{
			MaxDelay: 50 * time.Millisecond,
			Jitter:   false,
		}

		baseDelay := 100 * time.Millisecond
		delay := calculateDelay(baseDelay, config, &testError{"test error"})

		testutil.AssertEqual(t, config.MaxDelay, delay)
	})
}

func TestNewCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker("test", 5, 30*time.Second)

	testutil.AssertEqual(t, "test", cb.name)
	testutil.AssertEqual(t, 5, cb.maxFailures)
	testutil.AssertEqual(t, 30*time.Second, cb.resetTimeout)
	testutil.AssertEqual(t, 15*time.Second, cb.halfOpenTimeout)
	testutil.AssertEqual(t, CircuitClosed, cb.state)
	testutil.AssertEqual(t, 0, cb.failures)
}

func TestCircuitBreaker_Execute(t *testing.T) {
	t.Run("SuccessfulExecution", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 3, 100*time.Millisecond)

		executions := 0
		fn := func(ctx context.Context) error {
			executions++
			return nil
		}

		ctx := context.Background()
		err := cb.Execute(ctx, fn)

		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 1, executions)
		testutil.AssertEqual(t, CircuitClosed, cb.state)
		testutil.AssertEqual(t, 0, cb.failures)
	})

	t.Run("NonRetryableFailure", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 3, 100*time.Millisecond)

		executions := 0
		expectedErr := New(ErrorTypeBadRequest, "bad request")
		fn := func(ctx context.Context) error {
			executions++
			return expectedErr
		}

		ctx := context.Background()
		err := cb.Execute(ctx, fn)

		testutil.AssertEqual(t, expectedErr, err)
		testutil.AssertEqual(t, 1, executions)
		testutil.AssertEqual(t, CircuitClosed, cb.state)
		testutil.AssertEqual(t, 0, cb.failures) // Non-retryable errors don't count as failures
	})

	t.Run("RetryableFailuresOpenCircuit", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 3, 100*time.Millisecond)

		expectedErr := New(ErrorTypeInternal, "internal error")
		fn := func(ctx context.Context) error {
			return expectedErr
		}

		ctx := context.Background()

		// Execute multiple times to trigger circuit opening
		for i := 0; i < 3; i++ {
			err := cb.Execute(ctx, fn)
			testutil.AssertEqual(t, expectedErr, err)
		}

		testutil.AssertEqual(t, CircuitOpen, cb.state)
		testutil.AssertEqual(t, 3, cb.failures)

		// Next execution should fail immediately without calling function
		executions := 0
		quickFn := func(ctx context.Context) error {
			executions++
			return nil
		}

		err := cb.Execute(ctx, quickFn)
		testutil.AssertEqual(t, 0, executions) // Function should not be called

		ccErr, ok := err.(*CCProxyError)
		testutil.AssertTrue(t, ok)
		testutil.AssertEqual(t, ErrorTypeServiceUnavailable, ccErr.Type)
		testutil.AssertEqual(t, "Circuit breaker is open", ccErr.Message)
	})

	t.Run("HalfOpenSuccess", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 2, 50*time.Millisecond)

		// Trigger circuit opening
		failFn := func(ctx context.Context) error {
			return New(ErrorTypeInternal, "internal error")
		}

		ctx := context.Background()
		for i := 0; i < 2; i++ {
			cb.Execute(ctx, failFn)
		}

		testutil.AssertEqual(t, CircuitOpen, cb.state)

		// Wait for transition to half-open
		time.Sleep(60 * time.Millisecond)

		// Execute successful function
		successFn := func(ctx context.Context) error {
			return nil
		}

		err := cb.Execute(ctx, successFn)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, CircuitClosed, cb.state)
		testutil.AssertEqual(t, 0, cb.failures)
	})

	t.Run("HalfOpenFailure", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 2, 50*time.Millisecond)

		// Trigger circuit opening
		failFn := func(ctx context.Context) error {
			return New(ErrorTypeInternal, "internal error")
		}

		ctx := context.Background()
		for i := 0; i < 2; i++ {
			cb.Execute(ctx, failFn)
		}

		testutil.AssertEqual(t, CircuitOpen, cb.state)

		// Wait for transition to half-open
		time.Sleep(60 * time.Millisecond)

		// Execute failing function again
		err := cb.Execute(ctx, failFn)
		testutil.AssertEqual(t, false, err == nil)
		testutil.AssertEqual(t, CircuitOpen, cb.state) // Should go back to open
	})
}

func TestCircuitBreaker_GetState(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, 50*time.Millisecond)

	testutil.AssertEqual(t, "closed", cb.GetState())

	// Trigger opening
	failFn := func(ctx context.Context) error {
		return New(ErrorTypeInternal, "internal error")
	}

	ctx := context.Background()
	for i := 0; i < 2; i++ {
		cb.Execute(ctx, failFn)
	}

	testutil.AssertEqual(t, "open", cb.GetState())

	// Wait for half-open transition
	time.Sleep(60 * time.Millisecond)
	testutil.AssertEqual(t, "half-open", cb.GetState())
}

func TestCircuitBreaker_GetStats(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, 50*time.Millisecond)

	stats := cb.GetStats()

	testutil.AssertEqual(t, "test", stats["name"])
	testutil.AssertEqual(t, "closed", stats["state"])
	testutil.AssertEqual(t, 0, stats["failures"])
	testutil.AssertEqual(t, false, stats["last_failure"] == nil)
	testutil.AssertEqual(t, false, stats["state_since"] == nil)
}

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, 100*time.Millisecond)

	// Initial state
	testutil.AssertEqual(t, CircuitClosed, cb.getState())

	// Trigger failures to open circuit
	failFn := func(ctx context.Context) error {
		return New(ErrorTypeInternal, "internal error")
	}

	ctx := context.Background()
	for i := 0; i < 2; i++ {
		cb.Execute(ctx, failFn)
	}

	testutil.AssertEqual(t, CircuitOpen, cb.getState())

	// Wait for reset timeout to trigger half-open
	time.Sleep(110 * time.Millisecond)
	testutil.AssertEqual(t, CircuitHalfOpen, cb.getState())

	// Wait for half-open timeout to reset to closed
	time.Sleep(60 * time.Millisecond)
	testutil.AssertEqual(t, CircuitClosed, cb.getState())
	testutil.AssertEqual(t, 0, cb.failures)
}

func TestCircuitBreaker_OnSuccess(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, 100*time.Millisecond)

	// Set some failures first
	cb.failures = 5
	cb.state = CircuitOpen

	cb.onSuccess()

	testutil.AssertEqual(t, 0, cb.failures)
	testutil.AssertEqual(t, CircuitClosed, cb.state)
}

func TestCircuitBreaker_OnFailure(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, 100*time.Millisecond)

	// First failure
	cb.onFailure()
	testutil.AssertEqual(t, 1, cb.failures)
	testutil.AssertEqual(t, CircuitClosed, cb.state)

	// Second failure should open circuit
	cb.onFailure()
	testutil.AssertEqual(t, 2, cb.failures)
	testutil.AssertEqual(t, CircuitOpen, cb.state)
}
