package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	err := New(ErrorTypeBadRequest, "Invalid input")
	
	if err.Type != ErrorTypeBadRequest {
		t.Errorf("Expected type %s, got %s", ErrorTypeBadRequest, err.Type)
	}
	
	if err.Message != "Invalid input" {
		t.Errorf("Expected message 'Invalid input', got %s", err.Message)
	}
	
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, err.StatusCode)
	}
	
	if err.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
	
	if err.Retryable {
		t.Error("Bad request should not be retryable")
	}
}

func TestNewf(t *testing.T) {
	err := Newf(ErrorTypeNotFound, "Resource %s not found", "user")
	
	if err.Message != "Resource user not found" {
		t.Errorf("Expected formatted message, got %s", err.Message)
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("connection timeout")
	wrapped := Wrap(originalErr, ErrorTypeGatewayTimeout, "Provider timeout")
	
	if wrapped.Type != ErrorTypeGatewayTimeout {
		t.Errorf("Expected type %s, got %s", ErrorTypeGatewayTimeout, wrapped.Type)
	}
	
	if wrapped.Message != "Provider timeout" {
		t.Errorf("Expected message 'Provider timeout', got %s", wrapped.Message)
	}
	
	if !wrapped.Retryable {
		t.Error("Gateway timeout should be retryable")
	}
	
	// Test unwrap
	if unwrapped := wrapped.Unwrap(); unwrapped != originalErr {
		t.Error("Unwrap should return original error")
	}
}

func TestWrapNil(t *testing.T) {
	wrapped := Wrap(nil, ErrorTypeInternal, "Should not happen")
	
	if wrapped != nil {
		t.Error("Wrapping nil error should return nil")
	}
}

func TestWrapCCProxyError(t *testing.T) {
	original := New(ErrorTypeBadRequest, "Original error").
		WithDetails(map[string]interface{}{"field": "email"}).
		WithProvider("anthropic").
		WithRequestID("req-123")
	
	wrapped := Wrap(original, ErrorTypeValidationError, "Validation failed")
	
	// Should preserve metadata
	if wrapped.Provider != "anthropic" {
		t.Errorf("Expected provider to be preserved, got %s", wrapped.Provider)
	}
	
	if wrapped.RequestID != "req-123" {
		t.Errorf("Expected request ID to be preserved, got %s", wrapped.RequestID)
	}
	
	if wrapped.Details["field"] != "email" {
		t.Error("Expected details to be preserved")
	}
}

func TestWithMethods(t *testing.T) {
	err := New(ErrorTypeProviderError, "API error").
		WithProvider("openai").
		WithRequestID("req-456").
		WithDetails(map[string]interface{}{
			"model": "gpt-4",
			"usage": map[string]int{
				"prompt_tokens": 100,
			},
		}).
		WithRetryAfter(30 * time.Second)
	
	if err.Provider != "openai" {
		t.Errorf("Expected provider 'openai', got %s", err.Provider)
	}
	
	if err.RequestID != "req-456" {
		t.Errorf("Expected request ID 'req-456', got %s", err.RequestID)
	}
	
	if err.Details["model"] != "gpt-4" {
		t.Error("Expected model in details")
	}
	
	if !err.Retryable {
		t.Error("Expected error to be retryable after WithRetryAfter")
	}
	
	if err.RetryAfter == nil || *err.RetryAfter != 30*time.Second {
		t.Error("Expected retry after to be 30 seconds")
	}
}

func TestFromProviderResponse(t *testing.T) {
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
			body:       []byte(`{"error":{"type":"invalid_request_error","message":"Invalid API key"}}`),
			provider:   "anthropic",
			wantType:   ErrorTypeBadRequest,
			wantMsg:    "Invalid API key",
		},
		{
			name:       "no body",
			statusCode: 503,
			body:       []byte{},
			provider:   "openai",
			wantType:   ErrorTypeServiceUnavailable,
			wantMsg:    "Provider returned status 503",
		},
		{
			name:       "invalid json",
			statusCode: 500,
			body:       []byte("Internal Server Error"),
			provider:   "google",
			wantType:   ErrorTypeInternal,
			wantMsg:    "Provider returned status 500",
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
}

func TestToJSON(t *testing.T) {
	err := New(ErrorTypeBadRequest, "Invalid request").
		WithCode("invalid_model").
		WithProvider("anthropic").
		WithRequestID("req-789").
		WithDetails(map[string]interface{}{
			"field": "model",
			"value": "unknown",
		})
	
	data, jsonErr := err.ToJSON()
	if jsonErr != nil {
		t.Fatalf("Failed to marshal error: %v", jsonErr)
	}
	
	var result map[string]interface{}
	if jsonErr := json.Unmarshal(data, &result); jsonErr != nil {
		t.Fatalf("Failed to unmarshal result: %v", jsonErr)
	}
	
	errorObj := result["error"].(map[string]interface{})
	
	if errorObj["type"] != string(ErrorTypeBadRequest) {
		t.Errorf("Expected type in JSON to be %s", ErrorTypeBadRequest)
	}
	
	if errorObj["message"] != "Invalid request" {
		t.Error("Expected message in JSON")
	}
	
	if errorObj["code"] != "invalid_model" {
		t.Error("Expected code in JSON")
	}
	
	if errorObj["provider"] != "anthropic" {
		t.Error("Expected provider in JSON")
	}
	
	if errorObj["request_id"] != "req-789" {
		t.Error("Expected request_id in JSON")
	}
	
	details := errorObj["details"].(map[string]interface{})
	if details["field"] != "model" {
		t.Error("Expected details to be included")
	}
}

func TestWriteHTTPResponse(t *testing.T) {
	err := New(ErrorTypeTooManyRequests, "Rate limit exceeded").
		WithRetryAfter(60 * time.Second)
	
	w := httptest.NewRecorder()
	err.WriteHTTPResponse(w)
	
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status code %d, got %d", http.StatusTooManyRequests, w.Code)
	}
	
	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Expected Content-Type to be application/json")
	}
	
	if w.Header().Get("Retry-After") != "60" {
		t.Error("Expected Retry-After header to be set")
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	if _, ok := result["error"]; !ok {
		t.Error("Expected error object in response")
	}
}

func TestIs(t *testing.T) {
	err1 := New(ErrorTypeBadRequest, "Error 1")
	err2 := New(ErrorTypeBadRequest, "Error 2")
	err3 := New(ErrorTypeInternal, "Error 3")
	
	if !err1.Is(err2) {
		t.Error("Errors of same type should match")
	}
	
	if err1.Is(err3) {
		t.Error("Errors of different types should not match")
	}
	
	if err1.Is(errors.New("not a CCProxyError")) {
		t.Error("Should not match non-CCProxyError")
	}
}

func TestGetStatusCodeForType(t *testing.T) {
	tests := []struct {
		errorType  ErrorType
		wantStatus int
	}{
		{ErrorTypeBadRequest, http.StatusBadRequest},
		{ErrorTypeUnauthorized, http.StatusUnauthorized},
		{ErrorTypeForbidden, http.StatusForbidden},
		{ErrorTypeNotFound, http.StatusNotFound},
		{ErrorTypeMethodNotAllowed, http.StatusMethodNotAllowed},
		{ErrorTypeConflict, http.StatusConflict},
		{ErrorTypeUnprocessableEntity, http.StatusUnprocessableEntity},
		{ErrorTypeTooManyRequests, http.StatusTooManyRequests},
		{ErrorTypeRateLimitError, http.StatusTooManyRequests},
		{ErrorTypeInternal, http.StatusInternalServerError},
		{ErrorTypeNotImplemented, http.StatusNotImplemented},
		{ErrorTypeBadGateway, http.StatusBadGateway},
		{ErrorTypeProviderError, http.StatusBadGateway},
		{ErrorTypeServiceUnavailable, http.StatusServiceUnavailable},
		{ErrorTypeGatewayTimeout, http.StatusGatewayTimeout},
		{ErrorTypeValidationError, http.StatusBadRequest},
		{ErrorTypeTransformError, http.StatusInternalServerError},
	}
	
	for _, tt := range tests {
		t.Run(string(tt.errorType), func(t *testing.T) {
			got := getStatusCodeForType(tt.errorType)
			if got != tt.wantStatus {
				t.Errorf("getStatusCodeForType(%s) = %d, want %d", tt.errorType, got, tt.wantStatus)
			}
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		want      bool
	}{
		{ErrorTypeInternal, true},
		{ErrorTypeBadGateway, true},
		{ErrorTypeServiceUnavailable, true},
		{ErrorTypeGatewayTimeout, true},
		{ErrorTypeTooManyRequests, true},
		{ErrorTypeRateLimitError, true},
		{ErrorTypeBadRequest, false},
		{ErrorTypeUnauthorized, false},
		{ErrorTypeNotFound, false},
		{ErrorTypeValidationError, false},
	}
	
	for _, tt := range tests {
		t.Run(string(tt.errorType), func(t *testing.T) {
			got := isRetryable(tt.errorType)
			if got != tt.want {
				t.Errorf("isRetryable(%s) = %v, want %v", tt.errorType, got, tt.want)
			}
		})
	}
}

func TestCommonErrorConstructors(t *testing.T) {
	t.Run("ErrBadRequest", func(t *testing.T) {
		err := ErrBadRequest("Invalid input")
		if err.Type != ErrorTypeBadRequest {
			t.Errorf("Expected type %s, got %s", ErrorTypeBadRequest, err.Type)
		}
		if err.Message != "Invalid input" {
			t.Errorf("Expected message 'Invalid input', got %s", err.Message)
		}
	})
	
	t.Run("ErrUnauthorized", func(t *testing.T) {
		err := ErrUnauthorized("Invalid token")
		if err.Type != ErrorTypeUnauthorized {
			t.Errorf("Expected type %s, got %s", ErrorTypeUnauthorized, err.Type)
		}
	})
	
	t.Run("ErrNotFound", func(t *testing.T) {
		err := ErrNotFound("user")
		if err.Type != ErrorTypeNotFound {
			t.Errorf("Expected type %s, got %s", ErrorTypeNotFound, err.Type)
		}
		if err.Message != "user not found" {
			t.Errorf("Expected message 'user not found', got %s", err.Message)
		}
	})
	
	t.Run("ErrInternal", func(t *testing.T) {
		err := ErrInternal("Database error")
		if err.Type != ErrorTypeInternal {
			t.Errorf("Expected type %s, got %s", ErrorTypeInternal, err.Type)
		}
	})
	
	t.Run("ErrProviderError", func(t *testing.T) {
		origErr := errors.New("connection failed")
		err := ErrProviderError("openai", origErr)
		if err.Type != ErrorTypeProviderError {
			t.Errorf("Expected type %s, got %s", ErrorTypeProviderError, err.Type)
		}
		if err.Provider != "openai" {
			t.Errorf("Expected provider 'openai', got %s", err.Provider)
		}
		if err.Unwrap() != origErr {
			t.Error("Expected wrapped error")
		}
	})
	
	t.Run("ErrRateLimit", func(t *testing.T) {
		err := ErrRateLimit(30 * time.Second)
		if err.Type != ErrorTypeRateLimitError {
			t.Errorf("Expected type %s, got %s", ErrorTypeRateLimitError, err.Type)
		}
		if !err.Retryable {
			t.Error("Rate limit error should be retryable")
		}
		if err.RetryAfter == nil || *err.RetryAfter != 30*time.Second {
			t.Error("Expected retry after to be set")
		}
	})
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "CCProxyError retryable",
			err:  New(ErrorTypeGatewayTimeout, "Timeout"),
			want: true,
		},
		{
			name: "CCProxyError not retryable",
			err:  New(ErrorTypeBadRequest, "Bad input"),
			want: false,
		},
		{
			name: "connection refused",
			err:  errors.New("dial tcp: connection refused"),
			want: true,
		},
		{
			name: "connection reset",
			err:  errors.New("read: connection reset by peer"),
			want: true,
		},
		{
			name: "timeout",
			err:  errors.New("request timeout"),
			want: true,
		},
		{
			name: "service unavailable",
			err:  errors.New("503 Service Unavailable"),
			want: true,
		},
		{
			name: "bad gateway",
			err:  errors.New("502 Bad Gateway"),
			want: true,
		},
		{
			name: "no such host",
			err:  errors.New("dial tcp: lookup example.com: no such host"),
			want: true,
		},
		{
			name: "generic error",
			err:  errors.New("something went wrong"),
			want: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetryableError(tt.err)
			if got != tt.want {
				t.Errorf("IsRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRetryAfter(t *testing.T) {
	duration := 45 * time.Second
	
	t.Run("CCProxyError with retry after", func(t *testing.T) {
		err := New(ErrorTypeTooManyRequests, "Rate limited").WithRetryAfter(duration)
		got := GetRetryAfter(err)
		if got == nil {
			t.Fatal("Expected retry after duration")
		}
		if *got != duration {
			t.Errorf("Expected %v, got %v", duration, *got)
		}
	})
	
	t.Run("CCProxyError without retry after", func(t *testing.T) {
		err := New(ErrorTypeInternal, "Server error")
		got := GetRetryAfter(err)
		if got != nil {
			t.Error("Expected nil retry after")
		}
	})
	
	t.Run("non-CCProxyError", func(t *testing.T) {
		err := errors.New("generic error")
		got := GetRetryAfter(err)
		if got != nil {
			t.Error("Expected nil retry after for non-CCProxyError")
		}
	})
}