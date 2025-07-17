// Package common provides comprehensive error scenario tests for all providers
package common

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ccproxy/internal/constants"
	"ccproxy/internal/models"
)

// TestErrorScenarios provides comprehensive error scenario testing
func TestErrorScenarios(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		responseBody  string
		expectedError string
		shouldRetry   bool
		errorType     string
	}{
		{
			name:          "Unauthorized - Invalid API Key",
			statusCode:    http.StatusUnauthorized,
			responseBody:  `{"error": {"message": "Invalid API key", "type": "invalid_request_error"}}`,
			expectedError: "HTTP 401: 401 Unauthorized",
			shouldRetry:   false,
			errorType:     "auth_error",
		},
		{
			name:          "Rate Limited",
			statusCode:    http.StatusTooManyRequests,
			responseBody:  `{"error": {"message": "Rate limit exceeded", "type": "rate_limit_error"}}`,
			expectedError: "HTTP 429: 429 Too Many Requests",
			shouldRetry:   true,
			errorType:     "rate_limit_error",
		},
		{
			name:          "Internal Server Error",
			statusCode:    http.StatusInternalServerError,
			responseBody:  `{"error": {"message": "Internal server error", "type": "server_error"}}`,
			expectedError: "HTTP 500: 500 Internal Server Error",
			shouldRetry:   true,
			errorType:     "server_error",
		},
		{
			name:          "Service Unavailable",
			statusCode:    http.StatusServiceUnavailable,
			responseBody:  `{"error": {"message": "Service temporarily unavailable", "type": "server_error"}}`,
			expectedError: "HTTP 503: 503 Service Unavailable",
			shouldRetry:   true,
			errorType:     "server_error",
		},
		{
			name:          "Bad Gateway",
			statusCode:    http.StatusBadGateway,
			responseBody:  `{"error": {"message": "Bad gateway", "type": "server_error"}}`,
			expectedError: "HTTP 502: 502 Bad Gateway",
			shouldRetry:   true,
			errorType:     "server_error",
		},
		{
			name:          "Gateway Timeout",
			statusCode:    http.StatusGatewayTimeout,
			responseBody:  `{"error": {"message": "Gateway timeout", "type": "server_error"}}`,
			expectedError: "HTTP 504: 504 Gateway Timeout",
			shouldRetry:   true,
			errorType:     "server_error",
		},
		{
			name:          "Bad Request - Invalid Input",
			statusCode:    http.StatusBadRequest,
			responseBody:  `{"error": {"message": "Invalid request format", "type": "invalid_request_error"}}`,
			expectedError: "HTTP 400: 400 Bad Request",
			shouldRetry:   false,
			errorType:     "client_error",
		},
		{
			name:          "Forbidden - Insufficient Permissions",
			statusCode:    http.StatusForbidden,
			responseBody:  `{"error": {"message": "Insufficient permissions", "type": "permission_error"}}`,
			expectedError: "HTTP 403: 403 Forbidden",
			shouldRetry:   false,
			errorType:     "permission_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				//nolint:errcheck // Error is intentionally ignored in test server
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Test HTTP error creation
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Status:     http.StatusText(tt.statusCode),
			}

			err := NewHTTPError("test_provider", resp, nil)
			if err == nil {
				t.Fatalf("Expected error but got nil")
			}

			var providerErr *ProviderError
			if !errors.As(err, &providerErr) {
				t.Fatalf("Expected ProviderError but got %T", err)
			}

			// Verify error properties
			if providerErr.Code != tt.statusCode {
				t.Errorf("Expected status code %d but got %d", tt.statusCode, providerErr.Code)
			}

			if providerErr.IsRetryable() != tt.shouldRetry {
				t.Errorf("Expected retryable=%v but got %v", tt.shouldRetry, providerErr.IsRetryable())
			}

			if providerErr.Provider != "test_provider" {
				t.Errorf("Expected provider='test_provider' but got '%s'", providerErr.Provider)
			}
		})
	}
}

// TestNetworkTimeoutScenarios tests various network timeout scenarios
func TestNetworkTimeoutScenarios(t *testing.T) {
	tests := []struct {
		name          string
		serverDelay   time.Duration
		clientTimeout time.Duration
		expectTimeout bool
	}{
		{
			name:          "Fast Response - No Timeout",
			serverDelay:   100 * time.Millisecond,
			clientTimeout: 1 * time.Second,
			expectTimeout: false,
		},
		{
			name:          "Slow Response - Client Timeout",
			serverDelay:   2 * time.Second,
			clientTimeout: 500 * time.Millisecond,
			expectTimeout: true,
		},
		{
			name:          "Very Slow Response - Timeout",
			serverDelay:   5 * time.Second,
			clientTimeout: 1 * time.Second,
			expectTimeout: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server with delay
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(tt.serverDelay)
				w.WriteHeader(http.StatusOK)
				//nolint:errcheck // Error is intentionally ignored in test server
				_, _ = w.Write([]byte(`{"test": "response"}`))
			}))
			defer server.Close()

			// Create HTTP client with timeout
			client := NewConfiguredHTTPClient(tt.clientTimeout)

			// Make request
			ctx := context.Background()
			req, err := http.NewRequestWithContext(ctx, "POST", server.URL, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			resp, err := client.Do(req)
			if resp != nil {
				//nolint:errcheck // Error is intentionally ignored
				defer func() { _ = resp.Body.Close() }()
			}

			if tt.expectTimeout {
				if err == nil {
					t.Errorf("Expected timeout error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestConfigurationErrors tests provider configuration validation
func TestConfigurationErrors(t *testing.T) {
	tests := []struct {
		name        string
		provider    string
		field       string
		message     string
		expectError bool
	}{
		{
			name:        "Missing API Key",
			provider:    "openai",
			field:       "OPENAI_API_KEY",
			message:     "API key is required",
			expectError: true,
		},
		{
			name:        "Missing Base URL",
			provider:    "groq",
			field:       "GROQ_BASE_URL",
			message:     "base URL is required",
			expectError: true,
		},
		{
			name:        "Missing Model",
			provider:    "mistral",
			field:       "MISTRAL_MODEL",
			message:     "model is required",
			expectError: true,
		},
		{
			name:        "Invalid Max Tokens",
			provider:    "gemini",
			field:       "GEMINI_MAX_TOKENS",
			message:     "max tokens must be greater than 0",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewConfigError(tt.provider, tt.field, tt.message)
			if err == nil && tt.expectError {
				t.Fatalf("Expected error but got nil")
			}

			if err != nil {
				var providerErr *ProviderError
				if !errors.As(err, &providerErr) {
					t.Fatalf("Expected ProviderError but got %T", err)
				}

				if providerErr.Provider != tt.provider {
					t.Errorf("Expected provider '%s' but got '%s'", tt.provider, providerErr.Provider)
				}

				if providerErr.IsRetryable() {
					t.Errorf("Config errors should not be retryable")
				}
			}
		})
	}
}

// TestMalformedResponseScenarios tests handling of malformed API responses
func TestMalformedResponseScenarios(t *testing.T) {
	tests := []struct {
		name         string
		responseBody string
		expectError  bool
	}{
		{
			name:         "Invalid JSON",
			responseBody: `{"invalid": json}`,
			expectError:  true,
		},
		{
			name:         "Empty Response",
			responseBody: ``,
			expectError:  true,
		},
		{
			name:         "Null Response",
			responseBody: `null`,
			expectError:  false, // null is valid JSON
		},
		{
			name:         "Incomplete JSON",
			responseBody: `{"choices": [{"message":`,
			expectError:  true,
		},
		{
			name:         "Wrong Schema",
			responseBody: `{"completely": "different", "schema": true}`,
			expectError:  false, // Valid JSON, will result in zero values
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp models.ChatCompletionResponse
			err := UnmarshalJSONResponse([]byte(tt.responseBody), &resp, "test_provider")

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if err != nil {
				var providerErr *ProviderError
				if !errors.As(err, &providerErr) {
					t.Errorf("Expected ProviderError but got %T", err)
				} else if providerErr.Provider != "test_provider" {
					t.Errorf("Expected provider 'test_provider' but got '%s'", providerErr.Provider)
				}
			}
		})
	}
}

// TestConcurrentErrorScenarios tests error handling under concurrent load
func TestConcurrentErrorScenarios(t *testing.T) {
	const numGoroutines = 50
	const numRequestsPerGoroutine = 10

	// Create mock server that randomly returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate various error conditions
		switch r.URL.Query().Get("error_type") {
		case "timeout":
			time.Sleep(2 * time.Second)
		case "rate_limit":
			w.WriteHeader(http.StatusTooManyRequests)
		case "server_error":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusOK)
			//nolint:errcheck // Error is intentionally ignored in test server
			_, _ = w.Write([]byte(`{"test": "success"}`))
		}
	}))
	defer server.Close()

	client := NewConfiguredHTTPClient(1 * time.Second)
	errorChan := make(chan error, numGoroutines*numRequestsPerGoroutine)

	// Launch concurrent requests
	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numRequestsPerGoroutine; j++ {
				errorType := []string{"success", "rate_limit", "server_error", "timeout"}[j%4]
				url := server.URL + "?error_type=" + errorType

				req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
				if err != nil {
					errorChan <- err
					continue
				}

				resp, err := client.Do(req)
				if resp != nil {
					//nolint:errcheck // Error is intentionally ignored
					_ = resp.Body.Close()
				}
				errorChan <- err
			}
		}()
	}

	// Collect errors
	errorCount := 0
	timeoutCount := 0
	successCount := 0

	for i := 0; i < numGoroutines*numRequestsPerGoroutine; i++ {
		err := <-errorChan
		if err != nil {
			errorCount++
			if errors.Is(err, context.DeadlineExceeded) {
				timeoutCount++
			}
		} else {
			successCount++
		}
	}

	t.Logf("Concurrent test results: %d errors, %d timeouts, %d successes",
		errorCount, timeoutCount, successCount)

	// Verify we got a mix of results as expected
	if errorCount == 0 {
		t.Error("Expected some errors in concurrent test")
	}
	if successCount == 0 {
		t.Error("Expected some successes in concurrent test")
	}
}

// TestErrorPropagation tests that errors are properly propagated through the system
func TestErrorPropagation(t *testing.T) {
	// Test that wrapped errors maintain original error information
	originalErr := errors.New("network connection failed")
	providerErr := NewProviderError("test_provider", "connection failed", originalErr)

	// Test error unwrapping
	if !errors.Is(providerErr, originalErr) {
		t.Error("Provider error should wrap original error")
	}

	unwrapped := errors.Unwrap(providerErr)
	if unwrapped != originalErr {
		t.Error("Unwrapped error should be the original error")
	}

	// Test error chain
	wrappedAgain := NewProviderError("outer_provider", "outer error", providerErr)
	if !errors.Is(wrappedAgain, originalErr) {
		t.Error("Should be able to find original error in error chain")
	}
}

// TestSpecialErrorConditions tests edge cases and special error conditions
func TestSpecialErrorConditions(t *testing.T) {
	t.Run("Nil Response", func(t *testing.T) {
		err := NewHTTPError("test", nil, errors.New("nil response"))
		if err == nil {
			t.Error("Should handle nil response gracefully")
		}
	})

	t.Run("Empty Provider Name", func(t *testing.T) {
		err := NewProviderError("", "test error", nil)
		var providerErr *ProviderError
		if errors.As(err, &providerErr) && providerErr.Provider != "" {
			t.Error("Should handle empty provider name")
		}
	})

	t.Run("Request ID Context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), constants.RequestIDKey, "test-request-123")
		requestID := GetRequestID(ctx)
		if requestID != "test-request-123" {
			t.Errorf("Expected request ID 'test-request-123' but got '%s'", requestID)
		}

		// Test missing request ID
		emptyCtx := context.Background()
		defaultID := GetRequestID(emptyCtx)
		if defaultID != constants.DefaultRequestID {
			t.Errorf("Expected default request ID '%s' but got '%s'", constants.DefaultRequestID, defaultID)
		}
	})

	t.Run("Invalid Context Value Type", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), constants.RequestIDKey, 123) // Wrong type
		requestID := GetRequestID(ctx)
		if requestID != constants.DefaultRequestID {
			t.Errorf("Should return default ID for invalid context value type")
		}
	})
}
