package errors

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestComprehensiveErrorHandling tests all aspects of the error handling system
func TestComprehensiveErrorHandling(t *testing.T) {
	t.Run("Error Creation", testErrorCreation)
	t.Run("Error Wrapping", testErrorWrapping)
	t.Run("Error Metadata", testErrorMetadata)
	t.Run("HTTP Response", testHTTPResponse)
	t.Run("Error Type Mapping", testErrorTypeMapping)
	t.Run("Retry Logic", testRetryLogic)
	t.Run("Circuit Breaker", testCircuitBreaker)
	t.Run("Provider Error Handling", testProviderErrorHandling)
}

func testErrorCreation(t *testing.T) {
	tests := []struct {
		name       string
		errorType  ErrorType
		message    string
		wantStatus int
		wantRetry  bool
	}{
		{
			name:       "bad request error",
			errorType:  ErrorTypeBadRequest,
			message:    "Invalid input",
			wantStatus: http.StatusBadRequest,
			wantRetry:  false,
		},
		{
			name:       "internal error",
			errorType:  ErrorTypeInternal,
			message:    "Server error",
			wantStatus: http.StatusInternalServerError,
			wantRetry:  true,
		},
		{
			name:       "rate limit error",
			errorType:  ErrorTypeRateLimitError,
			message:    "Too many requests",
			wantStatus: http.StatusTooManyRequests,
			wantRetry:  true,
		},
		{
			name:       "gateway timeout",
			errorType:  ErrorTypeGatewayTimeout,
			message:    "Request timeout",
			wantStatus: http.StatusGatewayTimeout,
			wantRetry:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.errorType, tt.message)

			if err.Type != tt.errorType {
				t.Errorf("Expected type %s, got %s", tt.errorType, err.Type)
			}

			if err.Message != tt.message {
				t.Errorf("Expected message %q, got %q", tt.message, err.Message)
			}

			if err.StatusCode != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, err.StatusCode)
			}

			if err.Retryable != tt.wantRetry {
				t.Errorf("Expected retryable=%v, got %v", tt.wantRetry, err.Retryable)
			}

			if err.Timestamp.IsZero() {
				t.Error("Expected timestamp to be set")
			}
		})
	}
}

func testErrorWrapping(t *testing.T) {
	original := errors.New("original error")
	
	t.Run("simple wrap", func(t *testing.T) {
		wrapped := Wrap(original, ErrorTypeInternal, "Wrapped error")
		
		if wrapped.Unwrap() != original {
			t.Error("Expected wrapped error to contain original")
		}
		
		if !errors.Is(wrapped, original) {
			t.Error("Expected errors.Is to work with wrapped error")
		}
		
		errStr := wrapped.Error()
		if !strings.Contains(errStr, "original error") {
			t.Errorf("Expected error string to contain original error, got %s", errStr)
		}
	})
	
	t.Run("wrap CCProxyError", func(t *testing.T) {
		original := New(ErrorTypeBadRequest, "Bad input").
			WithCode("INVALID_FIELD").
			WithProvider("test-provider").
			WithRequestID("req-123")
		
		wrapped := Wrap(original, ErrorTypeValidationError, "Validation failed")
		
		// Should preserve metadata
		if wrapped.Code != original.Code {
			t.Errorf("Expected code %s, got %s", original.Code, wrapped.Code)
		}
		
		if wrapped.Provider != original.Provider {
			t.Errorf("Expected provider %s, got %s", original.Provider, wrapped.Provider)
		}
		
		if wrapped.RequestID != original.RequestID {
			t.Errorf("Expected request ID %s, got %s", original.RequestID, wrapped.RequestID)
		}
	})
	
	t.Run("wrap nil", func(t *testing.T) {
		wrapped := Wrap(nil, ErrorTypeInternal, "Should not wrap")
		if wrapped != nil {
			t.Error("Expected nil when wrapping nil error")
		}
	})
}

func testErrorMetadata(t *testing.T) {
	err := New(ErrorTypeProviderError, "Provider failed")
	
	t.Run("with code", func(t *testing.T) {
		err = err.WithCode("PROVIDER_UNAVAILABLE")
		if err.Code != "PROVIDER_UNAVAILABLE" {
			t.Errorf("Expected code PROVIDER_UNAVAILABLE, got %s", err.Code)
		}
	})
	
	t.Run("with provider", func(t *testing.T) {
		err = err.WithProvider("openai")
		if err.Provider != "openai" {
			t.Errorf("Expected provider openai, got %s", err.Provider)
		}
	})
	
	t.Run("with request ID", func(t *testing.T) {
		err = err.WithRequestID("req-456")
		if err.RequestID != "req-456" {
			t.Errorf("Expected request ID req-456, got %s", err.RequestID)
		}
	})
	
	t.Run("with details", func(t *testing.T) {
		details := map[string]interface{}{
			"model": "gpt-4",
			"tokens": 1000,
		}
		err = err.WithDetails(details)
		
		if err.Details["model"] != "gpt-4" {
			t.Error("Expected model detail to be preserved")
		}
		
		if err.Details["tokens"] != 1000 {
			t.Error("Expected tokens detail to be preserved")
		}
	})
	
	t.Run("with retry after", func(t *testing.T) {
		duration := 30 * time.Second
		err = err.WithRetryAfter(duration)
		
		if !err.Retryable {
			t.Error("Expected error to be retryable after WithRetryAfter")
		}
		
		if err.RetryAfter == nil || *err.RetryAfter != duration {
			t.Error("Expected retry after duration to be set")
		}
	})
}

func testHTTPResponse(t *testing.T) {
	tests := []struct {
		name         string
		err          *CCProxyError
		wantStatus   int
		wantHeaders  map[string]string
		wantBodyContains []string
	}{
		{
			name: "basic error",
			err: New(ErrorTypeBadRequest, "Invalid request"),
			wantStatus: http.StatusBadRequest,
			wantBodyContains: []string{
				`"type":"bad_request"`,
				`"message":"Invalid request"`,
			},
		},
		{
			name: "error with retry after",
			err: New(ErrorTypeRateLimitError, "Rate limited").
				WithRetryAfter(60 * time.Second),
			wantStatus: http.StatusTooManyRequests,
			wantHeaders: map[string]string{
				"Retry-After": "60",
			},
			wantBodyContains: []string{
				`"type":"rate_limit_error"`,
			},
		},
		{
			name: "error with all metadata",
			err: New(ErrorTypeProviderError, "Provider failed").
				WithCode("PROVIDER_DOWN").
				WithProvider("anthropic").
				WithRequestID("req-789").
				WithDetails(map[string]interface{}{
					"region": "us-east-1",
				}),
			wantStatus: http.StatusBadGateway,
			wantBodyContains: []string{
				`"code":"PROVIDER_DOWN"`,
				`"provider":"anthropic"`,
				`"request_id":"req-789"`,
				`"region":"us-east-1"`,
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tt.err.WriteHTTPResponse(w)
			
			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
			
			if w.Header().Get("Content-Type") != "application/json" {
				t.Error("Expected Content-Type to be application/json")
			}
			
			for header, value := range tt.wantHeaders {
				if w.Header().Get(header) != value {
					t.Errorf("Expected header %s=%s, got %s", header, value, w.Header().Get(header))
				}
			}
			
			body := w.Body.String()
			for _, expected := range tt.wantBodyContains {
				if !strings.Contains(body, expected) {
					t.Errorf("Expected body to contain %s, got %s", expected, body)
				}
			}
		})
	}
}

func testErrorTypeMapping(t *testing.T) {
	tests := []struct {
		statusCode int
		wantType   ErrorType
	}{
		{http.StatusBadRequest, ErrorTypeBadRequest},
		{http.StatusUnauthorized, ErrorTypeUnauthorized},
		{http.StatusForbidden, ErrorTypeForbidden},
		{http.StatusNotFound, ErrorTypeNotFound},
		{http.StatusTooManyRequests, ErrorTypeTooManyRequests},
		{http.StatusInternalServerError, ErrorTypeInternal},
		{http.StatusBadGateway, ErrorTypeBadGateway},
		{http.StatusServiceUnavailable, ErrorTypeServiceUnavailable},
		{http.StatusGatewayTimeout, ErrorTypeGatewayTimeout},
		{499, ErrorTypeBadRequest}, // Other 4xx
		{599, ErrorTypeInternal},    // Other 5xx
	}
	
	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.statusCode), func(t *testing.T) {
			gotType := getErrorTypeFromStatusCode(tt.statusCode)
			if gotType != tt.wantType {
				t.Errorf("Expected type %s for status %d, got %s", tt.wantType, tt.statusCode, gotType)
			}
		})
	}
}

func testRetryLogic(t *testing.T) {
	t.Run("successful retry", func(t *testing.T) {
		attempts := 0
		fn := func(ctx context.Context) error {
			attempts++
			if attempts < 3 {
				return New(ErrorTypeGatewayTimeout, "Timeout")
			}
			return nil
		}
		
		config := &RetryConfig{
			MaxAttempts:  5,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     100 * time.Millisecond,
			Multiplier:   2.0,
			Jitter:       false,
		}
		
		err := RetryWithConfig(context.Background(), config, fn)
		if err != nil {
			t.Errorf("Expected success after retry, got %v", err)
		}
		
		if attempts != 3 {
			t.Errorf("Expected 3 attempts, got %d", attempts)
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
	
	t.Run("retry with backoff timing", func(t *testing.T) {
		attempts := 0
		var delays []time.Duration
		lastTime := time.Now()
		
		fn := func(ctx context.Context) error {
			attempts++
			if attempts > 1 {
				delays = append(delays, time.Since(lastTime))
			}
			lastTime = time.Now()
			
			if attempts < 3 {
				return New(ErrorTypeServiceUnavailable, "Service down")
			}
			return nil
		}
		
		config := &RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 50 * time.Millisecond,
			MaxDelay:     200 * time.Millisecond,
			Multiplier:   2.0,
			Jitter:       false,
		}
		
		err := RetryWithConfig(context.Background(), config, fn)
		if err != nil {
			t.Errorf("Expected success, got %v", err)
		}
		
		// Check delays are approximately correct
		if len(delays) != 2 {
			t.Fatalf("Expected 2 delays, got %d", len(delays))
		}
		
		// First retry should be ~50ms
		if delays[0] < 45*time.Millisecond || delays[0] > 55*time.Millisecond {
			t.Errorf("Expected first delay ~50ms, got %v", delays[0])
		}
		
		// Second retry should be ~100ms (50ms * 2)
		if delays[1] < 95*time.Millisecond || delays[1] > 105*time.Millisecond {
			t.Errorf("Expected second delay ~100ms, got %v", delays[1])
		}
	})
}

func testCircuitBreaker(t *testing.T) {
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
		
		// Next request should be rejected immediately
		start := time.Now()
		err := cb.Execute(context.Background(), func(ctx context.Context) error {
			t.Error("Function should not be called when circuit is open")
			return nil
		})
		duration := time.Since(start)
		
		if err == nil {
			t.Error("Expected error when circuit is open")
		}
		
		// Should return immediately without calling function
		if duration > 10*time.Millisecond {
			t.Errorf("Circuit breaker should reject immediately, took %v", duration)
		}
		
		var ccErr *CCProxyError
		if errors.As(err, &ccErr) {
			if ccErr.Type != ErrorTypeServiceUnavailable {
				t.Errorf("Expected service unavailable error, got %s", ccErr.Type)
			}
			if !strings.Contains(ccErr.Message, "Circuit breaker is open") {
				t.Errorf("Expected circuit breaker message, got %s", ccErr.Message)
			}
		}
	})
	
	t.Run("circuit recovers after timeout", func(t *testing.T) {
		cb := NewCircuitBreaker("test-recovery", 1, 100*time.Millisecond)
		
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
		
		// Circuit should remain closed for subsequent requests
		for i := 0; i < 3; i++ {
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
	
	t.Run("non-retryable errors don't trip circuit", func(t *testing.T) {
		cb := NewCircuitBreaker("test-non-retryable", 2, 100*time.Millisecond)
		
		// Multiple non-retryable errors shouldn't trip the circuit
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
		
		stats := cb.GetStats()
		if stats["failures"].(int) != 0 {
			t.Errorf("Expected 0 failures for non-retryable errors, got %d", stats["failures"].(int))
		}
	})
}

func testProviderErrorHandling(t *testing.T) {
	t.Run("from provider response", func(t *testing.T) {
		tests := []struct {
			name       string
			statusCode int
			body       []byte
			provider   string
			wantType   ErrorType
			wantMsg    string
		}{
			{
				name:       "anthropic error format",
				statusCode: 400,
				body:       []byte(`{"error":{"type":"invalid_request_error","message":"Invalid API key","code":"invalid_api_key"}}`),
				provider:   "anthropic",
				wantType:   ErrorTypeBadRequest,
				wantMsg:    "Invalid API key",
			},
			{
				name:       "rate limit error",
				statusCode: 429,
				body:       []byte(`{"error":{"message":"Rate limit exceeded"}}`),
				provider:   "openai",
				wantType:   ErrorTypeTooManyRequests,
				wantMsg:    "Rate limit exceeded",
			},
			{
				name:       "no error body",
				statusCode: 503,
				body:       []byte{},
				provider:   "groq",
				wantType:   ErrorTypeServiceUnavailable,
				wantMsg:    "Provider returned status 503",
			},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := FromProviderResponse(tt.statusCode, tt.body, tt.provider)
				
				if err.Type != tt.wantType {
					t.Errorf("Expected type %s, got %s", tt.wantType, err.Type)
				}
				
				if err.Message != tt.wantMsg {
					t.Errorf("Expected message %q, got %q", tt.wantMsg, err.Message)
				}
				
				if err.Provider != tt.provider {
					t.Errorf("Expected provider %s, got %s", tt.provider, err.Provider)
				}
				
				if err.StatusCode != tt.statusCode {
					t.Errorf("Expected status code %d, got %d", tt.statusCode, err.StatusCode)
				}
			})
		}
	})
	
	t.Run("common error constructors", func(t *testing.T) {
		tests := []struct {
			name      string
			fn        func() *CCProxyError
			wantType  ErrorType
			wantMsg   string
			wantRetry bool
		}{
			{
				name:      "bad request",
				fn:        func() *CCProxyError { return ErrBadRequest("Invalid input") },
				wantType:  ErrorTypeBadRequest,
				wantMsg:   "Invalid input",
				wantRetry: false,
			},
			{
				name:      "unauthorized",
				fn:        func() *CCProxyError { return ErrUnauthorized("Invalid token") },
				wantType:  ErrorTypeUnauthorized,
				wantMsg:   "Invalid token",
				wantRetry: false,
			},
			{
				name:      "not found",
				fn:        func() *CCProxyError { return ErrNotFound("user") },
				wantType:  ErrorTypeNotFound,
				wantMsg:   "user not found",
				wantRetry: false,
			},
			{
				name:      "internal",
				fn:        func() *CCProxyError { return ErrInternal("Database error") },
				wantType:  ErrorTypeInternal,
				wantMsg:   "Database error",
				wantRetry: true,
			},
			{
				name:      "rate limit",
				fn:        func() *CCProxyError { return ErrRateLimit(30 * time.Second) },
				wantType:  ErrorTypeRateLimitError,
				wantMsg:   "Rate limit exceeded",
				wantRetry: true,
			},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.fn()
				
				if err.Type != tt.wantType {
					t.Errorf("Expected type %s, got %s", tt.wantType, err.Type)
				}
				
				if err.Message != tt.wantMsg {
					t.Errorf("Expected message %q, got %q", tt.wantMsg, err.Message)
				}
				
				if err.Retryable != tt.wantRetry {
					t.Errorf("Expected retryable=%v, got %v", tt.wantRetry, err.Retryable)
				}
			})
		}
	})
}