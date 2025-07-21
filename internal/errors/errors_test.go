package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		expected  string
	}{
		{"BadRequest", ErrorTypeBadRequest, "bad_request"},
		{"Unauthorized", ErrorTypeUnauthorized, "unauthorized"},
		{"Forbidden", ErrorTypeForbidden, "forbidden"},
		{"NotFound", ErrorTypeNotFound, "not_found"},
		{"MethodNotAllowed", ErrorTypeMethodNotAllowed, "method_not_allowed"},
		{"Conflict", ErrorTypeConflict, "conflict"},
		{"UnprocessableEntity", ErrorTypeUnprocessableEntity, "unprocessable_entity"},
		{"TooManyRequests", ErrorTypeTooManyRequests, "too_many_requests"},
		{"Internal", ErrorTypeInternal, "internal_error"},
		{"NotImplemented", ErrorTypeNotImplemented, "not_implemented"},
		{"BadGateway", ErrorTypeBadGateway, "bad_gateway"},
		{"ServiceUnavailable", ErrorTypeServiceUnavailable, "service_unavailable"},
		{"GatewayTimeout", ErrorTypeGatewayTimeout, "gateway_timeout"},
		{"ProviderError", ErrorTypeProviderError, "provider_error"},
		{"TransformError", ErrorTypeTransformError, "transform_error"},
		{"RoutingError", ErrorTypeRoutingError, "routing_error"},
		{"StreamingError", ErrorTypeStreamingError, "streaming_error"},
		{"ConfigError", ErrorTypeConfigError, "config_error"},
		{"ValidationError", ErrorTypeValidationError, "validation_error"},
		{"RateLimitError", ErrorTypeRateLimitError, "rate_limit_error"},
		{"ProxyError", ErrorTypeProxyError, "proxy_error"},
		{"ToolError", ErrorTypeToolError, "tool_error"},
		{"ResourceExhausted", ErrorTypeResourceExhausted, "resource_exhausted"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.AssertEqual(t, tt.expected, string(tt.errorType))
		})
	}
}

func TestNew(t *testing.T) {
	message := "test error message"
	err := New(ErrorTypeBadRequest, message)

	testutil.AssertEqual(t, ErrorTypeBadRequest, err.Type)
	testutil.AssertEqual(t, message, err.Message)
	testutil.AssertEqual(t, http.StatusBadRequest, err.StatusCode)
	testutil.AssertEqual(t, false, err.Retryable)
	testutil.AssertTrue(t, !err.Timestamp.IsZero())
	testutil.AssertEqual(t, "", err.Code)
	testutil.AssertEqual(t, "", err.Provider)
	testutil.AssertEqual(t, "", err.RequestID)
	testutil.AssertEqual(t, true, err.Details == nil)
	testutil.AssertEqual(t, true, err.RetryAfter == nil)
}

func TestNewf(t *testing.T) {
	format := "error with value: %d"
	value := 42
	err := Newf(ErrorTypeBadRequest, format, value)

	expectedMessage := fmt.Sprintf(format, value)
	testutil.AssertEqual(t, ErrorTypeBadRequest, err.Type)
	testutil.AssertEqual(t, expectedMessage, err.Message)
	testutil.AssertEqual(t, http.StatusBadRequest, err.StatusCode)
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := Wrap(originalErr, ErrorTypeInternal, "wrapped message")

	testutil.AssertEqual(t, ErrorTypeInternal, wrappedErr.Type)
	testutil.AssertEqual(t, "wrapped message", wrappedErr.Message)
	testutil.AssertEqual(t, originalErr, wrappedErr.wrapped)
	testutil.AssertEqual(t, http.StatusInternalServerError, wrappedErr.StatusCode)
}

func TestWrapNilError(t *testing.T) {
	wrappedErr := Wrap(nil, ErrorTypeInternal, "wrapped message")
	testutil.AssertEqual(t, (*CCProxyError)(nil), wrappedErr)
}

func TestWrapCCProxyError(t *testing.T) {
	originalErr := New(ErrorTypeBadRequest, "original").
		WithCode("CODE123").
		WithProvider("test-provider").
		WithRequestID("req-123").
		WithDetails(map[string]interface{}{"key": "value"})

	wrappedErr := Wrap(originalErr, ErrorTypeInternal, "wrapped message")

	testutil.AssertEqual(t, ErrorTypeInternal, wrappedErr.Type)
	testutil.AssertEqual(t, "wrapped message", wrappedErr.Message)
	testutil.AssertEqual(t, "CODE123", wrappedErr.Code)
	testutil.AssertEqual(t, "test-provider", wrappedErr.Provider)
	testutil.AssertEqual(t, "req-123", wrappedErr.RequestID)
	testutil.AssertEqual(t, "value", wrappedErr.Details["key"])
	testutil.AssertEqual(t, originalErr, wrappedErr.wrapped)
}

func TestWrapf(t *testing.T) {
	originalErr := errors.New("original error")
	format := "wrapped with value: %d"
	value := 42
	wrappedErr := Wrapf(originalErr, ErrorTypeInternal, format, value)

	expectedMessage := fmt.Sprintf(format, value)
	testutil.AssertEqual(t, ErrorTypeInternal, wrappedErr.Type)
	testutil.AssertEqual(t, expectedMessage, wrappedErr.Message)
	testutil.AssertEqual(t, originalErr, wrappedErr.wrapped)
}

func TestCCProxyError_Error(t *testing.T) {
	t.Run("WithoutWrappedError", func(t *testing.T) {
		err := New(ErrorTypeBadRequest, "test message")
		expected := "bad_request: test message"
		testutil.AssertEqual(t, expected, err.Error())
	})

	t.Run("WithWrappedError", func(t *testing.T) {
		originalErr := errors.New("original error")
		err := Wrap(originalErr, ErrorTypeInternal, "wrapped message")
		expected := "internal_error: wrapped message (wrapped: original error)"
		testutil.AssertEqual(t, expected, err.Error())
	})
}

func TestCCProxyError_Unwrap(t *testing.T) {
	t.Run("WithoutWrappedError", func(t *testing.T) {
		err := New(ErrorTypeBadRequest, "test message")
		testutil.AssertEqual(t, true, err.Unwrap() == nil)
	})

	t.Run("WithWrappedError", func(t *testing.T) {
		originalErr := errors.New("original error")
		err := Wrap(originalErr, ErrorTypeInternal, "wrapped message")
		testutil.AssertEqual(t, originalErr, err.Unwrap())
	})
}

func TestCCProxyError_Is(t *testing.T) {
	err := New(ErrorTypeBadRequest, "test message")

	t.Run("SameType", func(t *testing.T) {
		target := New(ErrorTypeBadRequest, "different message")
		testutil.AssertTrue(t, err.Is(target))
	})

	t.Run("DifferentType", func(t *testing.T) {
		target := New(ErrorTypeInternal, "test message")
		testutil.AssertFalse(t, err.Is(target))
	})

	t.Run("NotCCProxyError", func(t *testing.T) {
		target := errors.New("regular error")
		testutil.AssertFalse(t, err.Is(target))
	})
}

func TestCCProxyError_WithDetails(t *testing.T) {
	err := New(ErrorTypeBadRequest, "test message")
	details := map[string]interface{}{"key": "value", "number": 42}

	result := err.WithDetails(details)

	testutil.AssertEqual(t, err, result) // Should return same instance
	testutil.AssertEqual(t, "value", err.Details["key"])
	testutil.AssertEqual(t, 42, err.Details["number"])
}

func TestCCProxyError_WithProvider(t *testing.T) {
	err := New(ErrorTypeBadRequest, "test message")
	provider := "test-provider"

	result := err.WithProvider(provider)

	testutil.AssertEqual(t, err, result) // Should return same instance
	testutil.AssertEqual(t, provider, err.Provider)
}

func TestCCProxyError_WithRequestID(t *testing.T) {
	err := New(ErrorTypeBadRequest, "test message")
	requestID := "req-123"

	result := err.WithRequestID(requestID)

	testutil.AssertEqual(t, err, result) // Should return same instance
	testutil.AssertEqual(t, requestID, err.RequestID)
}

func TestCCProxyError_WithRetryAfter(t *testing.T) {
	err := New(ErrorTypeBadRequest, "test message")
	duration := 5 * time.Second

	result := err.WithRetryAfter(duration)

	testutil.AssertEqual(t, err, result) // Should return same instance
	testutil.AssertEqual(t, true, err.Retryable)
	testutil.AssertEqual(t, duration, *err.RetryAfter)
}

func TestCCProxyError_WithCode(t *testing.T) {
	err := New(ErrorTypeBadRequest, "test message")
	code := "ERROR_CODE_123"

	result := err.WithCode(code)

	testutil.AssertEqual(t, err, result) // Should return same instance
	testutil.AssertEqual(t, code, err.Code)
}

func TestSanitizeErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "API Key",
			input:    "Error with API_KEY: sk-1234567890abcdef",
			expected: "Error with [REDACTED]",
		},
		{
			name:     "Token",
			input:    "Authentication failed: token abc123def456",
			expected: "Authentication failed: [REDACTED]",
		},
		{
			name:     "Password",
			input:    "Login failed with password: secret123",
			expected: "Login failed with [REDACTED]",
		},
		{
			name:     "Bearer Token",
			input:    "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expected: "[REDACTED] eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		},
		{
			name:     "Email",
			input:    "User email: user@example.com not found",
			expected: "User email: [REDACTED] not found",
		},
		{
			name:     "IP Address",
			input:    "Connection failed to 192.168.1.1:8080",
			expected: "Connection failed to [REDACTED]",
		},
		{
			name:     "Long Message",
			input:    strings.Repeat("a", 600),
			expected: strings.Repeat("a", 497) + "...",
		},
		{
			name:     "Clean Message",
			input:    "Regular error message",
			expected: "Regular error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeErrorMessage(tt.input)
			testutil.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestFromProviderResponse(t *testing.T) {
	t.Run("ValidJSON", func(t *testing.T) {
		body := []byte(`{"error":{"type":"invalid_request","message":"Bad request","code":"400"}}`)
		err := FromProviderResponse(400, body, "test-provider")

		testutil.AssertEqual(t, ErrorTypeBadRequest, err.Type)
		testutil.AssertEqual(t, "Bad request", err.Message)
		testutil.AssertEqual(t, "400", err.Code)
		testutil.AssertEqual(t, 400, err.StatusCode)
		testutil.AssertEqual(t, "test-provider", err.Provider)
		testutil.AssertEqual(t, false, err.Retryable)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		body := []byte(`invalid json`)
		err := FromProviderResponse(500, body, "test-provider")

		testutil.AssertEqual(t, ErrorTypeInternal, err.Type)
		testutil.AssertEqual(t, "Provider returned status 500", err.Message)
		testutil.AssertEqual(t, "", err.Code)
		testutil.AssertEqual(t, 500, err.StatusCode)
		testutil.AssertEqual(t, "test-provider", err.Provider)
		testutil.AssertEqual(t, true, err.Retryable)
	})

	t.Run("RetryableStatusCode", func(t *testing.T) {
		body := []byte(`{"error":{"message":"Rate limited"}}`)
		err := FromProviderResponse(429, body, "test-provider")

		testutil.AssertEqual(t, ErrorTypeTooManyRequests, err.Type)
		testutil.AssertEqual(t, "Rate limited", err.Message)
		testutil.AssertEqual(t, 429, err.StatusCode)
		testutil.AssertEqual(t, true, err.Retryable)
	})
}

func TestCCProxyError_ToJSON(t *testing.T) {
	t.Run("MinimalError", func(t *testing.T) {
		err := New(ErrorTypeBadRequest, "test message")
		jsonData, jsonErr := err.ToJSON()

		testutil.AssertNoError(t, jsonErr)

		var response map[string]interface{}
		unmarshalErr := json.Unmarshal(jsonData, &response)
		testutil.AssertNoError(t, unmarshalErr)

		errorObj := response["error"].(map[string]interface{})
		testutil.AssertEqual(t, "bad_request", errorObj["type"])
		testutil.AssertEqual(t, "test message", errorObj["message"])
	})

	t.Run("FullError", func(t *testing.T) {
		err := New(ErrorTypeBadRequest, "test message").
			WithCode("CODE123").
			WithProvider("test-provider").
			WithRequestID("req-123").
			WithDetails(map[string]interface{}{"key": "value"})

		jsonData, jsonErr := err.ToJSON()
		testutil.AssertNoError(t, jsonErr)

		var response map[string]interface{}
		unmarshalErr := json.Unmarshal(jsonData, &response)
		testutil.AssertNoError(t, unmarshalErr)

		errorObj := response["error"].(map[string]interface{})
		testutil.AssertEqual(t, "bad_request", errorObj["type"])
		testutil.AssertEqual(t, "test message", errorObj["message"])
		testutil.AssertEqual(t, "CODE123", errorObj["code"])
		testutil.AssertEqual(t, "test-provider", errorObj["provider"])
		testutil.AssertEqual(t, "req-123", errorObj["request_id"])

		details := errorObj["details"].(map[string]interface{})
		testutil.AssertEqual(t, "value", details["key"])
	})
}

func TestCCProxyError_WriteHTTPResponse(t *testing.T) {
	t.Run("BasicResponse", func(t *testing.T) {
		err := New(ErrorTypeBadRequest, "test message")
		recorder := httptest.NewRecorder()

		err.WriteHTTPResponse(recorder)

		testutil.AssertEqual(t, http.StatusBadRequest, recorder.Code)
		testutil.AssertEqual(t, "application/json", recorder.Header().Get("Content-Type"))

		var response map[string]interface{}
		unmarshalErr := json.Unmarshal(recorder.Body.Bytes(), &response)
		testutil.AssertNoError(t, unmarshalErr)

		errorObj := response["error"].(map[string]interface{})
		testutil.AssertEqual(t, "bad_request", errorObj["type"])
		testutil.AssertEqual(t, "test message", errorObj["message"])
	})

	t.Run("WithRetryAfter", func(t *testing.T) {
		retryAfter := 30 * time.Second
		err := New(ErrorTypeTooManyRequests, "rate limited").WithRetryAfter(retryAfter)
		recorder := httptest.NewRecorder()

		err.WriteHTTPResponse(recorder)

		testutil.AssertEqual(t, http.StatusTooManyRequests, recorder.Code)
		testutil.AssertEqual(t, "30", recorder.Header().Get("Retry-After"))
	})
}

func TestGetStatusCodeForType(t *testing.T) {
	tests := []struct {
		errorType  ErrorType
		statusCode int
	}{
		{ErrorTypeBadRequest, http.StatusBadRequest},
		{ErrorTypeValidationError, http.StatusBadRequest},
		{ErrorTypeUnauthorized, http.StatusUnauthorized},
		{ErrorTypeForbidden, http.StatusForbidden},
		{ErrorTypeNotFound, http.StatusNotFound},
		{ErrorTypeMethodNotAllowed, http.StatusMethodNotAllowed},
		{ErrorTypeConflict, http.StatusConflict},
		{ErrorTypeUnprocessableEntity, http.StatusUnprocessableEntity},
		{ErrorTypeTooManyRequests, http.StatusTooManyRequests},
		{ErrorTypeRateLimitError, http.StatusTooManyRequests},
		{ErrorTypeNotImplemented, http.StatusNotImplemented},
		{ErrorTypeBadGateway, http.StatusBadGateway},
		{ErrorTypeProviderError, http.StatusBadGateway},
		{ErrorTypeServiceUnavailable, http.StatusServiceUnavailable},
		{ErrorTypeGatewayTimeout, http.StatusGatewayTimeout},
		{ErrorTypeInternal, http.StatusInternalServerError},
		{ErrorTypeTransformError, http.StatusInternalServerError}, // Default case
	}

	for _, tt := range tests {
		t.Run(string(tt.errorType), func(t *testing.T) {
			result := getStatusCodeForType(tt.errorType)
			testutil.AssertEqual(t, tt.statusCode, result)
		})
	}
}

func TestGetErrorTypeFromStatusCode(t *testing.T) {
	tests := []struct {
		statusCode int
		errorType  ErrorType
	}{
		{http.StatusBadRequest, ErrorTypeBadRequest},
		{http.StatusUnauthorized, ErrorTypeUnauthorized},
		{http.StatusForbidden, ErrorTypeForbidden},
		{http.StatusNotFound, ErrorTypeNotFound},
		{http.StatusMethodNotAllowed, ErrorTypeMethodNotAllowed},
		{http.StatusConflict, ErrorTypeConflict},
		{http.StatusUnprocessableEntity, ErrorTypeUnprocessableEntity},
		{http.StatusTooManyRequests, ErrorTypeTooManyRequests},
		{http.StatusNotImplemented, ErrorTypeNotImplemented},
		{http.StatusBadGateway, ErrorTypeBadGateway},
		{http.StatusServiceUnavailable, ErrorTypeServiceUnavailable},
		{http.StatusGatewayTimeout, ErrorTypeGatewayTimeout},
		{http.StatusInternalServerError, ErrorTypeInternal},
		{http.StatusTeapot, ErrorTypeBadRequest},         // 4xx default
		{http.StatusBadGateway + 100, ErrorTypeInternal}, // 5xx default
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Status%d", tt.statusCode), func(t *testing.T) {
			result := getErrorTypeFromStatusCode(tt.statusCode)
			testutil.AssertEqual(t, tt.errorType, result)
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		expected  bool
	}{
		{"Internal", ErrorTypeInternal, true},
		{"BadGateway", ErrorTypeBadGateway, true},
		{"ServiceUnavailable", ErrorTypeServiceUnavailable, true},
		{"GatewayTimeout", ErrorTypeGatewayTimeout, true},
		{"TooManyRequests", ErrorTypeTooManyRequests, true},
		{"RateLimitError", ErrorTypeRateLimitError, true},
		{"BadRequest", ErrorTypeBadRequest, false},
		{"Unauthorized", ErrorTypeUnauthorized, false},
		{"NotFound", ErrorTypeNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryable(tt.errorType)
			testutil.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestErrBadRequest(t *testing.T) {
	message := "bad request message"
	err := ErrBadRequest(message)

	testutil.AssertEqual(t, ErrorTypeBadRequest, err.Type)
	testutil.AssertEqual(t, message, err.Message)
	testutil.AssertEqual(t, http.StatusBadRequest, err.StatusCode)
}

func TestErrUnauthorized(t *testing.T) {
	message := "unauthorized message"
	err := ErrUnauthorized(message)

	testutil.AssertEqual(t, ErrorTypeUnauthorized, err.Type)
	testutil.AssertEqual(t, message, err.Message)
	testutil.AssertEqual(t, http.StatusUnauthorized, err.StatusCode)
}

func TestErrNotFound(t *testing.T) {
	resource := "user"
	err := ErrNotFound(resource)

	testutil.AssertEqual(t, ErrorTypeNotFound, err.Type)
	testutil.AssertEqual(t, "user not found", err.Message)
	testutil.AssertEqual(t, http.StatusNotFound, err.StatusCode)
}

func TestErrInternal(t *testing.T) {
	message := "internal error message"
	err := ErrInternal(message)

	testutil.AssertEqual(t, ErrorTypeInternal, err.Type)
	testutil.AssertEqual(t, message, err.Message)
	testutil.AssertEqual(t, http.StatusInternalServerError, err.StatusCode)
}

func TestErrProviderError(t *testing.T) {
	provider := "test-provider"
	originalErr := errors.New("original error")
	err := ErrProviderError(provider, originalErr)

	testutil.AssertEqual(t, ErrorTypeProviderError, err.Type)
	testutil.AssertEqual(t, "Provider test-provider error", err.Message)
	testutil.AssertEqual(t, provider, err.Provider)
	testutil.AssertEqual(t, originalErr, err.wrapped)
}

func TestErrRateLimit(t *testing.T) {
	retryAfter := 60 * time.Second
	err := ErrRateLimit(retryAfter)

	testutil.AssertEqual(t, ErrorTypeRateLimitError, err.Type)
	testutil.AssertEqual(t, "Rate limit exceeded", err.Message)
	testutil.AssertEqual(t, true, err.Retryable)
	testutil.AssertEqual(t, retryAfter, *err.RetryAfter)
}

func TestIsRetryableError(t *testing.T) {
	t.Run("CCProxyError", func(t *testing.T) {
		retryableErr := New(ErrorTypeInternal, "retryable error")
		nonRetryableErr := New(ErrorTypeBadRequest, "non-retryable error")

		testutil.AssertTrue(t, IsRetryableError(retryableErr))
		testutil.AssertFalse(t, IsRetryableError(nonRetryableErr))
	})

	t.Run("RegularError", func(t *testing.T) {
		tests := []struct {
			name     string
			err      error
			expected bool
		}{
			{"Connection refused", errors.New("connection refused"), true},
			{"Connection reset", errors.New("connection reset by peer"), true},
			{"No such host", errors.New("no such host"), true},
			{"Timeout", errors.New("request timeout"), true},
			{"Temporary failure", errors.New("temporary failure in name resolution"), true},
			{"Service unavailable", errors.New("service unavailable"), true},
			{"Bad gateway", errors.New("bad gateway"), true},
			{"Regular error", errors.New("regular error"), false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := IsRetryableError(tt.err)
				testutil.AssertEqual(t, tt.expected, result)
			})
		}
	})
}

func TestGetRetryAfter(t *testing.T) {
	t.Run("CCProxyErrorWithRetryAfter", func(t *testing.T) {
		duration := 30 * time.Second
		err := New(ErrorTypeRateLimitError, "rate limited").WithRetryAfter(duration)

		result := GetRetryAfter(err)
		testutil.AssertEqual(t, false, result == nil)
		testutil.AssertEqual(t, duration, *result)
	})

	t.Run("CCProxyErrorWithoutRetryAfter", func(t *testing.T) {
		err := New(ErrorTypeBadRequest, "bad request")

		result := GetRetryAfter(err)
		testutil.AssertEqual(t, true, result == nil)
	})

	t.Run("RegularError", func(t *testing.T) {
		err := errors.New("regular error")

		result := GetRetryAfter(err)
		testutil.AssertEqual(t, true, result == nil)
	})
}

func TestNewAuthError(t *testing.T) {
	t.Run("WithoutCause", func(t *testing.T) {
		message := "authentication failed"
		err := NewAuthError(message, nil)

		testutil.AssertEqual(t, ErrorTypeUnauthorized, err.Type)
		testutil.AssertEqual(t, message, err.Message)
		testutil.AssertEqual(t, true, err.wrapped == nil)
	})

	t.Run("WithCause", func(t *testing.T) {
		message := "authentication failed"
		cause := errors.New("invalid token")
		err := NewAuthError(message, cause)

		testutil.AssertEqual(t, ErrorTypeUnauthorized, err.Type)
		testutil.AssertEqual(t, message, err.Message)
		testutil.AssertEqual(t, cause, err.wrapped)
	})
}

func TestNewValidationError(t *testing.T) {
	t.Run("WithoutCause", func(t *testing.T) {
		message := "validation failed"
		err := NewValidationError(message, nil)

		testutil.AssertEqual(t, ErrorTypeValidationError, err.Type)
		testutil.AssertEqual(t, message, err.Message)
		testutil.AssertEqual(t, true, err.wrapped == nil)
	})

	t.Run("WithCause", func(t *testing.T) {
		message := "validation failed"
		cause := errors.New("required field missing")
		err := NewValidationError(message, cause)

		testutil.AssertEqual(t, ErrorTypeValidationError, err.Type)
		testutil.AssertEqual(t, message, err.Message)
		testutil.AssertEqual(t, cause, err.wrapped)
	})
}

func TestNewRateLimitError(t *testing.T) {
	t.Run("WithoutCause", func(t *testing.T) {
		message := "rate limit exceeded"
		err := NewRateLimitError(message, nil)

		testutil.AssertEqual(t, ErrorTypeRateLimitError, err.Type)
		testutil.AssertEqual(t, message, err.Message)
		testutil.AssertEqual(t, true, err.wrapped == nil)
	})

	t.Run("WithCause", func(t *testing.T) {
		message := "rate limit exceeded"
		cause := errors.New("too many requests")
		err := NewRateLimitError(message, cause)

		testutil.AssertEqual(t, ErrorTypeRateLimitError, err.Type)
		testutil.AssertEqual(t, message, err.Message)
		testutil.AssertEqual(t, cause, err.wrapped)
	})
}

func TestNewForbiddenError(t *testing.T) {
	t.Run("WithoutCause", func(t *testing.T) {
		message := "access forbidden"
		err := NewForbiddenError(message, nil)

		testutil.AssertEqual(t, ErrorTypeForbidden, err.Type)
		testutil.AssertEqual(t, message, err.Message)
		testutil.AssertEqual(t, true, err.wrapped == nil)
	})

	t.Run("WithCause", func(t *testing.T) {
		message := "access forbidden"
		cause := errors.New("insufficient permissions")
		err := NewForbiddenError(message, cause)

		testutil.AssertEqual(t, ErrorTypeForbidden, err.Type)
		testutil.AssertEqual(t, message, err.Message)
		testutil.AssertEqual(t, cause, err.wrapped)
	})
}

func TestNewNotFoundError(t *testing.T) {
	t.Run("WithoutCause", func(t *testing.T) {
		message := "resource not found"
		err := NewNotFoundError(message, nil)

		testutil.AssertEqual(t, ErrorTypeNotFound, err.Type)
		testutil.AssertEqual(t, message, err.Message)
		testutil.AssertEqual(t, true, err.wrapped == nil)
	})

	t.Run("WithCause", func(t *testing.T) {
		message := "resource not found"
		cause := errors.New("id does not exist")
		err := NewNotFoundError(message, cause)

		testutil.AssertEqual(t, ErrorTypeNotFound, err.Type)
		testutil.AssertEqual(t, message, err.Message)
		testutil.AssertEqual(t, cause, err.wrapped)
	})
}

func TestErrorsInterface(t *testing.T) {
	// Test that CCProxyError implements the error interface
	var err error = New(ErrorTypeBadRequest, "test message")
	testutil.AssertEqual(t, "bad_request: test message", err.Error())
}

func TestErrorsUnwrapInterface(t *testing.T) {
	// Test that CCProxyError can be unwrapped with errors.Unwrap
	originalErr := errors.New("original error")
	wrappedErr := Wrap(originalErr, ErrorTypeInternal, "wrapped message")

	unwrapped := errors.Unwrap(wrappedErr)
	testutil.AssertEqual(t, originalErr, unwrapped)
}

func TestErrorsIsInterface(t *testing.T) {
	// Test that CCProxyError works with errors.Is
	err1 := New(ErrorTypeBadRequest, "test message")
	err2 := New(ErrorTypeBadRequest, "different message")
	err3 := New(ErrorTypeInternal, "test message")

	testutil.AssertTrue(t, errors.Is(err1, err2))
	testutil.AssertFalse(t, errors.Is(err1, err3))
}

func TestErrorsAsInterface(t *testing.T) {
	// Test that CCProxyError works with errors.As
	originalErr := New(ErrorTypeBadRequest, "test message")
	wrappedErr := Wrap(originalErr, ErrorTypeInternal, "wrapped message")

	var ccErr *CCProxyError
	testutil.AssertTrue(t, errors.As(wrappedErr, &ccErr))
	testutil.AssertEqual(t, ErrorTypeInternal, ccErr.Type)
	testutil.AssertEqual(t, "wrapped message", ccErr.Message)
}

// Edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	t.Run("EmptyMessage", func(t *testing.T) {
		err := New(ErrorTypeBadRequest, "")
		testutil.AssertEqual(t, "", err.Message)
		testutil.AssertEqual(t, "bad_request: ", err.Error())
	})

	t.Run("NilDetails", func(t *testing.T) {
		err := New(ErrorTypeBadRequest, "test").WithDetails(nil)
		testutil.AssertEqual(t, true, err.Details == nil)
	})

	t.Run("EmptyDetails", func(t *testing.T) {
		err := New(ErrorTypeBadRequest, "test").WithDetails(map[string]interface{}{})
		testutil.AssertEqual(t, 0, len(err.Details))
	})

	t.Run("ZeroRetryAfter", func(t *testing.T) {
		err := New(ErrorTypeBadRequest, "test").WithRetryAfter(0)
		testutil.AssertEqual(t, true, err.Retryable)
		zero := time.Duration(0)
		testutil.AssertEqual(t, zero, *err.RetryAfter)
	})
}
