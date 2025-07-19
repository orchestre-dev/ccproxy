package errors

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRetryWithConfig(t *testing.T) {
	t.Run("successful on first attempt", func(t *testing.T) {
		attempts := 0
		fn := func(ctx context.Context) error {
			attempts++
			return nil
		}
		
		config := &RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     100 * time.Millisecond,
			Multiplier:   2.0,
		}
		
		err := RetryWithConfig(context.Background(), config, fn)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if attempts != 1 {
			t.Errorf("Expected 1 attempt, got %d", attempts)
		}
	})
	
	t.Run("successful after retries", func(t *testing.T) {
		attempts := 0
		fn := func(ctx context.Context) error {
			attempts++
			if attempts < 3 {
				return New(ErrorTypeGatewayTimeout, "Timeout")
			}
			return nil
		}
		
		config := &RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     100 * time.Millisecond,
			Multiplier:   2.0,
		}
		
		err := RetryWithConfig(context.Background(), config, fn)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if attempts != 3 {
			t.Errorf("Expected 3 attempts, got %d", attempts)
		}
	})
	
	t.Run("max attempts exceeded", func(t *testing.T) {
		attempts := 0
		fn := func(ctx context.Context) error {
			attempts++
			return New(ErrorTypeServiceUnavailable, "Service down")
		}
		
		config := &RetryConfig{
			MaxAttempts:  2,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     100 * time.Millisecond,
			Multiplier:   2.0,
		}
		
		err := RetryWithConfig(context.Background(), config, fn)
		if err == nil {
			t.Error("Expected error")
		}
		if attempts != 2 {
			t.Errorf("Expected 2 attempts, got %d", attempts)
		}
		
		// Check that it's wrapped properly
		var ccErr *CCProxyError
		if errors.As(err, &ccErr) {
			if !strings.Contains(ccErr.Error(), "Max retry attempts exceeded") {
				t.Error("Expected max retry error message")
			}
		}
	})
	
	t.Run("non-retryable error", func(t *testing.T) {
		attempts := 0
		fn := func(ctx context.Context) error {
			attempts++
			return New(ErrorTypeBadRequest, "Invalid input")
		}
		
		config := DefaultRetryConfig()
		
		err := RetryWithConfig(context.Background(), config, fn)
		if err == nil {
			t.Error("Expected error")
		}
		if attempts != 1 {
			t.Errorf("Expected 1 attempt for non-retryable error, got %d", attempts)
		}
	})
	
	t.Run("context cancelled", func(t *testing.T) {
		attempts := 0
		fn := func(ctx context.Context) error {
			attempts++
			return New(ErrorTypeGatewayTimeout, "Timeout")
		}
		
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		config := &RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     100 * time.Millisecond,
			Multiplier:   2.0,
		}
		
		err := RetryWithConfig(ctx, config, fn)
		if err == nil {
			t.Error("Expected error")
		}
		if attempts != 0 {
			t.Errorf("Expected 0 attempts when context cancelled, got %d", attempts)
		}
	})
	
	t.Run("context cancelled during retry", func(t *testing.T) {
		attempts := 0
		fn := func(ctx context.Context) error {
			attempts++
			return New(ErrorTypeGatewayTimeout, "Timeout")
		}
		
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		
		config := &RetryConfig{
			MaxAttempts:  5,
			InitialDelay: 100 * time.Millisecond, // Longer than context timeout
			MaxDelay:     1 * time.Second,
			Multiplier:   2.0,
		}
		
		err := RetryWithConfig(ctx, config, fn)
		if err == nil {
			t.Error("Expected error")
		}
		if attempts > 2 {
			t.Errorf("Expected at most 2 attempts before context timeout, got %d", attempts)
		}
	})
	
	t.Run("retry with specific error types", func(t *testing.T) {
		attempts := 0
		fn := func(ctx context.Context) error {
			attempts++
			if attempts == 1 {
				return New(ErrorTypeRateLimitError, "Rate limited")
			}
			return New(ErrorTypeBadRequest, "Bad request")
		}
		
		config := &RetryConfig{
			MaxAttempts:     3,
			InitialDelay:    10 * time.Millisecond,
			MaxDelay:        100 * time.Millisecond,
			Multiplier:      2.0,
			RetryableErrors: []ErrorType{ErrorTypeRateLimitError},
		}
		
		err := RetryWithConfig(context.Background(), config, fn)
		if err == nil {
			t.Error("Expected error")
		}
		
		// Should retry once for rate limit, then stop on bad request
		if attempts != 2 {
			t.Errorf("Expected 2 attempts, got %d", attempts)
		}
		
		var ccErr *CCProxyError
		if errors.As(err, &ccErr) {
			if ccErr.Type != ErrorTypeBadRequest {
				t.Errorf("Expected bad request error, got %s", ccErr.Type)
			}
		}
	})
}

func TestRetry(t *testing.T) {
	attempts := 0
	fn := func(ctx context.Context) error {
		attempts++
		if attempts < 2 {
			return New(ErrorTypeGatewayTimeout, "Timeout")
		}
		return nil
	}
	
	err := Retry(context.Background(), fn)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetryWithBackoff(t *testing.T) {
	attempts := 0
	fn := func(ctx context.Context) error {
		attempts++
		if attempts < 2 {
			return New(ErrorTypeServiceUnavailable, "Service down")
		}
		return nil
	}
	
	start := time.Now()
	err := RetryWithBackoff(context.Background(), 3, 50*time.Millisecond, fn)
	duration := time.Since(start)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
	
	// Should have delayed at least 37.5ms (50ms - 25% jitter)
	minDelay := time.Duration(float64(50*time.Millisecond) * 0.75)
	if duration < minDelay {
		t.Errorf("Expected delay of at least %v, got %v", minDelay, duration)
	}
}

func TestCalculateDelay(t *testing.T) {
	config := &RetryConfig{
		MaxDelay: 30 * time.Second,
		Jitter:   false,
	}
	
	t.Run("basic delay", func(t *testing.T) {
		delay := calculateDelay(5*time.Second, config, errors.New("error"))
		if delay != 5*time.Second {
			t.Errorf("Expected 5s delay, got %v", delay)
		}
	})
	
	t.Run("with retry after", func(t *testing.T) {
		err := New(ErrorTypeRateLimitError, "Rate limited").WithRetryAfter(10 * time.Second)
		delay := calculateDelay(5*time.Second, config, err)
		if delay != 10*time.Second {
			t.Errorf("Expected 10s delay from retry after, got %v", delay)
		}
	})
	
	t.Run("max delay cap", func(t *testing.T) {
		delay := calculateDelay(60*time.Second, config, errors.New("error"))
		if delay != 30*time.Second {
			t.Errorf("Expected max delay of 30s, got %v", delay)
		}
	})
	
	t.Run("with jitter", func(t *testing.T) {
		config.Jitter = true
		baseDelay := 1 * time.Second
		
		// Run multiple times to test jitter
		delays := make([]time.Duration, 10)
		for i := 0; i < 10; i++ {
			delays[i] = calculateDelay(baseDelay, config, errors.New("error"))
		}
		
		// Check that we get some variation
		allSame := true
		for i := 1; i < len(delays); i++ {
			if delays[i] != delays[0] {
				allSame = false
				break
			}
		}
		
		if allSame {
			t.Error("Expected jitter to produce different delays")
		}
		
		// Check that delays are within expected range (±25%)
		for _, d := range delays {
			if d < 750*time.Millisecond || d > 1250*time.Millisecond {
				t.Errorf("Delay %v outside expected jitter range", d)
			}
		}
	})
}

func TestCircuitBreaker(t *testing.T) {
	t.Run("successful requests", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 3, 1*time.Second)
		
		for i := 0; i < 5; i++ {
			err := cb.Execute(context.Background(), func(ctx context.Context) error {
				return nil
			})
			if err != nil {
				t.Errorf("Expected no error on attempt %d, got %v", i+1, err)
			}
		}
		
		if cb.GetState() != "closed" {
			t.Errorf("Expected circuit to remain closed, got %s", cb.GetState())
		}
	})
	
	t.Run("circuit opens after failures", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 2, 100*time.Millisecond)
		
		// First two failures should trip the circuit
		for i := 0; i < 2; i++ {
			err := cb.Execute(context.Background(), func(ctx context.Context) error {
				return New(ErrorTypeServiceUnavailable, "Service down")
			})
			if err == nil {
				t.Error("Expected error")
			}
		}
		
		if cb.GetState() != "open" {
			t.Errorf("Expected circuit to be open, got %s", cb.GetState())
		}
		
		// Next request should be rejected
		err := cb.Execute(context.Background(), func(ctx context.Context) error {
			t.Error("Function should not be called when circuit is open")
			return nil
		})
		
		if err == nil {
			t.Error("Expected error when circuit is open")
		}
		
		var ccErr *CCProxyError
		if errors.As(err, &ccErr) {
			if ccErr.Type != ErrorTypeServiceUnavailable {
				t.Errorf("Expected service unavailable error, got %s", ccErr.Type)
			}
		}
	})
	
	t.Run("circuit transitions to half-open", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 1, 100*time.Millisecond)
		
		// Trip the circuit
		cb.Execute(context.Background(), func(ctx context.Context) error {
			return New(ErrorTypeGatewayTimeout, "Timeout")
		})
		
		if cb.GetState() != "open" {
			t.Fatal("Expected circuit to be open")
		}
		
		// Wait for reset timeout
		time.Sleep(150 * time.Millisecond)
		
		// Should be half-open now
		if cb.GetState() != "half-open" {
			t.Errorf("Expected circuit to be half-open, got %s", cb.GetState())
		}
		
		// Successful request should close the circuit
		err := cb.Execute(context.Background(), func(ctx context.Context) error {
			return nil
		})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		if cb.GetState() != "closed" {
			t.Errorf("Expected circuit to be closed after success, got %s", cb.GetState())
		}
	})
	
	t.Run("half-open failure reopens circuit", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 1, 100*time.Millisecond)
		
		// Trip the circuit
		cb.Execute(context.Background(), func(ctx context.Context) error {
			return New(ErrorTypeGatewayTimeout, "Timeout")
		})
		
		// Wait for half-open
		time.Sleep(150 * time.Millisecond)
		
		// Failure in half-open should reopen
		err := cb.Execute(context.Background(), func(ctx context.Context) error {
			return New(ErrorTypeGatewayTimeout, "Still timing out")
		})
		if err == nil {
			t.Error("Expected error")
		}
		
		if cb.GetState() != "open" {
			t.Errorf("Expected circuit to be open after half-open failure, got %s", cb.GetState())
		}
	})
	
	t.Run("non-retryable errors don't count", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 2, 100*time.Millisecond)
		
		// Non-retryable errors shouldn't trip the circuit
		for i := 0; i < 5; i++ {
			err := cb.Execute(context.Background(), func(ctx context.Context) error {
				return New(ErrorTypeBadRequest, "Bad request")
			})
			if err == nil {
				t.Error("Expected error")
			}
		}
		
		if cb.GetState() != "closed" {
			t.Errorf("Expected circuit to remain closed for non-retryable errors, got %s", cb.GetState())
		}
	})
	
	t.Run("get stats", func(t *testing.T) {
		cb := NewCircuitBreaker("test-stats", 2, 100*time.Millisecond)
		
		// Cause a failure
		cb.Execute(context.Background(), func(ctx context.Context) error {
			return New(ErrorTypeGatewayTimeout, "Timeout")
		})
		
		stats := cb.GetStats()
		
		if stats["name"] != "test-stats" {
			t.Error("Expected name in stats")
		}
		if stats["state"] != "closed" {
			t.Error("Expected state in stats")
		}
		if stats["failures"].(int) != 1 {
			t.Error("Expected failure count in stats")
		}
	})
}

func TestConcurrentRetry(t *testing.T) {
	// Test that retry is thread-safe
	// Each goroutine has its own counter
	config := &RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}
	
	// Run multiple retries concurrently
	done := make(chan struct {
		err      error
		attempts int
	}, 5)
	
	for i := 0; i < 5; i++ {
		go func() {
			attempts := 0
			fn := func(ctx context.Context) error {
				attempts++
				if attempts < 2 {
					return New(ErrorTypeGatewayTimeout, "Timeout")
				}
				return nil
			}
			
			err := RetryWithConfig(context.Background(), config, fn)
			done <- struct {
				err      error
				attempts int
			}{err, attempts}
		}()
	}
	
	// Collect results
	totalAttempts := 0
	for i := 0; i < 5; i++ {
		result := <-done
		if result.err != nil {
			t.Errorf("Expected no error on goroutine %d, got %v", i, result.err)
		}
		if result.attempts != 2 {
			t.Errorf("Expected 2 attempts on goroutine %d, got %d", i, result.attempts)
		}
		totalAttempts += result.attempts
	}
	
	// Each goroutine should have made 2 attempts, so total 10
	if totalAttempts != 10 {
		t.Errorf("Expected 10 total attempts, got %d", totalAttempts)
	}
}