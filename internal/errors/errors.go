package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// Client errors (4xx)
	ErrorTypeBadRequest          ErrorType = "bad_request"
	ErrorTypeUnauthorized        ErrorType = "unauthorized"
	ErrorTypeForbidden           ErrorType = "forbidden"
	ErrorTypeNotFound            ErrorType = "not_found"
	ErrorTypeMethodNotAllowed    ErrorType = "method_not_allowed"
	ErrorTypeConflict            ErrorType = "conflict"
	ErrorTypeUnprocessableEntity ErrorType = "unprocessable_entity"
	ErrorTypeTooManyRequests     ErrorType = "too_many_requests"

	// Server errors (5xx)
	ErrorTypeInternal           ErrorType = "internal_error"
	ErrorTypeNotImplemented     ErrorType = "not_implemented"
	ErrorTypeBadGateway         ErrorType = "bad_gateway"
	ErrorTypeServiceUnavailable ErrorType = "service_unavailable"
	ErrorTypeGatewayTimeout     ErrorType = "gateway_timeout"

	// Custom errors
	ErrorTypeProviderError     ErrorType = "provider_error"
	ErrorTypeTransformError    ErrorType = "transform_error"
	ErrorTypeRoutingError      ErrorType = "routing_error"
	ErrorTypeStreamingError    ErrorType = "streaming_error"
	ErrorTypeConfigError       ErrorType = "config_error"
	ErrorTypeValidationError   ErrorType = "validation_error"
	ErrorTypeRateLimitError    ErrorType = "rate_limit_error"
	ErrorTypeProxyError        ErrorType = "proxy_error"
	ErrorTypeToolError         ErrorType = "tool_error"
	ErrorTypeResourceExhausted ErrorType = "resource_exhausted"
)

// CCProxyError represents a standardized error for ccproxy
type CCProxyError struct {
	Type       ErrorType              `json:"type"`
	Message    string                 `json:"message"`
	Code       string                 `json:"code,omitempty"`
	StatusCode int                    `json:"-"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Provider   string                 `json:"provider,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Retryable  bool                   `json:"retryable"`
	RetryAfter *time.Duration         `json:"-"`
	wrapped    error
}

// Error implements the error interface
func (e *CCProxyError) Error() string {
	if e.wrapped != nil {
		return fmt.Sprintf("%s: %s (wrapped: %v)", e.Type, e.Message, e.wrapped)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the wrapped error
func (e *CCProxyError) Unwrap() error {
	return e.wrapped
}

// Is checks if the error is of a specific type
func (e *CCProxyError) Is(target error) bool {
	t, ok := target.(*CCProxyError)
	if !ok {
		return false
	}
	return e.Type == t.Type
}

// WithDetails adds details to the error
func (e *CCProxyError) WithDetails(details map[string]interface{}) *CCProxyError {
	e.Details = details
	return e
}

// WithProvider adds provider information to the error
func (e *CCProxyError) WithProvider(provider string) *CCProxyError {
	e.Provider = provider
	return e
}

// WithRequestID adds request ID to the error
func (e *CCProxyError) WithRequestID(requestID string) *CCProxyError {
	e.RequestID = requestID
	return e
}

// WithRetryAfter sets retry information
func (e *CCProxyError) WithRetryAfter(duration time.Duration) *CCProxyError {
	e.Retryable = true
	e.RetryAfter = &duration
	return e
}

// WithCode adds a code to the error
func (e *CCProxyError) WithCode(code string) *CCProxyError {
	e.Code = code
	return e
}

// New creates a new CCProxyError
func New(errorType ErrorType, message string) *CCProxyError {
	return &CCProxyError{
		Type:       errorType,
		Message:    message,
		StatusCode: getStatusCodeForType(errorType),
		Timestamp:  time.Now(),
		Retryable:  isRetryable(errorType),
	}
}

// Newf creates a new CCProxyError with formatted message
func Newf(errorType ErrorType, format string, args ...interface{}) *CCProxyError {
	return New(errorType, fmt.Sprintf(format, args...))
}

// Wrap wraps an existing error
func Wrap(err error, errorType ErrorType, message string) *CCProxyError {
	if err == nil {
		return nil
	}

	// If it's already a CCProxyError, preserve some information
	if ccErr, ok := err.(*CCProxyError); ok {
		return &CCProxyError{
			Type:       errorType,
			Message:    message,
			Code:       ccErr.Code,
			StatusCode: getStatusCodeForType(errorType),
			Details:    ccErr.Details,
			Provider:   ccErr.Provider,
			RequestID:  ccErr.RequestID,
			Timestamp:  time.Now(),
			Retryable:  isRetryable(errorType),
			wrapped:    err,
		}
	}

	return &CCProxyError{
		Type:       errorType,
		Message:    message,
		StatusCode: getStatusCodeForType(errorType),
		Timestamp:  time.Now(),
		Retryable:  isRetryable(errorType),
		wrapped:    err,
	}
}

// Wrapf wraps an existing error with formatted message
func Wrapf(err error, errorType ErrorType, format string, args ...interface{}) *CCProxyError {
	return Wrap(err, errorType, fmt.Sprintf(format, args...))
}

// sanitizeErrorMessage removes sensitive information from error messages
func sanitizeErrorMessage(message string) string {
	// Remove potential API keys, tokens, and other sensitive data
	patterns := []string{
		`(?i)api[_-]?key[_-]?[:\s]*[a-zA-Z0-9_\-/+=]+`,       // API keys
		`(?i)token[_-]?[:\s]*[a-zA-Z0-9_\-/+=.]+`,            // Tokens
		`(?i)secret[_-]?[:\s]*[a-zA-Z0-9_\-/+=]+`,            // Secrets
		`(?i)password[_-]?[:\s]*\S+`,                         // Passwords
		`(?i)authorization[_-]?[:\s]*\S+`,                    // Authorization headers
		`(?i)bearer[_-]?\s*[a-zA-Z0-9_\-/+=.]+`,              // Bearer tokens
		`\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`, // Email addresses
		`\b(?:\d{1,3}\.){3}\d{1,3}(?::\d+)?\b`,               // IP addresses with ports
		`(?i)x-api-key[_-]?[:\s]*[a-zA-Z0-9_\-/+=]+`,         // x-api-key headers
	}

	sanitized := message
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		sanitized = re.ReplaceAllString(sanitized, "[REDACTED]")
	}

	// Limit message length to prevent log flooding
	if len(sanitized) > 500 {
		sanitized = sanitized[:497] + "..."
	}

	return sanitized
}

// FromProviderResponse creates an error from provider response
func FromProviderResponse(statusCode int, body []byte, provider string) *CCProxyError {
	var providerError struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
			Code    string `json:"code"`
		} `json:"error"`
	}

	message := fmt.Sprintf("Provider returned status %d", statusCode)

	if err := json.Unmarshal(body, &providerError); err == nil && providerError.Error.Message != "" {
		message = sanitizeErrorMessage(providerError.Error.Message)
	}

	errorType := getErrorTypeFromStatusCode(statusCode)

	return &CCProxyError{
		Type:       errorType,
		Message:    message,
		Code:       providerError.Error.Code,
		StatusCode: statusCode,
		Provider:   provider,
		Timestamp:  time.Now(),
		Retryable:  isRetryable(errorType),
		Details: map[string]interface{}{
			"provider_error_type": providerError.Error.Type,
		},
	}
}

// ToJSON converts error to JSON
func (e *CCProxyError) ToJSON() ([]byte, error) {
	// Create response structure matching Anthropic's error format
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"type":    string(e.Type),
			"message": e.Message,
		},
	}

	// Add optional fields
	if e.Code != "" {
		response["error"].(map[string]interface{})["code"] = e.Code
	}

	if len(e.Details) > 0 {
		response["error"].(map[string]interface{})["details"] = e.Details
	}

	if e.Provider != "" {
		response["error"].(map[string]interface{})["provider"] = e.Provider
	}

	if e.RequestID != "" {
		response["error"].(map[string]interface{})["request_id"] = e.RequestID
	}

	return json.Marshal(response)
}

// WriteHTTPResponse writes the error as HTTP response
func (e *CCProxyError) WriteHTTPResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")

	// Add retry header if applicable
	if e.Retryable && e.RetryAfter != nil {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", int(e.RetryAfter.Seconds())))
	}

	w.WriteHeader(e.StatusCode)

	data, err := e.ToJSON()
	if err != nil {
		// Fallback error response
		w.Write([]byte(`{"error":{"type":"internal_error","message":"Failed to serialize error"}}`))
		return
	}

	w.Write(data)
}

// Helper functions

func getStatusCodeForType(errorType ErrorType) int {
	switch errorType {
	case ErrorTypeBadRequest, ErrorTypeValidationError:
		return http.StatusBadRequest
	case ErrorTypeUnauthorized:
		return http.StatusUnauthorized
	case ErrorTypeForbidden:
		return http.StatusForbidden
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeMethodNotAllowed:
		return http.StatusMethodNotAllowed
	case ErrorTypeConflict:
		return http.StatusConflict
	case ErrorTypeUnprocessableEntity:
		return http.StatusUnprocessableEntity
	case ErrorTypeTooManyRequests, ErrorTypeRateLimitError:
		return http.StatusTooManyRequests
	case ErrorTypeNotImplemented:
		return http.StatusNotImplemented
	case ErrorTypeBadGateway, ErrorTypeProviderError:
		return http.StatusBadGateway
	case ErrorTypeServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrorTypeGatewayTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

func getErrorTypeFromStatusCode(statusCode int) ErrorType {
	switch statusCode {
	case http.StatusBadRequest:
		return ErrorTypeBadRequest
	case http.StatusUnauthorized:
		return ErrorTypeUnauthorized
	case http.StatusForbidden:
		return ErrorTypeForbidden
	case http.StatusNotFound:
		return ErrorTypeNotFound
	case http.StatusMethodNotAllowed:
		return ErrorTypeMethodNotAllowed
	case http.StatusConflict:
		return ErrorTypeConflict
	case http.StatusUnprocessableEntity:
		return ErrorTypeUnprocessableEntity
	case http.StatusTooManyRequests:
		return ErrorTypeTooManyRequests
	case http.StatusNotImplemented:
		return ErrorTypeNotImplemented
	case http.StatusBadGateway:
		return ErrorTypeBadGateway
	case http.StatusServiceUnavailable:
		return ErrorTypeServiceUnavailable
	case http.StatusGatewayTimeout:
		return ErrorTypeGatewayTimeout
	default:
		if statusCode >= 400 && statusCode < 500 {
			return ErrorTypeBadRequest
		}
		return ErrorTypeInternal
	}
}

func isRetryable(errorType ErrorType) bool {
	switch errorType {
	case ErrorTypeInternal,
		ErrorTypeBadGateway,
		ErrorTypeServiceUnavailable,
		ErrorTypeGatewayTimeout,
		ErrorTypeTooManyRequests,
		ErrorTypeRateLimitError:
		return true
	default:
		return false
	}
}

// Common error constructors

// ErrBadRequest creates a bad request error
func ErrBadRequest(message string) *CCProxyError {
	return New(ErrorTypeBadRequest, message)
}

// ErrUnauthorized creates an unauthorized error
func ErrUnauthorized(message string) *CCProxyError {
	return New(ErrorTypeUnauthorized, message)
}

// ErrNotFound creates a not found error
func ErrNotFound(resource string) *CCProxyError {
	return Newf(ErrorTypeNotFound, "%s not found", resource)
}

// ErrInternal creates an internal error
func ErrInternal(message string) *CCProxyError {
	return New(ErrorTypeInternal, message)
}

// ErrProviderError creates a provider error
func ErrProviderError(provider string, err error) *CCProxyError {
	return Wrap(err, ErrorTypeProviderError, fmt.Sprintf("Provider %s error", provider)).
		WithProvider(provider)
}

// ErrRateLimit creates a rate limit error
func ErrRateLimit(retryAfter time.Duration) *CCProxyError {
	return New(ErrorTypeRateLimitError, "Rate limit exceeded").
		WithRetryAfter(retryAfter)
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	var ccErr *CCProxyError
	if errors.As(err, &ccErr) {
		return ccErr.Retryable
	}

	// Check for common retryable error patterns
	errStr := err.Error()
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"no such host",
		"timeout",
		"temporary failure",
		"service unavailable",
		"bad gateway",
	}

	errLower := strings.ToLower(errStr)
	for _, pattern := range retryablePatterns {
		if strings.Contains(errLower, pattern) {
			return true
		}
	}

	return false
}

// GetRetryAfter extracts retry after duration from error
func GetRetryAfter(err error) *time.Duration {
	var ccErr *CCProxyError
	if errors.As(err, &ccErr) {
		return ccErr.RetryAfter
	}
	return nil
}

// Security-specific error constructors

// NewAuthError creates an authentication error
func NewAuthError(message string, cause error) *CCProxyError {
	if cause != nil {
		return Wrap(cause, ErrorTypeUnauthorized, message)
	}
	return New(ErrorTypeUnauthorized, message)
}

// NewValidationError creates a validation error
func NewValidationError(message string, cause error) *CCProxyError {
	if cause != nil {
		return Wrap(cause, ErrorTypeValidationError, message)
	}
	return New(ErrorTypeValidationError, message)
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(message string, cause error) *CCProxyError {
	if cause != nil {
		return Wrap(cause, ErrorTypeRateLimitError, message)
	}
	return New(ErrorTypeRateLimitError, message)
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string, cause error) *CCProxyError {
	if cause != nil {
		return Wrap(cause, ErrorTypeForbidden, message)
	}
	return New(ErrorTypeForbidden, message)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(message string, cause error) *CCProxyError {
	if cause != nil {
		return Wrap(cause, ErrorTypeNotFound, message)
	}
	return New(ErrorTypeNotFound, message)
}
